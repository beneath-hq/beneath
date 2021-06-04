package secret

import (
	"context"

	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/secrettoken"
)

// CreateUserSecret creates a new secret to manage a user
func (s *Service) CreateUserSecret(ctx context.Context, userID uuid.UUID, description string, publicOnly, readOnly bool) (*models.UserSecret, error) {
	// create
	secret := &models.UserSecret{}
	secret.Token = secrettoken.New(TokenFlagsUser)
	secret.Prefix = secret.Token.Prefix()
	secret.HashedToken = secret.Token.Hashed()
	secret.Description = description
	secret.UserID = userID
	secret.ReadOnly = readOnly
	secret.PublicOnly = publicOnly

	// validate
	err := secret.Validate()
	if err != nil {
		return nil, err
	}

	// insert
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, secret).Insert()
	if err != nil {
		return nil, err
	}

	// done
	return secret, nil
}

// CreateServiceSecret creates a new secret to manage a service
func (s *Service) CreateServiceSecret(ctx context.Context, serviceID uuid.UUID, description string) (*models.ServiceSecret, error) {
	// create
	secret := &models.ServiceSecret{}
	secret.Token = secrettoken.New(TokenFlagsService)
	secret.Prefix = secret.Token.Prefix()
	secret.HashedToken = secret.Token.Hashed()
	secret.Description = description
	secret.ServiceID = serviceID

	// validate
	err := secret.Validate()
	if err != nil {
		return nil, err
	}

	// insert
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, secret).Insert()
	if err != nil {
		return nil, err
	}

	// done
	return secret, nil
}

// RevokeUserSecret revokes a user secret
func (s *Service) RevokeUserSecret(ctx context.Context, secret *models.UserSecret) {
	// delete from db
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, secret).WherePK().Delete()
	if err != nil && err != pg.ErrNoRows {
		panic(err)
	}

	// remove from redis (ignore error)
	s.cache.Clear(ctx, secret.HashedToken)
}

// RevokeServiceSecret revokes a service secret
func (s *Service) RevokeServiceSecret(ctx context.Context, secret *models.ServiceSecret) {
	// delete from db
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, secret).WherePK().Delete()
	if err != nil && err != pg.ErrNoRows {
		panic(err)
	}

	// remove from redis (ignore error)
	s.cache.Clear(ctx, secret.HashedToken)
}
