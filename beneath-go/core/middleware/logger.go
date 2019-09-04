package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	chimiddleware "github.com/go-chi/chi/middleware"

	"github.com/beneath-core/beneath-go/core/log"
)

// Logger is a middleware that logs each request, along with some useful data about what
// was requested, what the response status was, and how long it took to return.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()

		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		tags := GetTags(r.Context())
		l := loggerWithTags(log.L, tags)

		l.Info(
			"http request",
			zap.String("method", r.Method),
			zap.String("host", r.Host),
			zap.String("path", r.RequestURI),
			zap.String("ip", r.RemoteAddr),
			zap.Duration("time", time.Since(t1)),
			zap.Int("status", ww.Status()),
			zap.Int("size", ww.BytesWritten()),
		)
	})
}

// LoggerUnaryServerInterceptor is a gRPC interceptor that logs each request
func LoggerUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		t1 := time.Now()
		resp, err := handler(ctx, req)

		tags := GetTags(ctx)
		l := loggerWithTags(log.L, tags)

		l.Info(
			"grpc unary request",
			zap.String("method", info.FullMethod),
			zap.Uint32("code", uint32(status.Code(err))),
			zap.Duration("time", time.Since(t1)),
		)

		return resp, err
	}
}

// LoggerStreamServerInterceptor is a gRPC interceptor that logs each streaming request
func LoggerStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		t1 := time.Now()
		err := handler(srv, ss)

		tags := GetTags(ss.Context())
		l := loggerWithTags(log.L, tags)

		l.Info(
			"grpc stream request",
			zap.String("method", info.FullMethod),
			zap.Uint32("code", uint32(status.Code(err))),
			zap.Duration("time", time.Since(t1)),
		)

		return err
	}
}

func loggerWithTags(l *zap.Logger, tags *Tags) *zap.Logger {
	if tags.Secret != nil {
		if tags.Secret.UserID != nil {
			l = l.With(
				zap.String("secret", tags.Secret.SecretID.String()),
				zap.String("user", tags.Secret.UserID.String()),
			)
		} else if tags.Secret.ProjectID != nil {
			l = l.With(
				zap.String("secret", tags.Secret.SecretID.String()),
				zap.String("project", tags.Secret.ProjectID.String()),
			)
		}
	}

	if tags.Query != nil {
		l = l.With(
			zap.Reflect("query", tags.Query),
		)
	}

	return l
}
