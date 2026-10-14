package bar

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	sessionSlots  = 8
	sessionPrefix = "sess."
	tilePrefix    = "tile."
	tilesBracket  = "spaces"

	pipWorking = "●"
	pipIdle    = "○"
	pipNeedsMe = "✦"
)

var liveSessionStatuses = []string{"active", "creating", "pending", "completed"}

type claudeAgent struct {
	ID     uint   `json:"ID"`
	Status string `json:"status"`
}

func pip(status string) string {
	switch status {
	case "working":
		return pipWorking
	case "needs_attention", "ready_for_review":
		return pipNeedsMe
	default:
		return pipIdle
	}
}

type sessionChip struct {
	id       uint
	label    string
	attached bool
	needsMe  bool
}

func sessionChips(sessions []utenaSession) []sessionChip {
	var chips []sessionChip
	for _, s := range sessions {
		if !slices.Contains(liveSessionStatuses, s.Status) {
			continue
		}
		var pips strings.Builder
		needsMe := false
		for _, a := range s.ClaudeSessions {
			p := pip(a.Status)
			needsMe = needsMe || p == pipNeedsMe
			pips.WriteString(p)
		}
		label := s.Name
		if pips.Len() > 0 {
			label += " " + pips.String()
		}
		chips = append(chips, sessionChip{id: s.ID, label: label, attached: s.IsAttached, needsMe: needsMe})
	}
	slices.SortFunc(chips, func(a, b sessionChip) int { return int(a.id) - int(b.id) })
	if len(chips) > sessionSlots {
		chips = chips[:sessionSlots]
	}
	return chips
}

type Sessions struct {
	utena  *Utena
	colors Colors
	shelf  *Shelf
	mu     sync.Mutex
	ids    []uint
	agents map[uint]string
	last   lastValue
}

func NewSessions(baseURL string, colors Colors, shelf *Shelf) *Sessions {
	return &Sessions{utena: NewUtena(baseURL, colors), colors: colors, shelf: shelf, agents: map[uint]string{}}
}

func sessionItem(i int) string { return fmt.Sprintf("%s%d", sessionPrefix, i) }

func (s *Sessions) Setup(ctx context.Context, b *Batch) error {
	for i := range sessionSlots {
		b.Add(sessionItem(i), "left", append(island(s.colors), "drawing=off", "icon.drawing=off", P("label.padding_left", islandPadding))...)
		b.OnClick(sessionItem(i))
	}
	return nil
}

func (s *Sessions) Interval() time.Duration {
	return 3 * time.Second
}

func (s *Sessions) Poll(ctx context.Context) Update {
	var body struct {
		Sessions []utenaSession `json:"sessions"`
	}
	if err := s.utena.get(ctx, "/sessions", &body); err != nil {
		slog.Debug("bar: utena sessions unavailable", "error", err)
		body.Sessions = nil
	}
	chips := sessionChips(body.Sessions)
	s.trackAgents(body.Sessions)

	s.mu.Lock()
	s.ids = s.ids[:0]
	for _, c := range chips {
		s.ids = append(s.ids, c.id)
	}
	s.mu.Unlock()

	key := fmt.Sprint(chips)
	if !s.last.changed(key) {
		return nil
	}
	return func(b *Batch) {
		renderSessions(b, s.colors, chips)
	}
}

func (s *Sessions) trackAgents(sessions []utenaSession) {
	needsMe, finished := false, false
	seen := map[uint]string{}
	for _, sess := range sessions {
		if !slices.Contains(liveSessionStatuses, sess.Status) {
			continue
		}
		for _, a := range sess.ClaudeSessions {
			seen[a.ID] = a.Status
			needsMe = needsMe || a.Status == "needs_attention"
			if s.agents[a.ID] == "working" && a.Status != "working" {
				finished = true
			}
		}
	}
	s.agents = seen
	s.shelf.update(needsMe, finished, time.Now())
}

func renderSessions(b *Batch, c Colors, chips []sessionChip) {
	for i := range sessionSlots {
		if i >= len(chips) {
			b.Set(sessionItem(i), "drawing=off")
			continue
		}
		chip := chips[i]
		props := island(c)
		switch {
		case chip.attached:
			props = hotIsland(c)
		case chip.needsMe:
			props = append(props, P("background.color", c.IslandBorder), P("background.border_color", c.Hot))
		}
		b.Set(sessionItem(i), append(props, "drawing=on", "icon.drawing=off", P("label.padding_left", islandPadding), P("label", chip.label))...)
	}
}

func (s *Sessions) Prefix() string {
	return sessionPrefix
}

func (s *Sessions) Click(ctx context.Context, ev Event, b *Batch) {
	i, err := strconv.Atoi(strings.TrimPrefix(ev.Name, sessionPrefix))
	if err != nil {
		return
	}
	s.mu.Lock()
	if i < 0 || i >= len(s.ids) {
		s.mu.Unlock()
		return
	}
	id := s.ids[i]
	s.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/sessions/%d/activate", s.utena.baseURL, id), nil)
	if err != nil {
		return
	}
	resp, err := s.utena.client.Do(req)
	if err != nil {
		slog.Warn("bar: activate session failed", "id", id, "error", err)
		return
	}
	resp.Body.Close()
}

func (s *Sessions) resetCache() {
	s.last.reset()
}

type SpaceTiles struct {
	runner Runner
	colors Colors
}

func NewSpaceTiles(runner Runner, colors Colors) *SpaceTiles {
	return &SpaceTiles{runner: runner, colors: colors}
}

func tileItem(ws string) string { return tilePrefix + ws }

func (t *SpaceTiles) Setup(ctx context.Context, b *Batch) error {
	b.Raw("--add", "event", WorkspaceChangeEvent, "--add", "event", WorkspaceMonitorChangeEvent)
	spaces, err := listWorkspaces(ctx, t.runner)
	if err != nil {
		slog.Warn("bar: list workspaces failed, retrying on next update", "error", err)
		return nil
	}
	slices.Reverse(spaces)
	members := []string{"--add", "bracket", tilesBracket}
	for i, s := range spaces {
		left, right := 2, 2
		if i == 0 {
			right = itemPadding + 4
		}
		if i == len(spaces)-1 {
			left = itemPadding + 4
		}
		members = append(members, tileItem(s.Workspace))
		b.Add(tileItem(s.Workspace), "right",
			"drawing=off",
			P("icon", s.Workspace),
			P("icon.font", font("ExtraBold", 12)),
			"icon.padding_left=10",
			"icon.padding_right=4",
			P("label.font", appFont),
			"label.padding_right=10",
			"background.drawing=on",
			"background.color=0x00000000",
			"background.height=24",
			"background.corner_radius=12",
			P("padding_left", left),
			P("padding_right", right),
			P("click_script", "aerospace workspace "+s.Workspace),
		)
	}
	b.Raw(members...)
	b.Set(tilesBracket, bracketIsland(t.colors)...)
	return nil
}

func (t *SpaceTiles) Events() []string {
	return []string{WorkspaceChangeEvent, WorkspaceMonitorChangeEvent, "front_app_switched", "space_windows_change", "display_change", "system_woke"}
}

func (t *SpaceTiles) Handle(ctx context.Context, ev Event, b *Batch) {
	spaces, err := listWorkspaces(ctx, t.runner)
	if err != nil {
		slog.Warn("bar: list workspaces failed", "error", err)
		return
	}
	out, err := t.runner.Run(ctx, "aerospace", "list-windows", "--all", "--json", "--format", "%{workspace} %{app-name}")
	if err != nil {
		slog.Warn("bar: list windows failed", "error", err)
		return
	}
	windows, err := decodeJSON[[]aerospaceWindow](out, "aerospace windows")
	if err != nil {
		slog.Warn("bar: list windows failed", "error", err)
		return
	}
	renderTiles(b, t.colors, spaces, windows, ev.Env["FOCUSED_WORKSPACE"])
}

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

func listWorkspaces(ctx context.Context, r Runner) ([]aerospaceWorkspace, error) {
	out, err := r.Run(ctx, "aerospace", "list-workspaces", "--all", "--json",
		"--format", "%{workspace} %{monitor-appkit-nsscreen-screens-id} %{workspace-is-focused} %{workspace-is-visible}")
	if err != nil {
		return nil, err
	}
	return decodeJSON[[]aerospaceWorkspace](out, "aerospace workspaces")
}

func renderTiles(b *Batch, c Colors, spaces []aerospaceWorkspace, windows []aerospaceWindow, focused string) {
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
		current := isFocused || (s.Visible && focused == "")
		fill, text := "0x00000000", c.Text
		if current {
			fill, text = c.Hot, c.HotText
		}
		b.Set(tileItem(s.Workspace),
			onOff("drawing", current || s.Visible || label != ""),
			P("display", s.Display),
			P("label", label),
			onOff("label.drawing", label != ""),
			P("background.color", fill),
			P("icon.color", text),
			P("label.color", text),
		)
	}
}

func BottomItems(cfg Config, runner Runner) []Item {
	c := cfg.Colors
	shelf := &Shelf{}
	return []Item{
		NewLayout(runner, cfg.Margin, cfg.MarginOverride, true),
		NewSessions(utenaURL, c, shelf),
		NewPet(runner, c, shelf),
		NewSpaceTiles(runner, c),
	}
}
