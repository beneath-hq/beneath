package gateway

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
)

// parseAuthKey returns the (hashed and formatted) key passed in the Authorization header
func parseAuth(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", NewHTTPError(401, "authorization required")
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return "", NewHTTPError(400, "bearer authorization header required")
	}

	token := header[7:]
	if len(token) != 44 {
		return "", NewHTTPError(400, "invalid authorization token format")
	}

	hashed := sha256.Sum256([]byte(token))
	encoded := base64.StdEncoding.EncodeToString(hashed[:])

	return encoded, nil
}
