package resolver

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
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

// IntToInt64 converts an *int to an *int64
func IntToInt64(x *int) *int64 {
	if x == nil {
		return nil
	}
	tmp := int64(*x)
	return &tmp
}

// Int64ToInt converts an *int64 to an *int
func Int64ToInt(x *int64) *int {
	if x == nil {
		return nil
	}
	tmp := int(*x)
	return &tmp
}

// StrToPtr converts a normal string to a string pointer, making the empty string nil
func StrToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
