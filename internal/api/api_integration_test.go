package api_test

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/imabg/toggled/internal/api"
	"github.com/imabg/toggled/internal/store/postgres"
)

// databaseURL returns the Postgres DSN used by integration tests. Tests
// skip unless TOGGLED_TEST_DATABASE_URL points at a real database, e.g.:
//
//	TOGGLED_TEST_DATABASE_URL=postgres://toggled:toggled@localhost:5432/toggled?sslmode=disable go test ./...
func databaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TOGGLED_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TOGGLED_TEST_DATABASE_URL not set; skipping integration test")
	}
	return url
}

// TestHealthzIntegration exercises the full user flow: a real Postgres
// pool wired into the router, served over real HTTP.
func TestHealthzIntegration(t *testing.T) {
	t.Parallel()

	pool, err := postgres.NewPool(t.Context(), databaseURL(t))
	if err != nil {
		t.Fatalf("NewPool() unexpected error: %v", err)
	}
	t.Cleanup(pool.Close)

	srv := httptest.NewServer(api.NewRouter(pool))
	t.Cleanup(srv.Close)

	t.Run("healthz returns 200 ok over real HTTP", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/healthz", nil)
		if err != nil {
			t.Fatalf("NewRequest() unexpected error: %v", err)
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("GET /healthz unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if payload["status"] != "ok" {
			t.Errorf("status = %q, want %q", payload["status"], "ok")
		}
	})

	t.Run("unknown route returns 404 over real HTTP", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/nope", nil)
		if err != nil {
			t.Fatalf("NewRequest() unexpected error: %v", err)
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("GET /nope unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}
