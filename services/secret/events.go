package secret

import (
	"context"

	"gitlab.com/beneath-hq/beneath/models"
)

func (s *Service) userUpdated(ctx context.Context, msg *models.UserUpdatedEvent) error {
	s.cache.ClearForUser(ctx, msg.User.UserID)
	return nil
}

func (s *Service) organizationUpdated(ctx context.Context, msg *models.OrganizationUpdatedEvent) error {
	s.cache.ClearForOrganization(ctx, msg.Organization.OrganizationID)
	return nil
}

func (s *Service) serviceUpdated(ctx context.Context, msg *models.ServiceUpdatedEvent) error {
	s.cache.ClearForService(ctx, msg.Service.ServiceID)
	return nil
}

func (s *Service) serviceDeleted(ctx context.Context, msg *models.ServiceDeletedEvent) error {
	s.cache.ClearForService(ctx, msg.ServiceID)
	return nil
}
