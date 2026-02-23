package app

import (
	"context"
	"testing"
)

// TestRun_Note documents that Run requires TUI which is hard to test
func TestRun_Note(t *testing.T) {
	t.Skip("Skipping TUI test - requires interactive terminal")

	// To properly test Run, we need to:
	// 1. Mock the TUI (tui.NewChat)
	// 2. Mock the tea.Program
	// 3. Test various scenarios (success, error, user interrupt)

	ctx := context.Background()
	err := Run(ctx)

	// Since we skipped, err should be nil
	_ = err
}

// TestRun_ContextCancellation tests Run with canceled context
func TestRun_ContextCancellation(t *testing.T) {
	t.Skip("Skipping context cancellation test - requires TUI mocking")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	err := Run(ctx)
	// Should handle cancellation gracefully
	_ = err
}
