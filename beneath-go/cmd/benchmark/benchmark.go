package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/db"

	"golang.org/x/sync/errgroup"
)

type configSpecification struct {
	MQDriver         string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LogDriver        string `envconfig:"ENGINE_LOG_DRIVER" required:"true"`
	LookupDriver     string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver  string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`
}

var (
	// Config ...
	Config configSpecification
)

func init() {
	core.LoadConfig("beneath", &Config)
	db.InitPostgres(Config.PostgresHost, Config.PostgresUser, Config.PostgresPassword)
	db.InitRedis(Config.RedisURL)
	db.InitEngine(Config.MQDriver, Config.LogDriver, Config.LookupDriver, Config.WarehouseDriver)
}

func main() {
	group := new(errgroup.Group)

	user, err := entity.CreateOrUpdateUser(context.Background(), "abc", "", "test@example.com", "test", "Mr. Test", "")
	if err != nil {
		log.Fatal(err)
	}

	n := 500
	m := 10
	for i := 0; i < n; i++ {
		group.Go(func() error {
			k := i

			secret, err := entity.CreateUserSecret(context.Background(), user.UserID, "")
			if err != nil {
				return err
			}

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			for j := 0; j < m; j++ {
				req, err := http.NewRequest("GET", "http://localhost:5000/streams/instances/eacdd317-c2c5-4287-af44-4ab58ffc2166", nil)
				if err != nil {
					return err
				}
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secret.SecretString))
				res, err := client.Do(req)
				if err != nil {
					return err
				}

				res.Body.Close()

				if k%50 == 0 {
					if j%10 == 0 {
						log.Printf("%d -- %d\n", i, j)
					}
				}
			}
			return nil
		})
	}

	// run simultaneously
	log.Println(group.Wait())
}
