package entity

import (
	"context"
	"math"
	"time"

	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/metrics"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
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
	billingInfo := FindBillingInfo(ctx, t.OrganizationID)
	if billingInfo == nil {
		panic("organization's billing info not found")
	}

	seatBillTimes := calculateBillTimes(t.Timestamp, billingInfo.BillingPlan.Period, true)
	usageBillTimes := calculateBillTimes(t.Timestamp, billingInfo.BillingPlan.Period, false)

	// add "seat" line items
	err := commitSeatsToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, billingInfo.Organization.Users, seatBillTimes)
	if err != nil {
		return err
	}

	// add "usage" line items for users
	var userIDs []uuid.UUID
	var usernames []string
	for _, user := range billingInfo.Organization.Users {
		userIDs = append(userIDs, user.UserID)
		usernames = append(usernames, user.Username)
	}
	err = commitUsagesToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, UserEntityKind, userIDs, usernames, usageBillTimes)
	if err != nil {
		return err
	}

	// add "usage" line items for services
	var serviceIDs []uuid.UUID
	var servicenames []string
	for _, service := range billingInfo.Organization.Services {
		serviceIDs = append(serviceIDs, service.ServiceID)
		servicenames = append(servicenames, service.Name)
	}
	err = commitUsagesToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, ServiceEntityKind, serviceIDs, servicenames, usageBillTimes)
	if err != nil {
		return err
	}

	// if applicable, add "read overage" to bill
	err = commitOverageToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, ReadProduct, billingInfo.Organization.Name, usageBillTimes)
	if err != nil {
		return err
	}

	// if applicable, add "write overage" to bill
	err = commitOverageToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, WriteProduct, billingInfo.Organization.Name, usageBillTimes)
	if err != nil {
		return err
	}

	err = taskqueue.Submit(context.Background(), &SendInvoiceTask{
		OrganizationID: t.OrganizationID,
		BillingTime:    seatBillTimes.BillingTime,
	})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	return nil
}

func commitSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, users []*User, billTimes *billTimes) error {
	if len(users) == 0 {
		log.S.Info("no users in this organization -- this organization must be getting deleted")
		return nil
	}

	var billedResources []*BilledResource
	for _, user := range users {
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        user.UserID,
			EntityName:      user.Username,
			EntityKind:      UserEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         SeatProduct,
			Quantity:        1,
			TotalPriceCents: billingPlan.SeatPriceCents,
			Currency:        billingPlan.Currency,
		})
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func commitProratedSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, userIDs []uuid.UUID, usernames []string, credit bool) error {
	if billingPlan == nil {
		panic("could not find the organization's billing plan")
	}

	now := time.Now()
	p := billingPlan.Period
	billTimes := &billTimes{
		BillingTime: timeutil.Next(now, p),
		StartTime:   now,
		EndTime:     timeutil.Next(now, p),
	}

	proratedFraction := float64(timeutil.DaysLeftInPeriod(now, p)) / float64(timeutil.TotalDaysInPeriod(now, p))
	proratedPrice := int32(math.Round(float64(billingPlan.SeatPriceCents) * proratedFraction))
	product := SeatProratedProduct

	if credit {
		proratedPrice *= -1
		product = SeatProratedCreditProduct
	}

	var billedResources []*BilledResource
	for i, userID := range userIDs {
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        userID,
			EntityName:      usernames[i],
			EntityKind:      UserEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         product,
			Quantity:        1,
			TotalPriceCents: proratedPrice,
			Currency:        billingPlan.Currency,
		})
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func commitUsagesToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, entityKind Kind, entityIDs []uuid.UUID, entityNames []string, billTimes *billTimes) error {
	var billedResources []*BilledResource

	for i, entityID := range entityIDs {
		_, monthlyMetrics, err := metrics.GetHistoricalUsage(ctx, entityID, billingPlan.Period, billTimes.StartTime, billTimes.BillingTime) // when adding annual plans, remember this function only accepts hourly or monthly periods
		if err != nil {
			return err
		}

		if len(monthlyMetrics) == 1 {
			// add reads
			billedResources = append(billedResources, &BilledResource{
				OrganizationID:  organizationID,
				BillingTime:     billTimes.BillingTime,
				EntityID:        entityID,
				EntityName:      entityNames[i],
				EntityKind:      entityKind,
				StartTime:       billTimes.StartTime,
				EndTime:         billTimes.EndTime,
				Product:         ReadProduct,
				Quantity:        monthlyMetrics[0].ReadBytes,
				TotalPriceCents: 0,
				Currency:        billingPlan.Currency,
			})

			// add writes
			billedResources = append(billedResources, &BilledResource{
				OrganizationID:  organizationID,
				BillingTime:     billTimes.BillingTime,
				EntityID:        entityID,
				EntityName:      entityNames[i],
				EntityKind:      entityKind,
				StartTime:       billTimes.StartTime,
				EndTime:         billTimes.EndTime,
				Product:         WriteProduct,
				Quantity:        monthlyMetrics[0].WriteBytes,
				TotalPriceCents: 0,
				Currency:        billingPlan.Currency,
			})
		} else if len(monthlyMetrics) > 1 {
			panic("expected a maximum of one item in monthlyMetrics")
		}
	}

	if len(billedResources) > 0 {
		err := CreateOrUpdateBilledResources(ctx, billedResources)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	}

	// done
	return nil
}

func commitCurrentUsageToNextBill(ctx context.Context, organizationID uuid.UUID, entityKind Kind, entityID uuid.UUID, entityName string, credit bool) error {
	billingInfo := FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		panic("organization not found")
	}

	now := time.Now()
	p := billingInfo.BillingPlan.Period
	billTimes := &billTimes{
		BillingTime: timeutil.Next(now, p),
		StartTime:   timeutil.Floor(now, p),
		EndTime:     now,
	}

	var billedResources []*BilledResource

	_, monthlyMetrics, err := metrics.GetHistoricalUsage(ctx, entityID, billingInfo.BillingPlan.Period, billTimes.StartTime, billTimes.BillingTime) // when adding annual plans, remember this function only accepts hourly or monthly periods
	if err != nil {
		return err
	}

	if len(monthlyMetrics) == 1 {
		readQuantity := monthlyMetrics[0].ReadBytes
		writeQuantity := monthlyMetrics[0].WriteBytes
		readProduct := ReadProduct
		writeProduct := WriteProduct

		if credit {
			readQuantity *= -1
			writeQuantity *= -1
			readProduct = ReadCreditProduct
			writeProduct = WriteCreditProduct
		}

		// add reads
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        entityID,
			EntityName:      entityName,
			EntityKind:      entityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         readProduct,
			Quantity:        readQuantity,
			TotalPriceCents: 0,
			Currency:        billingInfo.BillingPlan.Currency,
		})

		// add writes
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        entityID,
			EntityName:      entityName,
			EntityKind:      entityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         writeProduct,
			Quantity:        writeQuantity,
			TotalPriceCents: 0,
			Currency:        billingInfo.BillingPlan.Currency,
		})

		err = CreateOrUpdateBilledResources(ctx, billedResources)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	} else if len(monthlyMetrics) > 1 {
		panic("monthlyMetrics can't have more than one item")
	}

	// done
	return nil
}

func commitOverageToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, product Product, organizationName string, billTimes *billTimes) error {
	var creditProduct Product
	if product == ReadProduct {
		creditProduct = ReadCreditProduct
	} else if product == WriteProduct {
		creditProduct = WriteCreditProduct
	} else {
		panic("overage only applies to read and write products")
	}

	// fetch the organization's billed resources for the period
	var billedResources []*BilledResource
	err := hub.DB.ModelContext(ctx, &billedResources).
		Where("organization_id = ?", organizationID).
		Where("billing_time = ?", billTimes.BillingTime).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("product = ?", product).
				WhereOr("product = ?", creditProduct)
			return q, nil
		}).
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
	var billProduct Product
	var overageBytes int64
	var overageGB int64
	var price int32

	if product == ReadProduct {
		billProduct = ReadOverageProduct
		overageBytes := usage - billingPlan.BaseReadQuota
		overageGB := overageBytes/10 ^ 6
		price = int32(overageGB) * billingPlan.ReadOveragePriceCents // assuming unit of ReadOveragePriceCents is GB
	} else if product == WriteProduct {
		billProduct = WriteOverageProduct
		overageBytes := usage - billingPlan.BaseWriteQuota
		overageGB := overageBytes/10 ^ 6
		price = int32(overageGB) * billingPlan.WriteOveragePriceCents // assuming unit of WriteOveragePriceCents is GB
	} else {
		panic("overage only applies to read and write products")
	}

	if overageBytes > 0 {
		newBilledResource := []*BilledResource{&BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        organizationID,
			EntityName:      organizationName,
			EntityKind:      OrganizationEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         billProduct,
			Quantity:        overageGB,
			TotalPriceCents: price,
			Currency:        billingPlan.Currency,
		}}

		err := CreateOrUpdateBilledResources(ctx, newBilledResource)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	}

	// done
	return nil
}

// on the first of each month, we charge for: the upcoming month's seats; the previous month's overage
func calculateBillTimes(ts time.Time, p timeutil.Period, isSeatProduct bool) *billTimes {
	billingTime := timeutil.Floor(ts, p)
	startTime := timeutil.Last(ts, p)
	endTime := timeutil.Floor(ts, p)

	if isSeatProduct {
		if p == timeutil.PeriodMonth {
			startTime = startTime.AddDate(0, 1, 0)
			endTime = endTime.AddDate(0, 1, 0)
		} else if p == timeutil.PeriodYear {
			startTime = startTime.AddDate(1, 0, 0)
			endTime = endTime.AddDate(1, 0, 0)
		} else {
			panic("billing period is not supported")
		}
	}

	return &billTimes{
		BillingTime: billingTime,
		StartTime:   startTime,
		EndTime:     endTime,
	}
}
