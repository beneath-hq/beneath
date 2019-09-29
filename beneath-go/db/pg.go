package db

import (
	"fmt"

	"github.com/go-pg/pg/v9"
)

func newDatabase(host, username, password string) *pg.DB {
	db := pg.Connect(&pg.Options{
		Addr:     host + ":5432",
		User:     username,
		Password: password,
	})

	// Uncomment to log database queries
	// db.AddQueryHook(queryLoggerHook{})

	return db
}

// queryLoggerHook logs every database query
type queryLoggerHook struct{}

func (h queryLoggerHook) BeforeQuery(q *pg.QueryEvent) {
	fmt.Print(q.FormattedQuery())
}

func (h queryLoggerHook) AfterQuery(q *pg.QueryEvent) {
}
