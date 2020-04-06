package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/internal/hub"

	uuid "github.com/satori/go.uuid"
)

// BillingInfo encapsulates an organization's billing information
type BillingInfo struct {
	BillingInfoID  uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Organization   *Organization
	BillingPlanID  uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	BillingPlan    *BillingPlan
	PaymentsDriver PaymentsDriver
	DriverPayload  map[string]interface{}
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
}

// FindBillingInfo finds an organization's billing info by organizationID
func FindBillingInfo(ctx context.Context, organizationID uuid.UUID) *BillingInfo {
	billingInfo := &BillingInfo{
		OrganizationID: organizationID,
	}
	err := hub.DB.ModelContext(ctx, billingInfo).
		Where("billing_info.organization_id = ?", organizationID).
		Column("billing_info.*").
		Relation("Organization.Users").
		Relation("BillingPlan").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingInfo
}

// UpdateBillingInfo updates an organization's billing info
func UpdateBillingInfo(ctx context.Context, organizationID uuid.UUID, billingPlanID uuid.UUID, paymentsDriver PaymentsDriver, driverPayload map[string]interface{}) (*BillingInfo, error) {
	bi := FindBillingInfo(ctx, organizationID)
	if bi == nil {
		panic("could not get existing billing info")
	}
	prevBillingPlan := bi.BillingPlan
	users := bi.Organization.Users

	newBillingPlan := FindBillingPlan(ctx, billingPlanID)
	if newBillingPlan == nil {
		panic("could not get new billing plan")
	}

	// if switching billing plans:
	// - reconcile next month's bill
	// - update all parameters related to the new billing plan's features
	if billingPlanID != prevBillingPlan.BillingPlanID {
		var userIDs []uuid.UUID
		var usernames []string
		for _, user := range users {
			userIDs = append(userIDs, user.UserID)
			usernames = append(usernames, user.Username)
		}

		// give organization pro-rated credit for seats at old billing plan price
		err := commitProratedSeatsToBill(ctx, organizationID, prevBillingPlan, userIDs, usernames, true)
		if err != nil {
			panic("could not commit prorated seat credits to bill")
		}

		// charge organization the pro-rated amount for seats at new billing plan price
		err = commitProratedSeatsToBill(ctx, organizationID, newBillingPlan, userIDs, usernames, false)
		if err != nil {
			panic("could not commit prorated seats to bill")
		}

		// update the organization's users' quotas
		for _, u := range users {
			u.ReadQuota = newBillingPlan.SeatReadQuota
			u.WriteQuota = newBillingPlan.SeatWriteQuota

			// clear cache for the user's secrets
			secrets := FindUserSecrets(ctx, u.UserID)
			if secrets == nil {
				panic("could not find user secrets")
			}

			for _, s := range secrets {
				getSecretCache().Delete(ctx, s.HashedToken)
			}
		}

		_, err = hub.DB.ModelContext(ctx, &users).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
		if err != nil {
			return nil, err
		}

		// update organization's personal status
		organization := FindOrganization(ctx, organizationID)
		if organization == nil {
			panic("could not get organization")
		}

		err = organization.UpdatePersonalStatus(ctx, newBillingPlan.Personal)
		if err != nil {
			return nil, err
		}

		// if downgrading from private projects, lock down organization's outstanding private projects
		if !newBillingPlan.PrivateProjects && prevBillingPlan.PrivateProjects {
			for _, p := range organization.Projects {
				if !p.Public {
					err = p.SetLock(ctx, true)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	bi.OrganizationID = organizationID
	bi.BillingPlanID = billingPlanID
	bi.PaymentsDriver = paymentsDriver
	bi.DriverPayload = driverPayload

	// upsert
	_, err := hub.DB.ModelContext(ctx, bi).OnConflict("(billing_info_id) DO UPDATE").Insert()
	if err != nil {
		return nil, err
	}

	// TODO: if going from Free -> Pro plan, trigger bill
	// need to be charged the one prorated seat, nothing more
	// probably, skip right to the IssueInvoice() task
	// probably, the prorated seat needs to have a billing time of LAST period, so that we can IssueInvoice() and pretend we're at the beginning of last period, where there won't be product conflicts
	// problem: when adding users to an organization, prorated seats have billing time of NEXT period
	// err = taskqueue.Submit(context.Background(), &SendInvoiceTask{
	// 	OrganizationID: organizationID,
	// 	BillingTime:    timeutil.Floor(time.Now(), newBillingPlan.Period), // simulate that the bill was sent at the beginning of this period
	// })
	// if err != nil {
	// 	log.S.Errorw("Error creating task", err)
	// }

	return bi, nil
}

// GetOrganizationID implements payments/driver.BillingInfo
func (bi *BillingInfo) GetOrganizationID() uuid.UUID {
	return bi.OrganizationID
}

// GetBillingPlanCurrency implements payments/driver.BillingInfo
func (bi *BillingInfo) GetBillingPlanCurrency() string {
	return string(bi.BillingPlan.Currency)
}

// GetDriverPayload implements payments/driver.BillingInfo
func (bi *BillingInfo) GetDriverPayload() map[string]interface{} {
	return bi.DriverPayload
}

// GetPaymentsDriver implements payments/driver.BillingInfo
func (bi *BillingInfo) GetPaymentsDriver() string {
	return string(bi.PaymentsDriver)
}
