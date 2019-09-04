package db

import (
	"fmt"

	"github.com/go-pg/pg"
)

func newDatabase(postgresURL string) *pg.DB {
	opts, err := pg.ParseURL(postgresURL)
	if err != nil {
		panic(err)
	}

	db := pg.Connect(opts)

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
