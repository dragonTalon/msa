package renderer

import (
	"context"
	"fmt"
	"io"

	"msa/pkg/core/event"
)

// CLIRenderer writes Events to a terminal (os.Stdout in production, replaceable in tests).
type CLIRenderer struct {
	out     io.Writer
	verbose bool // when true, shows thinking content
}

// NewCLI creates a CLIRenderer.
// out: output destination (use os.Stdout in production)
// verbose: whether to display thinking/reasoning content
func NewCLI(out io.Writer, verbose bool) *CLIRenderer {
	return &CLIRenderer{out: out, verbose: verbose}
}

// Handle writes the event to the terminal.
func (r *CLIRenderer) Handle(ctx context.Context, e event.Event) error {
	switch e.Type {
	case event.EventTextChunk:
		if e.Text != "" {
			fmt.Fprint(r.out, e.Text)
		}

	case event.EventTextDone:
		fmt.Fprintln(r.out)

	case event.EventThinking:
		if r.verbose && e.Text != "" {
			fmt.Fprint(r.out, e.Text)
		}

	case event.EventToolStart:
		fmt.Fprintf(r.out, "\n⚙ 正在调用 %s...\n", e.Tool.Name)

	case event.EventToolResult:
		fmt.Fprintf(r.out, "✓ %s 完成\n", e.Result.Name)

	case event.EventToolError:
		fmt.Fprintf(r.out, "✗ %s 失败: %v\n", e.Result.Name, e.Err)

	case event.EventRoundDone:
		// conversation ended normally, no output needed

	case event.EventError:
		return e.Err // bubble up to Runner
	}
	return nil
}
