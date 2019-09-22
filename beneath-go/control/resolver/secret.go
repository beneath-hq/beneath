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
	if !secret.IsPersonal() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal secret")
	}

	if userID != *secret.UserID {
		return nil, gqlerror.Errorf("Not allowed to read user's secrets")
	}

	return entity.FindUserSecrets(ctx, userID), nil
}

func (r *queryResolver) SecretsForService(ctx context.Context, serviceID uuid.UUID) ([]*entity.Secret, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(serviceID) {
		return nil, gqlerror.Errorf("Not allowed to read service secrets")
	}

	return entity.FindServiceSecrets(ctx, serviceID), nil
}

func (r *mutationResolver) IssueUserSecret(ctx context.Context, description string) (*gql.NewSecret, error) {
	authSecret := middleware.GetSecret(ctx)
	if !authSecret.IsPersonal() {
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
	if !authSecret.EditsOrganization(serviceID) {
		return nil, gqlerror.Errorf("Not allowed to edit service")
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
	secret := entity.FindSecret(ctx, secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	authSecret := middleware.GetSecret(ctx)
	if secret.ServiceID != nil && !authSecret.EditsProject(*secret.ServiceID) {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	} else if secret.UserID != nil && *authSecret.UserID != *secret.UserID {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret.Revoke(ctx)
	return true, nil
}
