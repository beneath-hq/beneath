package model

import (
	"regexp"
	"time"

	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/control/db"
)

// Project represents a Beneath project
type Project struct {
	ProjectID   uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name        string    `sql:",unique,notnull",validate:"required,gte=3,lte=16"`
	DisplayName string    `sql:",notnull",validate:"required,gte=3,lte=40"`
	Site        string    `validate:"omitempty,url,lte=255"`
	Description string    `validate:"omitempty,lte=255"`
	PhotoURL    string    `validate:"omitempty,url,lte=255"`
	Public      bool      `sql:",notnull,default:true"`
	CreatedOn   time.Time `sql:",default:now()"`
	UpdatedOn   time.Time `sql:",default:now()"`
	Keys        []*Key
	Streams     []*Stream
	Users       []*User `pg:"many2many:projects_users,fk:project_id,joinFK:user_id"`
}

// ProjectToUser represnts the many-to-many relationship between users and projects
type ProjectToUser struct {
	tableName struct{}  `sql:"projects_users,alias:up"`
	ProjectID uuid.UUID `sql:",pk,type:uuid"`
	Project   *Project
	UserID    uuid.UUID `sql:",pk,type:uuid"`
	User      *User
}

var (
	// regex used in validation
	projectNameRegex *regexp.Regexp

	// redis cache for project data
	projectCache *cache.Codec
)

// configure constants and validator
func init() {
	projectNameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	GetValidator().RegisterStructValidation(validateProject, Project{})
}

// custom project validation
func validateProject(sl validator.StructLevel) {
	p := sl.Current().Interface().(Project)

	if !projectNameRegex.MatchString(p.Name) {
		sl.ReportError(p.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// AddUser makes user a member of project
func (p *Project) AddUser(userID uuid.UUID) error {
	return db.DB.Insert(&ProjectToUser{
		ProjectID: p.ProjectID,
		UserID:    userID,
	})
}

// RemoveUser removes a member from the project
func (p *Project) RemoveUser(userID uuid.UUID) error {
	// TODO remove from cache
	return db.DB.Delete(&ProjectToUser{
		ProjectID: p.ProjectID,
		UserID:    userID,
	})
}

func getProjectCache() *cache.Codec {
	if projectCache == nil {
		projectCache = &cache.Codec{
			Redis:     db.Redis,
			Marshal:   msgpack.Marshal,
			Unmarshal: msgpack.Unmarshal,
		}
	}
	return projectCache
}
