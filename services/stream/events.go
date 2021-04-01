package stream

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/models"
)

const (
	keepRecentInstances = 3
)

func (s *Service) organizationUpdated(ctx context.Context, msg *models.OrganizationUpdatedEvent) error {
	s.instanceCache.ClearForOrganization(ctx, msg.Organization.OrganizationID)
	return nil
}

func (s *Service) projectUpdated(ctx context.Context, msg *models.ProjectUpdatedEvent) error {
	s.instanceCache.ClearForProject(ctx, msg.Project.ProjectID)
	return nil
}

func (s *Service) streamUpdated(ctx context.Context, msg *models.StreamUpdatedEvent) error {
	instances := s.FindStreamInstances(ctx, msg.Stream.StreamID, nil, nil)
	for _, instance := range instances {
		s.instanceCache.Clear(ctx, instance.StreamInstanceID)
		err := s.Engine.RegisterInstance(ctx, msg.Stream, instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) streamDeleted(ctx context.Context, msg *models.StreamDeletedEvent) error {
	err := s.Engine.Usage.ClearUsage(ctx, msg.StreamID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) streamInstanceCreated(ctx context.Context, msg *models.StreamInstanceCreatedEvent) error {
	err := s.Engine.RegisterInstance(ctx, msg.Stream, msg.StreamInstance)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) streamInstanceUpdated(ctx context.Context, msg *models.StreamInstanceUpdatedEvent) error {
	// clear instance cache
	s.instanceCache.Clear(ctx, msg.StreamInstance.StreamInstanceID)

	// clear name cache
	if msg.MakePrimary {
		s.nameCache.Clear(ctx, msg.OrganizationName, msg.ProjectName, msg.Stream.Name)
	}

	// delete earlier versions
	if msg.MakePrimary {
		upToVersion := msg.StreamInstance.Version - keepRecentInstances
		if upToVersion > 0 {
			instances := s.FindStreamInstances(ctx, msg.Stream.StreamID, nil, &upToVersion)
			for _, instance := range instances {
				err := s.DeleteStreamInstance(ctx, msg.Stream, instance)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) streamInstanceDeleted(ctx context.Context, msg *models.StreamInstanceDeletedEvent) error {
	time.Sleep(s.instanceCache.cacheLRUTime())

	s.instanceCache.Clear(ctx, msg.StreamInstance.StreamInstanceID)

	err := s.Engine.RemoveInstance(ctx, msg.Stream, msg.StreamInstance)
	if err != nil {
		return err
	}

	err = s.Engine.Usage.ClearUsage(ctx, msg.StreamInstance.StreamInstanceID)
	if err != nil {
		return err
	}

	return nil
}
