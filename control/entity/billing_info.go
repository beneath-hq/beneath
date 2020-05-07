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
		Relation("BillingPlan").
		Relation("BillingMethod").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingInfo
}

// FindAllPayingBillingInfos returns all billing infos where the organization is on a paid plan
func FindAllPayingBillingInfos(ctx context.Context) []*BillingInfo {
	var billingInfos []*BillingInfo
	err := hub.DB.ModelContext(ctx, &billingInfos).
		Column("billing_info.*").
		Relation("Organization.Users").
		Relation("BillingPlan").
		Relation("BillingMethod").
		Where("billing_plan.default = false").
		Select()
	if err != nil {
		panic(err)
	}
	return billingInfos
}

// Update updates an organization's billing method and billing plan
// if switching billing plans:
// - trigger bill
// - update all parameters related to the new billing plan's features
func (bi *BillingInfo) Update(ctx context.Context, billingMethodID *uuid.UUID, billingPlanID uuid.UUID, country string, region *string, companyName *string, taxNumber *string) (*BillingInfo, error) {
	// TODO: start a big postgres transaction that will encompass all the updates in this function
	prevBillingPlan := bi.BillingPlan
	var newBillingPlan *BillingPlan
	// check for switching billing plans
	if billingPlanID != prevBillingPlan.BillingPlanID {
		newBillingPlan = FindBillingPlan(ctx, billingPlanID)
		if newBillingPlan == nil {
			panic("could not get new billing plan")
		}

		// downgrading to Free plan
		if newBillingPlan.Default {
			// trigger bill for overage assessment
			err := taskqueue.Submit(ctx, &SendInvoiceTask{
				BillingInfo: bi,
				BillingTime: time.Now(),
			})
			if err != nil {
				log.S.Errorw("Error creating task", err)
			}

			// TODO: LockOrganizationPrivateProjects()
		}
	}

	// make billing info updates
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

	if billingPlanID != prevBillingPlan.BillingPlanID {
		// Upgrades
		if !newBillingPlan.Default {
			billingTime := time.Now()
			err = commitProratedSeatsToBill(ctx, bi.OrganizationID, billingTime, newBillingPlan, bi.Organization.Users, false)
			if err != nil {
				panic("could not commit prorated seats to bill")
			}

			// upgrading from Free plan
			if prevBillingPlan.Default {
				// trigger bill for prorated seat
				err = taskqueue.Submit(ctx, &SendInvoiceTask{
					BillingInfo: bi,
					BillingTime: billingTime,
				})
				if err != nil {
					log.S.Errorw("Error creating task", err)
				}
			}
		}

		// set all user quotas to nil
		// if the new billing plan permits setting user-level quotas, then the admins should do that explicitly
		for _, u := range bi.Organization.Users {
			u.ReadQuota = nil
			u.WriteQuota = nil
		}

		_, err = hub.DB.ModelContext(ctx, &bi.Organization.Users).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
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
