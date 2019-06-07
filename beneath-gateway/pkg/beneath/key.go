package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Key models the keys table
type Key struct {
	KeyID       uuid.UUID `sql:",pk,type:uuid"`
	Description string
	Prefix      string
	HashedKey   string
	Role        string
	CreatedOn   time.Time
	UpdatedOn   time.Time
	UserID      uuid.UUID `sql:",type:uuid"`
	ProjectID   uuid.UUID `sql:",type:uuid"`
}

/*
	EXAMPLE OF HOW TO QUERY:

	var res []struct {
		Role string
	}
	err := beneath.DB.Model((*models.Key)(nil)).
		Column("key.role").
		Join("JOIN projects AS p").
		JoinOn("p.project_id = key.project_id").
		Where("p.name = ?", projectName).
		Where("key.hashed_key = ?", hashedKey).
		Limit(1).
		Select(&res)
*/
