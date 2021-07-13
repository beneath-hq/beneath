package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// User represents a Beneath user
type User struct {
	_msgpack              struct{}  `msgpack:",omitempty"`
	UserID                uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Email                 string    `sql:",notnull",validate:"required,email"`
	GoogleID              string    `sql:",unique",validate:"omitempty,lte=255"`
	GithubID              string    `sql:",unique",validate:"omitempty,lte=255"`
	CreatedOn             time.Time `sql:",default:now()"`
	UpdatedOn             time.Time `sql:",default:now()"`
	Master                bool      `sql:",notnull,default:false"` // NOTE: when updating value, clear secret cache
	ConsentTerms          bool      `sql:",notnull,default:false"`
	ConsentNewsletter     bool      `sql:",notnull,default:false"`
	QuotaEpoch            time.Time
	ReadQuota             *int64          // NOTE: when updating value, clear secret cache
	WriteQuota            *int64          // NOTE: when updating value, clear secret cache
	ScanQuota             *int64          // NOTE: when updating value, clear secret cache
	BillingOrganizationID uuid.UUID       `sql:",on_delete:restrict,notnull,type:uuid"` // NOTE: when updating value, clear secret cache
	BillingOrganization   *Organization   `msgpack:"-"`
	Projects              []*Project      `pg:"many2many:permissions_users_projects,fk:user_id,joinFK:project_id",msgpack:"-"`
	Organizations         []*Organization `pg:"many2many:permissions_users_organizations,fk:user_id,joinFK:organization_id",msgpack:"-"`
	Secrets               []*UserSecret   `msgpack:"-"`
}

// Validate runs validation on the user
func (u *User) Validate() error {
	return Validator.Struct(u)
}

// AuthTicket coordinates an authentication request issued by the CLI
type AuthTicket struct {
	_msgpack       struct{}   `msgpack:",omitempty"`
	AuthTicketID   uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	RequesterName  string     `sql:",notnull" validate:"gte=3,lte=100"`
	ApproverUserID *uuid.UUID `sql:",type:uuid,on_delete:cascade"`
	CreatedOn      time.Time  `sql:",default:now()"`
	UpdatedOn      time.Time  `sql:",default:now()"`

	IssuedSecret *UserSecret `sql:"-"`
}

// Validate runs validation on the ticket
func (t *AuthTicket) Validate() error {
	return Validator.Struct(t)
}

// ---------------
// Events

// UserCreatedEvent is sent when a user is created
type UserCreatedEvent struct {
	User *User
}

// UserUpdatedEvent is sent when a user is updated
type UserUpdatedEvent struct {
	User *User
}
