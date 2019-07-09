package gql

import (
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	uuid "github.com/satori/go.uuid"
)

// MarshalUUID marshals the UUID custom scalar
func MarshalUUID(id uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		w.Write([]byte(id.String()))
	})
}

// UnmarshalUUID unmarshals the UUID custom scalar
func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case string:
		return uuid.FromString(v)
	default:
		return uuid.Nil, fmt.Errorf("Cannot unmarshal %T to UUID (expected string)", v)
	}
}
