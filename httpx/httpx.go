// Package httpx holds shared HTTP plumbing for services deployed behind the
// IAM gateway. The gateway authenticates every request and forwards the
// caller's identity in the X-User-Id header, after stripping any value the
// client tried to supply itself.
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"uuid"
)

const HeaderUserID = "X-User-Id"

var (
	// ErrNoUserID means the gateway forwarded no identity: the service is
	// reachable but the request bypassed authentication.
	ErrNoUserID = errors.New("httpx: no " + HeaderUserID + " header")
	// ErrInvalidUserID means the header is present but not a UUID, which
	// points at a gateway bug rather than a routing gap.
	ErrInvalidUserID = errors.New("httpx: malformed " + HeaderUserID + " header")
)

// UserID returns the authenticated caller's IAM identity id.
func UserID(r *http.Request) (uuid.UUID, error) {
	// TrimSpace: proxies have been seen to pad forwarded values.
	raw := strings.TrimSpace(r.Header.Get(HeaderUserID))
	if raw == "" {
		return uuid.UUID{}, ErrNoUserID
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%w: %q", ErrInvalidUserID, raw)
	}

	return id, nil
}

// RequireUserID answers 401 itself when the request carries no valid identity
// and otherwise passes the parsed id to next, so handlers never see an
// unauthenticated request.
func RequireUserID(next func(w http.ResponseWriter, r *http.Request, id uuid.UUID)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := UserID(r)
		if err != nil {
			writeUnauthorized(w)
			return
		}
		next(w, r, id)
	})
}

const errUnauthenticated = "unauthenticated"

// The same error envelope the services' own handlers write, so a client sees
// one shape of failure regardless of which layer rejected the request.
func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": errUnauthenticated}); err != nil {
		http.Error(w, errUnauthenticated, http.StatusUnauthorized)
	}
}
