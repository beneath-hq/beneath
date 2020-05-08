package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/metrics"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

// PublicOrganization returns the gql.PublicOrganizationResolver
func (r *Resolver) PublicOrganization() gql.PublicOrganizationResolver {
	return &publicOrganizationResolver{r}
}

type publicOrganizationResolver struct{ *Resolver }

func (r *publicOrganizationResolver) OrganizationID(ctx context.Context, obj *entity.Organization) (string, error) {
	return obj.OrganizationID.String(), nil
}

func (r *publicOrganizationResolver) PersonalUserID(ctx context.Context, obj *entity.Organization) (*uuid.UUID, error) {
	return obj.UserID, nil
}

func (r *queryResolver) Me(ctx context.Context) (*gql.PrivateOrganization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'Me'")
	}

	org := entity.FindOrganizationByUserID(ctx, secret.GetOwnerID())
	if org == nil {
		return nil, gqlerror.Errorf("Couldn't find user! This is highly irregular.")
	}

	return organizationToPrivateOrganization(ctx, org, entity.OrganizationPermissions{
		View:   true,
		Create: true,
		Admin:  true,
	}), nil
}

func (r *queryResolver) OrganizationByName(ctx context.Context, name string) (gql.Organization, error) {
	org := entity.FindOrganizationByName(ctx, name)
	if org == nil {
		return nil, gqlerror.Errorf("Organization %s not found", name)
	}

	secret := middleware.GetSecret(ctx)
	perms := organizationPermissions(ctx, secret, org)
	if !perms.View {
		// remove private projects and return a PublicOrganization
		org.StripPrivateProjects()
		return org, nil
	}

	return organizationToPrivateOrganization(ctx, org, perms), nil
}

func (r *queryResolver) OrganizationByID(ctx context.Context, organizationID uuid.UUID) (gql.Organization, error) {
	org := entity.FindOrganization(ctx, organizationID)
	if org == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := organizationPermissions(ctx, secret, org)
	if !perms.View {
		// remove private projects and return a PublicOrganization
		org.StripPrivateProjects()
		return org, nil
	}

	return organizationToPrivateOrganization(ctx, org, perms), nil
}

func (r *queryResolver) OrganizationByUserID(ctx context.Context, userID uuid.UUID) (gql.Organization, error) {
	org := entity.FindOrganizationByUserID(ctx, userID)
	if org == nil {
		return nil, gqlerror.Errorf("Organization not found for userID %s", userID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := organizationPermissions(ctx, secret, org)
	if !perms.View {
		// remove private projects and return a PublicOrganization
		org.StripPrivateProjects()
		return org, nil
	}

	return organizationToPrivateOrganization(ctx, org, perms), nil
}

func (r *queryResolver) OrganizationMembers(ctx context.Context, organizationID uuid.UUID) ([]*entity.OrganizationMember, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You're not allowed to see the members of organization %s", organizationID.String())
	}

	return entity.FindOrganizationMembers(ctx, organizationID)
}

func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*gql.PrivateOrganization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsMaster() {
		return nil, gqlerror.Errorf("Only Beneath masters can create new organizations")
	}

	organization := &entity.Organization{Name: name}
	err := organization.CreateWithUser(ctx, secret.GetOwnerID(), true, true, true)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to create organization: %s", err.Error())
	}

	perms := entity.OrganizationPermissions{
		View:   true,
		Create: true,
		Admin:  true,
	}

	return organizationToPrivateOrganization(ctx, organization, perms), nil
}

func (r *mutationResolver) UpdateOrganization(ctx context.Context, organizationID uuid.UUID, name *string, displayName *string, description *string, photoURL *string) (*gql.PrivateOrganization, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	err := organization.UpdateDetails(ctx, name, displayName, description, photoURL)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update organization name: %s", err.Error())
	}

	return organizationToPrivateOrganization(ctx, organization, perms), nil
}

func (r *mutationResolver) UpdateOrganizationQuotas(ctx context.Context, organizationID uuid.UUID, readQuota *int, writeQuota *int) (*gql.PrivateOrganization, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	if !secret.IsMaster() {
		return nil, gqlerror.Errorf("Only Beneath masters can update organization quotas directly")
	}

	// TODO: invalidate cached quotas

	err := organization.UpdateQuotas(ctx, IntToInt64(readQuota), IntToInt64(writeQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating quotas: %s", err.Error())
	}

	perms := secret.OrganizationPermissions(ctx, organizationID)
	return organizationToPrivateOrganization(ctx, organization, perms), nil
}

func (r *mutationResolver) InviteUserToOrganization(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, view bool, create bool, admin bool) (bool, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return false, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	if !organization.IsMulti() {
		return false, gqlerror.Errorf("You cannot add other users to your personal account. Instead, create a new multi-user organization (you may have to upgrade to a new plan).")
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	user := entity.FindUser(ctx, userID)
	if user == nil {
		return false, gqlerror.Errorf("User %s not found", userID.String())
	}

	if user.BillingOrganizationID == organizationID {
		return false, gqlerror.Errorf("User is already a member of the organization")
	}

	invite := entity.OrganizationInvite{
		OrganizationID: organizationID,
		UserID:         userID,
		View:           view,
		Create:         create,
		Admin:          admin,
	}
	err := invite.Upsert(ctx)
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

	invite := entity.FindOrganizationInvite(ctx, organizationID, secret.GetOwnerID())
	if invite == nil {
		return false, gqlerror.Errorf("You do not have an invitation for organization %s", organizationID.String())
	}

	err := invite.User.BillingOrganization.TransferUser(ctx, invite.User, invite.Organization)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't accept invite: %s", err.Error())
	}

	// grant access permissions
	puo := &entity.PermissionsUsersOrganizations{
		UserID:         invite.UserID,
		OrganizationID: organizationID,
	}
	err = puo.Update(ctx, &invite.View, &invite.Create, &invite.Admin)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't grant permissions: %s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) LeaveBillingOrganization(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	personalOrg := entity.FindOrganizationByUserID(ctx, userID)
	if personalOrg == nil {
		return nil, gqlerror.Errorf("User not found")
	}

	if personalOrg.OrganizationID == personalOrg.User.BillingOrganizationID {
		return nil, gqlerror.Errorf("User is not in a multi-person organization")
	}

	// organization admins or users themselves can trigger this
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		perms := secret.OrganizationPermissions(ctx, personalOrg.User.BillingOrganizationID)
		if !perms.Admin {
			return nil, gqlerror.Errorf("Must be authenticated as the user or an admin of its organization")
		}
	}

	leavingOrg := entity.FindOrganization(ctx, personalOrg.User.BillingOrganizationID)
	err := leavingOrg.TransferUser(ctx, personalOrg.User, personalOrg)
	if err != nil {
		return nil, err
	}

	// remove access permissions
	falsy := false
	puo := &entity.PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: leavingOrg.OrganizationID,
	}
	err = puo.Update(ctx, &falsy, &falsy, &falsy)
	if err != nil {
		return nil, gqlerror.Errorf("Couldn't remove permissions: %s", err.Error())
	}

	return personalOrg.User, nil
}

func (r *mutationResolver) TransferProjectToOrganization(ctx context.Context, projectID uuid.UUID, organizationID uuid.UUID) (*entity.Project, error) {
	secret := middleware.GetSecret(ctx)
	projPerms := secret.ProjectPermissions(ctx, projectID, false)
	if !projPerms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}
	orgPerms := secret.OrganizationPermissions(ctx, organizationID)
	if !orgPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to create resources in organization %s", organizationID.String())
	}

	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	if project.OrganizationID == organizationID {
		return nil, gqlerror.Errorf("Project %s is already in organization %s", projectID.String(), organizationID.String())
	}

	targetOrganization := entity.FindOrganization(ctx, organizationID)
	if targetOrganization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID)
	}

	err := project.Organization.TransferProject(ctx, project, targetOrganization)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *mutationResolver) TransferServiceToOrganization(ctx context.Context, serviceID uuid.UUID, organizationID uuid.UUID) (*entity.Service, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service %s not found", serviceID.String())
	}

	if service.OrganizationID == organizationID {
		return nil, gqlerror.Errorf("Service %s is already in organization %s", serviceID.String(), organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	servicePerms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !servicePerms.Admin {
		return nil, gqlerror.Errorf("Not allowed to administrate the service's organization %s", service.OrganizationID.String())
	}
	orgPerms := secret.OrganizationPermissions(ctx, organizationID)
	if !orgPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to create resources in the target organization %s", organizationID.String())
	}

	targetOrganization := entity.FindOrganization(ctx, organizationID)
	if targetOrganization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID)
	}

	err := service.Organization.TransferService(ctx, service, targetOrganization)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func organizationPermissions(ctx context.Context, secret entity.Secret, org *entity.Organization) entity.OrganizationPermissions {
	perms := secret.OrganizationPermissions(ctx, org.OrganizationID)

	// if you're not allowed to view the org
	if !perms.View {
		// but, it's the personal org of a user that's paid for by another organization
		if org.User != nil && org.OrganizationID != org.User.BillingOrganizationID {
			// and, you're the admin paying the bills
			billingPerms := secret.OrganizationPermissions(ctx, org.User.BillingOrganizationID)
			if billingPerms.Admin {
				// you do get to take a peek
				perms.View = true
			}
		}
	}

	return perms
}

func organizationToPrivateOrganization(ctx context.Context, o *entity.Organization, p entity.OrganizationPermissions) *gql.PrivateOrganization {
	po := &gql.PrivateOrganization{
		OrganizationID:    o.OrganizationID.String(),
		Name:              o.Name,
		DisplayName:       o.DisplayName,
		Description:       StrToPtr(o.Description),
		PhotoURL:          StrToPtr(o.PhotoURL),
		CreatedOn:         o.CreatedOn,
		UpdatedOn:         o.UpdatedOn,
		PrepaidReadQuota:  Int64ToInt(o.ReadQuota),
		PrepaidWriteQuota: Int64ToInt(o.WriteQuota),
		ReadQuota:         Int64ToInt(o.ReadQuota),
		WriteQuota:        Int64ToInt(o.WriteQuota),
		Projects:          o.Projects,
		Services:          o.Services,
	}

	usage := metrics.GetCurrentUsage(ctx, o.OrganizationID)
	po.ReadUsage = int(usage.ReadBytes)
	po.WriteUsage = int(usage.WriteBytes)

	if o.UserID != nil {
		po.PersonalUserID = o.UserID
		po.PersonalUser = o.User
	}

	if p.View || p.Create || p.Admin {
		po.Permissions = &entity.PermissionsUsersOrganizations{
			View:   p.View,
			Create: p.Create,
			Admin:  p.Admin,
		}
	}

	return po
}
