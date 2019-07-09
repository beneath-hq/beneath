package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/cache"
	"github.com/vmihailenco/msgpack"

	"github.com/beneath-core/beneath-go/control/db"
	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Key represents an access token to read/write data or a user session token
type Key struct {
	_msgpack    struct{}   `msgpack:",omitempty"`
	KeyID       uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Description string     `validate:"omitempty,lte=32"`
	Prefix      string     `sql:",notnull",validate:"required,len=4"`
	HashedKey   string     `sql:",unique,notnull",validate:"required,lte=64"`
	Role        KeyRole    `sql:",notnull",validate:"required,lte=3"`
	UserID      *uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	User        *User
	ProjectID   *uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Project     *Project
	CreatedOn   *time.Time `sql:",default:now()"`
	UpdatedOn   *time.Time `sql:",default:now()"`
	KeyString   string     `sql:"-"`
}

// KeyRole represents a role in a Key
type KeyRole string

const (
	// KeyRoleReadonly can only read data
	KeyRoleReadonly KeyRole = "r"

	// KeyRoleReadWrite can read and write data
	KeyRoleReadWrite KeyRole = "rw"

	// KeyRoleManage can edit a user
	KeyRoleManage KeyRole = "m"

	// number of keys to cache in local memory for extra speed
	keyCacheLocalSize = 10000
)

var (
	// redis cache for authenticated keys
	keyCache *cache.Codec
)

func init() {
	// configure validator
	GetValidator().RegisterStructValidation(validateKey, Key{})
}

// custom key validation
func validateKey(sl validator.StructLevel) {
	k := sl.Current().Interface().(Key)

	if k.UserID == nil && k.ProjectID == nil {
		sl.ReportError(k.ProjectID, "ProjectID", "", "projectid_userid_empty", "")
	}

	if k.Role != "r" && k.Role != "rw" && k.Role != "m" {
		sl.ReportError(k.Role, "Role", "", "bad_role", "")
	}
}

func getKeyCache() *cache.Codec {
	if keyCache == nil {
		keyCache = &cache.Codec{
			Redis:     db.Redis,
			Marshal:   msgpack.Marshal,
			Unmarshal: msgpack.Unmarshal,
		}
		keyCache.UseLocalCache(keyCacheLocalSize, 1*time.Minute)
	}
	return keyCache
}

// GenerateKeyString returns a random generated key string
func GenerateKeyString() string {
	// generate 32 random bytes
	dest := make([]byte, 32)
	if _, err := rand.Read(dest); err != nil {
		log.Fatalf("rand.Read: %v", err.Error())
	}

	// encode as base64 string
	encoded := base64.StdEncoding.EncodeToString(dest)

	// done
	return encoded
}

// HashKeyString safely hashes keyString
func HashKeyString(keyString string) string {
	// decode bytes from base64
	bytes, err := base64.StdEncoding.DecodeString(keyString)
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

// NewKey creates a new, unconfigured key -- use NewUserKey or NewProjectKey instead
func NewKey() *Key {
	keyStr := GenerateKeyString()
	return &Key{
		KeyString: keyStr,
		HashedKey: HashKeyString(keyStr),
		Prefix:    keyStr[0:4],
	}
}

// CreateUserKey creates a new key to manage a user
func CreateUserKey(userID uuid.UUID, role KeyRole, description string) (*Key, error) {
	// create
	key := NewKey()
	key.Description = description
	key.Role = role
	key.UserID = &userID

	// validate
	err := GetValidator().Struct(key)
	if err != nil {
		return nil, err
	}

	// insert
	err = db.DB.Insert(key)
	if err != nil {
		return nil, err
	}

	// done
	return key, nil
}

// CreateProjectKey creates a new read or readwrite key for a project
func CreateProjectKey(projectID uuid.UUID, role KeyRole, description string) (*Key, error) {
	// create
	key := NewKey()
	key.Description = description
	key.Role = role
	key.ProjectID = &projectID

	// validate
	err := GetValidator().Struct(key)
	if err != nil {
		return nil, err
	}

	// insert
	err = db.DB.Insert(key)
	if err != nil {
		return nil, err
	}

	// done
	return key, nil
}

// AuthenticateKeyString returns the key object matching keyString or nil
func AuthenticateKeyString(keyString string) *Key {
	// note: we're also caching empty keys (i.e. where keyString doesn't match a key)
	// to prevent database crash if someone is spamming with a bad key

	hashed := HashKeyString(keyString)

	key := &Key{}
	err := getKeyCache().Once(&cache.Item{
		Key:        redisKeyForHashedKey(hashed),
		Object:     key,
		Expiration: 1 * time.Hour,
		Func: func() (interface{}, error) {
			selectedKey := &Key{}
			err := db.DB.Model(selectedKey).
				Column("key_id", "role", "user_id", "project_id").
				Where("hashed_key = ?", hashed).
				Select()
			if err != nil && err != pg.ErrNoRows {
				return nil, err
			}
			return selectedKey, nil
		},
	})

	if err != nil {
		log.Panic(err.Error())
		return nil
	}

	// see note above
	if key.KeyID == uuid.Nil {
		return nil
	}

	// not cached in redis
	key.HashedKey = hashed

	return key
}

// Revoke deletes the key
func (k *Key) Revoke() {
	// delete from db
	err := db.DB.Delete(k)
	if err != nil && err != pg.ErrNoRows {
		log.Panic(err.Error())
	}

	// remove from redis (ignore error)
	err = getKeyCache().Delete(redisKeyForHashedKey(k.HashedKey))
	if err != nil {
		log.Panic(err.Error())
	}
}

func redisKeyForHashedKey(hashedKey string) string {
	return fmt.Sprintf("key:%s", hashedKey)
}

// IsPersonal returns true iff the key gives manage rights on a user
func (k *Key) IsPersonal() bool {
	return k != nil && k.UserID != nil && k.Role == KeyRoleManage
}
