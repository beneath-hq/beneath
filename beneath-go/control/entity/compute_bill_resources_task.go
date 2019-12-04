package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/metrics"
	"github.com/beneath-core/beneath-go/taskqueue"
	uuid "github.com/satori/go.uuid"
)

type configSpecification struct {
	FreeBillingPlanID    string `envconfig:"CONTROL_FREE_BILLING_PLAN_ID" required:"true"`
	BeneathBillingPlanID string `envconfig:"CONTROL_BENEATH_BILLING_PLAN_ID" required:"true"`
}

// ComputeBillResourcesTask computes all items on an organization's bill
type ComputeBillResourcesTask struct {
	OrganizationID uuid.UUID
	Timestamp      time.Time
}

// register task
func init() {
	taskqueue.RegisterTask(&ComputeBillResourcesTask{})
}

// Run triggers the task
func (t *ComputeBillResourcesTask) Run(ctx context.Context) error {
	var config configSpecification
	core.LoadConfig("beneath", &config)

	organization := FindOrganization(ctx, t.OrganizationID)
	if organization == nil {
		panic("organization not found")
	}

	isMidPeriod := false
	billingTime, startTime, endTime := calculateBillTimes(t.Timestamp, organization.BillingPlan.Period, isMidPeriod)

	var billedResources []*BilledResource

	for _, user := range organization.Users {
		// add a "seat" line item
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  t.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        user.UserID,
			EntityKind:      UserEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         SeatProduct,
			Quantity:        1,
			TotalPriceCents: organization.BillingPlan.SeatPriceCents,
			Currency:        organization.BillingPlan.Currency,
		})

		// add a "usage" line item
		err := commitUsageToBill(ctx, t.OrganizationID, UserEntityKind, user.UserID, isMidPeriod)
		if err != nil {
			return err
		}
	}

	for _, service := range organization.Services {
		// add a "usage" line item
		err := commitUsageToBill(ctx, t.OrganizationID, ServiceEntityKind, service.ServiceID, isMidPeriod)
		if err != nil {
			return err
		}
	}

	// if applicable, add "read overage" to bill
	readOverageBytes := calculateOverage(ctx, t.OrganizationID, billingTime, ReadProduct)
	readOverageGB := readOverageBytes/10 ^ 6
	if readOverageBytes > 0 {
		price := int32(readOverageGB) * organization.BillingPlan.ReadOveragePriceCents // assuming unit of ReadOveragePriceCents is GB
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  t.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        t.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         ReadOverageProduct,
			Quantity:        readOverageGB,
			TotalPriceCents: price,
			Currency:        organization.BillingPlan.Currency,
		})
	}

	// if applicable, add "write overage" to bill
	writeOverageBytes := calculateOverage(ctx, t.OrganizationID, billingTime, WriteProduct)
	writeOverageGB := writeOverageBytes/10 ^ 6
	if writeOverageBytes > 0 {
		price := int32(writeOverageGB) * organization.BillingPlan.WriteOveragePriceCents // assuming unit of WriteOveragePriceCents is GB
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  t.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        t.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         WriteOverageProduct,
			Quantity:        writeOverageGB,
			TotalPriceCents: price,
			Currency:        organization.BillingPlan.Currency,
		})
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// only send invoices to paying customers (i.e. organizations not on the Free plan or Beneath plan)
	if (organization.BillingPlanID.String() != config.FreeBillingPlanID) && (organization.BillingPlanID.String() != config.BeneathBillingPlanID) {
		err := taskqueue.Submit(context.Background(), &SendInvoiceTask{
			OrganizationID: t.OrganizationID,
			BillingTime:    billingTime,
		})
		if err != nil {
			log.S.Errorw("Error creating task", err)
		}
	}

	return nil
}

func commitUsageToBill(ctx context.Context, organizationID uuid.UUID, entityKind Kind, entityID uuid.UUID, isMidPeriod bool) error {
	organization := FindOrganization(ctx, organizationID)
	if organization == nil {
		panic("organization not found")
	}

	billingTime, startTime, endTime := calculateBillTimes(time.Now(), organization.BillingPlan.Period, isMidPeriod)

	_, monthlyMetrics, err := metrics.GetHistoricalUsage(ctx, entityID, organization.BillingPlan.Period, startTime, billingTime) // when adding annual plans, remember this function only accepts hourly or monthly periods
	if err != nil {
		return err
	}
	if len(monthlyMetrics) > 1 {
		panic("expected a maximum of one item in monthlyMetrics")
	}

	var billedResources []*BilledResource

	if len(monthlyMetrics) == 1 {
		// add reads
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billingTime,
			EntityID:        entityID,
			EntityKind:      entityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         ReadProduct,
			Quantity:        monthlyMetrics[0].ReadBytes,
			TotalPriceCents: 0,
			Currency:        organization.BillingPlan.Currency,
		})

		// add writes
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billingTime,
			EntityID:        entityID,
			EntityKind:      entityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         WriteProduct,
			Quantity:        monthlyMetrics[0].WriteBytes,
			TotalPriceCents: 0,
			Currency:        organization.BillingPlan.Currency,
		})
	}

	if len(billedResources) > 0 {
		err = CreateOrUpdateBilledResources(ctx, billedResources)
		if err != nil {
			panic("unable to write billed resources to table")
		}
	}

	// done
	return nil
}

func calculateOverage(ctx context.Context, organizationID uuid.UUID, billingTime time.Time, product Product) int64 {
	organization := FindOrganization(ctx, organizationID)
	if organization == nil {
		panic("organization not found")
	}

	var billedResources []*BilledResource
	err := db.DB.ModelContext(ctx, &billedResources).
		Where("organization_id = ?", organizationID).
		Where("billing_time = ?", billingTime).
		Where("product = ?", product).
		Select()
	if err != nil {
		panic(err)
	}

	organizationUsage := int64(0)
	for _, user := range billedResources {
		organizationUsage += user.Quantity
	}

	organizationOverage := int64(0)
	if product == ReadProduct {
		organizationOverage = organizationUsage - organization.BillingPlan.SeatReadQuota*int64(len(organization.Users))
	} else if product == WriteProduct {
		organizationOverage = organizationUsage - organization.BillingPlan.SeatWriteQuota*int64(len(organization.Users))
	} else {
		panic("overage only applies to read and write products")
	}

	return organizationOverage
}

func calculateBillTimes(ts time.Time, p timeutil.Period, isMidPeriod bool) (time.Time, time.Time, time.Time) {
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
	return billingTime, startTime, endTime
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
