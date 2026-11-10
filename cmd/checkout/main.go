// Command checkout prices a cart.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/mbaykara/grafana-assistant-agentic-demo/internal/pricing"
	"github.com/mbaykara/grafana-assistant-agentic-demo/internal/telemetry"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdown, err := telemetry.Setup(ctx, "checkout", version)
	if err != nil {
		slog.Error("telemetry setup failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdown(context.Background()) }()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.Handle("POST /checkout", telemetry.Recover(http.HandlerFunc(handleCheckout)))

	srv := &http.Server{Addr: addr(), Handler: otelhttp.NewHandler(mux, "checkout"), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	slog.Info("checkout listening", "addr", srv.Addr, "version", version)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	var cart pricing.Cart
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		http.Error(w, "invalid cart", http.StatusBadRequest)
		return
	}
	total, err := pricing.Total(cart)
	if err != nil {
		slog.WarnContext(r.Context(), "cart rejected", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slog.InfoContext(r.Context(), "order priced",
		"items", len(cart.Items), "discount_code", cart.DiscountCode, "total_cents", total)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int64{"total_cents": total})
}

func addr() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8081"
}
