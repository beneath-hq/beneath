package project

import (
	"context"

	"github.com/beneath-hq/beneath/models"
)

// CreateUserStarterProject creates a starter project when a user is created
func (s *Service) CreateUserStarterProject(ctx context.Context, msg *models.UserCreatedEvent) error {
	// stage initial project
	starterProject := &models.Project{
		Name:           "starter_project",
		DisplayName:    "Starter project",
		Description:    "We automatically created this project for you to help you get started",
		Public:         true,
		OrganizationID: msg.User.BillingOrganizationID,
	}

	err := s.CreateWithUser(ctx, starterProject, nil, nil, nil, nil, nil, msg.User.UserID, models.ProjectPermissions{
		View:   true,
		Create: true,
		Admin:  true,
	})
	if err != nil {
		return err
	}

	return nil
}
