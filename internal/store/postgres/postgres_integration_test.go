package postgres_test

import (
	"os"
	"testing"

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

func TestNewPoolIntegration(t *testing.T) {
	t.Parallel()

	t.Run("connects and pings a real database", func(t *testing.T) {
		t.Parallel()

		pool, err := postgres.NewPool(t.Context(), databaseURL(t))
		if err != nil {
			t.Fatalf("NewPool() unexpected error: %v", err)
		}
		t.Cleanup(pool.Close)

		if err := pool.Ping(t.Context()); err != nil {
			t.Fatalf("Ping() unexpected error: %v", err)
		}
	})

	t.Run("executes a query against a real database", func(t *testing.T) {
		t.Parallel()

		pool, err := postgres.NewPool(t.Context(), databaseURL(t))
		if err != nil {
			t.Fatalf("NewPool() unexpected error: %v", err)
		}
		t.Cleanup(pool.Close)

		var one int
		if err := pool.QueryRow(t.Context(), "select 1").Scan(&one); err != nil {
			t.Fatalf("QueryRow() unexpected error: %v", err)
		}
		if one != 1 {
			t.Errorf("select 1 = %d, want 1", one)
		}
	})

	t.Run("fails on unreachable database", func(t *testing.T) {
		t.Parallel()

		// Port 1 is never listening; the ping must fail and the pool must
		// be closed rather than leaked.
		_, err := postgres.NewPool(t.Context(), "postgres://toggled:toggled@127.0.0.1:1/toggled?sslmode=disable")
		if err == nil {
			t.Fatal("NewPool() error = nil, want ping error")
		}
	})

	t.Run("fails on malformed url", func(t *testing.T) {
		t.Parallel()

		_, err := postgres.NewPool(t.Context(), "not-a-postgres-url")
		if err == nil {
			t.Fatal("NewPool() error = nil, want parse error")
		}
	})
}
