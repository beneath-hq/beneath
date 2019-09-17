package task

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

// CleanupProject is a task that removes all data related to a project
type CleanupProject struct {
	ProjectID uuid.UUID
}

// Run triggers the task
func (t *CleanupProject) Run(ctx context.Context) error {
	return nil
}
