package task

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

// CleanupStream is a task that removes all data related to a stream
type CleanupStream struct {
	StreamID  uuid.UUID
	ProjectID uuid.UUID
}

// Run triggers the task
func (t *CleanupStream) Run(ctx context.Context) error {
	return nil
}
