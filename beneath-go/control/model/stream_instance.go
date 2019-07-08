package model

import (
	"time"

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
}
