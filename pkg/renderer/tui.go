package renderer

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"msa/pkg/core/event"
)

// TUIRenderer forwards Events to the Bubble Tea program.
// The TUI model decides how to render each event type.
type TUIRenderer struct {
	send func(tea.Msg)
}

// NewTUI creates a TUIRenderer.
// send: the bubbletea program.Send function
func NewTUI(send func(tea.Msg)) *TUIRenderer {
	return &TUIRenderer{send: send}
}

// Handle forwards the event to the Bubble Tea event loop.
func (r *TUIRenderer) Handle(ctx context.Context, e event.Event) error {
	r.send(e)
	return nil
}
