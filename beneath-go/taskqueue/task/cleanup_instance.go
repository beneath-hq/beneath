package task

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

// CleanupInstance is a task that removes all data and tables related to an instance
type CleanupInstance struct {
	InstanceID uuid.UUID
	StreamID   uuid.UUID
	ProjectID  uuid.UUID
}

// Run triggers the task
func (t *CleanupInstance) Run(ctx context.Context) error {
	return nil
}
