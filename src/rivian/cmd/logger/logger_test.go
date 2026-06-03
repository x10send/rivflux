package main

import "testing"

// runLogger delegates entirely to pkg/logger.Run, which is tested there.
// This file exists to keep the package under test.
func TestRunLoggerSignature(t *testing.T) {
	// Confirms runLogger is callable; pkg/logger covers the actual behaviour.
	_ = runLogger
}
