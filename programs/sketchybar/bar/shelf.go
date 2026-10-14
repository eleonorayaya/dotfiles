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
	sessionSlots   = 8
	sessionPrefix  = "sess."
	sessionPopup   = "sesspop."
	maxNameRunes   = 14
	maxPopupAgents = 6
	tilePrefix     = "tile."
	tilesBracket   = "spaces"

	pipWorking = "●"
	pipIdle    = "○"
	pipNeedsMe = "✦"
)

var liveSessionStatuses = []string{"active", "creating", "pending", "completed"}

var agentStatusLabels = map[string]string{
	"working":          "working",
	"idle":             "idle",
	"needs_attention":  "needs you",
	"ready_for_review": "ready for review",
	"done":             "done",
}

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
	name     string
	agents   []string
	attached bool
	needsMe  bool
	more     int
}

func (c sessionChip) label() string {
	if c.more > 0 {
		return fmt.Sprintf("+%d", c.more)
	}
	var pips strings.Builder
	for _, status := range c.agents {
		pips.WriteString(pip(status))
	}
	if pips.Len() == 0 {
		return c.name
	}
	return c.name + " " + pips.String()
}

func truncateName(name string) string {
	r := []rune(name)
	if len(r) <= maxNameRunes {
		return name
	}
	return string(r[:maxNameRunes-1]) + "…"
}

func sessionChips(sessions []utenaSession) []sessionChip {
	var chips []sessionChip
	for _, s := range sessions {
		if !slices.Contains(liveSessionStatuses, s.Status) {
			continue
		}
		chip := sessionChip{id: s.ID, name: truncateName(s.Name), attached: s.IsAttached}
		for _, a := range s.ClaudeSessions {
			chip.agents = append(chip.agents, a.Status)
			chip.needsMe = chip.needsMe || pip(a.Status) == pipNeedsMe
		}
		chips = append(chips, chip)
	}
	slices.SortFunc(chips, func(a, b sessionChip) int { return int(a.id) - int(b.id) })
	if len(chips) > sessionSlots {
		more := len(chips) - (sessionSlots - 1)
		chips = append(chips[:sessionSlots-1], sessionChip{more: more})
	}
	return chips
}

type Sessions struct {
	utena  *Utena
	colors Colors
	shelf  *Shelf
	mu     sync.Mutex
	chips  []sessionChip
	agents map[uint]string
	last   lastValue
}

func NewSessions(baseURL string, colors Colors, shelf *Shelf) *Sessions {
	return &Sessions{utena: NewUtena(baseURL, colors), colors: colors, shelf: shelf, agents: map[uint]string{}}
}

func sessionItem(i int) string { return fmt.Sprintf("%s%d", sessionPrefix, i) }

func (s *Sessions) Setup(ctx context.Context, b *Batch) error {
	for i := range sessionSlots {
		b.Add(sessionItem(i), "left", append(append(island(s.colors), large()...), "drawing=off", "icon.drawing=off", "popup.align=left")...)
		b.OnClick(sessionItem(i), "mouse.entered", "mouse.exited", "mouse.exited.global")
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
	finished := s.trackAgents(body.Sessions)

	s.mu.Lock()
	s.chips = chips
	s.mu.Unlock()

	var bounce []int
	for i, c := range chips {
		if slices.Contains(finished, c.id) {
			bounce = append(bounce, i)
		}
	}
	if !s.last.changed(fmt.Sprint(chips)) && len(bounce) == 0 {
		return nil
	}
	return func(b *Batch) {
		renderSessions(b, s.colors, chips)
		for _, i := range bounce {
			b.Raw("--animate", "sin", "14", "--set", sessionItem(i), "y_offset=6", "y_offset=0")
		}
	}
}

func (s *Sessions) trackAgents(sessions []utenaSession) []uint {
	needsMe := false
	var finished []uint
	seen := map[uint]string{}
	for _, sess := range sessions {
		if !slices.Contains(liveSessionStatuses, sess.Status) {
			continue
		}
		for _, a := range sess.ClaudeSessions {
			seen[a.ID] = a.Status
			needsMe = needsMe || a.Status == "needs_attention"
			if s.agents[a.ID] == "working" && a.Status != "working" && !slices.Contains(finished, sess.ID) {
				finished = append(finished, sess.ID)
			}
		}
	}
	s.agents = seen
	s.shelf.update(needsMe, len(finished) > 0, time.Now())
	return finished
}

func chipProps(c Colors, chip sessionChip, hovered bool) []string {
	props := island(c)
	switch {
	case chip.attached:
		props = hotIsland(c)
	case chip.needsMe:
		props = append(props, P("background.color", c.IslandBorder), P("background.border_color", c.Hot))
	}
	if hovered && !chip.attached {
		props = append(props, P("background.border_color", c.Hot))
	}
	return append(append(props, large()...), "drawing=on", "icon.drawing=off", P("label", chip.label()))
}

func renderSessions(b *Batch, c Colors, chips []sessionChip) {
	for i := range sessionSlots {
		if i >= len(chips) {
			b.Set(sessionItem(i), "drawing=off")
			continue
		}
		b.Set(sessionItem(i), chipProps(c, chips[i], false)...)
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
	if i < 0 || i >= len(s.chips) {
		s.mu.Unlock()
		return
	}
	chip := s.chips[i]
	s.mu.Unlock()

	item := sessionItem(i)
	switch ev.Sender {
	case "mouse.entered":
		b.Set(item, chipProps(s.colors, chip, true)...)
		renderAgentPopup(b, item, chip)
	case "mouse.exited", "mouse.exited.global":
		b.Set(item, append(chipProps(s.colors, chip, false), "popup.drawing=off")...)
	case "mouse.clicked":
		if chip.more == 0 {
			s.activate(ctx, chip.id)
		}
	}
}

func renderAgentPopup(b *Batch, host string, chip sessionChip) {
	b.Remove("/" + strings.ReplaceAll(sessionPopup, ".", `\.`) + ".*/")
	if len(chip.agents) == 0 {
		b.Set(host, "popup.drawing=off")
		return
	}
	for j, status := range chip.agents[:min(len(chip.agents), maxPopupAgents)] {
		label := agentStatusLabels[status]
		if label == "" {
			label = status
		}
		b.Add(fmt.Sprintf("%s%d", sessionPopup, j), "popup."+host,
			P("icon", pip(status)),
			"icon.padding_left=12",
			"icon.padding_right=6",
			P("label", fmt.Sprintf("agent %d · %s", j+1, label)),
			"label.padding_right=14",
		)
	}
	b.Set(host, "popup.drawing=on")
}

func (s *Sessions) activate(ctx context.Context, id uint) {
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
	runner  Runner
	colors  Colors
	mu      sync.Mutex
	current map[string]bool
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
			P("icon.font", font("ExtraBold", 15)),
			"icon.padding_left=12",
			"icon.padding_right=5",
			P("label.font", largeAppFont),
			"label.padding_right=12",
			"background.drawing=on",
			"background.color=0x00000000",
			"background.height=30",
			"background.corner_radius=15",
			P("padding_left", left),
			P("padding_right", right),
			P("click_script", "aerospace workspace "+s.Workspace),
		)
		b.Listen(tileItem(s.Workspace), "mouse.entered", "mouse.exited")
	}
	b.Raw(members...)
	b.Set(tilesBracket, append(bracketIsland(t.colors), P("background.height", largeHeight), P("background.corner_radius", largeRadius))...)
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
	current := renderTiles(b, t.colors, spaces, windows, ev.Env["FOCUSED_WORKSPACE"])
	t.mu.Lock()
	t.current = current
	t.mu.Unlock()
}

func (t *SpaceTiles) Prefix() string {
	return tilePrefix
}

func (t *SpaceTiles) Click(ctx context.Context, ev Event, b *Batch) {
	ws := strings.TrimPrefix(ev.Name, tilePrefix)
	t.mu.Lock()
	current := t.current[ws]
	t.mu.Unlock()
	if current {
		return
	}
	fill := "0x00000000"
	if ev.Sender == "mouse.entered" {
		fill = t.colors.IslandBorder
	}
	b.Set(ev.Name, P("background.color", fill))
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

func renderTiles(b *Batch, c Colors, spaces []aerospaceWorkspace, windows []aerospaceWindow, focused string) map[string]bool {
	currents := map[string]bool{}
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
		currents[s.Workspace] = current
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
	return currents
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
