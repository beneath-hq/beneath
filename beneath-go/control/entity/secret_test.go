package entity

import (
	"testing"

	"github.com/beneath-core/beneath-go/db"
	"github.com/stretchr/testify/assert"
)

func init() {
	db.InitPostgres("postgresql://postgres@localhost:5432/postgres?sslmode=disable")
	db.InitRedis("redis://localhost/")
}

func TestSecretIntegration(t *testing.T) {
	// create a user
	user, err := CreateOrUpdateUser("tmp", "", "test@example.org", "Test Test", "")
	assert.Nil(t, err)
	assert.NotNil(t, user)

	secret1, err := CreateUserSecret(user.UserID, SecretRoleManage, "Test secret")
	assert.Nil(t, err)
	assert.NotNil(t, secret1)
	assert.NotEmpty(t, secret1.SecretString)

	secret2 := AuthenticateSecretString(secret1.SecretString)
	assert.NotNil(t, secret2)
	assert.Equal(t, secret1.UserID, secret2.UserID)

	secret3 := AuthenticateSecretString("")
	assert.Nil(t, secret3)

	secret4 := AuthenticateSecretString("notasecret")
	assert.Nil(t, secret4)

	secret1.Revoke()
	secret2 = AuthenticateSecretString(secret1.HashedSecret)
	assert.Nil(t, secret2)

	// cleanup
	secretCache = nil
	db.Redis.FlushAll()
	assert.Nil(t, user.Delete())
}
