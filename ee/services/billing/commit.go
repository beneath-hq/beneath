package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/ee/models"
	"github.com/beneath-hq/beneath/infra/engine/driver"
	nee_models "github.com/beneath-hq/beneath/models"
)

// IMPORTANT: READ THIS BEFORE EDITING!
// 1. This file contains most of the hairy logic related to billing. We've tried to keep the logic in the other
// files relatively trivial, to isolate the complicated logic here. We want to keep it that way if possible.
// 2. RunNormalBilling and UpdateBillingPlan are similar, but sufficiently different that we assess they're
// easier to reason about by duplicating code (versus abstracting to helper functions). IF YOU EDIT ONE OF THEM,
// MAKE SURE TO REFLECT THE CHANGES IN THE OTHER.

// CommitBilledResources commits billed resources and advances NextBillingTime. It fails if called before NextBillingTime.
// It doesn't directly trigger an invoice, but publishes BillingCommittedEvent, which async triggers actual invoicing.
func (s *Service) CommitBilledResources(ctx context.Context, bi *models.BillingInfo) error {
	// NOTE: Read note at the top of the file first.

	// used to calculate billing periods
	quotaPeriod := s.Usage.GetQuotaPeriod(bi.Organization.QuotaEpoch)

	// get billing time we're processing for
	billingTime := bi.NextBillingTime
	if billingTime.After(time.Now()) {
		return fmt.Errorf("expected RunBilling to get called for NextBillingTime < now")
	}
	if billingTime != quotaPeriod.Floor(billingTime) {
		return fmt.Errorf("expected NextBillingTime to be at the start of a quota period")
	}

	// compute relevant timestamps for billing
	prevStart := quotaPeriod.PrevFloor(billingTime)
	nextStart := billingTime
	nextEnd := quotaPeriod.Next(billingTime)

	// get usage in prev period
	usage, err := s.Usage.GetHistoricalUsageSingle(ctx, bi.OrganizationID, driver.UsageLabelQuotaMonth, prevStart)
	if err != nil {
		return err
	}

	// get members
	members, err := s.Organizations.FindOrganizationMembers(ctx, bi.OrganizationID)
	if err != nil {
		return err
	}

	// no need for transaction because commits are idempotent

	// save overage in prev period
	err = s.CommitOverageBilledResources(ctx, bi, usage, billingTime, prevStart, nextStart)
	if err != nil {
		return err
	}

	// save base fee and seats
	err = s.CommitBaseBilledResource(ctx, bi, billingTime, nextStart, nextEnd)
	if err != nil {
		return err
	}
	err = s.CommitSeatBilledResources(ctx, bi, members, billingTime, nextStart, nextEnd)
	if err != nil {
		return err
	}

	// set next billing time
	bi.NextBillingTime = nextEnd
	bi.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, bi).
		WherePK().
		Column("next_billing_time", "updated_on").
		Update()
	if err != nil {
		return err
	}

	// recompute organization's prepaid quotas
	// necessary to account for a) a user leaving an organization mid-period b) a billing plan's parameters changing mid-period and
	err = s.ComputeAndUpdateOrganizationQuotas(ctx, bi, len(members))
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.ShouldInvoiceBilledResourcesEvent{
		OrganizationID: bi.OrganizationID,
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateBillingPlan updates an organization's billing plan, probably triggering a bill, changing billing cycles, and updating allowed quotas.
func (s *Service) UpdateBillingPlan(ctx context.Context, bi *models.BillingInfo, billingPlan *models.BillingPlan) error {
	// NOTE: Read note at the top of the file first.

	// prepare
	now := time.Now()
	oldBillingPlan := bi.BillingPlan
	newBillingPlan := billingPlan

	// done if we're already on billingPlan
	if oldBillingPlan.BillingPlanID == newBillingPlan.BillingPlanID {
		return nil
	}

	// cannot change to paid plan if no billing method set
	if !newBillingPlan.IsFree() && bi.BillingMethodID == nil {
		return fmt.Errorf("cannot upgrade to paid billing plan without a registered billing method")
	}

	// get all members, for seat-related calculations
	members, err := s.Organizations.FindOrganizationMembers(ctx, bi.OrganizationID)
	if err != nil {
		return err
	}

	// check new plan allows multiple members if relevant
	if !newBillingPlan.MultipleUsers && len(members) > 1 {
		return fmt.Errorf("cannot change plan because the new plan doesn't allow multiple users")
	}

	// check there's no billing pending (e.g. if changing plans right after the end of the period)
	// (pretty crude)
	if now.After(bi.NextBillingTime) {
		err := s.CommitBilledResources(ctx, bi)
		if err != nil {
			return err
		}
		return fmt.Errorf("billing is running for this organization, please try again in a few minutes")
	}

	// used to calculate billing periods
	oldQuotaPeriod := s.Usage.GetQuotaPeriod(bi.Organization.QuotaEpoch)
	newQuotaPeriod := s.Usage.GetQuotaPeriod(now)

	// compute relevant timestamps for billing
	billingTime := now
	prevStart := oldQuotaPeriod.Floor(billingTime)
	nextStart := billingTime
	nextEnd := newQuotaPeriod.Next(billingTime)

	// get usage in prev period
	usage, err := s.Usage.GetHistoricalUsageSingle(ctx, bi.OrganizationID, driver.UsageLabelQuotaMonth, prevStart)
	if err != nil {
		return err
	}

	// run transaction that updates billing info, and also commits billed resources and updates quotas
	err = s.DB.InTransaction(ctx, func(ctx context.Context) error {
		// add current overage to billed resources (under old plan)
		err = s.CommitOverageBilledResources(ctx, bi, usage, billingTime, prevStart, nextStart)
		if err != nil {
			return err
		}

		// update billing info
		bi.NextBillingTime = nextEnd
		bi.BillingPlan = newBillingPlan
		bi.BillingPlanID = newBillingPlan.BillingPlanID
		bi.UpdatedOn = time.Now()
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, bi).
			WherePK().
			Column("billing_plan_id", "next_billing_time", "updated_on").
			Update()
		if err != nil {
			return err
		}

		// update quota epoch
		err = s.Organizations.UpdateQuotaEpoch(ctx, bi.Organization, billingTime)
		if err != nil {
			return err
		}

		// save base fee and seats
		err = s.CommitBaseBilledResource(ctx, bi, billingTime, nextStart, nextEnd)
		if err != nil {
			return err
		}
		err = s.CommitSeatBilledResources(ctx, bi, members, billingTime, nextStart, nextEnd)
		if err != nil {
			return err
		}

		// TODO (future): credit unused resources from old plan (time-constrained)

		// update organization quotas based on new plan
		err = s.ComputeAndUpdateOrganizationQuotas(ctx, bi, len(members))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.ShouldInvoiceBilledResourcesEvent{
		OrganizationID: bi.OrganizationID,
	})
	if err != nil {
		return err
	}

	return nil
}

// ComputeAndUpdateOrganizationQuotas derives quotas for the organization (based on its billing plan), and updates them
func (s *Service) ComputeAndUpdateOrganizationQuotas(ctx context.Context, bi *models.BillingInfo, numSeats int) error {
	org := bi.Organization
	plan := bi.BillingPlan

	// update prepaid quota field (only used when showing billing info)
	org.PrepaidReadQuota = int64ToPtr(plan.BaseReadQuota + plan.SeatReadQuota*int64(numSeats))
	org.PrepaidWriteQuota = int64ToPtr(plan.BaseWriteQuota + plan.SeatWriteQuota*int64(numSeats))
	org.PrepaidScanQuota = int64ToPtr(plan.BaseScanQuota + plan.SeatScanQuota*int64(numSeats))
	org.UpdatedOn = time.Now()
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, org).
		Column("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	// update normal quotas (important, sets limits on activity)
	err = s.Organizations.UpdateQuotas(ctx, org, &plan.ReadQuota, &plan.WriteQuota, &plan.ScanQuota)
	if err != nil {
		return err
	}

	return nil
}

func int64ToPtr(val int64) *int64 {
	return &val
}

// AddProratedUser adds a prorated seat to an organization mid-period
func (s *Service) AddProratedUser(ctx context.Context, bi *models.BillingInfo, user *nee_models.User, now time.Time) error {
	quotaPeriod := s.Usage.GetQuotaPeriod(bi.Organization.QuotaEpoch)

	billingTime := now
	startTime := now
	endTime := quotaPeriod.Next(now)
	rate := float64(endTime.Sub(startTime)) / float64(quotaPeriod.PeriodDuration)

	plan := bi.BillingPlan
	deltaRead := int64(float64(plan.SeatReadQuota) * rate)
	deltaWrite := int64(float64(plan.SeatWriteQuota) * rate)
	deltaScan := int64(float64(plan.SeatScanQuota) * rate)

	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		err := s.CommitProratedSeatBilledResource(ctx, bi, user, billingTime, startTime, endTime, rate)
		if err != nil {
			return err
		}

		// NOTE: Won't get retried/double-added because it's (currently) called from a sync listener published in a tx
		_, err = s.DB.GetDB(ctx).ModelContext(ctx, bi.Organization).
			WherePK().
			Set("prepaid_read_quota = prepaid_read_quota + ?", deltaRead).
			Set("prepaid_write_quota = prepaid_write_quota + ?", deltaWrite).
			Set("prepaid_scan_quota = prepaid_scan_quota + ?", deltaScan).
			Set("updated_on = ?", now).
			Returning("prepaid_read_quota", "prepaid_write_quota", "prepaid_scan_quota", "updated_on").
			Update()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
