package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type StreamInstance struct {
	StreamInstanceID uuid.UUID `sql:",pk,type:uuid"`
	StreamID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Stream           *Stream
	CreatedOn        time.Time `sql:",default:now()"`
	UpdatedOn        time.Time `sql:",default:now()"`
}