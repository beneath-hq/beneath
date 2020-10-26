package data

import (
	"context"
	"net/http"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"gitlab.com/beneath-hq/beneath/pkg/grpcutil"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	gwgrpc "gitlab.com/beneath-hq/beneath/server/data/grpc"
	gwhttp "gitlab.com/beneath-hq/beneath/server/data/http"
	"gitlab.com/beneath-hq/beneath/services/data"
	"gitlab.com/beneath-hq/beneath/services/middleware"
	"gitlab.com/beneath-hq/beneath/services/secret"
	"gitlab.com/beneath-hq/beneath/services/stream"
)

// ServerOptions are the options for creating a data server
type ServerOptions struct {
	HTTPPort int `mapstructure:"http_port"`
	GRPCPort int `mapstructure:"grpc_port"`
}

// Server is the data server
type Server struct {
	Opts        *ServerOptions
	DataService *data.Service
	HTTP        *http.Server
	GRPC        *grpc.Server
}

// NewServer initializes a new data-plane server that supports HTTP and GRPC
func NewServer(opts *ServerOptions, data *data.Service, middleware *middleware.Service, secret *secret.Service, stream *stream.Service) *Server {
	s := &Server{
		Opts:        opts,
		DataService: data,
		GRPC:        gwgrpc.NewServer(data, middleware),
		HTTP:        gwhttp.NewServer(data, middleware, secret, stream),
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

	log.S.Infof("serving data server on http port %d and grpc port %d", s.Opts.HTTPPort, s.Opts.GRPCPort)
	return group.Wait()
}
