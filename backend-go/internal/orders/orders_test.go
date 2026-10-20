package orders

import (
	"errors"
	"testing"
	"time"
)

type recPub struct {
	events []Event
	fails  int
}

func (r *recPub) Publish(e Event) error {
	if r.fails > 0 {
		r.fails--
		return errors.New("unavailable")
	}
	r.events = append(r.events, e)
	return nil
}

func newTest(p *recPub) *Service {
	s := NewService(p)
	s.sleep = func(time.Duration) {}
	return s
}

func TestCheckoutIsIdempotentPerTenant(t *testing.T) {
	p := &recPub{}
	s := newTest(p)
	req := CheckoutRequest{TenantID: "a", ProductID: "p1", Quantity: 10, IdempotencyKey: "k"}
	o1, created1, err := s.Checkout(req)
	if err != nil || !created1 {
		t.Fatalf("first checkout: created=%v err=%v", created1, err)
	}
	o2, created2, _ := s.Checkout(req)
	if created2 || o2.ID != o1.ID {
		t.Fatalf("repeat key should return same order")
	}
	req.TenantID = "b"
	o3, created3, _ := s.Checkout(req)
	if !created3 || o3.ID == o1.ID {
		t.Fatalf("same key in another tenant must create a new order")
	}
	if len(p.events) != 2 {
		t.Fatalf("want 2 events, got %d", len(p.events))
	}
}

func TestTenantIsolation(t *testing.T) {
	s := newTest(&recPub{})
	o, _, _ := s.Checkout(CheckoutRequest{TenantID: "a", ProductID: "p", Quantity: 1})
	if _, err := s.Get("b", o.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("tenant b must not read tenant a's order")
	}
	if _, err := s.Get("a", o.ID); err != nil {
		t.Fatal(err)
	}
}

func TestValidation(t *testing.T) {
	s := newTest(&recPub{})
	for _, r := range []CheckoutRequest{{ProductID: "p", Quantity: 1}, {TenantID: "a", Quantity: 1}, {TenantID: "a", ProductID: "p"}} {
		if _, _, err := s.Checkout(r); !errors.Is(err, ErrInvalid) {
			t.Fatalf("want ErrInvalid for %+v", r)
		}
	}
}

func TestPublishRetriesThenSucceeds(t *testing.T) {
	p := &recPub{fails: 2}
	s := newTest(p)
	if _, _, err := s.Checkout(CheckoutRequest{TenantID: "a", ProductID: "p", Quantity: 1}); err != nil {
		t.Fatalf("expected success on third attempt: %v", err)
	}
	p2 := &recPub{fails: 5}
	if _, _, err := newTest(p2).Checkout(CheckoutRequest{TenantID: "a", ProductID: "p", Quantity: 1}); err == nil {
		t.Fatal("expected error after 3 failed attempts")
	}
}

func TestBackoff(t *testing.T) {
	if Backoff(1) != 100*time.Millisecond || Backoff(10) != 2*time.Second {
		t.Fatalf("unexpected backoff values")
	}
}
