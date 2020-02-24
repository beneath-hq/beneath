package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/beneath-core/control"
	"github.com/beneath-core/control/auth"
	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/migrations"
	"github.com/beneath-core/control/payments"
	"github.com/beneath-core/control/taskqueue/worker"
	"github.com/beneath-core/gateway"
	gwgrpc "github.com/beneath-core/gateway/grpc"
	pb "github.com/beneath-core/gateway/grpc/proto"
	gwhttp "github.com/beneath-core/gateway/http"
	"github.com/beneath-core/gateway/pipeline"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/pkg/envutil"
	"github.com/beneath-core/pkg/log"
)

type configSpecification struct {
	RedisURL         string `envconfig:"CONTROL_REDIS_URL" required:"true"`
	PostgresHost     string `envconfig:"CONTROL_POSTGRES_HOST" required:"true"`
	PostgresDB       string `envconfig:"CONTROL_POSTGRES_DB" required:"true"`
	PostgresUser     string `envconfig:"CONTROL_POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"CONTROL_POSTGRES_PASSWORD" required:"true"`

	MQDriver        string `envconfig:"ENGINE_MQ_DRIVER" required:"true"`
	LookupDriver    string `envconfig:"ENGINE_LOOKUP_DRIVER" required:"true"`
	WarehouseDriver string `envconfig:"ENGINE_WAREHOUSE_DRIVER" required:"true"`
}

var (
	testUser    *entity.User
	testSecret  *entity.UserSecret
	controlHTTP *httptest.Server
	gatewayHTTP *httptest.Server
	gatewayGRPC pb.GatewayClient
)

func TestMain(m *testing.M) {
	// must run at project root to detect configs/
	os.Chdir("../..")

	// load config
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	// Init logging
	log.InitLogger()

	// connect postgres, redis, engine, and payment drivers
	hub.InitPostgres(config.PostgresHost, config.PostgresDB, config.PostgresUser, config.PostgresPassword)
	hub.InitRedis(config.RedisURL)
	hub.InitEngine(config.MQDriver, config.LookupDriver, config.WarehouseDriver)
	hub.SetPaymentDrivers(payments.InitDrivers([]string{"anarchism"}))

	// reset engine
	err := hub.Engine.Reset(context.Background())
	panicIf(err)

	// Init gateway globals
	gateway.InitMetrics()
	gateway.InitSubscriptions(hub.Engine)

	// configure auth (empty config, so it doesn't actually work)
	auth.InitGoth(&auth.GothConfig{})

	// run migrations
	migrations.MustRunResetAndUp(hub.DB)

	// flush redis
	_, err = hub.Redis.FlushAll().Result()
	panicIf(err)

	// create control server
	controlHTTP = httptest.NewServer(control.Handler("", ""))
	defer controlHTTP.Close()

	// create gateway HTTP server
	gatewayHTTP = httptest.NewServer(gwhttp.Handler())
	defer gatewayHTTP.Close()

	// create gateway GRPC server
	grpcListener := bufconn.Listen(1024 * 1024)
	grpcServer := gwgrpc.Server()
	go func() {
		err := grpcServer.Serve(grpcListener)
		panicIf(err)
	}()
	defer grpcServer.Stop()

	// create gateway GRPC client
	dialer := grpc.WithContextDialer(func(ctx context.Context, url string) (net.Conn, error) {
		return grpcListener.Dial()
	})
	gatewayClient, err := grpc.DialContext(context.Background(), "", dialer, grpc.WithInsecure())
	panicIf(err)
	defer gatewayClient.Close()
	gatewayGRPC = pb.NewGatewayClient(gatewayClient)

	// start pipeline
	go func() {
		err := pipeline.Run()
		panicIf(err)
	}()

	// start taskqueue
	go func() {
		err := worker.Work()
		panicIf(err)
	}()

	// create a user and secret
	ctx := context.Background()
	testUser, err = entity.CreateOrUpdateUser(ctx, "google", "", "test@example.org", "test", "Test Testeson", "")
	panicIf(err)
	testSecret, err = entity.CreateUserSecret(ctx, testUser.UserID, "", false, false)
	panicIf(err)

	// run tests
	code := m.Run()

	// reset database
	_, _, err = migrations.Run(hub.DB, "reset")
	panicIf(err)

	// flush redis
	_, err = hub.Redis.FlushAll().Result()
	panicIf(err)

	// exit with code
	os.Exit(code)
}

type gqlResponse struct {
	Data   map[string]map[string]interface{}
	Errors []map[string]interface{}
}

func queryGQL(query string, variables interface{}) gqlResponse {
	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	panicIf(err)

	url := fmt.Sprintf("%s/graphql", controlHTTP.URL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	panicIf(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authString())

	client := &http.Client{}
	res, err := client.Do(req)
	panicIf(err)

	var resp gqlResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	panicIf(err)

	return resp
}

func queryGatewayHTTP(method string, path string, data interface{}) (int, map[string]interface{}) {
	body, err := json.Marshal(data)
	panicIf(err)
	// log.S.Infof("HERE %v", string(body))

	url := fmt.Sprintf("%s/%s", gatewayHTTP.URL, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	panicIf(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authString())

	client := &http.Client{}
	res, err := client.Do(req)
	panicIf(err)

	var resp interface{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil && err != io.EOF {
		panic(err)
	}

	if resp == nil {
		return res.StatusCode, nil
	}
	return res.StatusCode, resp.(map[string]interface{})
}

func grpcContext() context.Context {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authString())
	return ctx
}

func authString() string {
	return fmt.Sprintf("Bearer %s", testSecret.Token.String())
}
