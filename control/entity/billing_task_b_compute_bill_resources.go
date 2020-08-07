package entity

import (
	"context"
	"math"
	"time"

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

// register task
func init() {
	taskqueue.RegisterTask(&ComputeBillResourcesTask{})
}

// Run triggers the task
func (t *ComputeBillResourcesTask) Run(ctx context.Context) error {
	// if applicable, add "prepaid usage" line item
	err := commitPrepaidQuotaToBill(ctx, t.BillingInfo, t.Timestamp)
	if err != nil {
		return err
	}

	// if applicable, add "seat" line items
	err = commitSeatsToBill(ctx, t.BillingInfo, t.Timestamp)
	if err != nil {
		return err
	}

	// if applicable, add overages to bill
	err = commitOverageToBill(ctx, t.BillingInfo, t.Timestamp)
	if err != nil {
		return err
	}

	// recompute organization's prepaid quotas
	// necessary to account for a) a user leaving an organization mid-period b) a billing plan's parameters changing mid-period and
	err = recomputeOrganizationPrepaidQuotas(ctx, t.BillingInfo)
	if err != nil {
		return err
	}

	err = taskqueue.Submit(context.Background(), &SendInvoiceTask{
		BillingInfo: t.BillingInfo,
		BillingTime: timeutil.Floor(t.Timestamp, t.BillingInfo.BillingPlan.Period),
	})
	if err != nil {
		log.S.Errorw("Error creating task", err)
	}

	return nil
}

func commitPrepaidQuotaToBill(ctx context.Context, bi *BillingInfo, ts time.Time) error {
	if bi.BillingPlan.BasePriceCents == 0 {
		return nil
	}

	billingTime := timeutil.Floor(ts, bi.BillingPlan.Period)
	startTime := timeutil.Floor(ts, bi.BillingPlan.Period)
	endTime := timeutil.Next(ts, bi.BillingPlan.Period)

	var billedResources []*BilledResource
	billedResources = append(billedResources, &BilledResource{
		OrganizationID:  bi.OrganizationID,
		BillingTime:     billingTime,
		EntityID:        bi.OrganizationID,
		EntityKind:      OrganizationEntityKind,
		StartTime:       startTime,
		EndTime:         endTime,
		Product:         PrepaidQuotaProduct,
		Quantity:        1,
		TotalPriceCents: bi.BillingPlan.BasePriceCents,
		Currency:        bi.BillingPlan.Currency,
	})

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func commitSeatsToBill(ctx context.Context, bi *BillingInfo, ts time.Time) error {
	if bi.BillingPlan.SeatPriceCents == 0 || len(bi.Organization.Users) == 0 {
		return nil
	}

	billingTime := timeutil.Floor(ts, bi.BillingPlan.Period)
	startTime := timeutil.Floor(ts, bi.BillingPlan.Period)
	endTime := timeutil.Next(ts, bi.BillingPlan.Period)

	var billedResources []*BilledResource
	for _, user := range bi.Organization.Users {
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        user.UserID,
			EntityKind:      UserEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         SeatProduct,
			Quantity:        1,
			TotalPriceCents: bi.BillingPlan.SeatPriceCents,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func commitOverageToBill(ctx context.Context, bi *BillingInfo, ts time.Time) error {
	billingTime := timeutil.Floor(ts, bi.BillingPlan.Period)
	startTime := timeutil.Last(ts, bi.BillingPlan.Period)
	endTime := timeutil.Floor(ts, bi.BillingPlan.Period)

	_, usages, err := metrics.GetHistoricalUsage(ctx, bi.OrganizationID, bi.BillingPlan.Period, startTime, endTime)
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
	readOverageGB := float32(readOverageBytes) / float32(1e9)
	readOveragePrice := int32(readOverageGB * float32(bi.BillingPlan.ReadOveragePriceCents)) // assuming unit of ReadOveragePriceCents is GB

	writeOverageBytes := usages[0].WriteBytes - *bi.Organization.PrepaidWriteQuota
	writeOverageGB := float32(writeOverageBytes) / float32(1e9)
	writeOveragePrice := int32(writeOverageGB * float32(bi.BillingPlan.WriteOveragePriceCents)) // assuming unit of WriteOveragePriceCents is GB

	scanOverageBytes := usages[0].ScanBytes - *bi.Organization.PrepaidScanQuota
	scanOverageGB := float32(scanOverageBytes) / float32(1e9)
	scanOveragePrice := int32(scanOverageGB * float32(bi.BillingPlan.ScanOveragePriceCents)) // assuming unit of ScanOveragePriceCents is GB

	var newBilledResources []*BilledResource

	if readOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        bi.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         ReadOverageProduct,
			Quantity:        readOverageGB,
			TotalPriceCents: readOveragePrice,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	if writeOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        bi.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         WriteOverageProduct,
			Quantity:        writeOverageGB,
			TotalPriceCents: writeOveragePrice,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	if scanOverageBytes > 0 {
		newBilledResources = append(newBilledResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        bi.OrganizationID,
			EntityKind:      OrganizationEntityKind,
			StartTime:       startTime,
			EndTime:         endTime,
			Product:         ScanOverageProduct,
			Quantity:        scanOverageGB,
			TotalPriceCents: scanOveragePrice,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	if len(newBilledResources) == 0 {
		return nil
	}

	err = CreateOrUpdateBilledResources(ctx, newBilledResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func recomputeOrganizationPrepaidQuotas(ctx context.Context, bi *BillingInfo) error {
	org := bi.Organization
	numSeats := int64(len(org.Users))

	if org.PrepaidReadQuota != nil {
		newPrepaidReadQuota := bi.BillingPlan.BaseReadQuota + bi.BillingPlan.SeatReadQuota*numSeats
		org.PrepaidReadQuota = &newPrepaidReadQuota
	}
	if org.PrepaidWriteQuota != nil {
		newPrepaidWriteQuota := bi.BillingPlan.BaseWriteQuota + bi.BillingPlan.SeatWriteQuota*numSeats
		org.PrepaidWriteQuota = &newPrepaidWriteQuota
	}
	if org.PrepaidScanQuota != nil {
		newPrepaidScanQuota := bi.BillingPlan.BaseScanQuota + bi.BillingPlan.SeatScanQuota*numSeats
		org.PrepaidScanQuota = &newPrepaidScanQuota
	}

	if org.PrepaidReadQuota != nil || org.PrepaidWriteQuota != nil || org.PrepaidScanQuota != nil {
		org.UpdatedOn = time.Now()
		_, err := hub.DB.WithContext(ctx).Model(org).
			Column("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
			WherePK().
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}

func commitProratedPrepaidQuotaToBill(ctx context.Context, bi *BillingInfo, billingTime time.Time) error {
	if bi.BillingPlan.BasePriceCents == 0 {
		return nil
	}

	now := time.Now()
	p := bi.BillingPlan.Period

	proratedFraction := float64(timeutil.DaysLeftInPeriod(now, p)) / float64(timeutil.TotalDaysInPeriod(now, p))
	proratedPrice := int32(math.Round(float64(bi.BillingPlan.BasePriceCents) * proratedFraction))

	var billedResources []*BilledResource
	billedResources = append(billedResources, &BilledResource{
		OrganizationID:  bi.OrganizationID,
		BillingTime:     billingTime,
		EntityID:        bi.OrganizationID,
		EntityKind:      OrganizationEntityKind,
		StartTime:       now,
		EndTime:         timeutil.Next(now, p),
		Product:         PrepaidQuotaProratedProduct,
		Quantity:        1,
		TotalPriceCents: proratedPrice,
		Currency:        bi.BillingPlan.Currency,
	})

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}

func commitProratedSeatsToBill(ctx context.Context, bi *BillingInfo, billingTime time.Time, users []*User) error {
	if bi.BillingPlan.SeatPriceCents == 0 || len(users) == 0 {
		return nil
	}

	now := time.Now()
	p := bi.BillingPlan.Period

	proratedFraction := float64(timeutil.DaysLeftInPeriod(now, p)) / float64(timeutil.TotalDaysInPeriod(now, p))
	proratedPrice := int32(math.Round(float64(bi.BillingPlan.SeatPriceCents) * proratedFraction))

	var billedResources []*BilledResource
	for _, user := range users {
		billedResources = append(billedResources, &BilledResource{
			OrganizationID:  bi.OrganizationID,
			BillingTime:     billingTime,
			EntityID:        user.UserID,
			EntityKind:      UserEntityKind,
			StartTime:       now,
			EndTime:         timeutil.Next(now, p),
			Product:         SeatProratedProduct,
			Quantity:        1,
			TotalPriceCents: proratedPrice,
			Currency:        bi.BillingPlan.Currency,
		})
	}

	err := CreateOrUpdateBilledResources(ctx, billedResources)
	if err != nil {
		panic("unable to write billed resources to table")
	}

	// done
	return nil
}
