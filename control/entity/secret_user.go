package entity

import (
	"context"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/secrettoken"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
)

// UserSecret implements Secret for User entities
type UserSecret struct {
	_msgpack     struct{}  `msgpack:",omitempty"`
	UserSecretID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	BaseSecret
	UserID     uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	User       *User
	ReadOnly   bool `sql:",notnull"`
	PublicOnly bool `sql:",notnull"`
}

// CreateUserSecret creates a new secret to manage a user
func CreateUserSecret(ctx context.Context, userID uuid.UUID, description string, publicOnly, readOnly bool) (*UserSecret, error) {
	// create
	secret := &UserSecret{}
	secret.Token = secrettoken.New(TokenFlagsUser)
	secret.Prefix = secret.Token.Prefix()
	secret.HashedToken = secret.Token.Hashed()
	secret.Description = description
	secret.UserID = userID
	secret.ReadOnly = readOnly
	secret.PublicOnly = publicOnly

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

// FindUserSecret finds a secret
func FindUserSecret(ctx context.Context, secretID uuid.UUID) *UserSecret {
	secret := &UserSecret{
		UserSecretID: secretID,
	}
	err := hub.DB.ModelContext(ctx, secret).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return secret
}

// FindUserSecrets finds all the user's secrets
func FindUserSecrets(ctx context.Context, userID uuid.UUID) []*UserSecret {
	var secrets []*UserSecret
	err := hub.DB.ModelContext(ctx, &secrets).Where("user_id = ?", userID).Limit(1000).Select()
	if err != nil {
		panic(err)
	}
	return secrets
}

// GetSecretID implements the Secret interface
func (s *UserSecret) GetSecretID() uuid.UUID {
	return s.UserSecretID
}

// GetOwnerID implements the Secret interface
func (s *UserSecret) GetOwnerID() uuid.UUID {
	return s.UserID
}

// IsAnonymous implements the Secret interface
func (s *UserSecret) IsAnonymous() bool {
	return false
}

// IsUser implements the Secret interface
func (s *UserSecret) IsUser() bool {
	return true
}

// IsService implements the Secret interface
func (s *UserSecret) IsService() bool {
	return false
}

// StreamPermissions implements the Secret interface
func (s *UserSecret) StreamPermissions(ctx context.Context, streamID uuid.UUID, projectID uuid.UUID, public bool, external bool) StreamPermissions {
	projectPerms := CachedUserProjectPermissions(ctx, s.UserID, projectID)
	return StreamPermissions{
		Read:  (projectPerms.View && !s.PublicOnly) || public,
		Write: projectPerms.Create && external && !s.ReadOnly && (!s.PublicOnly || public),
	}
}

// ProjectPermissions implements the Secret interface
func (s *UserSecret) ProjectPermissions(ctx context.Context, projectID uuid.UUID, public bool) ProjectPermissions {
	if s.PublicOnly && !public {
		return ProjectPermissions{}
	}
	if s.ReadOnly && public {
		return ProjectPermissions{View: true}
	}
	perms := CachedUserProjectPermissions(ctx, s.UserID, projectID)
	if public {
		perms.View = true
	}
	if s.ReadOnly {
		perms.Admin = false
		perms.Create = false
	}
	return perms
}

// OrganizationPermissions implements the Secret interface
func (s *UserSecret) OrganizationPermissions(ctx context.Context, organizationID uuid.UUID) OrganizationPermissions {
	if s.ReadOnly || s.PublicOnly {
		return OrganizationPermissions{}
	}
	return CachedUserOrganizationPermissions(ctx, s.UserID, organizationID)
}

// ManagesModelBatches implements the Secret interface
func (s *UserSecret) ManagesModelBatches(model *Model) bool {
	return false
}

// Revoke implements the Secret interface
func (s *UserSecret) Revoke(ctx context.Context) {
	// delete from db
	err := hub.DB.WithContext(ctx).Delete(s)
	if err != nil && err != pg.ErrNoRows {
		panic(err)
	}

	// remove from redis (ignore error)
	getSecretCache().Clear(ctx, s.HashedToken)
}
