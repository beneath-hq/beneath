package entity

import (
	"context"
	"regexp"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/go-redis/cache/v7"
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
	DeletedOn   time.Time
	Secrets     []*Secret
	Streams     []*Stream
	Models      []*Model
	Users       []*User `pg:"many2many:projects_users,fk:project_id,joinFK:user_id"`
}

// ProjectToUser represnts the many-to-many relationship between users and projects
type ProjectToUser struct {
	tableName struct{}  `sql:"projects_users,alias:pu"`
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

func init() {
	orm.RegisterTable((*ProjectToUser)(nil))
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
func FindProject(ctx context.Context, projectID uuid.UUID) *Project {
	project := &Project{
		ProjectID: projectID,
	}
	err := db.DB.ModelContext(ctx, project).WherePK().Column("project.*", "Streams", "Users").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return project
}

// FindProjects returns a sample of projects
func FindProjects(ctx context.Context) []*Project {
	var projects []*Project
	err := db.DB.ModelContext(ctx, &projects).Where("project.public = true").Limit(200).Select()
	if err != nil {
		panic(err)
	}
	return projects
}

// FindProjectByName finds a project by name
func FindProjectByName(ctx context.Context, name string) *Project {
	project := &Project{}
	err := db.DB.ModelContext(ctx, project).
		Where("lower(project.name) = lower(?)", name).
		Column("project.*", "Streams", "Users").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return project
}

// CreateWithUser creates a project and makes user a member
func (p *Project) CreateWithUser(ctx context.Context, userID uuid.UUID) error {
	// validate
	err := GetValidator().Struct(p)
	if err != nil {
		return err
	}

	// create project and ProjectToUser in one transaction
	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
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

		err = db.Engine.Warehouse.RegisterProject(ctx, p.ProjectID, p.Public, p.Name, p.DisplayName, p.Description)
		if err != nil {
			return err
		}

		return nil
	})
}

// AddUser makes user a member of project
func (p *Project) AddUser(ctx context.Context, userID uuid.UUID) error {
	return db.DB.WithContext(ctx).Insert(&ProjectToUser{
		ProjectID: p.ProjectID,
		UserID:    userID,
	})
}

// RemoveUser removes a member from the project
func (p *Project) RemoveUser(ctx context.Context, userID uuid.UUID) error {
	// TODO remove from cache
	// TODO only if not last user (there's a check in resolver, but it should be part of db tx)
	return db.DB.WithContext(ctx).Delete(&ProjectToUser{
		ProjectID: p.ProjectID,
		UserID:    userID,
	})
}

// UpdateDetails updates projects user-facing details
func (p *Project) UpdateDetails(ctx context.Context, displayName *string, site *string, description *string, photoURL *string) error {
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

	// update in tx with call to bigquery
	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		_, err = db.DB.WithContext(ctx).Model(p).
			Column("display_name", "site", "description", "photo_url").
			WherePK().
			Update()
		if err != nil {
			return err
		}

		// update in warehouse
		err = db.Engine.Warehouse.UpdateProject(ctx, p.ProjectID, p.Public, p.Name, p.DisplayName, p.Description)
		if err != nil {
			return err
		}

		return nil
	})
}

// Delete safely deletes the project (fails if the project still has content)
func (p *Project) Delete(ctx context.Context) error {
	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		err := db.DB.WithContext(ctx).Delete(p)
		if err != nil {
			return err
		}

		err = db.Engine.Warehouse.DeregisterProject(ctx, p.ProjectID, p.Name)
		if err != nil {
			return err
		}

		return nil
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
