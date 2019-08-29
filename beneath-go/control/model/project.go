package model

import (
	"log"
	"regexp"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/db"
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
	Secrets     []*Secret
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

// FindProject finds a project by ID
func FindProject(projectID uuid.UUID) *Project {
	project := &Project{
		ProjectID: projectID,
	}
	err := db.DB.Model(project).WherePK().Column("project.*", "Streams", "Users").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return project
}

// FindProjects returns a sample of projects
func FindProjects() []*Project {
	var projects []*Project
	err := db.DB.Model(&projects).Where("project.public = true").Limit(200).Select()
	if err != nil {
		log.Panic(err.Error())
	}
	return projects
}

// FindProjectByName finds a project by name
func FindProjectByName(name string) *Project {
	project := &Project{}
	err := db.DB.Model(project).
		Where("lower(project.name) = lower(?)", name).
		Column("project.*", "Streams", "Users").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return project
}

// CreateWithUser creates a project and makes user a member
func (p *Project) CreateWithUser(userID uuid.UUID) error {
	// validate
	err := GetValidator().Struct(p)
	if err != nil {
		return err
	}

	// create project and ProjectToUser in one transaction
	return db.DB.RunInTransaction(func(tx *pg.Tx) error {
		// insert project
		_, err := tx.Model(p).Insert()
		if err != nil {
			return err
		}

		// connect project to userID
		err = tx.Insert(&ProjectToUser{
			ProjectID: p.ProjectID,
			UserID:    userID,
		})
		if err != nil {
			return err
		}

		err = db.Engine.Warehouse.RegisterProject(p.ProjectID, p.Public, p.Name, p.DisplayName, p.Description)
		if err != nil {
			return err
		}

		return nil
	})
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
	// TODO only if not last user (there's a check in resolver, but it should be part of db tx)
	return db.DB.Delete(&ProjectToUser{
		ProjectID: p.ProjectID,
		UserID:    userID,
	})
}

// UpdateDetails updates projects user-facing details
func (p *Project) UpdateDetails(displayName *string, site *string, description *string, photoURL *string) error {
	// set fields
	if displayName != nil {
		p.DisplayName = *displayName
	}
	if site != nil {
		p.Site = *site
	}
	if description != nil {
		p.Description = *description
	}
	if photoURL != nil {
		p.PhotoURL = *photoURL
	}

	// validate
	err := GetValidator().Struct(p)
	if err != nil {
		return err
	}

	// TODO: Also update in BigQuery

	// update
	_, err = db.DB.Model(p).
		Column("display_name", "site", "description", "photo_url").
		WherePK().
		Update()
	return err
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
