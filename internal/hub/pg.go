package hub

import (
	"context"
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

func (h queryLoggerHook) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	fmt.Print(q.FormattedQuery())
	return ctx, nil
}

func (h queryLoggerHook) AfterQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}
