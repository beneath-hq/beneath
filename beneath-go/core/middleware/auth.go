package middleware

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/httputil"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

// GetSecret extracts the auth object from ctx
func GetSecret(ctx context.Context) *entity.Secret {
	tags := GetTags(ctx)
	return tags.Secret
}

// Auth reads bearer token and injects auth into the context of an HTTP request
// Sets ContextKey to nil if no authorization passed (contrary to gRPC)
func Auth(next http.Handler) http.Handler {
	return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) error {
		var secret *entity.Secret

		header := r.Header.Get("Authorization")
		if header != "" {
			if len(header) < 6 || !strings.EqualFold(header[0:6], "bearer") {
				return httputil.NewError(400, "bearer authorization header required")
			}

			token := strings.TrimSpace(header[6:])

			secret = entity.AuthenticateSecretString(r.Context(), token)
			if secret == nil {
				return httputil.NewError(400, "unauthenticated")
			}
		}

		tags := GetTags(r.Context())
		tags.Secret = secret

		next.ServeHTTP(w, r)
		return nil
	})
}

// AuthInterceptor reads bearer token and injects auth into the context of a gRPC call
// Errors if no authorization passed (contrary to HTTP)
func AuthInterceptor(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %v", err)
	}

	secret := entity.AuthenticateSecretString(ctx, token)

	tags := GetTags(ctx)
	tags.Secret = secret

	return ctx, nil
}
