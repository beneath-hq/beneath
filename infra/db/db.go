package db

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

// DB represents the control-plane SQL (Postgres) database
type DB interface {
	// Returns the current transaction or the DB connection outright
	GetDB(ctx context.Context) orm.DB

	// InTransaction starts a DB transaction. Any call to GetDB with the derived ctx returns the transaction.
	InTransaction(ctx context.Context, fn func(context.Context) error) error
}

// Options for opening a new DB connection
type Options struct {
	Host     string
	Database string
	User     string
	Password string
}

// NewDB connects to the control-plane DB
func NewDB(opts *Options) DB {
	postgres := pg.Connect(&pg.Options{
		Addr:     opts.Host + ":5432",
		Database: opts.Database,
		User:     opts.User,
		Password: opts.Password,
	})

	// Uncomment to log database queries
	// postgres.AddQueryHook(queryLoggerHook{})

	return db{
		postgres: postgres,
	}
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

// DB implementation
// -----------------

// txContextKey is the transaction context key
type txContextKey struct{}

type db struct {
	postgres *pg.DB
}

func (d db) GetDB(ctx context.Context) orm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*pg.Tx); ok {
		return tx
	}
	return d.postgres
}

func (d db) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ctx.Value(txContextKey{}).(*pg.Tx); ok {
		return fn(ctx)
	}
	return d.postgres.RunInTransaction(func(tx *pg.Tx) error {
		txCtx := context.WithValue(ctx, txContextKey{}, tx)
		return fn(txCtx)
	})
}
