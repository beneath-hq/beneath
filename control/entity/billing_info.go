package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"

	uuid "github.com/satori/go.uuid"
)

// BillingInfo encapsulates an organization's billing method and billing plan
type BillingInfo struct {
	BillingInfoID   uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID  uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Organization    *Organization
	BillingMethodID *uuid.UUID `sql:"on_delete:RESTRICT,type:uuid"`
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
func (bi *BillingInfo) Update(ctx context.Context, billingMethodID *uuid.UUID, billingPlanID uuid.UUID, country string, region *string, companyName *string, taxNumber *string) (*BillingInfo, error) {
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

		var userIDs []uuid.UUID
		users := bi.Organization.Users
		for _, user := range users {
			userIDs = append(userIDs, user.UserID)
		}

		// charge organization the pro-rated amount for seats at new billing plan price
		billingTime := time.Now()
		err = commitProratedSeatsToBill(ctx, bi.OrganizationID, billingTime, newBillingPlan, userIDs, false)
		if err != nil {
			panic("could not commit prorated seats to bill")
		}

		// SCENARIO 1: Free->Pro
		// trigger bill
		if prevBillingPlan.Default {
			err = taskqueue.Submit(ctx, &SendInvoiceTask{
				OrganizationID: bi.OrganizationID,
				BillingTime:    billingTime,
			})
			if err != nil {
				log.S.Errorw("Error creating task", err)
			}
		}

		// set all user quotas to nil
		// if the new billing plan permits setting user-level quotas, then the admins should do that explicitly
		for _, u := range users {
			u.ReadQuota = nil
			u.WriteQuota = nil
		}

		_, err = hub.DB.ModelContext(ctx, &users).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
		if err != nil {
			return nil, err
		}

		// update the organization's quotas
		// additionally, the UpdateQuotas() function clears the cache for all the organization's users' secrets
		organization := FindOrganization(ctx, bi.OrganizationID)
		err = organization.UpdateQuotas(ctx, &newBillingPlan.ReadQuota, &newBillingPlan.WriteQuota)
		if err != nil {
			return nil, err
		}

		// SCENARIO 2: Pro->Free
		// LockOrganizationPrivateProjects()
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

// GetCountry implements payments/driver.BillingInfo
func (bi *BillingInfo) GetCountry() string {
	return bi.Country
}

// GetRegion implements payments/driver.BillingInfo
func (bi *BillingInfo) GetRegion() string {
	return bi.Region
}

// IsCompany implements payments/driver.BillingInfo
func (bi *BillingInfo) IsCompany() bool {
	if bi.CompanyName != "" {
		return true
	}
	return false
}
