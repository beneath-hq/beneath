package organization

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/models"
)

// CreateWithUser creates an organization and makes user a member
func (s *Service) CreateWithUser(ctx context.Context, name string, userID uuid.UUID, view bool, create bool, admin bool) (*models.Organization, error) {
	// validate name
	org := &models.Organization{Name: name}
	err := org.Validate()
	if err != nil {
		return nil, err
	}

	// create in tx
	err = s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		// insert org
		_, err := tx.Model(org).Insert()
		if err != nil {
			return err
		}

		// connect to user
		err = tx.Insert(&models.PermissionsUsersOrganizations{
			UserID:         userID,
			OrganizationID: org.OrganizationID,
			View:           view,
			Create:         create,
			Admin:          admin,
		})
		if err != nil {
			return err
		}

		// send event
		err = s.Bus.Publish(ctx, &models.OrganizationCreatedEvent{
			Organization: org,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return org, nil
}

// UpdateDetails updates the organization's name, display name, description and/or photo
func (s *Service) UpdateDetails(ctx context.Context, o *models.Organization, name *string, displayName *string, description *string, photoURL *string) error {
	if name != nil {
		o.Name = *name
	}
	if displayName != nil {
		o.DisplayName = *displayName
	}
	if description != nil {
		o.Description = *description
	}
	if photoURL != nil {
		o.PhotoURL = *photoURL
	}

	err := o.Validate()
	if err != nil {
		return err
	}

	o.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, o).
		Column(
			"name",
			"display_name",
			"description",
			"photo_url",
			"updated_on",
		).WherePK().Update()
	if err != nil {
		return err
	}

	// send event
	err = s.Bus.Publish(ctx, &models.OrganizationUpdatedEvent{
		Organization: o,
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateQuotas updates the quotas enforced upon the organization
func (s *Service) UpdateQuotas(ctx context.Context, o *models.Organization, readQuota *int64, writeQuota *int64, scanQuota *int64) error {
	// set fields
	o.ReadQuota = readQuota
	o.WriteQuota = writeQuota
	o.ScanQuota = scanQuota

	// validate
	err := o.Validate()
	if err != nil {
		return err
	}

	// update
	o.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, o).
		Column("read_quota", "write_quota", "scan_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	// send event
	err = s.Bus.Publish(ctx, &models.OrganizationUpdatedEvent{
		Organization: o,
	})
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the organization. It fails if there are any users, services, or projects
// that are still tied to the organization.
func (s *Service) Delete(ctx context.Context, o *models.Organization) error {
	// NOTE: effectively, this resolver doesn't work, as there will always be one admin member.
	// TODO: check it's empty and has just one admin user, trigger bill, make the admin leave, delete org.

	_, err := s.DB.GetDB(ctx).ModelContext(ctx, o).Delete()
	if err != nil {
		return err
	}

	return nil
}
