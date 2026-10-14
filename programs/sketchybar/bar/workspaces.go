package bar

import (
	"context"
	"log/slog"
)

const (
	WorkspaceChangeEvent        = "aerospace_workspace_change"
	WorkspaceMonitorChangeEvent = "aerospace_workspace_monitor_change"
)

var workspaceIcons = map[string]string{
	"1": "",
	"2": "",
	"3": "\U000F066F",
	"5": "\U000F0EBF",
	"7": "",
	"8": "\U000F0D8E",
	"T": "",
}

type aerospaceWorkspace struct {
	Workspace string `json:"workspace"`
	Display   int    `json:"monitor-appkit-nsscreen-screens-id"`
	Focused   bool   `json:"workspace-is-focused"`
	Visible   bool   `json:"workspace-is-visible"`
}

type aerospaceWindow struct {
	Workspace string `json:"workspace"`
}

type Workspaces struct {
	runner Runner
	icons  map[string]string
}

func NewWorkspaces(runner Runner) *Workspaces {
	return &Workspaces{runner: runner, icons: workspaceIcons}
}

func workspaceItem(id string) string {
	return "space." + id
}

func (w *Workspaces) Setup(ctx context.Context, b *Batch) error {
	spaces, err := w.listWorkspaces(ctx)
	if err != nil {
		return err
	}
	b.Raw("--add", "event", WorkspaceChangeEvent, "--add", "event", WorkspaceMonitorChangeEvent)
	for _, s := range spaces {
		props := []string{
			P("click_script", "aerospace workspace "+s.Workspace),
			"icon.padding_left=8",
			"icon.padding_right=8",
			"label.padding_right=8",
		}
		if icon, ok := w.icons[s.Workspace]; ok {
			props = append(props, P("icon", icon), "label.drawing=off")
		} else {
			props = append(props, "icon.drawing=off", P("label", s.Workspace), "label.padding_left=8")
		}
		b.Add(workspaceItem(s.Workspace), "left", props...)
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
	out, err := w.runner.Run(ctx, "aerospace", "list-windows", "--all", "--json", "--format", "%{workspace}")
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
	occupied := map[string]bool{}
	for _, win := range windows {
		occupied[win.Workspace] = true
	}
	for _, s := range spaces {
		isFocused := s.Focused
		if focused != "" {
			isFocused = s.Workspace == focused
		}
		props := []string{
			onOff("drawing", isFocused || s.Visible || occupied[s.Workspace]),
			P("display", s.Display),
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
