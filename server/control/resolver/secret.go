package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// ServiceSecret returns the gql.ServiceSecretResolver
func (r *Resolver) ServiceSecret() gql.ServiceSecretResolver {
	return &serviceSecretResolver{r}
}

type serviceSecretResolver struct{ *Resolver }

func (r *serviceSecretResolver) ServiceSecretID(ctx context.Context, obj *models.ServiceSecret) (string, error) {
	return obj.ServiceSecretID.String(), nil
}

// UserSecret returns the gql.UserSecretResolver
func (r *Resolver) UserSecret() gql.UserSecretResolver {
	return &userSecretResolver{r}
}

type userSecretResolver struct{ *Resolver }

func (r *userSecretResolver) UserSecretID(ctx context.Context, obj *models.UserSecret) (string, error) {
	return obj.UserSecretID.String(), nil
}

func (r *queryResolver) SecretsForUser(ctx context.Context, userID uuid.UUID) ([]*models.UserSecret, error) {
	secret := middleware.GetSecret(ctx)
	if secret.GetOwnerID() != userID {
		return nil, MakeUnauthenticatedError("Must be authenticated as userID")
	}

	return r.Secrets.FindUserSecrets(ctx, userID), nil
}

func (r *queryResolver) SecretsForService(ctx context.Context, serviceID uuid.UUID) ([]*models.ServiceSecret, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service not found")
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, service.ProjectID, false)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read service secrets")
	}

	return r.Secrets.FindServiceSecrets(ctx, serviceID), nil
}

func (r *mutationResolver) IssueUserSecret(ctx context.Context, description string, readOnly bool, publicOnly bool) (*gql.NewUserSecret, error) {
	authSecret := middleware.GetSecret(ctx)
	if !authSecret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal secret")
	}

	secret, err := r.Secrets.CreateUserSecret(ctx, authSecret.GetOwnerID(), description, publicOnly, readOnly)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewUserSecret{
		Secret: secret,
		Token:  secret.Token.String(),
	}, nil
}

func (r *mutationResolver) IssueServiceSecret(ctx context.Context, serviceID uuid.UUID, description string) (*gql.NewServiceSecret, error) {
	service := r.Services.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service not found")
	}

	authSecret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, authSecret, service.ProjectID, false)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to perform create functions in the organization")
	}

	secret, err := r.Secrets.CreateServiceSecret(ctx, serviceID, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewServiceSecret{
		Secret: secret,
		Token:  secret.Token.String(),
	}, nil
}

func (r *mutationResolver) RevokeServiceSecret(ctx context.Context, secretID uuid.UUID) (bool, error) {
	authSecret := middleware.GetSecret(ctx)
	if authSecret.IsAnonymous() {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret := r.Secrets.FindServiceSecret(ctx, secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	service := r.Services.FindService(ctx, secret.ServiceID)
	if service == nil {
		return false, gqlerror.Errorf("Service not found")
	}

	perms := r.Permissions.ProjectPermissionsForSecret(ctx, authSecret, service.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to edit service")
	}

	r.Secrets.RevokeServiceSecret(ctx, secret)
	return true, nil
}

func (r *mutationResolver) RevokeUserSecret(ctx context.Context, secretID uuid.UUID) (bool, error) {
	authSecret := middleware.GetSecret(ctx)
	if authSecret.IsAnonymous() {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret := r.Secrets.FindUserSecret(ctx, secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	if authSecret.GetOwnerID() != secret.UserID {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	r.Secrets.RevokeUserSecret(ctx, secret)
	return true, nil
}
