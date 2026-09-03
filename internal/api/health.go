package api

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// healthPingTimeout bounds the DB connectivity check per health request.
const healthPingTimeout = 2 * time.Second

// healthz reports 200 {"status":"ok"} when the database is reachable,
// 503 otherwise.
func healthz(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthPingTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := db.Ping(ctx); err != nil {
			zap.L().Warn("health check failed", zap.Error(err))
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.MarshalWrite(w, map[string]string{"status": "unavailable"})
			return
		}
		_ = json.MarshalWrite(w, map[string]string{"status": "ok"})
	}
}
