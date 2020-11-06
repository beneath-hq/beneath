package project

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// Service manages Beneath projects
type Service struct {
	Bus *bus.Bus
	DB  db.DB
}

// New creates a project service
func New(bus *bus.Bus, db db.DB) *Service {
	s := &Service{
		Bus: bus,
		DB:  db,
	}
	bus.AddSyncListener(s.CreateUserStarterProject)
	return s
}

// FindProject finds a project by ID
func (s *Service) FindProject(ctx context.Context, projectID uuid.UUID) *models.Project {
	project := &models.Project{
		ProjectID: projectID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, project).
		WherePK().
		Column("project.*", "Streams", "Services", "Organization").
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return project
}

// FindProjectByOrganizationAndName finds a project by organization name and project name
func (s *Service) FindProjectByOrganizationAndName(ctx context.Context, organizationName string, projectName string) *models.Project {
	project := &models.Project{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, project).
		Relation("Organization", func(q *orm.Query) (*orm.Query, error) {
			return q.Where("lower(organization.name) = lower(?)", organizationName), nil
		}).
		Where("lower(project.name) = lower(?)", projectName).
		Column("project.*", "Streams", "Services").
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return project
}

// FindProjectsForUser finds the projects that the user has been granted access to
func (s *Service) FindProjectsForUser(ctx context.Context, userID uuid.UUID) []*models.Project {
	var projects []*models.Project
	err := s.DB.GetDB(ctx).ModelContext(ctx, &projects).
		Column("project.*", "Organization").
		Join("JOIN permissions_users_projects AS pup ON pup.project_id = project.project_id").
		Where("pup.user_id = ?", userID).
		Order("project.name").
		Limit(200).
		Select()
	if err != nil {
		panic(err)
	}
	return projects
}

// ExploreProjects returns a list of featured projects
func (s *Service) ExploreProjects(ctx context.Context) []*models.Project {
	var projects []*models.Project
	err := s.DB.GetDB(ctx).ModelContext(ctx, &projects).
		Where("project.explore_rank IS NOT NULL").
		Limit(200).
		Order("explore_rank").
		Relation("Organization").
		Select()
	if err != nil {
		panic(err)
	}
	return projects
}

// FindProjectMembers is an effective way to get info about a project's members (represented with ProjectMember instead of User)
func (s *Service) FindProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error) {
	var result []*models.ProjectMember
	_, err := s.DB.GetDB(ctx).QueryContext(ctx, &result, `
		select
			p.project_id,
			p.user_id,
			o.name,
			o.display_name,
			o.photo_url,
			p.view,
			p."create",
			p.admin
		from permissions_users_projects p
		join organizations o on p.user_id = o.user_id
		where p.project_id = ?
	`, projectID)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return result, nil
}

// StageWithUser updates the project if it already exists or creates it with the given user as a member
func (s *Service) StageWithUser(ctx context.Context, p *models.Project, displayName *string, public *bool, description *string, site *string, photoURL *string, userID uuid.UUID, perms models.ProjectPermissions) error {
	// determine whether to insert or update
	update := (p.ProjectID != uuid.Nil)

	// tracks whether a save is necessary
	save := !update

	if displayName != nil && p.DisplayName != *displayName {
		p.DisplayName = *displayName
		save = true
	}

	if public != nil && p.Public != *public {
		p.Public = *public
		save = true
	}

	if description != nil && p.Description != *description {
		p.Description = *description
		save = true
	}

	if site != nil && p.Site != *site {
		p.Site = *site
		save = true
	}

	if photoURL != nil && p.PhotoURL != *photoURL {
		p.PhotoURL = *photoURL
		save = true
	}

	// quit if no changes
	if !save {
		return nil
	}

	// validate
	err := p.Validate()
	if err != nil {
		return err
	}

	// note: if we ever support renaming projects, must invalidate stream cache for all instances in project

	err = s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)
		if update {
			// update
			p.UpdatedOn = time.Now()
			_, err = tx.Model(p).
				Column("display_name", "public", "description", "site", "photo_url", "updated_on").
				WherePK().
				Update()
			if err != nil {
				return err
			}

			err = s.Bus.Publish(ctx, &models.ProjectUpdatedEvent{
				Project: p,
			})
			if err != nil {
				return err
			}
		} else {
			// insert
			_, err := tx.Model(p).Insert()
			if err != nil {
				return err
			}

			// connect project to userID
			err = tx.Insert(&models.PermissionsUsersProjects{
				UserID:    userID,
				ProjectID: p.ProjectID,
				View:      perms.View,
				Create:    perms.Create,
				Admin:     perms.Admin,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Delete safely deletes the project (fails if the project still has content)
func (s *Service) Delete(ctx context.Context, project *models.Project) error {
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, project).WherePK().Delete()
	if err != nil {
		return err
	}

	err = s.Bus.Publish(ctx, &models.ProjectDeletedEvent{ProjectID: project.ProjectID})
	if err != nil {
		return err
	}

	return nil
}
