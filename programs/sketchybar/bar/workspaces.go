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

type aerospaceWorkspace struct {
	Workspace string `json:"workspace"`
	Display   int    `json:"monitor-appkit-nsscreen-screens-id"`
	Visible   bool   `json:"workspace-is-visible"`
}

type aerospaceWindow struct {
	Workspace string `json:"workspace"`
	AppName   string `json:"app-name"`
}

func addWorkspaceEvents(b *Batch) {
	b.Raw("--add", "event", WorkspaceChangeEvent, "--add", "event", WorkspaceMonitorChangeEvent)
}

func listWorkspaces(ctx context.Context, r Runner) ([]aerospaceWorkspace, error) {
	out, err := r.Run(ctx, "aerospace", "list-workspaces", "--all", "--json",
		"--format", "%{workspace} %{monitor-appkit-nsscreen-screens-id} %{workspace-is-visible}")
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]aerospaceWorkspace](out, "aerospace workspaces")
}

func listWindows(ctx context.Context, r Runner) ([]aerospaceWindow, error) {
	out, err := r.Run(ctx, "aerospace", "list-windows", "--all", "--json", "--format", "%{workspace} %{app-name}")
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]aerospaceWindow](out, "aerospace windows")
}

type Workspace struct {
	runner Runner
	colors Colors
	last   string
}

func NewWorkspace(runner Runner, colors Colors) *Workspace {
	return &Workspace{runner: runner, colors: colors}
}

func (w *Workspace) Setup(ctx context.Context, b *Batch) error {
	addWorkspaceEvents(b)
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
	if w.last != "" && focused != w.last {
		b.Bounce(workspaceItem, 5)
	}
	w.last = focused
}
