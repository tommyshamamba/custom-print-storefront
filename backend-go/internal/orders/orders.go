// Package orders implements idempotent checkout for a multi-tenant print storefront.
package orders

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// Status of an order.
type Status string

const (
	StatusPending Status = "pending_payment"
)

// Order is a checkout result scoped to one tenant.
type Order struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CheckoutRequest is the client payload.
type CheckoutRequest struct {
	TenantID       string `json:"tenant_id"`
	ProductID      string `json:"product_id"`
	Quantity       int    `json:"quantity"`
	IdempotencyKey string `json:"idempotency_key"`
}

// Event is emitted after a new order is created.
type Event struct {
	Topic string
	Order Order
}

// Publisher delivers events, e.g. to Redis pub/sub or a queue.
type Publisher interface {
	Publish(Event) error
}

var (
	ErrInvalid  = errors.New("tenant_id, product_id and a quantity of 1-10000 are required")
	ErrNotFound = errors.New("order not found")
)

// Service stores orders in memory, keyed per tenant so tenants never see each other's data.
type Service struct {
	mu       sync.Mutex
	orders   map[string]map[string]Order // tenant -> id -> order
	idem     map[string]string           // tenant|key -> id
	pub      Publisher
	now      func() time.Time
	attempts int
	backoff  func(int) time.Duration
	sleep    func(time.Duration)
}

// NewService returns a Service that publishes through pub with up to 3 attempts.
func NewService(pub Publisher) *Service {
	return &Service{
		orders: map[string]map[string]Order{}, idem: map[string]string{},
		pub: pub, now: time.Now, attempts: 3, backoff: Backoff, sleep: time.Sleep,
	}
}

// Backoff returns an exponential delay (50ms, 100ms, 200ms, ...) capped at 2s.
func Backoff(attempt int) time.Duration {
	d := 50 * time.Millisecond << attempt
	if d > 2*time.Second || d <= 0 {
		return 2 * time.Second
	}
	return d
}

// Checkout creates an order, or returns the existing one for a repeated idempotency key.
// The bool reports whether a new order was created.
func (s *Service) Checkout(req CheckoutRequest) (Order, bool, error) {
	if req.TenantID == "" || req.ProductID == "" || req.Quantity < 1 || req.Quantity > 10000 {
		return Order{}, false, ErrInvalid
	}
	s.mu.Lock()
	if req.IdempotencyKey != "" {
		if id, ok := s.idem[req.TenantID+"|"+req.IdempotencyKey]; ok {
			o := s.orders[req.TenantID][id]
			s.mu.Unlock()
			return o, false, nil
		}
	}
	o := Order{ID: newID(), TenantID: req.TenantID, ProductID: req.ProductID,
		Quantity: req.Quantity, Status: StatusPending, CreatedAt: s.now().UTC()}
	if s.orders[o.TenantID] == nil {
		s.orders[o.TenantID] = map[string]Order{}
	}
	s.orders[o.TenantID][o.ID] = o
	if req.IdempotencyKey != "" {
		s.idem[req.TenantID+"|"+req.IdempotencyKey] = o.ID
	}
	s.mu.Unlock()
	return o, true, s.publish(Event{Topic: "orders.created", Order: o})
}

func (s *Service) publish(e Event) error {
	var err error
	for i := 0; i < s.attempts; i++ {
		if i > 0 {
			s.sleep(s.backoff(i))
		}
		if err = s.pub.Publish(e); err == nil {
			return nil
		}
	}
	return err
}

// Get returns one tenant's order.
func (s *Service) Get(tenantID, id string) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[tenantID][id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return o, nil
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
