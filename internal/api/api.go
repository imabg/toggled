// Package api contains the HTTP router, handlers, and middleware.
package api

import "context"

// Pinger verifies that a dependency is reachable.
// *pgxpool.Pool satisfies this interface.
type Pinger interface {
	Ping(ctx context.Context) error
}
