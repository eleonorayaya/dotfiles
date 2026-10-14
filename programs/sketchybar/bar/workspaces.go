package bar

import (
	"context"
	"log/slog"
	"strings"
)

const (
	WorkspaceChangeEvent        = "aerospace_workspace_change"
	WorkspaceMonitorChangeEvent = "aerospace_workspace_monitor_change"

	workspaceItem = "space"
	iconFlower    = "✿"
)

type Workspace struct {
	runner Runner
	colors Colors
}

func NewWorkspace(runner Runner, colors Colors) *Workspace {
	return &Workspace{runner: runner, colors: colors}
}

func (w *Workspace) Setup(ctx context.Context, b *Batch) error {
	b.Raw("--add", "event", WorkspaceChangeEvent, "--add", "event", WorkspaceMonitorChangeEvent)
	b.Add(workspaceItem, "left", append(hotIsland(w.colors),
		P("icon", iconFlower),
		P("icon.font", symbolFont+":Bold:14.0"),
		"icon.padding_right=5",
		P("label.font", font("ExtraBold", 13)),
		"display=active",
	)...)
	return nil
}

func (w *Workspace) Events() []string {
	return []string{WorkspaceChangeEvent, WorkspaceMonitorChangeEvent, "display_change", "system_woke"}
}

func (w *Workspace) Handle(ctx context.Context, ev Event, b *Batch) {
	focused := ev.Env["FOCUSED_WORKSPACE"]
	if focused == "" {
		out, err := w.runner.Run(ctx, "aerospace", "list-workspaces", "--focused")
		if err != nil {
			slog.Warn("bar: focused workspace lookup failed", "error", err)
			return
		}
		focused = strings.TrimSpace(string(out))
	}
	b.Set(workspaceItem, P("label", focused))
}
