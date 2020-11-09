package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"gitlab.com/beneath-hq/beneath/pkg/httputil"
)

// RecovererMiddleware is a HTTP middleware that catches any downstream panic calls, and
// logs them without halting execution of the entire process
func (s *Service) RecovererMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) (httpErr error) {
			defer func() {
				if r := recover(); r != nil {
					rerr, _ := r.(error)
					logger.Error(
						"http recovered panic",
						zap.Error(rerr),
					)
					httpErr = httputil.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				}
			}()
			next.ServeHTTP(w, r)
			return nil
		})
	}
}

// RecovererUnaryServerInterceptor is a gRPC interceptor similar to Recoverer
func (s *Service) RecovererUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				rerr, _ := r.(error)
				logger.Error(
					"grpc unary recovered panic",
					zap.Error(rerr),
				)
				err = grpc.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// RecovererStreamServerInterceptor is a gRPC interceptor similar to Recoverer
func (s *Service) RecovererStreamServerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				rerr, _ := r.(error)
				logger.Error(
					"grpc stream recovered panic",
					zap.Error(rerr),
				)
				err = grpc.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}
