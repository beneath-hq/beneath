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
	ShortID       int      `sql:",notnull",validate:"required"`
	Fields        []string `sql:",notnull",validate:"required,gte=1"`
	Primary       bool     `sql:",notnull"`
	Normalize     bool     `sql:",notnull"`
}

// GetIndexID implements codec.Index
func (i *StreamIndex) GetIndexID() uuid.UUID {
	return i.StreamIndexID
}

// GetShortID implements codec.Index
func (i *StreamIndex) GetShortID() int {
	return i.ShortID
}

// GetFields implements codec.Index
func (i *StreamIndex) GetFields() []string {
	return i.Fields
}

// GetNormalize implements codec.Index
func (i *StreamIndex) GetNormalize() bool {
	return i.Normalize
}
