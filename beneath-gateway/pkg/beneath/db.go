package beneath

import (
	"log"

	"github.com/go-pg/pg"
)

var (
	// DB is the postgres connection
	DB pg.DB
)

func init() {
	opts, err := pg.ParseURL(Config.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %s", err.Error())
	}

	DB := pg.Connect(opts)
	defer DB.Close()
}
