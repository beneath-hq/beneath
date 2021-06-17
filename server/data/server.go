package data

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/beneath-hq/beneath/pkg/grpcutil"
	"github.com/beneath-hq/beneath/pkg/httputil"
	gwgrpc "github.com/beneath-hq/beneath/server/data/grpc"
	gwhttp "github.com/beneath-hq/beneath/server/data/http"
	"github.com/beneath-hq/beneath/services/data"
	"github.com/beneath-hq/beneath/services/middleware"
	"github.com/beneath-hq/beneath/services/secret"
	"github.com/beneath-hq/beneath/services/table"
)

// ServerOptions are the options for creating a data server
type ServerOptions struct {
	HTTPPort int `mapstructure:"http_port"`
	GRPCPort int `mapstructure:"grpc_port"`
}

// Server is the data server
type Server struct {
	Opts        *ServerOptions
	Logger      *zap.SugaredLogger
	DataService *data.Service
	HTTP        *http.Server
	GRPC        *grpc.Server
}

// NewServer initializes a new data-plane server that supports HTTP and GRPC
func NewServer(opts *ServerOptions, logger *zap.Logger, data *data.Service, middleware *middleware.Service, secret *secret.Service, table *table.Service) *Server {
	l := logger.Named("data.server")
	s := &Server{
		Opts:        opts,
		Logger:      l.Sugar(),
		DataService: data,
		GRPC:        gwgrpc.NewServer(l, data, middleware),
		HTTP:        gwhttp.NewServer(l, data, middleware, secret, table),
	}

	return s
}

// Run starts serving on HTTP and GRPC
func (s *Server) Run(ctx context.Context) error {
	group, cctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return httputil.ListenAndServeContext(cctx, s.HTTP, s.Opts.HTTPPort)
	})

	group.Go(func() error {
		return grpcutil.ListenAndServeContext(cctx, s.GRPC, s.Opts.GRPCPort)
	})

	group.Go(func() error {
		return s.DataService.ServeSubscriptions(cctx)
	})

	s.Logger.Infof("serving on http port %d and grpc port %d", s.Opts.HTTPPort, s.Opts.GRPCPort)
	return group.Wait()
}
