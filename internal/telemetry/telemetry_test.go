package telemetry

import (
	"context"
	"testing"
)

// Setup must succeed with exporting enabled. Exporters connect lazily, so no
// collector is needed for this test.
func TestSetup_ExportEnabled(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=test")

	shutdown, err := Setup(context.Background(), "test-svc", "v0")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // do not wait for exports to the fake endpoint
	_ = shutdown(ctx)
}
