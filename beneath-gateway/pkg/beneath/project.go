package beneath

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Project models the projects table
type Project struct {
	ProjectID   uuid.UUID `sql:",pk,type:uuid"`
	Name        string
	DisplayName string
	Site        string
	Description string
	PhotoURL    string
	CreatedOn   time.Time
	UpdatedOn   time.Time
}
