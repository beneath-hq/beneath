package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type authContextKey struct{}

// Auth contains information about the authorized user -- currently just the key
type Auth string

// getAuth extracts auth from a context
func getAuth(ctx context.Context) Auth {
	auth, ok := ctx.Value(authContextKey{}).(Auth)
	if !ok {
		log.Panicln("couldn't get auth from context")
	}
	return auth
}

// authInterceptor injects auth into the context of a gRPC call
func authInterceptor(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication error: %v", err)
	}

	auth, err := parseToken(token)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, err.Error())
	}

	newCtx := context.WithValue(ctx, authContextKey{}, auth)
	return newCtx, nil
}

// authMiddleware injects auth into the context of an HTTP request
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			WriteHTTPError(w, NewHTTPError(401, "authorization required"))
			return
		}

		if !strings.HasPrefix(header, "Bearer ") {
			WriteHTTPError(w, NewHTTPError(400, "bearer authorization header required"))
			return
		}

		token := header[6:]

		auth, err := parseToken(token)
		if err != nil {
			WriteHTTPError(w, NewHTTPError(400, err.Error()))
			return
		}

		ctx := context.WithValue(r.Context(), authContextKey{}, auth)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseToken(token string) (Auth, error) {
	token = strings.TrimSpace(token)
	if len(token) != 44 {
		return "", errors.New("invalid authorization token format")
	}

	hashed := sha256.Sum256([]byte(token))
	encoded := base64.StdEncoding.EncodeToString(hashed[:])

	return Auth(encoded), nil
}
