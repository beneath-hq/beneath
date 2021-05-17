package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/models"
)

// TransferProject transfers a project in fromOrg to targetOrg
func (s *Service) TransferProject(ctx context.Context, fromOrg *models.Organization, targetOrg *models.Organization, project *models.Project) error {
	project.Organization = targetOrg
	project.OrganizationID = targetOrg.OrganizationID
	project.UpdatedOn = time.Now()
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, project).Column("organization_id", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	err = s.Bus.Publish(ctx, &models.ProjectUpdatedEvent{
		Project: project,
	})
	if err != nil {
		return err
	}

	return nil
}

// TransferUser transfers a user in fromOrg to targetOrg.
// Can be from personal to multi, multi to personal, multi to multi.
func (s *Service) TransferUser(ctx context.Context, fromOrg *models.Organization, targetOrg *models.Organization, user *models.User) error {
	// Ensure the last admin member isn't leaving a multi-user org
	if fromOrg.IsMulti() {
		foundRemainingAdmin := false
		members, err := s.FindOrganizationMembers(ctx, fromOrg.OrganizationID)
		if err != nil {
			return err
		}
		for _, member := range members {
			if member.Admin && member.UserID != user.UserID {
				foundRemainingAdmin = true
				break
			}
		}
		if !foundRemainingAdmin {
			return fmt.Errorf("Cannot transfer user because it would leave the organization without an admin")
		}
	}

	return s.DB.InTransaction(ctx, func(ctx context.Context) error {
		// update user
		user.BillingOrganization = targetOrg
		user.BillingOrganizationID = targetOrg.OrganizationID
		user.ReadQuota = nil
		user.WriteQuota = nil
		user.ScanQuota = nil
		user.UpdatedOn = time.Now()
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, user).
			Column("billing_organization_id", "read_quota", "write_quota", "scan_quota", "updated_on").
			WherePK().
			Update()
		if err != nil {
			return err
		}

		// publish events
		err = s.Bus.Publish(ctx, &models.UserUpdatedEvent{
			User: user,
		})
		if err != nil {
			return err
		}

		err = s.Bus.Publish(ctx, &models.OrganizationTransferredUserEvent{
			Source: fromOrg,
			Target: targetOrg,
			User:   user,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
