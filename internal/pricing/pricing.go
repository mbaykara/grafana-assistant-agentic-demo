// Package pricing computes order totals for the checkout service.
package pricing

import "errors"

// Item is one line in a cart.
type Item struct {
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitCents int64  `json:"unit_cents"`
}

// Cart is what the storefront sends to checkout.
type Cart struct {
	Items        []Item `json:"items"`
	DiscountCode string `json:"discount_code,omitempty"`
}

// ErrUnknownCode is returned when the discount code does not exist.
var ErrUnknownCode = errors.New("unknown discount code")

// discounts maps a code to its percentage off.
var discounts = map[string]int64{
	"CAMP10":  10,
	"STARS20": 20,
}

// Total returns the order total in cents after any discount.
func Total(c Cart) (int64, error) {
	// Receipt lines per SKU, plus one line for the discount.
	var lines map[string]int64
	for _, it := range c.Items {
		if lines == nil {
			lines = make(map[string]int64, len(c.Items)+1)
		}
		lines[it.SKU] += int64(it.Qty) * it.UnitCents
	}

	if c.DiscountCode != "" {
		pct, ok := discounts[c.DiscountCode]
		if !ok {
			return 0, ErrUnknownCode
		}
		lines["discount"] = -subtotal(lines) * pct / 100
	}

	return subtotal(lines), nil
}

func subtotal(lines map[string]int64) int64 {
	var sum int64
	for _, v := range lines {
		sum += v
	}
	return sum
}
