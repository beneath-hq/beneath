package user

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/beneath-hq/beneath/models"
)

// UpdateConsent updates the user's consent preferences
func (s *Service) UpdateConsent(ctx context.Context, u *models.User, terms *bool, newsletter *bool) error {
	if !u.ConsentTerms && (terms == nil || !*terms) {
		return fmt.Errorf("Cannot continue without consent to the terms of service")
	}

	if terms != nil {
		if !*terms {
			return fmt.Errorf("You cannot withdraw consent to the terms of service; instead, delete your user")
		}
		u.ConsentTerms = *terms
	}

	if newsletter != nil {
		u.ConsentNewsletter = *newsletter
	}

	// validate and save
	u.UpdatedOn = time.Now()
	err := u.Validate()
	if err != nil {
		return err
	}

	_, err = s.DB.GetDB(ctx).ModelContext(ctx, u).
		Column("consent_terms", "consent_newsletter", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return nil
}

// UpdateQuotas change the user's quotas
func (s *Service) UpdateQuotas(ctx context.Context, u *models.User, readQuota *int64, writeQuota *int64, scanQuota *int64) error {
	// set fields
	u.ReadQuota = readQuota
	u.WriteQuota = writeQuota
	u.ScanQuota = scanQuota

	// validate and save
	u.UpdatedOn = time.Now()
	err := u.Validate()
	if err != nil {
		return err
	}
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, u).
		Column("read_quota", "write_quota", "scan_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, models.UserUpdatedEvent{
		User: u,
	})
	if err != nil {
		return err
	}

	return nil
}
