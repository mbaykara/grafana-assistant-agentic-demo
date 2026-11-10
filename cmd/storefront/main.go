// Command storefront takes orders and asks checkout for the total.
package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/mbaykara/grafana-assistant-agentic-demo/internal/telemetry"
)

var version = "dev"

var client = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport), Timeout: 5 * time.Second}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdown, err := telemetry.Setup(ctx, "storefront", version)
	if err != nil {
		slog.Error("telemetry setup failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdown(context.Background()) }()

	checkoutURL := getenv("CHECKOUT_URL", "http://localhost:8081") + "/checkout"

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("POST /buy", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, checkoutURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			slog.ErrorContext(r.Context(), "checkout unreachable", "error", err.Error())
			http.Error(w, "checkout unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		switch {
		case resp.StatusCode >= 500:
			slog.ErrorContext(r.Context(), "checkout failed", "status", resp.StatusCode)
		case resp.StatusCode >= 400:
			slog.WarnContext(r.Context(), "order rejected", "status", resp.StatusCode)
		default:
			slog.InfoContext(r.Context(), "order placed", "status", resp.StatusCode)
		}
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	srv := &http.Server{Addr: ":" + getenv("PORT", "8080"), Handler: otelhttp.NewHandler(mux, "storefront"), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	slog.Info("storefront listening", "addr", srv.Addr, "version", version)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
