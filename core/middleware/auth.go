package middleware

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/core/httputil"
	"github.com/beneath-core/core/secrettoken"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	uuid "github.com/satori/go.uuid"
)

// GetSecret extracts the auth object from ctx
func GetSecret(ctx context.Context) entity.Secret {
	tags := GetTags(ctx)
	return tags.Secret
}

// Auth reads bearer token and injects auth into the context of an HTTP request
// Sets ContextKey to nil if no authorization passed (contrary to gRPC)
func Auth(next http.Handler) http.Handler {
	return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) error {
		var secret entity.Secret
		secret = &entity.AnonymousSecret{}

		header := r.Header.Get("Authorization")
		if header != "" {
			if len(header) < 6 || !strings.EqualFold(header[0:6], "bearer") {
				return httputil.NewError(400, "authentication error: bearer authorization header required")
			}

			tokenStr := strings.TrimSpace(header[6:])
			token, err := secrettoken.FromString(tokenStr)
			if err != nil {
				return httputil.NewError(400, fmt.Sprintf("authentication error: %v", err.Error()))
			}

			secret = entity.AuthenticateWithToken(r.Context(), token)
			if secret == nil || reflect.ValueOf(secret).IsNil() {
				return httputil.NewError(400, "authentication error: token not found")
			}
		}

		tags := GetTags(r.Context())
		tags.Secret = secret
		tags.AnonymousID = uuid.FromStringOrNil(r.Header.Get("X-Beneath-Aid"))

		next.ServeHTTP(w, r)
		return nil
	})
}

// AuthInterceptor reads bearer token and injects auth into the context of a gRPC call
// Errors if no authorization passed (contrary to HTTP)
func AuthInterceptor(ctx context.Context) (context.Context, error) {
	tokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %v", err)
	}

	token, err := secrettoken.FromString(tokenStr)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %v", err)
	}

	secret := entity.AuthenticateWithToken(ctx, token)
	if secret == nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: secret not found")
	}

	tags := GetTags(ctx)
	tags.Secret = secret

	return ctx, nil
}
