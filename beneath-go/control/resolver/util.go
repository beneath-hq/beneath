package resolver

import (
	"github.com/vektah/gqlparser/gqlerror"
)

// DereferenceString does what it says it does
func DereferenceString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// MakeUnauthenticatedError does what it says it does
func MakeUnauthenticatedError(msg string) *gqlerror.Error {
	return &gqlerror.Error{
		Message: msg,
		Extensions: map[string]interface{}{
			"code": "UNAUTHENTICATED",
		},
	}
}
