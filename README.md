# Custom Print Storefront

[![CI](https://github.com/tommyshamamba/custom-print-storefront/actions/workflows/ci.yml/badge.svg)](https://github.com/tommyshamamba/custom-print-storefront/actions/workflows/ci.yml)

A multi-tenant checkout API in Go with a small React storefront. It covers idempotent order
creation, tenant isolation, and event publishing with retries.

> Portfolio project. Orders are kept in memory and no payment provider is connected.
> It claims no production use or traffic figures.

## What it does

- `POST /api/checkout` validates `tenant_id`, `product_id` and `quantity` (1–10,000) and creates an order.
  A repeated `idempotency_key` within the same tenant returns the original order with `200` instead of `201`.
- Orders are stored per tenant, so `GET /api/tenants/{tenant}/orders/{id}` never returns another tenant's order.
- Each new order emits an `orders.created` event through a `Publisher` interface. Publishing is
  retried up to 3 times with exponential backoff (50 ms, 100 ms, 200 ms, capped at 2 s). If every
  attempt fails, the API returns `503`.
- `GET /healthz` is used by the Kubernetes probes. Logs are structured JSON (`log/slog`).

## Architecture

```
frontend-react (Vite + React) ──/api──► backend-go/cmd/server (net/http, Go 1.22 routing)
                                              │
                                              ▼
                                  internal/orders.Service
                                  ├─ per-tenant order map + idempotency index (mutex)
                                  └─ Publisher interface ─► log publisher (swap for Redis/queue)
```

`infra/k8s/deployment.yaml` deploys the API with readiness/liveness probes and resource limits.
The Dockerfile builds a static binary into a distroless non-root image.

## Run

```sh
go run ./backend-go/cmd/server                       # API on :8080
cd frontend-react && npm install && npm run dev      # storefront, proxies /api to :8080
docker build -t storefront-api . && docker run -p 8080:8080 storefront-api
```

## Tests

```sh
go vet ./... && go test -race ./...
```

Unit tests cover idempotency across tenants, tenant isolation, validation, publish retry and
backoff. An HTTP test checks the 201/200/422 responses. CI runs gofmt, vet, race-enabled tests,
a Docker build and the frontend type-check/build.

## License

MIT
