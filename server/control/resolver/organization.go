package resolver

import (
	"context"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// PublicOrganization returns the gql.PublicOrganizationResolver
func (r *Resolver) PublicOrganization() gql.PublicOrganizationResolver {
	return &publicOrganizationResolver{r}
}

type publicOrganizationResolver struct{ *Resolver }

func (r *publicOrganizationResolver) OrganizationID(ctx context.Context, obj *models.Organization) (string, error) {
	return obj.OrganizationID.String(), nil
}

func (r *publicOrganizationResolver) PersonalUserID(ctx context.Context, obj *models.Organization) (*uuid.UUID, error) {
	return obj.UserID, nil
}

func (r *queryResolver) Me(ctx context.Context) (*gql.PrivateOrganization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'Me'")
	}

	org := r.Organizations.FindOrganizationByUserID(ctx, secret.GetOwnerID())
	if org == nil {
		return nil, gqlerror.Errorf("Couldn't find user! This is highly irregular.")
	}

	return r.organizationToPrivateOrganization(ctx, org, models.OrganizationPermissions{
		View:   true,
		Create: true,
		Admin:  true,
	}), nil
}

func (r *queryResolver) OrganizationByName(ctx context.Context, name string) (gql.Organization, error) {
	org := r.Organizations.FindOrganizationByName(ctx, name)
	if org == nil {
		return nil, gqlerror.Errorf("Organization %s not found", name)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.organizationPermissions(ctx, secret, org)
	if !perms.View {
		// remove private projects and return a PublicOrganization
		org.StripPrivateProjects()
		return org, nil
	}

	return r.organizationToPrivateOrganization(ctx, org, perms), nil
}

func (r *queryResolver) OrganizationByID(ctx context.Context, organizationID uuid.UUID) (gql.Organization, error) {
	org := r.Organizations.FindOrganization(ctx, organizationID)
	if org == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.organizationPermissions(ctx, secret, org)
	if !perms.View {
		// remove private projects and return a PublicOrganization
		org.StripPrivateProjects()
		return org, nil
	}

	return r.organizationToPrivateOrganization(ctx, org, perms), nil
}

func (r *queryResolver) OrganizationByUserID(ctx context.Context, userID uuid.UUID) (gql.Organization, error) {
	org := r.Organizations.FindOrganizationByUserID(ctx, userID)
	if org == nil {
		return nil, gqlerror.Errorf("Organization not found for userID %s", userID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.organizationPermissions(ctx, secret, org)
	if !perms.View {
		// remove private projects and return a PublicOrganization
		org.StripPrivateProjects()
		return org, nil
	}

	return r.organizationToPrivateOrganization(ctx, org, perms), nil
}

func (r *queryResolver) OrganizationMembers(ctx context.Context, organizationID uuid.UUID) ([]*models.OrganizationMember, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You're not allowed to see the members of organization %s", organizationID.String())
	}

	return r.Organizations.FindOrganizationMembers(ctx, organizationID)
}

func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*gql.PrivateOrganization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsMaster() {
		return nil, gqlerror.Errorf("Only Beneath masters can create new organizations")
	}

	organization, err := r.Organizations.CreateWithUser(ctx, name, secret.GetOwnerID(), true, true, true)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to create organization: %s", err.Error())
	}

	perms := models.OrganizationPermissions{
		View:   true,
		Create: true,
		Admin:  true,
	}

	return r.organizationToPrivateOrganization(ctx, organization, perms), nil
}

func (r *mutationResolver) UpdateOrganization(ctx context.Context, organizationID uuid.UUID, name *string, displayName *string, description *string, photoURL *string) (*gql.PrivateOrganization, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	organization := r.Organizations.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	err := r.Organizations.UpdateDetails(ctx, organization, name, displayName, description, photoURL)
	if err != nil {
		// specifically handle the common error that a username is already taken
		if strings.Contains(err.Error(), `duplicate key value violates unique constraint "organizations_name_key"`) {
			path := graphql.GetResolverContext(ctx).Path()
			path = append(path, ast.PathName("name"))
			if organization.UserID != nil {
				return nil, gqlerror.ErrorPathf(path, "Username already taken")
			}
			return nil, gqlerror.ErrorPathf(path, "Name already taken")
		}
		return nil, gqlerror.Errorf("Failed to update organization: %s", err.Error())
	}

	return r.organizationToPrivateOrganization(ctx, organization, perms), nil
}

func (r *mutationResolver) UpdateOrganizationQuotas(ctx context.Context, organizationID uuid.UUID, readQuota *int, writeQuota *int, scanQuota *int) (*gql.PrivateOrganization, error) {
	organization := r.Organizations.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	if !secret.IsMaster() {
		return nil, gqlerror.Errorf("Only Beneath masters can update organization quotas directly")
	}

	err := r.Organizations.UpdateQuotas(ctx, organization, IntToInt64(readQuota), IntToInt64(writeQuota), IntToInt64(scanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating quotas: %s", err.Error())
	}

	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	return r.organizationToPrivateOrganization(ctx, organization, perms), nil
}

func (r *mutationResolver) InviteUserToOrganization(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, view bool, create bool, admin bool) (bool, error) {
	organization := r.Organizations.FindOrganization(ctx, organizationID)
	if organization == nil {
		return false, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	if !organization.IsMulti() {
		return false, gqlerror.Errorf("You cannot add other users to your personal account. Instead, create a new multi-user organization (you may have to upgrade to a new plan).")
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	user := r.Users.FindUser(ctx, userID)
	if user == nil {
		return false, gqlerror.Errorf("User %s not found", userID.String())
	}

	if user.BillingOrganizationID == organizationID {
		return false, gqlerror.Errorf("User is already a member of the organization")
	}

	err := r.Organizations.CreateOrUpdateInvite(ctx, organizationID, userID, view, create, admin)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't create invite: %s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) AcceptOrganizationInvite(ctx context.Context, organizationID uuid.UUID) (bool, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return false, MakeUnauthenticatedError("Must be authenticated with a personal key")
	}

	invite := r.Organizations.FindOrganizationInvite(ctx, organizationID, secret.GetOwnerID())
	if invite == nil {
		return false, gqlerror.Errorf("You do not have an invitation for organization %s", organizationID.String())
	}

	err := r.Organizations.TransferUser(ctx, invite.User.BillingOrganization, invite.Organization, invite.User)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't accept invite: %s", err.Error())
	}

	// grant access permissions
	puo := &models.PermissionsUsersOrganizations{
		UserID:         invite.UserID,
		OrganizationID: organizationID,
	}
	err = r.Permissions.UpdateUserOrganizationPermission(ctx, puo, &invite.View, &invite.Create, &invite.Admin)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't grant permissions: %s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) LeaveBillingOrganization(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	personalOrg := r.Organizations.FindOrganizationByUserID(ctx, userID)
	if personalOrg == nil {
		return nil, gqlerror.Errorf("User not found")
	}

	if personalOrg.OrganizationID == personalOrg.User.BillingOrganizationID {
		return nil, gqlerror.Errorf("User is not in a multi-person organization")
	}

	// organization admins or users themselves can trigger this
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, personalOrg.User.BillingOrganizationID)
		if !perms.Admin {
			return nil, gqlerror.Errorf("Must be authenticated as the user or an admin of its organization")
		}
	}

	leavingOrg := r.Organizations.FindOrganization(ctx, personalOrg.User.BillingOrganizationID)
	err := r.Organizations.TransferUser(ctx, leavingOrg, personalOrg, personalOrg.User)
	if err != nil {
		return nil, err
	}

	// remove access permissions
	falsy := false
	puo := &models.PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: leavingOrg.OrganizationID,
	}
	err = r.Permissions.UpdateUserOrganizationPermission(ctx, puo, &falsy, &falsy, &falsy)
	if err != nil {
		return nil, gqlerror.Errorf("Couldn't remove permissions: %s", err.Error())
	}

	return personalOrg.User, nil
}

func (r *mutationResolver) TransferProjectToOrganization(ctx context.Context, projectID uuid.UUID, organizationID uuid.UUID) (*models.Project, error) {
	secret := middleware.GetSecret(ctx)
	projPerms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, projectID, false)
	if !projPerms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}
	orgPerms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !orgPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to create resources in organization %s", organizationID.String())
	}

	project := r.Projects.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	if project.OrganizationID == organizationID {
		return nil, gqlerror.Errorf("Project %s is already in organization %s", projectID.String(), organizationID.String())
	}

	targetOrganization := r.Organizations.FindOrganization(ctx, organizationID)
	if targetOrganization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID)
	}

	err := r.Organizations.TransferProject(ctx, project.Organization, targetOrganization, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *queryResolver) organizationPermissions(ctx context.Context, secret models.Secret, org *models.Organization) models.OrganizationPermissions {
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, org.OrganizationID)

	// if you're not allowed to view the org
	if !perms.View {
		// but, it's the personal org of a user that's paid for by another organization
		if org.User != nil && org.OrganizationID != org.User.BillingOrganizationID {
			// and, you're the admin paying the bills
			billingPerms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, org.User.BillingOrganizationID)
			if billingPerms.Admin {
				// you do get to take a peek
				perms.View = true
			}
		}
	}

	return perms
}

func (r *Resolver) organizationToPrivateOrganization(ctx context.Context, o *models.Organization, p models.OrganizationPermissions) *gql.PrivateOrganization {
	po := &gql.PrivateOrganization{
		OrganizationID: o.OrganizationID.String(),
		Name:           o.Name,
		DisplayName:    o.DisplayName,
		Description:    StrToPtr(o.Description),
		PhotoURL:       StrToPtr(o.PhotoURL),
		CreatedOn:      o.CreatedOn,
		UpdatedOn:      o.UpdatedOn,
		Projects:       o.Projects,
	}

	if o.UserID != nil {
		po.PersonalUserID = o.UserID
		po.PersonalUser = o.User
	}

	isNonBillingUser := o.User != nil && o.User.BillingOrganizationID != o.OrganizationID
	if isNonBillingUser {
		po.QuotaEpoch = o.User.QuotaEpoch
		po.QuotaStartTime = r.Usage.GetQuotaPeriod(po.QuotaEpoch).Floor(time.Now())
		po.QuotaEndTime = r.Usage.GetQuotaPeriod(po.QuotaEpoch).Next(time.Now())
		po.PrepaidReadQuota = nil
		po.PrepaidWriteQuota = nil
		po.PrepaidScanQuota = nil
		po.ReadQuota = Int64ToInt(o.User.ReadQuota)
		po.WriteQuota = Int64ToInt(o.User.WriteQuota)
		po.ScanQuota = Int64ToInt(o.User.ScanQuota)

		usage := r.Usage.GetCurrentQuotaUsage(ctx, o.User.UserID, o.User.QuotaEpoch)
		po.ReadUsage = int(usage.ReadBytes)
		po.WriteUsage = int(usage.WriteBytes)
		po.ScanUsage = int(usage.ScanBytes)
	} else {
		po.QuotaEpoch = o.QuotaEpoch
		po.QuotaStartTime = r.Usage.GetQuotaPeriod(o.QuotaEpoch).Floor(time.Now())
		po.QuotaEndTime = r.Usage.GetQuotaPeriod(o.QuotaEpoch).Next(time.Now())
		po.PrepaidReadQuota = Int64ToInt(o.PrepaidReadQuota)
		po.PrepaidWriteQuota = Int64ToInt(o.PrepaidWriteQuota)
		po.PrepaidScanQuota = Int64ToInt(o.PrepaidScanQuota)
		po.ReadQuota = Int64ToInt(o.ReadQuota)
		po.WriteQuota = Int64ToInt(o.WriteQuota)
		po.ScanQuota = Int64ToInt(o.ScanQuota)

		usage := r.Usage.GetCurrentQuotaUsage(ctx, o.OrganizationID, o.QuotaEpoch)
		po.ReadUsage = int(usage.ReadBytes)
		po.WriteUsage = int(usage.WriteBytes)
		po.ScanUsage = int(usage.ScanBytes)
	}

	if p.View || p.Create || p.Admin {
		po.Permissions = &models.PermissionsUsersOrganizations{
			View:   p.View,
			Create: p.Create,
			Admin:  p.Admin,
		}
	}

	return po
}
