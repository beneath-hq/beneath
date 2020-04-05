package entity

import (
	"context"

	"gitlab.com/beneath-org/beneath/pkg/secrettoken"

	"gitlab.com/beneath-org/beneath/internal/hub"
	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
)

// ServiceSecret implements Secret for Token entities
type ServiceSecret struct {
	_msgpack        struct{}  `msgpack:",omitempty"`
	ServiceSecretID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	BaseSecret
	ServiceID uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Service   *Service
}

// CreateServiceSecret creates a new secret to manage a service
func CreateServiceSecret(ctx context.Context, serviceID uuid.UUID, description string) (*ServiceSecret, error) {
	// create
	secret := &ServiceSecret{}
	secret.Token = secrettoken.New(TokenFlagsService)
	secret.Prefix = secret.Token.Prefix()
	secret.HashedToken = secret.Token.Hashed()
	secret.Description = description
	secret.ServiceID = serviceID

	// validate
	err := GetValidator().Struct(secret)
	if err != nil {
		return nil, err
	}

	// insert
	err = hub.DB.WithContext(ctx).Insert(secret)
	if err != nil {
		return nil, err
	}

	// done
	return secret, nil
}

// FindServiceSecret finds a secret
func FindServiceSecret(ctx context.Context, secretID uuid.UUID) *ServiceSecret {
	secret := &ServiceSecret{
		ServiceSecretID: secretID,
	}
	err := hub.DB.ModelContext(ctx, secret).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return secret
}

// FindServiceSecrets finds all the service's secrets
func FindServiceSecrets(ctx context.Context, serviceID uuid.UUID) []*ServiceSecret {
	var secrets []*ServiceSecret
	err := hub.DB.ModelContext(ctx, &secrets).Where("service_id = ?", serviceID).Limit(1000).Select()
	if err != nil {
		panic(err)
	}
	return secrets
}

// GetSecretID implements the Secret interface
func (s *ServiceSecret) GetSecretID() uuid.UUID {
	return s.ServiceSecretID
}

// GetOwnerID implements the Secret interface
func (s *ServiceSecret) GetOwnerID() uuid.UUID {
	return s.ServiceID
}

// IsAnonymous implements the Secret interface
func (s *ServiceSecret) IsAnonymous() bool {
	return false
}

// IsUser implements the Secret interface
func (s *ServiceSecret) IsUser() bool {
	return false
}

// IsService implements the Secret interface
func (s *ServiceSecret) IsService() bool {
	return true
}

// StreamPermissions implements the Secret interface
func (s *ServiceSecret) StreamPermissions(ctx context.Context, streamID uuid.UUID, projectID uuid.UUID, public bool, external bool) StreamPermissions {
	return CachedServiceStreamPermissions(ctx, s.ServiceID, streamID)
}

// ProjectPermissions implements the Secret interface
func (s *ServiceSecret) ProjectPermissions(ctx context.Context, projectID uuid.UUID, public bool) ProjectPermissions {
	return ProjectPermissions{}
}

// OrganizationPermissions implements the Secret interface
func (s *ServiceSecret) OrganizationPermissions(ctx context.Context, organizationID uuid.UUID) OrganizationPermissions {
	return OrganizationPermissions{}
}

// ManagesModelBatches implements the Secret interface
func (s *ServiceSecret) ManagesModelBatches(model *Model) bool {
	return s.ServiceID == model.ServiceID
}

// Revoke implements the Secret interface
func (s *ServiceSecret) Revoke(ctx context.Context) {
	// delete from db
	err := hub.DB.WithContext(ctx).Delete(s)
	if err != nil && err != pg.ErrNoRows {
		panic(err)
	}

	// remove from redis (ignore error)
	getSecretCache().Delete(ctx, s.HashedToken)
}
