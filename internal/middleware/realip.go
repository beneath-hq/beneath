package middleware

// Based on https://github.com/go-chi/chi/blob/master/middleware/realip.go

import (
	"net/http"
	"strings"

	"gitlab.com/beneath-hq/beneath/pkg/log"
)

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results
// of parsing the X-Forwarded-For header (only use behind a load balancer!).
func RealIP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if xff := r.Header.Get(xForwardedFor); xff != "" {
			i := strings.Index(xff, ", ")
			if i == -1 {
				i = len(xff)
			}
			ip := xff[:i]
			log.S.Infow("full_real_ip", "ip", xff)
			r.RemoteAddr = ip
		}
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
