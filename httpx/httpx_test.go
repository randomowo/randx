package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"uuid"
)

func TestUserIDParsesGatewayHeader(t *testing.T) {
	id := uuid.MustParse("0197c0de-0000-7000-8000-000000000001")

	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set(HeaderUserID, id.String())

	got, err := UserID(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != id {
		t.Fatalf("got %s, want %s", got, id)
	}
}

func TestUserIDAcceptsPaddedValue(t *testing.T) {
	id := uuid.MustParse("0197c0de-0000-7000-8000-000000000002")

	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set(HeaderUserID, "  "+id.String()+" ")

	got, err := UserID(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != id {
		t.Fatalf("got %s, want %s", got, id)
	}
}

func TestUserIDRejectsMissingAndMalformed(t *testing.T) {
	cases := map[string]string{
		"missing":   "",
		"empty":     "   ",
		"garbage":   "not-a-uuid",
		"truncated": "0197c0de-0000-7000-8000",
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			if value != "" {
				r.Header.Set(HeaderUserID, value)
			}

			_, err := UserID(r)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestUserIDSentinels(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	if _, err := UserID(r); !errors.Is(err, ErrNoUserID) {
		t.Fatalf("missing header: got %v, want ErrNoUserID", err)
	}

	r.Header.Set(HeaderUserID, "not-a-uuid")
	if _, err := UserID(r); !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("malformed header: got %v, want ErrInvalidUserID", err)
	}
}

func TestRequireUserIDPassesIdentityThrough(t *testing.T) {
	id := uuid.MustParse("0197c0de-0000-7000-8000-000000000003")
	var got uuid.UUID
	h := RequireUserID(func(w http.ResponseWriter, r *http.Request, uid uuid.UUID) {
		got = uid
		w.WriteHeader(http.StatusNoContent)
	})

	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set(HeaderUserID, id.String())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusNoContent)
	}
	if got != id {
		t.Fatalf("handler saw %s, want %s", got, id)
	}
}

func TestRequireUserIDRejectsUnauthenticated(t *testing.T) {
	cases := map[string]string{
		"missing":   "",
		"malformed": "not-a-uuid",
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			h := RequireUserID(func(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
				t.Error("handler must not run")
			})

			r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			if value != "" {
				r.Header.Set(HeaderUserID, value)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("got status %d, want %d", w.Code, http.StatusUnauthorized)
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decoding body %q: %v", w.Body.String(), err)
			}
			if body.Error != "unauthenticated" {
				t.Fatalf("got error %q, want %q", body.Error, "unauthenticated")
			}
		})
	}
}
