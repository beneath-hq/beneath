package entity

import (
	"context"
	"time"

	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/db"
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

// CreateStreamInstance creates a new instance
func CreateStreamInstance(ctx context.Context, streamID uuid.UUID) (res *StreamInstance, err error) {
	err = db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		res, err = CreateStreamInstanceWithTx(tx, streamID)
		return err
	})
	return res, err
}

// CreateStreamInstanceWithTx is the same as CreateStreamInstance, but in a database transaction
func CreateStreamInstanceWithTx(tx *pg.Tx, streamID uuid.UUID) (*StreamInstance, error) {
	si := &StreamInstance{StreamID: streamID}
	_, err := tx.Model(si).Insert()
	return si, err
}
