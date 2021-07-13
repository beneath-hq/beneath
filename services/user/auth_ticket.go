package user

import (
	"context"
	"time"

	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
	uuid "github.com/satori/go.uuid"
)

const authTicketTTL = 10 * time.Minute

// FindAuthTicket finds an authentication ticket by ID
func (s *Service) FindAuthTicket(ctx context.Context, authTicketID uuid.UUID) *models.AuthTicket {
	ticket := &models.AuthTicket{
		AuthTicketID: authTicketID,
	}

	err := s.DB.GetDB(ctx).ModelContext(ctx, ticket).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}

	// enforce TTL
	if time.Since(ticket.CreatedOn) > authTicketTTL {
		s.DeleteAuthTicket(ctx, ticket)
		return nil
	}

	return ticket
}

// CreateAuthTicket creates a new authentication ticket that can be approved
func (s *Service) CreateAuthTicket(ctx context.Context, ticket *models.AuthTicket) error {
	// validate
	err := ticket.Validate()
	if err != nil {
		return err
	}

	// insert
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, ticket).Insert()
	if err != nil {
		return err
	}

	return nil
}

// ApproveAuthTicket approves the authentication ticket, enabling the issuance of one secret for the user
func (s *Service) ApproveAuthTicket(ctx context.Context, ticket *models.AuthTicket, userID uuid.UUID) error {
	ticket.ApproverUserID = &userID

	// validate
	err := ticket.Validate()
	if err != nil {
		return err
	}

	// update
	ticket.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, ticket).
		Column("approver_user_id", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return nil
}

// DeleteAuthTicket deletes an authentication ticket (usually because it's denied or spent)
func (s *Service) DeleteAuthTicket(ctx context.Context, ticket *models.AuthTicket) error {
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, ticket).WherePK().Delete()
	if err != nil {
		return err
	}
	return nil
}
