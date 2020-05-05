package entity

import (
	"context"
	"math"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
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
	numSeats, err := commitSeatsToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, billingInfo.Organization.Users, seatBillTimes)
	if err != nil {
		return err
	}

	// if applicable, add overages to bill
	err = commitOverageToBill(ctx, t.OrganizationID, billingInfo.BillingPlan, numSeats, billingInfo.Organization.Name, usageBillTimes)
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

func commitSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, users []*User, billTimes *billTimes) (int, error) {
	var billedResources []*BilledResource
	var numSeats int
	for _, user := range users {
		// only bill those organization users who have it as their billing org
		if user.BillingOrganizationID == organizationID {
			billedResources = append(billedResources, &BilledResource{
				OrganizationID:  organizationID,
				BillingTime:     billTimes.BillingTime,
				EntityID:        user.UserID,
				EntityName:      user.Email,
				EntityKind:      UserEntityKind,
				StartTime:       billTimes.StartTime,
				EndTime:         billTimes.EndTime,
				Product:         SeatProduct,
				Quantity:        1,
				TotalPriceCents: billingPlan.SeatPriceCents,
				Currency:        billingPlan.Currency,
			})
			numSeats++
		}
	}

	if len(billedResources) == 0 {
		if !billingPlan.Personal {
			log.S.Info("The multi organization has no users and will be deleted.")
		}

		return 0, nil
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return numSeats, nil
}

func commitProratedSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingTime time.Time, billingPlan *BillingPlan, userIDs []uuid.UUID, usernames []string, credit bool) error {
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

func commitOverageToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, numSeats int, organizationName string, billTimes *billTimes) error {
	_, usages, err := metrics.GetHistoricalUsage(ctx, organizationID, billingPlan.Period, billTimes.StartTime, billTimes.EndTime)
	if err != nil {
		panic("unable to get historical usage for organization")
	}

	if len(usages) > 1 {
		panic("GetHistoricalUsage returned more periods than expected")
	} else if len(usages) == 0 {
		log.S.Infof("organization %s had no usage in the billing period", organizationID.String())
		return nil
	}

	readIncludedQuota := billingPlan.BaseReadQuota + billingPlan.SeatReadQuota*int64(numSeats)
	writeIncludedQuota := billingPlan.BaseWriteQuota + billingPlan.SeatWriteQuota*int64(numSeats)

	readOverageBytes := usages[0].ReadBytes - readIncludedQuota
	readOverageGB := readOverageBytes/10 ^ 6
	readOveragePrice := int32(readOverageGB) * billingPlan.ReadOveragePriceCents // assuming unit of ReadOveragePriceCents is GB

	writeOverageBytes := usages[0].WriteBytes - writeIncludedQuota
	writeOverageGB := writeOverageBytes/10 ^ 6
	writeOveragePrice := int32(writeOverageGB) * billingPlan.WriteOveragePriceCents // assuming unit of WriteOveragePriceCents is GB

	var newBilledResources []*BilledResource

	if readOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        organizationID,
			EntityName:      organizationName,
			EntityKind:      OrganizationEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         ReadOverageProduct,
			Quantity:        readOverageGB,
			TotalPriceCents: readOveragePrice,
			Currency:        billingPlan.Currency,
		})
	}

	if writeOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        organizationID,
			EntityName:      organizationName,
			EntityKind:      OrganizationEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         WriteOverageProduct,
			Quantity:        writeOverageGB,
			TotalPriceCents: writeOveragePrice,
			Currency:        billingPlan.Currency,
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
