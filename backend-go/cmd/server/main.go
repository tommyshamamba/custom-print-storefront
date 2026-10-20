// Command server exposes the storefront checkout API over HTTP.
package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/tommyshamamba/custom-print-storefront/backend-go/internal/orders"
)

type logPublisher struct{ log *slog.Logger }

func (p logPublisher) Publish(e orders.Event) error {
	p.log.Info("event", "topic", e.Topic, "order", e.Order.ID, "tenant", e.Order.TenantID)
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func newMux(svc *orders.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/checkout", func(w http.ResponseWriter, r *http.Request) {
		var req orders.CheckoutRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		o, created, err := svc.Checkout(req)
		switch {
		case errors.Is(err, orders.ErrInvalid):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		case err != nil:
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "event publish failed"})
		case created:
			writeJSON(w, http.StatusCreated, o)
		default:
			writeJSON(w, http.StatusOK, o)
		}
	})
	mux.HandleFunc("GET /api/tenants/{tenant}/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		o, err := svc.Get(r.PathValue("tenant"), r.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, o)
	})
	return mux
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           newMux(orders.NewService(logPublisher{log})),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
