package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/core/httputil"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

// ContextKey used as key in context.Context to set/get the auth object
type ContextKey struct{}

// GetSecret extracts the auth object from ctx
func GetSecret(ctx context.Context) *model.Secret {
	auth, ok := ctx.Value(ContextKey{}).(*model.Secret)
	if !ok {
		log.Panicln("couldn't get auth from context")
	}
	return auth
}

// GRPCInterceptor reads bearer token and injects auth into the context of a gRPC call
// Errors if no authorization passed (contrary to HTTP)
func GRPCInterceptor(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %v", err)
	}

	secret := model.AuthenticateSecretString(token)

	newCtx := context.WithValue(ctx, ContextKey{}, secret)
	return newCtx, nil
}

// HTTPMiddleware reads bearer token and injects auth into the context of an HTTP request
// Sets ContextKey to nil if no authorization passed (contrary to gRPC)
func HTTPMiddleware(next http.Handler) http.Handler {
	return httputil.AppHandler(func(w http.ResponseWriter, r *http.Request) error {
		var secret *model.Secret

		header := r.Header.Get("Authorization")
		if header != "" {
			if !strings.HasPrefix(header, "Bearer ") {
				return httputil.NewError(400, "bearer authorization header required")
			}

			token := strings.TrimSpace(header[6:])

			secret = model.AuthenticateSecretString(token)
			if secret == nil {
				return httputil.NewError(400, "unauthenticated")
			}
		}

		ctx := context.WithValue(r.Context(), ContextKey{}, secret)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}
