package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/beneath-core/pkg/httputil"
	"github.com/beneath-core/pkg/log"
)

// Recoverer is a middleware that catches any downstream panic calls, and logs them without
// halting execution of the entire process
func Recoverer(next http.Handler) http.Handler {
	return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) (httpErr error) {
		defer func() {
			if r := recover(); r != nil {
				err, _ := r.(error)
				log.L.Error(
					"http recovered panic",
					zap.Error(err),
				)
				httpErr = httputil.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}
		}()
		next.ServeHTTP(w, r)
		return nil
	})
}

// RecovererUnaryServerInterceptor is a gRPC interceptor similar to Recoverer
func RecovererUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err, _ := r.(error)
				log.L.Error(
					"grpc unary recovered panic",
					zap.Error(err),
				)
				err = grpc.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// RecovererStreamServerInterceptor is a gRPC interceptor similar to Recoverer
func RecovererStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err, _ := r.(error)
				log.L.Error(
					"grpc stream recovered panic",
					zap.Error(err),
				)
				err = grpc.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}
