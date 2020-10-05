package events

import (
	"context"
	"fmt"

	"gitlab.com/beneath-hq/beneath/bi/sendgrid"
	"gitlab.com/beneath-hq/beneath/control/entity"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/engine"
)

// HandleUserCreate processes a UserCreate event
func HandleUserCreate(ctx context.Context, event *engine.ControlEvent) error {
	// lookup user
	userIDstring := event.Data["user_id"]
	userID, err := uuid.FromString(userIDstring.(string))
	if err != nil {
		return fmt.Errorf("could not get user_id from the control event %v", event.ID)
	}
	user := entity.FindUser(ctx, userID)

	// send welcome email
	err = sendgrid.SendWelcomeEmail(user.Email)
	if err != nil {
		return fmt.Errorf("could not send welcome email %v", err)
	}

	return nil
}
