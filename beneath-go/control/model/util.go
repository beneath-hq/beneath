package model

import (
	"log"

	"github.com/go-pg/pg"
)

// AssertFoundOne uses the error from a QueryOne operation
// (which includes Select for a single object) to determine
// if a result was found or not. Panics on database errors
func AssertFoundOne(err error) bool {
	if err != nil {
		if err == pg.ErrNoRows {
			return false
		}
		log.Panic(err.Error())
		return false
	}
	return true
}
