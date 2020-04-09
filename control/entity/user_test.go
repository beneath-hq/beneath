package entity

import (
	"testing"

	"gitlab.com/beneath-hq/beneath/internal/hub"
)

func init() {
	hub.InitPostgres("localhost", "", "postgres", "")
	hub.InitRedis("redis://localhost/")
}

func TestUserCreate(t *testing.T) {
	// ctx := context.Background()

	// user10, err := CreateOrUpdateUser(ctx, "10", "", "benjamin@example10.org", "bem8", "1", "")
	// assert.Nil(t, err)
	// t.Logf("%v", user10)

	// user20, err := CreateOrUpdateUser(ctx, "20", "", "benjamin@example20.org", "", "Egelund", "")
	// assert.Nil(t, err)
	// t.Logf("%v", user20)

	// assert.Nil(t, user10.Delete(ctx))
	// assert.Nil(t, user20.Delete(ctx))

	// // cleanup
	// hub.Redis.FlushAll()
}
