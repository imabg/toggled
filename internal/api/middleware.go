package api

import "net/http"

// Middleware wraps an http.Handler, e.g. to add authentication or logging.
type Middleware func(http.Handler) http.Handler

// chain applies middleware to a handler, with the last middleware innermost:
// chain(h, a, b) == a(b(h)). Private routes will use it once auth lands.
func chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
