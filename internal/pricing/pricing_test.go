package pricing

import "testing"

func TestTotal_NoDiscount(t *testing.T) {
	got, err := Total(Cart{Items: []Item{{SKU: "a", Qty: 2, UnitCents: 500}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1000 {
		t.Fatalf("got %d, want 1000", got)
	}
}

func TestTotal_EmptyCart(t *testing.T) {
	got, err := Total(Cart{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
}

func TestTotal_Discount10(t *testing.T) {
	got, err := Total(Cart{Items: []Item{{SKU: "a", Qty: 1, UnitCents: 1000}}, DiscountCode: "CAMP10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 900 {
		t.Fatalf("got %d, want 900", got)
	}
}

func TestTotal_UnknownCode(t *testing.T) {
	_, err := Total(Cart{Items: []Item{{SKU: "a", Qty: 1, UnitCents: 1000}}, DiscountCode: "NOPE"})
	if err != ErrUnknownCode {
		t.Fatalf("got %v, want ErrUnknownCode", err)
	}
}
