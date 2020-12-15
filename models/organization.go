package models

import (
	"fmt"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Organization represents the entity that manages billing on behalf of its users
type Organization struct {
	_msgpack          struct{}   `msgpack:",omitempty"`
	OrganizationID    uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name              string     `sql:",unique,notnull",validate:"required,gte=3,lte=40"` // NOTE: when updating, clear stream cache
	DisplayName       string     `sql:",notnull",validate:"required,gte=1,lte=50"`
	Description       string     `validate:"omitempty,lte=255"`
	PhotoURL          string     `validate:"omitempty,url,lte=400"`
	UserID            *uuid.UUID `sql:",on_delete:restrict,type:uuid"`
	User              *User      `msgpack:"-"`
	CreatedOn         time.Time  `sql:",default:now()"`
	UpdatedOn         time.Time  `sql:",default:now()"`
	QuotaEpoch        time.Time
	ReadQuota         *int64     // bytes // NOTE: when updating value, clear secret cache
	WriteQuota        *int64     // bytes // NOTE: when updating value, clear secret cache
	ScanQuota         *int64     // bytes // NOTE: when updating value, clear secret cache
	PrepaidReadQuota  *int64     // bytes
	PrepaidWriteQuota *int64     // bytes
	PrepaidScanQuota  *int64     // bytes
	Projects          []*Project `msgpack:"-"`
	Users             []*User    `pg:"fk:billing_organization_id",msgpack:"-"`

	// used to indicate requestor's permissions in resolvers
	Permissions *PermissionsUsersOrganizations `sql:"-",msgpack:"-"`
}

// Validate runs validation on the organization
func (o *Organization) Validate() error {
	return Validator.Struct(o)
}

// IsMulti returns true if o is a multi-user organization
func (o *Organization) IsMulti() bool {
	return o.UserID == nil
}

// IsOrganization implements gql.Organization
func (o *Organization) IsOrganization() {}

// IsBillingOrganizationForUser returns true if o is also the billing org for the user it represents.
// It panics if called on a non-personal organization.
func (o *Organization) IsBillingOrganizationForUser() bool {
	if o.UserID == nil {
		panic(fmt.Errorf("Called IsBillingOrganizationForUser on non-personal organization"))
	}
	return o.User.BillingOrganizationID == o.OrganizationID
}

// StripPrivateProjects removes private projects from o.Projects (no changes in database, just the loaded object)
func (o *Organization) StripPrivateProjects() {
	for i, p := range o.Projects {
		if !p.Public {
			n := len(o.Projects)
			o.Projects[n-1], o.Projects[i] = o.Projects[i], o.Projects[n-1]
			o.Projects = o.Projects[:n-1]
		}
	}
}

// OrganizationMember is a convenience representation of organization membership
type OrganizationMember struct {
	OrganizationID        uuid.UUID
	UserID                uuid.UUID
	BillingOrganizationID uuid.UUID
	Name                  string
	DisplayName           string
	PhotoURL              string
	View                  bool
	Create                bool
	Admin                 bool
	QuotaEpoch            time.Time
	ReadQuota             *int
	WriteQuota            *int
	ScanQuota             *int
}

// OrganizationInvite represents invites to use an organization as your billing organization.
type OrganizationInvite struct {
	OrganizationInviteID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID       uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Organization         *Organization
	UserID               uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	User                 *User
	CreatedOn            time.Time `sql:",default:now()"`
	UpdatedOn            time.Time `sql:",default:now()"`
	View                 bool      `sql:",notnull"`
	Create               bool      `sql:",notnull"`
	Admin                bool      `sql:",notnull"`
}

// Validate runs validation on the organization invite
func (i *OrganizationInvite) Validate() error {
	return Validator.Struct(i)
}

// ---------------
// Events

// OrganizationCreatedEvent is sent when an organization is created
type OrganizationCreatedEvent struct {
	Organization *Organization
}

// OrganizationUpdatedEvent is sent when an organization is updated
type OrganizationUpdatedEvent struct {
	Organization *Organization
}

// OrganizationTransferredUserEvent is sent when a user is transferred from one org to another
type OrganizationTransferredUserEvent struct {
	Source *Organization
	Target *Organization
	User   *User
}

// ---------------
// Validation

var orgNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")

var orgNameBlacklist = []string{
	"admin",
	"api",
	"auth",
	"billing",
	"console",
	"control",
	"data",
	"docs",
	"documentation",
	"ee",
	"enterprise",
	"explore",
	"gateway",
	"gql",
	"graphql",
	"health",
	"healthz",
	"instance",
	"instances",
	"internal",
	"master",
	"ops",
	"organization",
	"organizations",
	"permissions",
	"platform",
	"project",
	"projects",
	"redirects",
	"secret",
	"secrets",
	"stream",
	"streams",
	"terminal",
	"user",
	"username",
	"users",
}

func init() {
	Validator.RegisterStructValidation(func(sl validator.StructLevel) {
		s := sl.Current().Interface().(Organization)
		if !orgNameRegex.MatchString(s.Name) {
			sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
		}
		for _, blacklisted := range orgNameBlacklist {
			if s.Name == blacklisted {
				sl.ReportError(s.Name, "Name", "", "blacklisted", "")
				break
			}
		}
	}, Organization{})
}
