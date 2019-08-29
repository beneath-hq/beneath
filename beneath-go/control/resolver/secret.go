package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Secret returns the gql.SecretResolver
func (r *Resolver) Secret() gql.SecretResolver {
	return &secretResolver{r}
}

type secretResolver struct{ *Resolver }

func (r *secretResolver) SecretID(ctx context.Context, obj *model.Secret) (string, error) {
	return obj.SecretID.String(), nil
}

func (r *secretResolver) Role(ctx context.Context, obj *model.Secret) (string, error) {
	return string(obj.Role), nil
}

func (r *queryResolver) SecretsForUser(ctx context.Context, userID uuid.UUID) ([]*model.Secret, error) {
	secret := auth.GetSecret(ctx)
	if !secret.IsPersonal() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal secret")
	}

	if userID != *secret.UserID {
		return nil, gqlerror.Errorf("Not allowed to read user's secrets")
	}

	return model.FindUserSecrets(userID), nil
}

func (r *queryResolver) SecretsForProject(ctx context.Context, projectID uuid.UUID) ([]*model.Secret, error) {
	secret := auth.GetSecret(ctx)
	if !secret.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to read project secrets")
	}

	return model.FindProjectSecrets(projectID), nil
}

func (r *mutationResolver) IssueUserSecret(ctx context.Context, readonly bool, description string) (*gql.NewSecret, error) {
	authSecret := auth.GetSecret(ctx)
	if !authSecret.IsPersonal() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal secret")
	}

	role := model.SecretRoleManage
	if readonly {
		role = model.SecretRoleReadonly
	}

	secret, err := model.CreateUserSecret(*authSecret.UserID, role, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewSecret{
		Secret:       secret,
		SecretString: secret.SecretString,
	}, nil
}

func (r *mutationResolver) IssueProjectSecret(ctx context.Context, projectID uuid.UUID, readonly bool, description string) (*gql.NewSecret, error) {
	authSecret := auth.GetSecret(ctx)
	if !authSecret.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project")
	}

	role := model.SecretRoleReadWrite
	if readonly {
		role = model.SecretRoleReadonly
	}

	secret, err := model.CreateProjectSecret(projectID, role, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewSecret{
		Secret:       secret,
		SecretString: secret.SecretString,
	}, nil
}

func (r *mutationResolver) RevokeSecret(ctx context.Context, secretID uuid.UUID) (bool, error) {
	secret := model.FindSecret(secretID)
	if secret == nil {
		return false, gqlerror.Errorf("Secret not found")
	}

	authSecret := auth.GetSecret(ctx)
	if secret.ProjectID != nil && !authSecret.EditsProject(*secret.ProjectID) {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	} else if secret.UserID != nil && *authSecret.UserID != *secret.UserID {
		return false, gqlerror.Errorf("Not allowed to edit secret")
	}

	secret.Revoke()
	return true, nil
}
