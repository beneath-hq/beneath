package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
)

// Secret returns the gql.SecretResolver
func (r *Resolver) Secret() gql.SecretResolver {
	return &secretResolver{r}
}

type secretResolver struct{ *Resolver }

func (r *secretResolver) SecretID(ctx context.Context, obj *entity.Secret) (string, error) {
	return obj.SecretID.String(), nil
}

func (r *queryResolver) SecretsForUser(ctx context.Context, userID uuid.UUID) ([]*entity.Secret, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUserID(userID) {
		return nil, MakeUnauthenticatedError("Must be authenticated as userID")
	}

	return entity.FindUserSecrets(ctx, userID), nil
}

func (r *queryResolver) SecretsForService(ctx context.Context, serviceID uuid.UUID) ([]*entity.Secret, error) {
	secret := middleware.GetSecret(ctx)
	service := entity.FindService(ctx, serviceID)
	perms := secret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read service secrets")
	}

	return entity.FindServiceSecrets(ctx, serviceID), nil
}

func (r *mutationResolver) IssueUserSecret(ctx context.Context, description string) (*gql.NewSecret, error) {
	authSecret := middleware.GetSecret(ctx)
	if !authSecret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal secret")
	}

	secret, err := entity.CreateUserSecret(ctx, *authSecret.UserID, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewSecret{
		Secret:       secret,
		SecretString: secret.SecretString,
	}, nil
}

func (r *mutationResolver) IssueServiceSecret(ctx context.Context, serviceID uuid.UUID, description string) (*gql.NewSecret, error) {
	authSecret := middleware.GetSecret(ctx)
	service := entity.FindService(ctx, serviceID)
	perms := authSecret.OrganizationPermissions(ctx, service.OrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in the organization")
	}

	secret, err := entity.CreateServiceSecret(ctx, serviceID, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewSecret{
		Secret:       secret,
		SecretString: secret.SecretString,
	}, nil
}

func (r *mutationResolver) RevokeSecret(ctx context.Context, secretID uuid.UUID) (bool, error) {
	authSecret := middleware.GetSecret(ctx)
	if authSecret == nil {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret := entity.FindSecret(ctx, secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	if secret.UserID != nil {
		// we're revoking a personal user secret
		if authSecret.UserID == nil || *authSecret.UserID != *secret.UserID {
			return false, gqlerror.Errorf("Not allowed to edit secret")
		}
	} else if secret.ServiceID != nil {
		// we're revoking a service secret
		service := entity.FindService(ctx, *secret.ServiceID)
		perms := authSecret.OrganizationPermissions(ctx, service.OrganizationID)
		if !perms.Admin {
			return false, gqlerror.Errorf("Not allowed to edit secret")
		}
	}

	secret.Revoke(ctx)
	return true, nil
}
