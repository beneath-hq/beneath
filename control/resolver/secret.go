package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/gql"
	"github.com/beneath-core/internal/middleware"
)

// ServiceSecret returns the gql.ServiceSecretResolver
func (r *Resolver) ServiceSecret() gql.ServiceSecretResolver {
	return &serviceSecretResolver{r}
}

type serviceSecretResolver struct{ *Resolver }

func (r *serviceSecretResolver) ServiceSecretID(ctx context.Context, obj *entity.ServiceSecret) (string, error) {
	return obj.ServiceSecretID.String(), nil
}

// UserSecret returns the gql.UserSecretResolver
func (r *Resolver) UserSecret() gql.UserSecretResolver {
	return &userSecretResolver{r}
}

type userSecretResolver struct{ *Resolver }

func (r *userSecretResolver) UserSecretID(ctx context.Context, obj *entity.UserSecret) (string, error) {
	return obj.UserSecretID.String(), nil
}

func (r *queryResolver) SecretsForUser(ctx context.Context, userID uuid.UUID) ([]*entity.UserSecret, error) {
	secret := middleware.GetSecret(ctx)
	if secret.GetOwnerID() != userID {
		return nil, MakeUnauthenticatedError("Must be authenticated as userID")
	}

	return entity.FindUserSecrets(ctx, userID), nil
}

func (r *queryResolver) SecretsForService(ctx context.Context, serviceID uuid.UUID) ([]*entity.ServiceSecret, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service not found")
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read service secrets")
	}

	return entity.FindServiceSecrets(ctx, serviceID), nil
}

func (r *mutationResolver) IssueUserSecret(ctx context.Context, description string, readOnly bool, publicOnly bool) (*gql.NewUserSecret, error) {
	authSecret := middleware.GetSecret(ctx)
	if !authSecret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal secret")
	}

	secret, err := entity.CreateUserSecret(ctx, authSecret.GetOwnerID(), description, publicOnly, readOnly)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewUserSecret{
		Secret: secret,
		Token:  secret.Token.String(),
	}, nil
}

func (r *mutationResolver) IssueServiceSecret(ctx context.Context, serviceID uuid.UUID, description string) (*gql.NewServiceSecret, error) {
	service := entity.FindService(ctx, serviceID)
	if service == nil {
		return nil, gqlerror.Errorf("Service not found")
	}

	if service.Kind != entity.ServiceKindExternal {
		return nil, gqlerror.Errorf("Can only create secrets for external services")
	}

	authSecret := middleware.GetSecret(ctx)
	perms := authSecret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in the organization")
	}

	secret, err := entity.CreateServiceSecret(ctx, serviceID, description)
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

	secret := entity.FindServiceSecret(ctx, secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	service := entity.FindService(ctx, secret.ServiceID)
	if service == nil {
		return false, gqlerror.Errorf("Service not found")
	}

	perms := authSecret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	if service.Kind != entity.ServiceKindExternal {
		return false, gqlerror.Errorf("Can only revoke secrets for external services")
	}

	secret.Revoke(ctx)
	return true, nil
}

func (r *mutationResolver) RevokeUserSecret(ctx context.Context, secretID uuid.UUID) (bool, error) {
	authSecret := middleware.GetSecret(ctx)
	if authSecret.IsAnonymous() {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret := entity.FindUserSecret(ctx, secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	if authSecret.GetOwnerID() != secret.UserID {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret.Revoke(ctx)
	return true, nil
}
