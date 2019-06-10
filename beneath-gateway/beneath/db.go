package beneath

import (
	"log"

	"github.com/go-pg/pg"
)

var (
	// DB is the postgres connection
	DB *pg.DB
)

func init() {
	opts, err := pg.ParseURL(Config.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %s", err.Error())
	}

	DB = pg.Connect(opts)

	// Uncomment to log database queries
	// DB.AddQueryHook(queryLoggerHook{})
}

// queryLoggerHook logs every database query
type queryLoggerHook struct{}

func (h queryLoggerHook) BeforeQuery(q *pg.QueryEvent) {
}

func (h queryLoggerHook) AfterQuery(q *pg.QueryEvent) {
	log.Println(q.FormattedQuery())
}
