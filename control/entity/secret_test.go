package entity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/pkg/secrettoken"
)

func init() {
	hub.InitPostgres("localhost", "", "postgres", "")
	hub.InitRedis("redis://localhost/")
}

func TestSecretIntegration(t *testing.T) {
	ctx := context.Background()

	// create a user
	user, err := CreateOrUpdateUser(ctx, "tmp", "", "test@example.org", "test", "Test Test", "")
	assert.Nil(t, err)
	assert.NotNil(t, user)

	secret1, err := CreateUserSecret(ctx, user.UserID, "Test secret", false, false)
	assert.Nil(t, err)
	assert.NotNil(t, secret1)
	assert.NotEqual(t, secret1.Token, secrettoken.Nil)

	secret2 := AuthenticateWithToken(ctx, secret1.Token)
	assert.NotNil(t, secret2)
	assert.Equal(t, secret1.UserID, secret2.GetOwnerID())

	secret3 := AuthenticateWithToken(ctx, secrettoken.FromStringOrNil(""))
	assert.Nil(t, secret3)

	secret4 := AuthenticateWithToken(ctx, secrettoken.FromStringOrNil("GBJApvATiuhoxTXksicC6ePzhVu9VDy7hWnWLvpzayhY"))
	assert.Nil(t, secret4)

	secret1.Revoke(ctx)
	secret2 = AuthenticateWithToken(ctx, secret1.Token)
	assert.Nil(t, secret2)

	// cleanup
	secretCache = nil
	hub.Redis.FlushAll()
}
