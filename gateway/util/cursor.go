package util

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// CursorType represents the engine service of the underlying cursor (payload)
type CursorType byte

const (
	// LogCursorType represents a LogService cursor
	LogCursorType CursorType = 1

	// IndexCursorType represents an IndexService cursor
	IndexCursorType CursorType = 2

	// WarehouseCursorType represents a WarehouseCursorType cursor
	WarehouseCursorType CursorType = 3
)

const (
	cursorTypeSize = 1
)

// Cursor wraps an underlying engine cursor with cursor type and UUID
type Cursor []byte

// NewCursor builds a new Cursor
func NewCursor(cursorType CursorType, id uuid.UUID, payload []byte) Cursor {
	if len(payload) == 0 {
		panic(fmt.Errorf("cannot create cursor with empty payload"))
	}

	n := cursorTypeSize + uuid.Size + len(payload)

	res := make([]byte, n)
	idx := 0

	res[idx] = byte(cursorType)
	idx += cursorTypeSize

	copy(res[idx:(idx+uuid.Size)], id[:])
	idx += uuid.Size

	copy(res[idx:n], payload)

	return Cursor(res)
}

// CursorFromBytes creates (and validates) a Cursor from the return value of cursor.GetBytes()
func CursorFromBytes(bytes []byte) (Cursor, error) {
	if len(bytes) <= cursorTypeSize+uuid.Size {
		return nil, fmt.Errorf("corrupted cursor (too short to include type, instance_id and payload)")
	}

	cursorType := CursorType(bytes[0])
	if cursorType != LogCursorType && cursorType != IndexCursorType && cursorType != WarehouseCursorType {
		return nil, fmt.Errorf("corrupted cursor (unknown type %v)", cursorType)
	}

	return Cursor(bytes), nil
}

// GetBytes returns a byte representation of the cursor. It can be reversed with CursorFromBytes.
func (c Cursor) GetBytes() []byte {
	return c
}

// GetType returns the cursor's type
func (c Cursor) GetType() CursorType {
	return CursorType(c[0])
}

// GetID returns the ID embedded in the cursor
func (c Cursor) GetID() uuid.UUID {
	return uuid.FromBytesOrNil(c[cursorTypeSize:(cursorTypeSize + uuid.Size)])
}

// GetPayload returns the cursor's payload
func (c Cursor) GetPayload() []byte {
	return c[(cursorTypeSize + uuid.Size):]
}
