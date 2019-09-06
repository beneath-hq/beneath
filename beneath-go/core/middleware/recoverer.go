package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/core/log"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc"
)

// Recoverer is a middleware that catches any downstream panic calls, and logs them without
// halting execution of the entire process
func Recoverer(next http.Handler) http.Handler {
	return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) (httpErr error) {
		defer func() {
			if r := recover(); r != nil {
				log.L.Error(
					"http recovered panic",
					zap.Reflect("error", r),
					zap.Stack("stack"),
				)
				httpErr = httputil.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}
		}()

		next.ServeHTTP(w, r)
		return httpErr
	})
}

// RecovererUnaryServerInterceptor is a gRPC interceptor similar to Recoverer
func RecovererUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.L.Error(
					"grpc unary recovered panic",
					zap.Reflect("error", r),
					zap.Stack("stack"),
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
				log.L.Error(
					"grpc stream recovered panic",
					zap.Reflect("error", r),
					zap.Stack("stack"),
				)
				err = grpc.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}
