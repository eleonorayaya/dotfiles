package bar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	icons  map[string]string
}

func NewWorkspaces(runner Runner, icons map[string]string) *Workspaces {
	return &Workspaces{runner: runner, icons: icons}
}

func workspaceItem(id string) string {
	return "space." + id
}

func (w *Workspaces) Setup(ctx context.Context, b *Batch) error {
	spaces, err := w.listWorkspaces(ctx)
	if err != nil {
		return err
	}
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
		b.Add("item", workspaceItem(s.Workspace), "left", props...)
	}
	return nil
}

func (w *Workspaces) Events() []string {
	return []string{
		"aerospace_workspace_change",
		"aerospace_workspace_monitor_change",
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
	windows, err := w.listWindows(ctx)
	if err != nil {
		slog.Warn("bar: list windows failed", "error", err)
		return
	}
	focused := ev.Env["FOCUSED_WORKSPACE"]
	renderWorkspaces(b, spaces, windows, focused)
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
		b.Set(workspaceItem(s.Workspace),
			onOff("drawing", isFocused || s.Visible || occupied[s.Workspace]),
			P("display", s.Display),
			onOff("icon.highlight", isFocused),
			onOff("label.highlight", isFocused),
		)
	}
}

func (w *Workspaces) listWorkspaces(ctx context.Context) ([]aerospaceWorkspace, error) {
	out, err := w.runner.Run(ctx, "aerospace", "list-workspaces", "--all", "--json",
		"--format", "%{workspace} %{monitor-appkit-nsscreen-screens-id} %{workspace-is-focused} %{workspace-is-visible}")
	if err != nil {
		return nil, err
	}
	return parseAerospaceWorkspaces(out)
}

func (w *Workspaces) listWindows(ctx context.Context) ([]aerospaceWindow, error) {
	out, err := w.runner.Run(ctx, "aerospace", "list-windows", "--all", "--json", "--format", "%{workspace} %{app-name}")
	if err != nil {
		return nil, err
	}
	return parseAerospaceWindows(out)
}

func parseAerospaceWorkspaces(data []byte) ([]aerospaceWorkspace, error) {
	var spaces []aerospaceWorkspace
	if err := json.Unmarshal(data, &spaces); err != nil {
		return nil, fmt.Errorf("failed to parse aerospace workspaces: %w", err)
	}
	return spaces, nil
}

func parseAerospaceWindows(data []byte) ([]aerospaceWindow, error) {
	var windows []aerospaceWindow
	if err := json.Unmarshal(data, &windows); err != nil {
		return nil, fmt.Errorf("failed to parse aerospace windows: %w", err)
	}
	return windows, nil
}
