package entity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-redis/cache/v7"
	"github.com/vmihailenco/msgpack"

	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Secret represents an access token to read/write data or a user session token
type Secret struct {
	_msgpack     struct{}   `msgpack:",omitempty"`
	SecretID     uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Description  string     `validate:"omitempty,lte=32"`
	Prefix       string     `sql:",notnull",validate:"required,len=4"`
	HashedSecret string     `sql:",unique,notnull",validate:"required,lte=64"`
	UserID       *uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	User         *User
	ServiceID    *uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Service      *Service
	CreatedOn    time.Time `sql:",notnull,default:now()"`
	UpdatedOn    time.Time `sql:",notnull,default:now()"`
	DeletedOn    time.Time
	SecretString string `sql:"-"`
}

const (
	// number of secrets to cache in local memory for extra speed
	secretCacheLocalSize = 10000
)

var (
	// redis cache for authenticated secrets
	secretCache *cache.Codec
)

func init() {
	// configure validator
	GetValidator().RegisterStructValidation(validateSecret, Secret{})
}

// custom secret validation
func validateSecret(sl validator.StructLevel) {
	k := sl.Current().Interface().(Secret)

	if k.UserID == nil && k.ServiceID == nil {
		sl.ReportError(k.ServiceID, "ServiceID", "", "serviceid_userid_empty", "")
	}
}

func getSecretCache() *cache.Codec {
	if secretCache == nil {
		secretCache = &cache.Codec{
			Redis:     db.Redis,
			Marshal:   msgpack.Marshal,
			Unmarshal: msgpack.Unmarshal,
		}
		secretCache.UseLocalCache(secretCacheLocalSize, 1*time.Minute)
	}
	return secretCache
}

// GenerateSecretString returns a random generated secret string
func GenerateSecretString() string {
	// generate 32 random bytes
	dest := make([]byte, 32)
	if _, err := rand.Read(dest); err != nil {
		panic(err.Error())
	}

	// encode as base64 string
	encoded := base64.StdEncoding.EncodeToString(dest)

	// done
	return encoded
}

// HashSecretString safely hashes secretString
func HashSecretString(secretString string) string {
	// decode bytes from base64
	bytes, err := base64.StdEncoding.DecodeString(secretString)
	if err != nil {
		return ""
	}

	// use sha256 digest
	hashed := sha256.Sum256(bytes)

	// encode hashed bytes as base64
	encoded := base64.StdEncoding.EncodeToString(hashed[:])

	// done
	return encoded
}

// FindSecret finds a secret
func FindSecret(ctx context.Context, secretID uuid.UUID) *Secret {
	secret := &Secret{
		SecretID: secretID,
	}
	err := db.DB.ModelContext(ctx, secret).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return secret
}

// FindUserSecrets finds all the user's secrets
func FindUserSecrets(ctx context.Context, userID uuid.UUID) []*Secret {
	var secrets []*Secret
	err := db.DB.ModelContext(ctx, &secrets).Where("user_id = ?", userID).Limit(1000).Select()
	if err != nil {
		panic(err)
	}
	return secrets
}

// FindServiceSecrets finds all the service's secrets
func FindServiceSecrets(ctx context.Context, serviceID uuid.UUID) []*Secret {
	var secrets []*Secret
	err := db.DB.ModelContext(ctx, &secrets).Where("service_id = ?", serviceID).Limit(1000).Select()
	if err != nil {
		panic(err)
	}
	return secrets
}

// NewSecret creates a new, unconfigured secret -- use CreateUserSecret or CreateProjectSecret instead
func NewSecret() *Secret {
	secretStr := GenerateSecretString()
	return &Secret{
		SecretString: secretStr,
		HashedSecret: HashSecretString(secretStr),
		Prefix:       secretStr[0:4],
	}
}

// CreateUserSecret creates a new secret to manage a user
func CreateUserSecret(ctx context.Context, userID uuid.UUID, description string) (*Secret, error) {
	// create
	secret := NewSecret()
	secret.Description = description
	secret.UserID = &userID
	secret.User.ReadQuota = DefaultServiceReadQuota // how best to accept an optional input to overwrite this?
	secret.User.WriteQuota = DefaultServiceWriteQuota

	// validate
	err := GetValidator().Struct(secret)
	if err != nil {
		return nil, err
	}

	// insert
	err = db.DB.WithContext(ctx).Insert(secret)
	if err != nil {
		return nil, err
	}

	// done
	return secret, nil
}

// CreateServiceSecret creates a new secret for a service
func CreateServiceSecret(ctx context.Context, serviceID uuid.UUID, description string) (*Secret, error) {
	// create
	secret := NewSecret()
	secret.Description = description
	secret.ServiceID = &serviceID
	secret.Service.ReadQuota = DefaultServiceReadQuota // how best to accept an optional input to overwrite this?
	secret.Service.WriteQuota = DefaultServiceWriteQuota

	// validate
	err := GetValidator().Struct(secret)
	if err != nil {
		return nil, err
	}

	// insert
	err = db.DB.WithContext(ctx).Insert(secret)
	if err != nil {
		return nil, err
	}

	// done
	return secret, nil
}

// AuthenticateSecretString returns the secret object matching secretString or nil
func AuthenticateSecretString(ctx context.Context, secretString string) *Secret {
	// note: we're also caching empty secrets (i.e. where secretString doesn't match a secret)
	// to prevent database crash if someone is spamming with a bad secret

	hashed := HashSecretString(secretString)

	secret := &Secret{}
	err := getSecretCache().Once(&cache.Item{
		Key:        redisKeyForHashedSecret(hashed),
		Object:     secret,
		Expiration: 1 * time.Hour,
		Func: func() (interface{}, error) {
			selectedSecret := &Secret{}
			err := db.DB.ModelContext(ctx, selectedSecret).
				Column("secret_id", "role", "user_id", "project_id").
				Where("hashed_secret = ?", hashed).
				Select()
			if err != nil && err != pg.ErrNoRows {
				return nil, err
			}
			return selectedSecret, nil
		},
	})

	if err != nil {
		panic(err)
	}

	// see note above
	if secret.SecretID == uuid.Nil {
		return nil
	}

	// not cached in redis
	secret.HashedSecret = hashed

	return secret
}

// Revoke deletes the secret
func (k *Secret) Revoke(ctx context.Context) {
	// delete from db
	err := db.DB.WithContext(ctx).Delete(k)
	if err != nil && err != pg.ErrNoRows {
		panic(err)
	}

	// remove from redis (ignore error)
	err = getSecretCache().Delete(redisKeyForHashedSecret(k.HashedSecret))
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

func redisKeyForHashedSecret(hashedSecret string) string {
	return fmt.Sprintf("secret:%s", hashedSecret)
}

// IsAnonymous returns true iff the secret doesn't exist
func (k *Secret) IsAnonymous() bool {
	return k == nil || k.SecretID == uuid.Nil
}

// IsPersonal returns true iff the secret gives manage rights on a user
func (k *Secret) IsPersonal() bool {
	return k != nil && k.UserID != nil
}

// ReadsProject returns true iff the secret gives permission to read the project
func (k *Secret) ReadsProject(projectID uuid.UUID) bool {
	// TODO
	if k == nil {
	}
	return true
}

// EditsOrganization returns true iff the secret gives permission to edit the organization
func (k *Secret) EditsOrganization(organizationID uuid.UUID) bool {
	// TODO
	if k == nil {
	}
	return true
}

// EditsProject returns true iff the secret gives permission to edit the project
func (k *Secret) EditsProject(projectID uuid.UUID) bool {
	// TODO
	if k == nil {
	}
	return true
}

// EditsService returns true iff the secret gives permission to edit the service
func (k *Secret) EditsService(serviceID uuid.UUID) bool {
	// TODO
	if k == nil {
	}
	return true
}

// WritesStream returns true iff the secret gives permission to write to the stream
func (k *Secret) WritesStream(stream *CachedStream) bool {
	// TODO
	// role, err := RoleCache.Get(string(auth), stream.ProjectID)
	// if err != nil {
	// 	return httputil.NewHTTPError(404, err.Error())
	// }

	// if !role.Write && !(stream.Manual && role.Manage) {
	// 	return httputil.NewHTTPError(403, "token doesn't grant right to write to this stream")
	// }
	return k != nil
}

// BillingID gets the BillingID based on the BillingEntity
func (k *Secret) BillingID() uuid.UUID {
	if k.UserID != nil {
		return *k.UserID
	} else if k.ServiceID != nil {
		return *k.ServiceID
	}
	panic(fmt.Errorf("neither user id nor service id set"))
}

// CheckReadQuota checks the user's read quota
func (k *Secret) CheckReadQuota(u pb.QuotaUsage) bool {
	if k.UserID != nil {
		// if any constraints are hit, the user has hit its quota
		if u.ReadBytes >= k.User.ReadQuota {
			return false
		}

		// the user still has resources available
		return true
	}
	if k.ServiceID != nil {
		// if any constraints are hit, the user has hit its quota
		if u.ReadBytes >= k.Service.ReadQuota {
			return false
		}

		// the user still has resources available
		return true
	}

	// this should never be hit
	return false
}

// CheckWriteQuota checks the user's write quota
func (k *Secret) CheckWriteQuota(u pb.QuotaUsage) bool {
	if k.UserID != nil {
		// if any constraints are hit, the user has hit its quota
		if u.WriteBytes >= k.User.WriteQuota {
			return false
		}

		// the user still has resources available
		return true
	}
	if k.ServiceID != nil {
		// if any constraints are hit, the user has hit its quota
		if u.WriteBytes >= k.Service.WriteQuota {
			return false
		}

		// the user still has resources available
		return true
	}

	// this should never be hit
	return false
}
