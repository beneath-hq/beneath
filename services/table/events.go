package table

import (
	"context"
	"time"

	"github.com/beneath-hq/beneath/models"
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

func (s *Service) tableUpdated(ctx context.Context, msg *models.TableUpdatedEvent) error {
	instances := s.FindTableInstances(ctx, msg.Table.TableID, nil, nil)
	for _, instance := range instances {
		s.instanceCache.Clear(ctx, instance.TableInstanceID)
		err := s.Engine.RegisterInstance(ctx, msg.Table, instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) tableDeleted(ctx context.Context, msg *models.TableDeletedEvent) error {
	err := s.Engine.Usage.ClearUsage(ctx, msg.TableID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) tableInstanceCreated(ctx context.Context, msg *models.TableInstanceCreatedEvent) error {
	err := s.Engine.RegisterInstance(ctx, msg.Table, msg.TableInstance)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) tableInstanceUpdated(ctx context.Context, msg *models.TableInstanceUpdatedEvent) error {
	// clear instance cache
	s.instanceCache.Clear(ctx, msg.TableInstance.TableInstanceID)

	// clear name cache
	if msg.MakePrimary {
		s.nameCache.Clear(ctx, msg.OrganizationName, msg.ProjectName, msg.Table.Name)
	}

	// delete earlier versions
	if msg.MakePrimary {
		upToVersion := msg.TableInstance.Version - keepRecentInstances
		if upToVersion > 0 {
			instances := s.FindTableInstances(ctx, msg.Table.TableID, nil, &upToVersion)
			for _, instance := range instances {
				err := s.DeleteTableInstance(ctx, msg.Table, instance)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) tableInstanceDeleted(ctx context.Context, msg *models.TableInstanceDeletedEvent) error {
	time.Sleep(s.instanceCache.cacheLRUTime())

	s.instanceCache.Clear(ctx, msg.TableInstance.TableInstanceID)

	err := s.Engine.RemoveInstance(ctx, msg.Table, msg.TableInstance)
	if err != nil {
		return err
	}

	err = s.Engine.Usage.ClearUsage(ctx, msg.TableInstance.TableInstanceID)
	if err != nil {
		return err
	}

	return nil
}
