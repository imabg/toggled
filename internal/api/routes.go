package api

import "net/http"

// NewRouter returns the root HTTP handler with all routes registered.
func NewRouter(db Pinger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz(db))
	registerPrivateRoutes(mux, db)
	return mux
}

// registerPrivateRoutes registers endpoints that require authentication.
//
// TODO: no private endpoints yet. Intended pattern for flag CRUD etc.
// (auth middleware is a placeholder, not yet implemented):
//
//	mux.Handle("GET /api/v1/flags", chain(listFlags(db), authMiddleware))
func registerPrivateRoutes(mux *http.ServeMux, db Pinger) {
}
