package model

import (
	"testing"

	"github.com/beneath-core/beneath-go/db"
	"github.com/stretchr/testify/assert"
)

func init() {
	db.InitPostgres("postgresql://postgres@localhost:5432/postgres?sslmode=disable")
	db.InitRedis("redis://localhost/")
}

func TestKeyIntegration(t *testing.T) {
	// create a user
	user, err := CreateOrUpdateUser("tmp", "", "test@example.org", "Test Test", "")
	assert.Nil(t, err)
	assert.NotNil(t, user)

	key1, err := CreateUserKey(user.UserID, KeyRoleManage, "Test key")
	assert.Nil(t, err)
	assert.NotNil(t, key1)
	assert.NotEmpty(t, key1.KeyString)

	key2 := AuthenticateKeyString(key1.KeyString)
	assert.NotNil(t, key2)
	assert.Equal(t, key1.UserID, key2.UserID)

	key3 := AuthenticateKeyString("")
	assert.Nil(t, key3)

	key4 := AuthenticateKeyString("notakey")
	assert.Nil(t, key4)

	key1.Revoke()
	key2 = AuthenticateKeyString(key1.HashedKey)
	assert.Nil(t, key2)

	// cleanup
	keyCache = nil
	db.Redis.FlushAll()
	assert.Nil(t, user.Delete())
}
