package entity

import (
	"context"
	"regexp"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/internal/hub"
)

// Project represents a Beneath project
type Project struct {
	ProjectID      uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `sql:",notnull",validate:"required,gte=3,lte=16"`
	DisplayName    string    `validate:"omitempty,lte=40"`
	Site           string    `validate:"omitempty,url,lte=255"`
	Description    string    `validate:"omitempty,lte=255"`
	PhotoURL       string    `validate:"omitempty,url,lte=255"`
	Public         bool      `sql:",notnull,default:true"`
	Locked         bool      `sql:",notnull,default:false"`
	OrganizationID uuid.UUID `sql:",on_delete:restrict,notnull,type:uuid"`
	Organization   *Organization
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	Streams        []*Stream
	Models         []*Model
	Users          []*User `pg:"many2many:permissions_users_projects,fk:project_id,joinFK:user_id"`

	// used to indicate requestor's permissions in resolvers
	Permissions *PermissionsUsersProjects `sql:"-"`
}

var (
	// regex used in validation
	projectNameRegex *regexp.Regexp
)

func init() {
	projectNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
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
	err := hub.DB.ModelContext(ctx, project).WherePK().Column("project.*", "Streams", "Users", "Organization").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return project
}

// FindProjects returns a sample of projects
func FindProjects(ctx context.Context) []*Project {
	var projects []*Project
	err := hub.DB.ModelContext(ctx, &projects).Where("project.public = true").Limit(200).Order("name").Relation("Organization").Select()
	if err != nil {
		panic(err)
	}
	return projects
}

// FindProjectByOrganizationAndName finds a project by organization name and project name
func FindProjectByOrganizationAndName(ctx context.Context, organizationName string, projectName string) *Project {
	project := &Project{}
	err := hub.DB.ModelContext(ctx, project).
		Relation("Organization", func(q *orm.Query) (*orm.Query, error) {
			return q.Where("lower(organization.name) = lower(?)", organizationName), nil
		}).
		Where("lower(project.name) = lower(?)", projectName).
		Column("project.*", "Streams", "Users").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return project
}

// GetProjectID implements engine/driver.Project
func (p *Project) GetProjectID() uuid.UUID {
	return p.ProjectID
}

// GetProjectName implements engine/driver.Project
func (p *Project) GetProjectName() string {
	return p.Name
}

// GetPublic implements engine/driver.Project
func (p *Project) GetPublic() bool {
	return p.Public
}

// CreateWithUser creates a project and makes user a member
func (p *Project) CreateWithUser(ctx context.Context, userID uuid.UUID, perms ProjectPermissions) error {
	// validate
	err := GetValidator().Struct(p)
	if err != nil {
		return err
	}

	// create project and PermissionsUsersProjects in one transaction
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert project
		_, err := tx.Model(p).Insert()
		if err != nil {
			return err
		}

		// connect project to userID
		err = tx.Insert(&PermissionsUsersProjects{
			UserID:    userID,
			ProjectID: p.ProjectID,
			View:      perms.View,
			Create:    perms.Create,
			Admin:     perms.Admin,
		})
		if err != nil {
			return err
		}

		// register in engine
		err = hub.Engine.RegisterProject(ctx, p)
		if err != nil {
			return err
		}

		return nil
	})
}

// UpdateDetails updates projects user-facing details
func (p *Project) UpdateDetails(ctx context.Context, displayName *string, public *bool, site *string, description *string, photoURL *string) error {
	// set fields
	if displayName != nil {
		p.DisplayName = *displayName
	}
	if public != nil {
		p.Public = *public
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

	// note: if we ever support renaming projects, must invalidate stream cache for all instances in project

	// update in tx with call to bigquery
	err = hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		p.UpdatedOn = time.Now()
		_, err = hub.DB.WithContext(ctx).Model(p).
			Column("display_name", "public", "site", "description", "photo_url", "updated_on").
			WherePK().
			Update()
		if err != nil {
			return err
		}

		// update in warehouse
		err = hub.Engine.RegisterProject(ctx, p)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// SetLock sets a project's "locked" status
func (p *Project) SetLock(ctx context.Context, isLocked bool) error {
	p.Locked = isLocked
	p.UpdatedOn = time.Now()

	_, err := hub.DB.ModelContext(ctx, p).
		Column("locked", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return nil
}

// Delete safely deletes the project (fails if the project still has content)
func (p *Project) Delete(ctx context.Context) error {
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		err := hub.DB.WithContext(ctx).Delete(p)
		if err != nil {
			return err
		}

		err = hub.Engine.RemoveProject(ctx, p)
		if err != nil {
			return err
		}

		return nil
	})
}
