package middleware

import (
	"context"
	"fmt"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"

	"github.com/beneath-core/beneath-go/control/entity"
)

// Tags represents annotations on a request made available both up and down the
// middleware/interceptor chain (e.g., secrets set down the chain becomes available
// to logging middleware at the top of the chain).
type Tags struct {
	AnonymousID uuid.UUID
	Secret      entity.Secret
	Payload     interface{}
}

// TagsContextKey is the request context key for the request's Tags object
type TagsContextKey struct{}

// InjectTags is an HTTP middleware that injects Tags into into the request context
func InjectTags(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags := &Tags{}
		ctx := context.WithValue(r.Context(), TagsContextKey{}, tags)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// InjectTagsUnaryServerInterceptor is like InjectTags, but for unary gRPC calls
func InjectTagsUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tags := &Tags{}
		newCtx := context.WithValue(ctx, TagsContextKey{}, tags)
		return handler(newCtx, req)
	}
}

// InjectTagsStreamServerInterceptor is like InjectTags, but for streaming gRPC calls
func InjectTagsStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		tags := &Tags{}
		newCtx := context.WithValue(ss.Context(), TagsContextKey{}, tags)

		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

// GetTags extracts the tags object from ctx
func GetTags(ctx context.Context) *Tags {
	tags, ok := ctx.Value(TagsContextKey{}).(*Tags)
	if !ok {
		panic(fmt.Errorf("couldn't get tags from context"))
	}
	return tags
}

// SetTagsPayload sets the Payload field of middleware.GetTags(ctx)
func SetTagsPayload(ctx context.Context, payload interface{}) {
	if payload == nil {
		return
	}

	tags := GetTags(ctx)
	tags.Payload = payload
}
