package billing

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/metrics"
	"github.com/beneath-core/beneath-go/taskqueue"
	uuid "github.com/satori/go.uuid"
)

// ComputeBillResourcesTask computes all items on an organization's bill
type ComputeBillResourcesTask struct {
	OrganizationID uuid.UUID
	Timestamp      time.Time
}

type billTimes struct {
	BillingTime time.Time
	StartTime   time.Time
	EndTime     time.Time
}

// register task
func init() {
	taskqueue.RegisterTask(&ComputeBillResourcesTask{})
}

// Run triggers the task
func (t *ComputeBillResourcesTask) Run(ctx context.Context) error {
	organization := entity.FindOrganization(ctx, t.OrganizationID)
	if organization == nil {
		panic("organization not found")
	}

	billingInfo := entity.FindBillingInfo(ctx, t.OrganizationID)
	if billingInfo == nil {
		panic("organization's billing info not found")
	}

	isMidPeriod := false
	billTimes := calculateBillTimes(t.Timestamp, billingInfo.BillingPlan.Period, isMidPeriod)

	// add "seat" line items
	err := commitSeatsToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, billingInfo.Users, billTimes)
	if err != nil {
		return err
	}

	// add "usage" line items for users
	var userIDs []uuid.UUID
	for _, user := range billingInfo.Users {
		userIDs = append(userIDs, user.UserID)
	}
	err = commitUsagesToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, entity.UserEntityKind, userIDs, billTimes)
	if err != nil {
		return err
	}

	// add "usage" line items for services
	var serviceIDs []uuid.UUID
	for _, service := range billingInfo.Services {
		serviceIDs = append(serviceIDs, service.ServiceID)
	}
	err = commitUsagesToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, entity.ServiceEntityKind, serviceIDs, billTimes)
	if err != nil {
		return err
	}

	// if applicable, add "read overage" to bill
	err = commitOverageToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, entity.ReadProduct, billTimes)

	// if applicable, add "write overage" to bill
	err = commitOverageToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, entity.WriteProduct, billTimes)

	err = taskqueue.Submit(context.Background(), &SendInvoiceTask{
		OrganizationID: t.OrganizationID,
		BillingTime:    billTimes.BillingTime,
	})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	return nil
}

func commitSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *entity.BillingPlan, users []*entity.User, billTimes *billTimes) error {
	var billedResources []*entity.BilledResource
	for _, user := range users {
		billedResources = append(billedResources, &entity.BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        user.UserID,
			EntityKind:      entity.UserEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         entity.SeatProduct,
			Quantity:        1,
			TotalPriceCents: billingPlan.SeatPriceCents,
			Currency:        billingPlan.Currency,
		})
	}

	err := entity.CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func commitUsagesToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *entity.BillingPlan, entityKind entity.Kind, entityIDs []uuid.UUID, billTimes *billTimes) error {
	var billedResources []*entity.BilledResource

	for _, entityID := range entityIDs {
		_, monthlyMetrics, err := metrics.GetHistoricalUsage(ctx, entityID, billingPlan.Period, billTimes.StartTime, billTimes.BillingTime) // when adding annual plans, remember this function only accepts hourly or monthly periods
		if err != nil {
			return err
		}

		if len(monthlyMetrics) == 1 {
			// add reads
			billedResources = append(billedResources, &entity.BilledResource{
				OrganizationID:  organizationID,
				BillingTime:     billTimes.BillingTime,
				EntityID:        entityID,
				EntityKind:      entityKind,
				StartTime:       billTimes.StartTime,
				EndTime:         billTimes.EndTime,
				Product:         entity.ReadProduct,
				Quantity:        monthlyMetrics[0].ReadBytes,
				TotalPriceCents: 0,
				Currency:        billingPlan.Currency,
			})

			// add writes
			billedResources = append(billedResources, &entity.BilledResource{
				OrganizationID:  organizationID,
				BillingTime:     billTimes.BillingTime,
				EntityID:        entityID,
				EntityKind:      entityKind,
				StartTime:       billTimes.StartTime,
				EndTime:         billTimes.EndTime,
				Product:         entity.WriteProduct,
				Quantity:        monthlyMetrics[0].WriteBytes,
				TotalPriceCents: 0,
				Currency:        billingPlan.Currency,
			})
		} else if len(monthlyMetrics) > 1 {
			panic("expected a maximum of one item in monthlyMetrics")
		}
	}

	if len(billedResources) > 0 {
		err := entity.CreateOrUpdateBilledResources(ctx, billedResources)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	}

	// done
	return nil
}

func commitOverageToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *entity.BillingPlan, product entity.Product, billTimes *billTimes) error {
	// fetch the organization's billed resources for the period
	var billedResources []*entity.BilledResource
	err := db.DB.ModelContext(ctx, &billedResources).
		Where("organization_id = ?", organizationID).
		Where("billing_time = ?", billTimes.BillingTime).
		Where("product = ?", product).
		Select()
	if err != nil {
		panic(err)
	}

	// calculate total usage across all the organization's users
	usage := int64(0)
	for _, billedResource := range billedResources {
		usage += billedResource.Quantity
	}

	// get product-specific variables
	var billProduct entity.Product
	var overageBytes int64
	var overageGB int64
	var price int32

	if product == entity.ReadProduct {
		billProduct = entity.ReadOverageProduct
		overageBytes := usage - billingPlan.BaseReadQuota
		overageGB := overageBytes/10 ^ 6
		price = int32(overageGB) * billingPlan.ReadOveragePriceCents // assuming unit of ReadOveragePriceCents is GB
	} else if product == entity.WriteProduct {
		billProduct = entity.WriteOverageProduct
		overageBytes := usage - billingPlan.BaseWriteQuota
		overageGB := overageBytes/10 ^ 6
		price = int32(overageGB) * billingPlan.WriteOveragePriceCents // assuming unit of WriteOveragePriceCents is GB
	} else {
		panic("overage only applies to read and write products")
	}

	var newBilledResources []*entity.BilledResource

	if overageBytes > 0 {
		newBilledResources = append(newBilledResources, &entity.BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        organizationID,
			EntityKind:      entity.OrganizationEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         billProduct,
			Quantity:        overageGB,
			TotalPriceCents: price,
			Currency:        billingPlan.Currency,
		})
	}

	if len(newBilledResources) > 0 {
		err := entity.CreateOrUpdateBilledResources(ctx, newBilledResources)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	}

	// done
	return nil
}

func calculateBillTimes(ts time.Time, p timeutil.Period, isMidPeriod bool) *billTimes {
	now := time.Now()

	var billingTime time.Time
	var startTime time.Time
	var endTime time.Time

	if isMidPeriod {
		billingTime = BeginningOfNextPeriod(p)
		startTime = BeginningOfThisPeriod(p)
		endTime = now
	} else {
		billingTime = BeginningOfThisPeriod(p)
		startTime = BeginningOfLastPeriod(p)
		endTime = EndOfLastPeriod(p)
		// FOR TESTING (since I don't have usage data from last month):
		startTime = startTime.AddDate(0, 1, 0)
		endTime = endTime.AddDate(0, 1, 0)
		billingTime = billingTime.AddDate(0, 1, 0)
	}
	return &billTimes{
		BillingTime: billingTime,
		StartTime:   startTime,
		EndTime:     endTime,
	}
}

// BeginningOfThisPeriod gets the beginning of this period
func BeginningOfThisPeriod(p timeutil.Period) time.Time {
	return timeutil.Floor(time.Now(), p)
}

// BeginningOfNextPeriod gets the beginning of the next period
func BeginningOfNextPeriod(p timeutil.Period) time.Time {
	ts := time.Now().UTC()
	return timeutil.Next(ts, p)
}

// BeginningOfLastPeriod gets the beginning of the last period
func BeginningOfLastPeriod(p timeutil.Period) time.Time {
	ts := time.Now().UTC()
	return timeutil.Last(ts, p)
}

// EndOfLastPeriod gets the end of the last period
func EndOfLastPeriod(p timeutil.Period) time.Time {
	ts := time.Now().UTC()
	return timeutil.Floor(ts, p)
}
