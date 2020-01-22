package entity

import (
	"context"
	"math"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/metrics"
	uuid "github.com/satori/go.uuid"
)

type billTimes struct {
	BillingTime time.Time
	StartTime   time.Time
	EndTime     time.Time
}

func commitProratedSeatsToBill(ctx context.Context, organizationID uuid.UUID, billingPlan *BillingPlan, userIDs []uuid.UUID, credit bool) error {
	if billingPlan == nil {
		panic("could not find the organization's billing plan")
	}

	now := time.Now()
	p := billingPlan.Period

	billTimes := &billTimes{
		BillingTime: timeutil.BeginningOfNextPeriod(p),
		StartTime:   now,
		EndTime:     timeutil.BeginningOfNextPeriod(p),
	}

	proratedFraction := float64(timeutil.DaysLeftInPeriod(now, p)) / float64(timeutil.TotalDaysInPeriod(now, p))
	proratedPrice := int32(math.Round(float64(billingPlan.SeatPriceCents) * proratedFraction))

	if credit {
		proratedPrice = -1 * proratedPrice
	}

	var billedResources []*BilledResource
	for _, userID := range userIDs {
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        userID,
			EntityKind:      UserEntityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         SeatProduct,
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

func commitCurrentUsageToNextBill(ctx context.Context, organizationID uuid.UUID, entityKind Kind, entityID uuid.UUID, credit bool) error {
	billingInfo := FindBillingInfo(ctx, organizationID)
	if billingInfo == nil {
		panic("organization not found")
	}

	billTimes := &billTimes{
		BillingTime: timeutil.BeginningOfNextPeriod(billingInfo.BillingPlan.Period),
		StartTime:   timeutil.BeginningOfThisPeriod(billingInfo.BillingPlan.Period),
		EndTime:     time.Now(),
	}

	var billedResources []*BilledResource

	_, monthlyMetrics, err := metrics.GetHistoricalUsage(ctx, entityID, billingInfo.BillingPlan.Period, billTimes.StartTime, billTimes.BillingTime) // when adding annual plans, remember this function only accepts hourly or monthly periods
	if err != nil {
		return err
	}

	readQuantity := int64(0)
	writeQuantity := int64(0)
	if credit {
		readQuantity = -1 * monthlyMetrics[0].ReadBytes
		writeQuantity = -1 * monthlyMetrics[0].WriteBytes
	} else {
		readQuantity = monthlyMetrics[0].ReadBytes
		writeQuantity = monthlyMetrics[0].WriteBytes
	}

	if len(monthlyMetrics) == 1 {
		// add reads
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        entityID,
			EntityKind:      entityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         ReadProduct,
			Quantity:        readQuantity,
			TotalPriceCents: 0,
			Currency:        billingInfo.BillingPlan.Currency,
		})

		// add writes
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  organizationID,
			BillingTime:     billTimes.BillingTime,
			EntityID:        entityID,
			EntityKind:      entityKind,
			StartTime:       billTimes.StartTime,
			EndTime:         billTimes.EndTime,
			Product:         WriteProduct,
			Quantity:        writeQuantity,
			TotalPriceCents: 0,
			Currency:        billingInfo.BillingPlan.Currency,
		})
	} else if len(monthlyMetrics) > 1 {
		panic("expected a maximum of one item in monthlyMetrics")
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
