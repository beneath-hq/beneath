package entity

import (
	uuid "github.com/satori/go.uuid"
)

// StreamIndex represents an index on a stream
type StreamIndex struct {
	tableName     struct{}  `sql:"stream_indexes"`
	StreamIndexID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	StreamID      uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Stream        *Stream
	Fields        []string `sql:",notnull",validate:"required,gte=1"`
	Primary       bool     `sql:",notnull"`
	Normalize     bool     `sql:",notnull"`
}
