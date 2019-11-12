package entity

import (
	"context"

	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/taskqueue"

	uuid "github.com/satori/go.uuid"
)

// CleanupInstanceTask is a task that removes all data and tables related to an instance
type CleanupInstanceTask struct {
	InstanceID  uuid.UUID
	StreamID    uuid.UUID
	StreamName  string
	ProjectID   uuid.UUID
	ProjectName string
}

// register task
func init() {
	taskqueue.RegisterTask(&CleanupInstanceTask{})
}

// Run triggers the task
func (t *CleanupInstanceTask) Run(ctx context.Context) error {
	// delete in bigquery
	err := db.Engine.Warehouse.DeregisterStreamInstance(
		ctx,
		t.ProjectName,
		t.StreamName,
		t.InstanceID,
	)
	if err != nil {
		return err
	}

	// delete in bigtable records
	err = db.Engine.Tables.ClearRecords(ctx, t.InstanceID)
	if err != nil {
		return err
	}

	return nil
}
