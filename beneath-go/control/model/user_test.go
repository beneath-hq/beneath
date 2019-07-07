package model

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	user := &User{
		UserID:   uuid.NewV4(),
		Username: "a_b",
		Email:    "test@example.com",
	}

	err := GetValidator().Struct(user)
	assert.Nil(t, err)
}
