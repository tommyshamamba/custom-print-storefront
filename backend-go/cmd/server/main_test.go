package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tommyshamamba/custom-print-storefront/backend-go/internal/orders"
)

func TestCheckoutHTTP(t *testing.T) {
	srv := httptest.NewServer(newMux(orders.NewService(logPublisher{slog.New(slog.NewTextHandler(io.Discard, nil))})))
	defer srv.Close()
	body := `{"tenant_id":"a","product_id":"p","quantity":5,"idempotency_key":"k"}`
	r1, _ := http.Post(srv.URL+"/api/checkout", "application/json", strings.NewReader(body))
	r2, _ := http.Post(srv.URL+"/api/checkout", "application/json", strings.NewReader(body))
	if r1.StatusCode != 201 || r2.StatusCode != 200 {
		t.Fatalf("want 201 then 200, got %d %d", r1.StatusCode, r2.StatusCode)
	}
	bad, _ := http.Post(srv.URL+"/api/checkout", "application/json", strings.NewReader(`{"tenant_id":"a"}`))
	if bad.StatusCode != 422 {
		t.Fatalf("want 422, got %d", bad.StatusCode)
	}
	h, _ := http.Get(srv.URL + "/healthz")
	if h.StatusCode != 200 {
		t.Fatalf("healthz %d", h.StatusCode)
	}
}
