package middleware

import (
	"context"
	"net"
	"net/http"
	"reflect"
	"time"

	chimiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// LoggerMiddleware is a HTTP middleware that logs each request, along with some useful data
// about what was requested, what the response status was, and how long it took to return.
func (s *Service) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()

		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		tags := GetTags(r.Context())
		l := loggerWithTags(log.L, tags)

		if r.RequestURI == "/healthz" {
			return
		}

		status := ww.Status()
		if status == 0 {
			status = 200
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		l.Info(
			"http request",
			zap.String("method", r.Method),
			zap.String("proto", r.Proto),
			zap.String("host", r.Host),
			zap.String("path", r.RequestURI),
			zap.String("ip", ip),
			zap.Duration("time", time.Since(t1)),
			zap.Int("status", status),
			zap.Int("size", ww.BytesWritten()),
		)
	})
}

// LoggerUnaryServerInterceptor is a gRPC interceptor that logs each request
func (s *Service) LoggerUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		t1 := time.Now()
		resp, err := handler(ctx, req)

		tags := GetTags(ctx)
		l := loggerWithTags(log.L, tags)

		l.Info(
			"grpc unary request",
			zap.String("method", info.FullMethod),
			zap.String("ip", gRPCRealIP(ctx)),
			zap.Uint32("code", uint32(status.Code(err))),
			zap.Duration("time", time.Since(t1)),
		)

		return resp, err
	}
}

// LoggerStreamServerInterceptor is a gRPC interceptor that logs each streaming request
func (s *Service) LoggerStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		t1 := time.Now()
		err := handler(srv, ss)

		tags := GetTags(ss.Context())
		l := loggerWithTags(log.L, tags)

		l.Info(
			"grpc stream request",
			zap.String("method", info.FullMethod),
			zap.String("ip", gRPCRealIP(ss.Context())),
			zap.Uint32("code", uint32(status.Code(err))),
			zap.Duration("time", time.Since(t1)),
		)

		return err
	}
}

func loggerWithTags(l *zap.Logger, tags *Tags) *zap.Logger {
	if tags.Secret != nil && !reflect.ValueOf(tags.Secret).IsNil() {
		if tags.Secret.IsUser() {
			l = l.With(
				zap.String("secret", tags.Secret.GetSecretID().String()),
				zap.String("user", tags.Secret.GetOwnerID().String()),
			)
		} else if tags.Secret.IsService() {
			l = l.With(
				zap.String("secret", tags.Secret.GetSecretID().String()),
				zap.String("service", tags.Secret.GetOwnerID().String()),
			)
		}
	}

	if tags.Payload != nil {
		l = l.With(
			zap.Reflect("payload", tags.Payload),
		)
	}

	return l
}
