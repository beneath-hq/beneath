package entity

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/beneath-hq/beneath/hub"

	uuid "github.com/satori/go.uuid"
)

// StreamInstance represents a single version of a stream (for a streaming stream,
// there will only be one instance; but for a batch stream, each update represents
// a new instance)
type StreamInstance struct {
	StreamInstanceID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	StreamID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Stream           *Stream
	CreatedOn        time.Time `sql:",default:now()"`
	UpdatedOn        time.Time `sql:",default:now()"`
	MadePrimaryOn    *time.Time
	MadeFinalOn      *time.Time
	Version          int `sql:",notnull,default:0"`
}

// EfficientStreamInstance can be used to efficiently make a UUID conform to engine/driver.StreamInstance
type EfficientStreamInstance uuid.UUID

// GetStreamInstanceID implements engine/driver.StreamInstance
func (si EfficientStreamInstance) GetStreamInstanceID() uuid.UUID {
	return uuid.UUID(si)
}

// GetStreamInstanceID implements engine/driver.StreamInstance
func (si *StreamInstance) GetStreamInstanceID() uuid.UUID {
	return si.StreamInstanceID
}

// FindStreamInstances finds the instances associated with the given stream
func FindStreamInstances(ctx context.Context, streamID uuid.UUID, fromVersion *int, toVersion *int) []*StreamInstance {
	var res []*StreamInstance

	// optimization
	if fromVersion != nil && toVersion != nil && *fromVersion == *toVersion {
		return res
	}

	// build query
	query := hub.DB.ModelContext(ctx, &res).Where("stream_id = ?", streamID)
	if fromVersion != nil {
		query = query.Where("version >= ?", *fromVersion)
	}
	if toVersion != nil {
		query = query.Where("version < ?", *toVersion)
	}

	// select
	err := query.Select()
	if err != nil {
		panic(fmt.Errorf("error fetching stream instances: %s", err.Error()))
	}

	return res
}

// FindStreamInstance finds an instance and related stream details
func FindStreamInstance(ctx context.Context, instanceID uuid.UUID) *StreamInstance {
	si := &StreamInstance{
		StreamInstanceID: instanceID,
	}
	err := hub.DB.ModelContext(ctx, si).Column(
		"stream_instance.*",
		"Stream",
		"Stream.StreamIndexes",
		"Stream.Project",
		"Stream.Project.Organization",
	).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return si
}
