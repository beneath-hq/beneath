package models

import (
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Project represents a Beneath project
type Project struct {
	_msgpack       struct{}  `msgpack:",omitempty"`
	ProjectID      uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `sql:",notnull",validate:"required,gte=3,lte=16"`
	DisplayName    string    `validate:"omitempty,lte=40"`
	Site           string    `validate:"omitempty,url,lte=255"`
	Description    string    `validate:"omitempty,lte=255"`
	PhotoURL       string    `validate:"omitempty,url,lte=255"`
	Public         bool      `sql:",notnull,default:false"`
	Locked         bool      `sql:",notnull,default:false"`
	ExploreRank    int
	OrganizationID uuid.UUID     `sql:",on_delete:restrict,notnull,type:uuid"`
	Organization   *Organization `msgpack:"-"`
	CreatedOn      time.Time     `sql:",default:now()"`
	UpdatedOn      time.Time     `sql:",default:now()"`
	Tables         []*Table      `msgpack:"-"`
	Services       []*Service    `msgpack:"-"`
	Users          []*User       `pg:"many2many:permissions_users_projects,fk:project_id,joinFK:user_id",msgpack:"-"`

	// used to indicate requestor's permissions in resolvers
	Permissions *PermissionsUsersProjects `sql:"-",msgpack:"-"`
}

// Validate runs validation on the project
func (p *Project) Validate() error {
	return Validator.Struct(p)
}

// GetProjectID implements engine/driver.Project
func (p *Project) GetProjectID() uuid.UUID {
	return p.ProjectID
}

// GetOrganizationName implements engine/driver.Project
func (p *Project) GetOrganizationName() string {
	return p.Organization.Name
}

// GetProjectName implements engine/driver.Project
func (p *Project) GetProjectName() string {
	return p.Name
}

// GetPublic implements engine/driver.Project
func (p *Project) GetPublic() bool {
	return p.Public
}

// ProjectMember is a convenience representation of a user in a project
type ProjectMember struct {
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	Name        string
	DisplayName string
	PhotoURL    string
	View        bool
	Create      bool
	Admin       bool
}

// ---------------
// Events

// ProjectUpdatedEvent is sent when a project is updated
type ProjectUpdatedEvent struct {
	Project *Project
}

// ProjectDeletedEvent is setn when a project is deleted
type ProjectDeletedEvent struct {
	ProjectID uuid.UUID
}

// ---------------
// Validation

var projectNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")

func init() {
	Validator.RegisterStructValidation(func(sl validator.StructLevel) {
		p := sl.Current().Interface().(Project)

		if !projectNameRegex.MatchString(p.Name) {
			sl.ReportError(p.Name, "Name", "", "alphanumericorunderscore", "")
		}
	}, Project{})
}
