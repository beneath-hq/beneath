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

	"github.com/beneath-core/beneath-go/core/httputil"
	"github.com/beneath-core/beneath-go/db"
)

var (
	limiter *redis_rate.Limiter
)

func initLimiter() *redis_rate.Limiter {
	if limiter == nil {
		limiter = redis_rate.NewLimiter(db.Redis, &redis_rate.Limit{
			Burst:  20,
			Rate:   20,
			Period: time.Second * 15,
		})
	}
	return limiter
}

// IPRateLimit is an HTTP middleware that checks the ipaddress for rate limiting
func IPRateLimit() func(http.Handler) http.Handler {
	initLimiter()
	return func(next http.Handler) http.Handler {
		return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) error {
			secret := GetSecret(r.Context())
			if secret == nil {
				// check rate limit
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					ip = r.RemoteAddr
				}

				res, err := limiter.Allow(ip)
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
}

// IPRateLimitUnaryServerInterceptor is like IPRateLimit, but for unary gRPC calls
func IPRateLimitUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	initLimiter()
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		secret := GetSecret(ctx)
		if secret == nil {
			// check rate limit
			ip := gRPCRealIP(ctx)
			res, err := limiter.Allow(ip)
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
func IPRateLimitStreamServerInterceptor() grpc.StreamServerInterceptor {
	initLimiter()
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		secret := GetSecret(ss.Context())
		if secret == nil {
			// check rate limit
			ip := gRPCRealIP(ss.Context())

			res, err := limiter.Allow(ip)
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
