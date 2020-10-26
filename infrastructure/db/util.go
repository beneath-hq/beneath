package db

import (
	"github.com/go-pg/pg/v9"
)

// AssertFoundOne uses the error from a QueryOne operation
// (which includes Select for a single object) to determine
// if a result was found or not. Panics on database errors
func AssertFoundOne(err error) bool {
	if err != nil {
		if err == pg.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}
