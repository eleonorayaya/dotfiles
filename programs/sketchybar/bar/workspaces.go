package bar

import (
	"context"
	"log/slog"
	"slices"
	"strings"
)

const (
	WorkspaceChangeEvent        = "aerospace_workspace_change"
	WorkspaceMonitorChangeEvent = "aerospace_workspace_monitor_change"
)

type aerospaceWorkspace struct {
	Workspace string `json:"workspace"`
	Display   int    `json:"monitor-appkit-nsscreen-screens-id"`
	Focused   bool   `json:"workspace-is-focused"`
	Visible   bool   `json:"workspace-is-visible"`
}

type aerospaceWindow struct {
	Workspace string `json:"workspace"`
	AppName   string `json:"app-name"`
}

type Workspaces struct {
	runner Runner
}

func NewWorkspaces(runner Runner) *Workspaces {
	return &Workspaces{runner: runner}
}

func workspaceItem(id string) string {
	return "space." + id
}

func (w *Workspaces) Setup(ctx context.Context, b *Batch) error {
	b.Raw("--add", "event", WorkspaceChangeEvent, "--add", "event", WorkspaceMonitorChangeEvent)
	spaces, err := w.listWorkspaces(ctx)
	if err != nil {
		slog.Warn("bar: list workspaces failed, retrying on next update", "error", err)
		return nil
	}
	for _, s := range spaces {
		b.Add(workspaceItem(s.Workspace), "left",
			P("click_script", "aerospace workspace "+s.Workspace),
			P("icon", s.Workspace),
			"icon.padding_left=8",
			P("label.font", appFont),
			"label.padding_right=8",
		)
	}
	return nil
}

func (w *Workspaces) Events() []string {
	return []string{
		WorkspaceChangeEvent,
		WorkspaceMonitorChangeEvent,
		"front_app_switched",
		"space_windows_change",
		"display_change",
		"system_woke",
	}
}

func (w *Workspaces) Handle(ctx context.Context, ev Event, b *Batch) {
	spaces, err := w.listWorkspaces(ctx)
	if err != nil {
		slog.Warn("bar: list workspaces failed", "error", err)
		return
	}
	out, err := w.runner.Run(ctx, "aerospace", "list-windows", "--all", "--json", "--format", "%{workspace} %{app-name}")
	if err != nil {
		slog.Warn("bar: list windows failed", "error", err)
		return
	}
	windows, err := decodeJSON[[]aerospaceWindow](out, "aerospace windows")
	if err != nil {
		slog.Warn("bar: list windows failed", "error", err)
		return
	}
	renderWorkspaces(b, spaces, windows, ev.Env["FOCUSED_WORKSPACE"])
}

func renderWorkspaces(b *Batch, spaces []aerospaceWorkspace, windows []aerospaceWindow, focused string) {
	apps := map[string][]string{}
	for _, win := range windows {
		icon := appIcon(win.AppName)
		if !slices.Contains(apps[win.Workspace], icon) {
			apps[win.Workspace] = append(apps[win.Workspace], icon)
		}
	}
	for _, s := range spaces {
		isFocused := s.Focused
		if focused != "" {
			isFocused = s.Workspace == focused
		}
		label := strings.Join(apps[s.Workspace], " ")
		props := []string{
			onOff("drawing", isFocused || s.Visible || label != ""),
			P("display", s.Display),
			P("label", label),
			onOff("label.drawing", label != ""),
		}
		b.Set(workspaceItem(s.Workspace), append(props, highlight(isFocused)...)...)
	}
}

func (w *Workspaces) listWorkspaces(ctx context.Context) ([]aerospaceWorkspace, error) {
	out, err := w.runner.Run(ctx, "aerospace", "list-workspaces", "--all", "--json",
		"--format", "%{workspace} %{monitor-appkit-nsscreen-screens-id} %{workspace-is-focused} %{workspace-is-visible}")
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]aerospaceWorkspace](out, "aerospace workspaces")
}
