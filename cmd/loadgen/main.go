// Command loadgen sends a steady stream of orders to the storefront.
package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mbaykara/grafana-assistant-agentic-demo/internal/pricing"
)

func main() {
	target := getenv("STOREFRONT_URL", "http://localhost:8080") + "/buy"
	rps, _ := strconv.Atoi(getenv("RPS", "5"))
	if rps <= 0 {
		rps = 5
	}
	client := &http.Client{Timeout: 5 * time.Second}
	tick := time.NewTicker(time.Second / time.Duration(rps))
	defer tick.Stop()

	slog.Info("loadgen started", "target", target, "rps", rps)
	for range tick.C {
		go func() {
			body, _ := json.Marshal(randomCart())
			resp, err := client.Post(target, "application/json", bytes.NewReader(body))
			if err != nil {
				slog.Warn("request failed", "error", err.Error())
				return
			}
			_ = resp.Body.Close()
		}()
	}
}

var skus = []pricing.Item{
	{SKU: "telescope", UnitCents: 49900},
	{SKU: "lens", UnitCents: 12900},
	{SKU: "tripod", UnitCents: 7900},
	{SKU: "star-map", UnitCents: 1500},
}

// randomCart returns a mix of orders: about 20% are "apply code first" carts
// with a discount code and no items yet, 30% are normal carts with a code.
func randomCart() pricing.Cart {
	var c pricing.Cart
	switch r := rand.IntN(100); {
	case r < 20:
		c.DiscountCode = "CAMP10"
		return c
	case r < 50:
		c.DiscountCode = []string{"CAMP10", "STARS20"}[rand.IntN(2)]
	}
	for i := 0; i < 1+rand.IntN(3); i++ {
		it := skus[rand.IntN(len(skus))]
		it.Qty = 1 + rand.IntN(2)
		c.Items = append(c.Items, it)
	}
	return c
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
