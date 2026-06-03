package logger

import (
	"context"
	"testing"
	"time"
)

// TestRunRespectsContextCancellation verifies that Run returns promptly when
// ctx is cancelled even if InfluxDB is unreachable (the health check will
// fail, but the retry loop must still exit on cancel).
func TestRunRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		// Point at a non-existent InfluxDB so the health check fails immediately.
		done <- Run(ctx, "/nonexistent/auth.json",
			"http://127.0.0.1:19999", "fake-token", "org", "bucket", 5)
	}()

	// Give the goroutine a moment to start, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// returned — good
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
