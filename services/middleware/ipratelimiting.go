package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v8"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
)

func (s *Service) initRateLimiter() {
	s.limiter = redis_rate.NewLimiter(hub.Redis, &redis_rate.Limit{
		Burst:  20,
		Rate:   20,
		Period: time.Second * 15,
	})
}

// IPRateLimitMiddleware is a HTTP middleware that uses the ipaddress for rate limiting.
// It skips rate limit checks if secret is nil.
func (s *Service) IPRateLimitMiddleware(next http.Handler) http.Handler {
	return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) error {
		secret := GetSecret(r.Context())
		if secret.IsAnonymous() {
			// check rate limit
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			res, err := s.limiter.Allow(ip)
			if err != nil {
				panic(err)
			}

			if !res.Allowed {
				return httputil.NewError(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
			}
		}
		next.ServeHTTP(w, r)
		return nil
	})
}

// IPRateLimitUnaryServerInterceptor is like IPRateLimit, but for unary gRPC calls
func (s *Service) IPRateLimitUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		secret := GetSecret(ctx)
		if secret.IsAnonymous() {
			// check rate limit
			ip := gRPCRealIP(ctx)
			res, err := s.limiter.Allow(ip)
			if err != nil {
				panic(err)
			}

			if !res.Allowed {
				return nil, status.Error(codes.ResourceExhausted, "Too many requests - please authenticate")
			}
		}
		return handler(ctx, req)
	}
}

// IPRateLimitStreamServerInterceptor is like IPRateLimit, but for streaming gRPC calls
func (s *Service) IPRateLimitStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		secret := GetSecret(ss.Context())
		if secret.IsAnonymous() {
			// check rate limit
			ip := gRPCRealIP(ss.Context())

			res, err := s.limiter.Allow(ip)
			if err != nil {
				panic(err)
			}

			if !res.Allowed {
				return status.Error(codes.ResourceExhausted, "Too many requests - please authenticate")
			}
		}
		return handler(srv, ss)
	}
}

func gRPCRealIP(ctx context.Context) string {
	md := metautils.ExtractIncoming(ctx)
	addr := md.Get("x-forwarded-for")
	if addr == "" {
		p, _ := peer.FromContext(ctx)
		addr = p.Addr.String()
	}

	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		ip = addr
	}

	return ip
}
