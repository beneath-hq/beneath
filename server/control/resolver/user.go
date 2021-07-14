package resolver

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/server/control/gql"
	"github.com/beneath-hq/beneath/services/middleware"
)

// PrivateUser returns the gql.UserResolver
func (r *Resolver) PrivateUser() gql.PrivateUserResolver {
	return &privateUserResolver{r}
}

type privateUserResolver struct{ *Resolver }

func (r *privateUserResolver) UserID(ctx context.Context, obj *models.User) (string, error) {
	return obj.UserID.String(), nil
}

func (r *privateUserResolver) QuotaStartTime(ctx context.Context, obj *models.User) (*time.Time, error) {
	t := r.Usage.GetQuotaPeriod(obj.QuotaEpoch).Floor(time.Now())
	return &t, nil
}

func (r *privateUserResolver) QuotaEndTime(ctx context.Context, obj *models.User) (*time.Time, error) {
	t := r.Usage.GetQuotaPeriod(obj.QuotaEpoch).Next(time.Now())
	return &t, nil
}

func (r *mutationResolver) RegisterUserConsent(ctx context.Context, userID uuid.UUID, terms *bool, newsletter *bool) (*models.User, error) {
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		return nil, gqlerror.Errorf("You can only register consent for yourself")
	}

	user := r.Users.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("Couldn't find user with ID %s", userID.String())
	}

	err := r.Users.UpdateConsent(ctx, user, terms, newsletter)
	if err != nil {
		return nil, gqlerror.Errorf("Error: %s", err.Error())
	}

	return user, nil
}

func (r *mutationResolver) UpdateUserQuotas(ctx context.Context, userID uuid.UUID, readQuota *int, writeQuota *int, scanQuota *int) (*models.User, error) {
	org := r.Organizations.FindOrganizationByUserID(ctx, userID)
	if org == nil {
		return nil, gqlerror.Errorf("User %s not found", userID.String())
	}

	if org.IsBillingOrganizationForUser() {
		return nil, gqlerror.Errorf("You cannot set user quotas in single-user organizations")
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, org.User.BillingOrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", org.User.BillingOrganizationID.String())
	}

	err := r.Users.UpdateQuotas(ctx, org.User, IntToInt64(readQuota), IntToInt64(writeQuota), IntToInt64(scanQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating quotas: %s", err.Error())
	}

	return org.User, nil
}

func (r *mutationResolver) UpdateUserProjectPermissions(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, view *bool, create *bool, admin *bool) (*models.PermissionsUsersProjects, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, projectID, false)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}

	pup := r.Permissions.FindPermissionsUsersProjects(ctx, userID, projectID)
	if pup == nil {
		pup = &models.PermissionsUsersProjects{
			UserID:    userID,
			ProjectID: projectID,
		}
	}

	err := r.Permissions.UpdateUserProjectPermission(ctx, pup, view, create, admin)
	if err != nil {
		return nil, err
	}

	return pup, nil
}

func (r *mutationResolver) UpdateUserOrganizationPermissions(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, view *bool, create *bool, admin *bool) (*models.PermissionsUsersOrganizations, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	organization := r.Organizations.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization not found")
	}

	if !organization.IsMulti() {
		return nil, gqlerror.Errorf("Cannot edit permissions of personal organization")
	}

	puo := r.Permissions.FindPermissionsUsersOrganizations(ctx, userID, organizationID)
	if puo == nil {
		puo = &models.PermissionsUsersOrganizations{
			UserID:         userID,
			OrganizationID: organizationID,
		}
	}

	err := r.Permissions.UpdateUserOrganizationPermission(ctx, puo, view, create, admin)
	if err != nil {
		return nil, err
	}

	return puo, nil
}

// AuthTicket returns a gql.AuthTicketResolver
func (r *Resolver) AuthTicket() gql.AuthTicketResolver { return &authTicketResolver{r} }

type authTicketResolver struct{ *Resolver }

func (r *authTicketResolver) IssuedSecret(ctx context.Context, obj *models.AuthTicket) (*gql.NewUserSecret, error) {
	if obj.IssuedSecret != nil {
		return &gql.NewUserSecret{
			Secret: obj.IssuedSecret,
			Token:  obj.IssuedSecret.Token.String(),
		}, nil
	}
	return nil, nil
}

func (r *queryResolver) AuthTicketByID(ctx context.Context, authTicketID uuid.UUID) (*models.AuthTicket, error) {
	ticket := r.Users.FindAuthTicket(ctx, authTicketID)
	if ticket == nil {
		return nil, gqlerror.Errorf("Authentication ticket not found (it was likely denied or timed out)")
	}

	if ticket.ApproverUserID != nil {
		secret, err := r.Secrets.CreateUserSecret(ctx, *ticket.ApproverUserID, ticket.RequesterName, false, false)
		if err != nil {
			return nil, gqlerror.Errorf("Error creating secret: %s", err.Error())
		}

		ticket.IssuedSecret = secret

		err = r.Users.DeleteAuthTicket(ctx, ticket)
		if err != nil {
			return nil, gqlerror.Errorf("Error creating ticket: %s", err.Error())
		}
	}

	return ticket, nil
}

func (r *mutationResolver) CreateAuthTicket(ctx context.Context, input gql.CreateAuthTicketInput) (*models.AuthTicket, error) {
	ticket := &models.AuthTicket{
		RequesterName: input.RequesterName,
	}

	err := r.Users.CreateAuthTicket(ctx, ticket)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating authentication ticket: %s", err.Error())
	}

	return ticket, nil
}

func (r *mutationResolver) UpdateAuthTicket(ctx context.Context, input gql.UpdateAuthTicketInput) (*models.AuthTicket, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, gqlerror.Errorf("You must be logged in to approve an authentication ticket")
	}

	ticket := r.Users.FindAuthTicket(ctx, input.AuthTicketID)
	if ticket == nil {
		return nil, gqlerror.Errorf("Authentication ticket not found (it was likely denied or timed out)")
	}

	if input.Approve {
		err := r.Users.ApproveAuthTicket(ctx, ticket, secret.GetOwnerID())
		if err != nil {
			return nil, gqlerror.Errorf("Error approving authentication ticket: %s", err.Error())
		}
	} else {
		err := r.Users.DeleteAuthTicket(ctx, ticket)
		if err != nil {
			return nil, gqlerror.Errorf("Error denying authentication ticket: %s", err.Error())
		}
		ticket = nil
	}

	return ticket, nil
}
