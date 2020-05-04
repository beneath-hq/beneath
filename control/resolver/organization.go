package resolver

import (
	"context"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/middleware"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Organization returns the gql.OrganizationResolver
func (r *Resolver) Organization() gql.OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

func (r *organizationResolver) OrganizationID(ctx context.Context, obj *entity.Organization) (string, error) {
	return obj.OrganizationID.String(), nil
}

func (r *queryResolver) OrganizationByName(ctx context.Context, name string) (*entity.Organization, error) {
	organization := entity.FindOrganizationByName(ctx, name)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", name)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organization.OrganizationID)
	user := entity.FindUser(ctx, secret.GetOwnerID()) // get the user's projects

	// if you're not a member of the organization, hide the services
	if !perms.View {
		organization.Services = nil
	}

	// if you don't have permission for a project, hide it
	for i, orgProject := range organization.Projects {
		if !orgProject.Public {
			hide := true
			if user != nil {
				for _, userProject := range user.Projects {
					if orgProject.ProjectID == userProject.ProjectID {
						hide = false
						break
					}
				}
			}
			if hide {
				sliceLength := len(organization.Projects)
				organization.Projects[sliceLength-1], organization.Projects[i] = organization.Projects[i], organization.Projects[sliceLength-1]
				organization.Projects = organization.Projects[:sliceLength-1]
			}
		}
	}

	return organization, nil
}

func (r *queryResolver) UsersOrganizationPermissions(ctx context.Context, organizationID uuid.UUID) ([]*entity.PermissionsUsersOrganizations, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You are not allowed to view organization %s", organizationID.String())
	}

	permissions := entity.FindOrganizationPermissions(ctx, organizationID)
	if permissions == nil {
		return nil, gqlerror.Errorf("Permissions not found for organization %s", organizationID.String())
	}

	return permissions, nil
}

func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*entity.Organization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsMaster() {
		return nil, gqlerror.Errorf("Only Beneath masters can create new organizations")
	}

	organization := &entity.Organization{
		Name:     name,
		Personal: false,
	}

	err := organization.CreateWithUser(ctx, secret.GetOwnerID(), true, true, true)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *mutationResolver) UpdateOrganizationName(ctx context.Context, organizationID uuid.UUID, name string) (*entity.Organization, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	if organization.Personal {
		return nil, gqlerror.Errorf("Cannot change the name of a personal organization. Instead, to achieve the same effect, update your username.")
	}

	organization, err := organization.UpdateName(ctx, name)
	if err != nil {
		return nil, gqlerror.Errorf("Failed to update organization name: %s", err.Error())
	}

	return organization, nil
}

func (r *mutationResolver) InviteUserToOrganization(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, view bool, create bool, admin bool) (bool, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return false, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	if organization.Personal {
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
	err := invite.Save(ctx)
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

func (r *mutationResolver) UpdateUserQuotas(ctx context.Context, userID uuid.UUID, readQuota *int, writeQuota *int) (*entity.User, error) {
	user := entity.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("User %s not found", userID.String())
	}

	organizationID := user.BillingOrganizationID

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	// TODO: invalidate cached quotas

	err := user.UpdateQuotas(ctx, readQuota, writeQuota)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating quotas: %s", err.Error())
	}

	return user, nil
}

func (r *mutationResolver) LeaveBillingOrganization(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user := entity.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("User not found")
	}

	// organization admins or users themselves can trigger this
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		perms := secret.OrganizationPermissions(ctx, user.BillingOrganizationID)
		if !perms.Admin {
			return nil, gqlerror.Errorf("Must be authenticated as the user or an admin of its organization")
		}
	}

	if user.BillingOrganizationID == user.PersonalOrganizationID {
		return nil, gqlerror.Errorf("User is not in a multi-person organization")
	}

	leavingOrg := user.BillingOrganization
	err := leavingOrg.TransferUser(ctx, user, user.PersonalOrganization)
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

	return user, nil
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
