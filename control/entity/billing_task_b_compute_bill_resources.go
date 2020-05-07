package entity

import (
	"context"
	"math"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/metrics"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// ComputeBillResourcesTask computes all items on an organization's bill
type ComputeBillResourcesTask struct {
	BillingInfo *BillingInfo
	Timestamp   time.Time
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
	seatBillTimes := calculateBillTimes(t.Timestamp, t.BillingInfo.BillingPlan.Period, true)
	usageBillTimes := calculateBillTimes(t.Timestamp, t.BillingInfo.BillingPlan.Period, false)

	// add "seat" line items
	numSeats, err := commitSeatsToBill(ctx, t.BillingInfo, seatBillTimes)
	if err != nil {
		return err
	}

	// if applicable, add overages to bill
	err = commitOverageToBill(ctx, t.BillingInfo, usageBillTimes)
	if err != nil {
		return err
	}

	// recompute organization's prepaid quotas
	// necessary to account for a) a user leaving an organization mid-period b) a billing plan's parameters changing mid-period and
	err = recomputeOrganizationPrepaidQuotas(ctx, t.BillingInfo, numSeats)
	if err != nil {
		return err
	}

	err = taskqueue.Submit(context.Background(), &SendInvoiceTask{
		BillingInfo: t.BillingInfo,
		BillingTime: seatBillTimes.BillingTime,
	})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	return nil
}

func commitSeatsToBill(ctx context.Context, bi *BillingInfo, billTimes *billTimes) (int64, error) {
	var billedResources []*BilledResource
	var numSeats int64
	for _, user := range bi.Organization.Users {
		// only bill those organization users who have it as their billing org
		if user.BillingOrganizationID == bi.OrganizationID {
			billedResources = append(billedResources, &BilledResource{
				OrganizationID:  bi.OrganizationID,
				BillingTime:     billTimes.BillingTime,
				EntityID:        user.UserID,
				EntityKind:      UserEntityKind,
				StartTime:       billTimes.StartTime,
				EndTime:         billTimes.EndTime,
				Product:         SeatProduct,
				Quantity:        1,
				TotalPriceCents: bi.BillingPlan.SeatPriceCents,
				Currency:        bi.BillingPlan.Currency,
			})
			numSeats++
		}
	}

	if len(billedResources) == 0 {
		if !bi.BillingPlan.Personal {
			log.S.Info("The multi organization has no users and will be deleted.")
		}

		return numSeats, nil
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return numSeats, nil
}

func commitProratedSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingTime time.Time, billingPlan *BillingPlan, users []*User, credit bool) error {
	if billingPlan == nil {
		panic("could not find the organization's billing plan")
	}

	now := time.Now()
	p := billingPlan.Period
	billTimes := &billTimes{
		BillingTime: billingTime,
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
	for _, user := range users {
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        user.UserID,
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

func commitOverageToBill(ctx context.Context, bi *BillingInfo, billTimes *billTimes) error {
	_, usages, err := metrics.GetHistoricalUsage(ctx, bi.OrganizationID, bi.BillingPlan.Period, billTimes.StartTime, billTimes.EndTime)
	if err != nil {
		panic("unable to get historical usage for organization")
	}

	if len(usages) > 1 {
		panic("GetHistoricalUsage returned more periods than expected")
	} else if len(usages) == 0 {
		log.S.Infof("organization %s had no usage in the billing period", bi.OrganizationID.String())
		return nil
	}

	readOverageBytes := usages[0].ReadBytes - *bi.Organization.PrepaidReadQuota
	readOverageGB := readOverageBytes/10 ^ 6
	readOveragePrice := int32(readOverageGB) * bi.BillingPlan.ReadOveragePriceCents // assuming unit of ReadOveragePriceCents is GB

	writeOverageBytes := usages[0].WriteBytes - *bi.Organization.PrepaidWriteQuota
	writeOverageGB := writeOverageBytes/10 ^ 6
	writeOveragePrice := int32(writeOverageGB) * bi.BillingPlan.WriteOveragePriceCents // assuming unit of WriteOveragePriceCents is GB

	var newBilledResources []*BilledResource

	if readOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        bi.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         ReadOverageProduct,
			Quantity:        readOverageGB,
			TotalPriceCents: readOveragePrice,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	if writeOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        bi.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         WriteOverageProduct,
			Quantity:        writeOverageGB,
			TotalPriceCents: writeOveragePrice,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	if len(newBilledResources) > 0 {
		err := CreateOrUpdateBilledResources(ctx, newBilledResources)
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

func recomputeOrganizationPrepaidQuotas(ctx context.Context, bi *BillingInfo, numSeats int64) error {
	org := bi.Organization

	if org.PrepaidReadQuota != nil {
		newPrepaidReadQuota := bi.BillingPlan.BaseReadQuota + bi.BillingPlan.SeatReadQuota*numSeats
		org.PrepaidReadQuota = &newPrepaidReadQuota
	}
	if org.PrepaidReadQuota != nil {
		newPrepaidWriteQuota := bi.BillingPlan.BaseWriteQuota + bi.BillingPlan.SeatWriteQuota*numSeats
		org.PrepaidWriteQuota = &newPrepaidWriteQuota
	}

	if org.PrepaidReadQuota != nil || org.PrepaidWriteQuota != nil {
		org.UpdatedOn = time.Now()
		_, err := hub.DB.WithContext(ctx).Model(org).
			Column("prepaid_read_quota", "prepaid_write_quota", "updated_on").
			WherePK().
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}
