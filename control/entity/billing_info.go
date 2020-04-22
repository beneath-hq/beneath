package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/internal/hub"

	uuid "github.com/satori/go.uuid"
)

// BillingInfo encapsulates an organization's billing method and billing plan
type BillingInfo struct {
	BillingInfoID   uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID  uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Organization    *Organization
	BillingMethodID uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	BillingMethod   *BillingMethod
	BillingPlanID   uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	BillingPlan     *BillingPlan
	Country         string
	Region          string
	CompanyName     string
	TaxNumber       string
	CreatedOn       time.Time `sql:",default:now()"`
	UpdatedOn       time.Time `sql:",default:now()"`
}

// FindBillingInfo finds an organization's billing info by organizationID
func FindBillingInfo(ctx context.Context, organizationID uuid.UUID) *BillingInfo {
	billingInfo := &BillingInfo{
		OrganizationID: organizationID,
	}
	err := hub.DB.ModelContext(ctx, billingInfo).
		Where("billing_info.organization_id = ?", organizationID).
		Column("billing_info.*").
		Relation("Organization.Users"). // TODO: this isn't needed for viewing an organization's billing page, but it's needed for computing the monthly bill
		Relation("BillingPlan").
		Relation("BillingMethod").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingInfo
}

// Update updates an organization's billing method and billing plan
func (bi *BillingInfo) Update(ctx context.Context, billingMethodID uuid.UUID, billingPlanID uuid.UUID, country string, region *string, companyName *string, taxNumber *string) (*BillingInfo, error) {
	// TODO: start a big postgres transaction that will encompass all the updates in this function
	bi.BillingMethodID = billingMethodID
	bi.BillingPlanID = billingPlanID
	bi.Country = country
	if region != nil {
		bi.Region = *region
	}
	if companyName != nil {
		bi.CompanyName = *companyName
	}
	if taxNumber != nil {
		bi.TaxNumber = *taxNumber
	}

	// upsert
	_, err := hub.DB.ModelContext(ctx, bi).OnConflict("(billing_info_id) DO UPDATE").Insert()
	if err != nil {
		return nil, err
	}

	// if switching billing plans:
	// - reconcile next month's bill
	// - update all parameters related to the new billing plan's features
	prevBillingPlan := bi.BillingPlan
	if billingPlanID != prevBillingPlan.BillingPlanID {
		newBillingPlan := FindBillingPlan(ctx, billingPlanID)
		if newBillingPlan == nil {
			panic("could not get new billing plan")
		}

		// SCENARIO 1: Free->Pro
		// TODO: trigger bill
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

		// SCENARIO 2: Pro->Free
		// LockOrganizationProjects()

		// SCENARIO 3: upgrading to enterprise plan
		// if prevBillingPlan.Personal && !newBillingPlan.Personal {
		// 	CreateOrganizationWithUser()
		// then user needs to:
		// -- 1) change the name of their organization, so its the enterprise's name, not the original user's name
		// -- 2) add teammates (which will call "InviteUser"; then the user will have to accept the invite, which will call "JoinOrganization")
		// }

		// SCENARIO 4: downgrading from enterprise plan
		// if newBillingPlan.Personal && !prevBillingPlan.Personal {
		// 	DeleteEnterpriseOrganization()
		//  LockOrganizationProjects()
		// }

		var userIDs []uuid.UUID
		var usernames []string
		users := bi.Organization.Users
		for _, user := range users {
			userIDs = append(userIDs, user.UserID)
			usernames = append(usernames, user.Username)
		}

		// give (previous) organization pro-rated credit for seats at old billing plan price
		err := commitProratedSeatsToBill(ctx, bi.OrganizationID, prevBillingPlan, userIDs, usernames, true)
		if err != nil {
			panic("could not commit prorated seat credits to bill")
		}

		// charge (new) organization the pro-rated amount for seats at new billing plan price
		err = commitProratedSeatsToBill(ctx, bi.OrganizationID, newBillingPlan, userIDs, usernames, false)
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
	}

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
	return bi.BillingMethod.DriverPayload
}

// GetPaymentsDriver implements payments/driver.BillingInfo
func (bi *BillingInfo) GetPaymentsDriver() string {
	return string(bi.BillingMethod.PaymentsDriver)
}
