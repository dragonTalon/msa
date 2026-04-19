// Package renderer provides output adapters for the MSA pipeline.
// CLI and TUI each implement the Renderer interface, allowing the Runner to be UI-agnostic.
package renderer

import (
	"context"

	"msa/pkg/core/event"
)

// Renderer defines the output adapter interface.
// CLI and TUI each implement this; Runner uses it to decouple from UI.
type Renderer interface {
	Handle(ctx context.Context, e event.Event) error
}
