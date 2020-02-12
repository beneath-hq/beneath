package entity

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-core/db"

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
	CommittedOn      time.Time
}

// GetStreamInstanceID implements engine/beneath/driver.StreamInstance
func (si *StreamInstance) GetStreamInstanceID() uuid.UUID {
	return si.StreamInstanceID
}

// FindStreamInstance finds an instance and related stream details
func FindStreamInstance(ctx context.Context, instanceID uuid.UUID) *StreamInstance {
	si := &StreamInstance{
		StreamInstanceID: instanceID,
	}
	err := db.DB.ModelContext(ctx, si).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	si.Stream = FindStream(ctx, si.StreamID)
	if si.Stream == nil {
		panic(fmt.Errorf("stream for instance is nil"))
	}
	return si
}
