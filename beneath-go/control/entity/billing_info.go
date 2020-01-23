package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/db"
	uuid "github.com/satori/go.uuid"
)

// BillingInfo encapsulates an organization's billing information
type BillingInfo struct {
	OrganizationID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	BillingPlanID  uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	BillingPlan    *BillingPlan
	PaymentsDriver PaymentsDriver // StripeCustomerID is currently the only thing that goes in the payload
	DriverPayload  map[string]interface{}
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	DeletedOn      time.Time
	Services       []*Service
	Users          []*User
}

// FindBillingInfo finds an organization's billing info by organizationID
func FindBillingInfo(ctx context.Context, organizationID uuid.UUID) *BillingInfo {
	billingInfo := &BillingInfo{
		OrganizationID: organizationID,
	}
	err := db.DB.ModelContext(ctx, billingInfo).WherePK().Column("billing_info.*", "BillingPlan", "Services", "Users").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return billingInfo
}

// UpdateBillingInfo creates or updates an organization's billing info
func UpdateBillingInfo(ctx context.Context, organizationID uuid.UUID, billingPlanID uuid.UUID, paymentsDriver PaymentsDriver, driverPayload map[string]interface{}) (*BillingInfo, error) {
	prevBillingInfo := FindBillingInfo(ctx, organizationID)
	if prevBillingInfo == nil {
		panic("could not get existing billing info")
	}
	prevBillingPlan := prevBillingInfo.BillingPlan
	users := prevBillingInfo.Users

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
		}

		_, err = db.DB.ModelContext(ctx, &users).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
		if err != nil {
			return nil, err
		}

		// update organization's personal status
		organization := FindOrganization(ctx, organizationID)
		if organization == nil {
			panic("could not get organization")
		}

		if newBillingPlan.Personal {
			err = organization.UpdatePersonalStatus(ctx, true)
			if err != nil {
				panic("Error updating organization")
			}
		}

		// TODO: if upgrading to a non-personal plan, trigger an email to the user:
		// -- if prevBillingPlan.Personal && !newBillingPlan.Personal, need to prompt the user to
		// -- 1) change the name of their organization, so its the enterprise's name, not the original user's name
		// -- 2) add teammates (which will call "InviteUser"; then the user will have to accept the invite, which will call "JoinOrganization")
		// Q: should this code go inside the organization.UpdatePersonal function?

		// TODO: if downgrading from private projects, lock down organization's outstanding private projects
		// if !newBillingPlan.PrivateProjects && prevBillingPlan.PrivateProjects {
		// 	// option1: lock projects so that they can't be viewed (e.g. create a new column Project.Locked)
		// 	// option2: clear existing project permissions, and ensure they can't be added back
		// }
	}

	bi := &BillingInfo{
		OrganizationID: organizationID,
		BillingPlanID:  billingPlanID,
		PaymentsDriver: paymentsDriver,
		DriverPayload:  driverPayload,
	}

	// validate
	err := GetValidator().Struct(bi)
	if err != nil {
		return nil, err
	}

	// insert
	_, err = db.DB.ModelContext(ctx, bi).OnConflict("(organization_id) DO UPDATE").Insert()
	if err != nil {
		return nil, err
	}

	// Q: do I need to fetch bi so that the joined tables are attached?

	return bi, nil
}
