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
	sessionSlots       = 8
	sessionPrefix      = "sess."
	sessionPopup       = "sesspop."
	sessionPopupRegex  = `/sesspop\..*/`
	maxNameRunes       = 14
	maxPopupAgents     = 6
	sessionBounce      = 6
	tilePrefix         = "tile."
	tilesBracket       = "spaces"
	tileOuterPadding   = itemPadding + 4
	tileInnerPadding   = 2
	statusWorking      = "working"
	statusNeedsYou     = "needs_attention"
	statusReadyReview  = "ready_for_review"
	pipWorking         = "●"
	pipIdle            = "○"
	pipNeedsMe         = "✦"
	agentLabelNeedsYou = "needs you"
	agentLabelReview   = "ready for review"
)

var liveSessionStatuses = []string{"active", "creating", "pending", "completed"}

type claudeAgent struct {
	ID     uint   `json:"ID"`
	Status string `json:"status"`
}

func wantsYou(status string) bool {
	return status == statusNeedsYou || status == statusReadyReview
}

func pip(status string) string {
	switch {
	case status == statusWorking:
		return pipWorking
	case wantsYou(status):
		return pipNeedsMe
	default:
		return pipIdle
	}
}

func agentLabel(status string) string {
	switch status {
	case statusNeedsYou:
		return agentLabelNeedsYou
	case statusReadyReview:
		return agentLabelReview
	}
	return status
}

func liveSessions(sessions []utenaSession) []utenaSession {
	return slices.DeleteFunc(sessions, func(s utenaSession) bool {
		return !slices.Contains(liveSessionStatuses, s.Status)
	})
}

type Shelf struct {
	mu         sync.Mutex
	urgent     bool
	finishedAt time.Time
}

func (s *Shelf) update(urgent, finished bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urgent = urgent
	if finished {
		s.finishedAt = time.Now()
	}
}

func (s *Shelf) snapshot() (bool, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.urgent, s.finishedAt
}

type sessionChip struct {
	id       uint
	name     string
	agents   []string
	attached bool
	wantsYou bool
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
		chip := sessionChip{id: s.ID, name: truncateName(s.Name), attached: s.IsAttached}
		for _, a := range s.ClaudeSessions {
			chip.agents = append(chip.agents, a.Status)
			chip.wantsYou = chip.wantsYou || wantsYou(a.Status)
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
	utena   utenaClient
	colors  Colors
	shelf   *Shelf
	mu      sync.Mutex
	chips   []sessionChip
	hovered int
	agents  map[uint]string
	last    lastValue
}

func NewSessions(baseURL string, colors Colors, shelf *Shelf) *Sessions {
	return &Sessions{utena: newUtenaClient(baseURL), colors: colors, shelf: shelf, hovered: -1, agents: map[uint]string{}}
}

func sessionItem(i int) string { return fmt.Sprintf("%s%d", sessionPrefix, i) }

func (s *Sessions) Setup(ctx context.Context, b *Batch) error {
	for i := range sessionSlots {
		b.Add(sessionItem(i), "left", "drawing=off", "popup.align=left")
		b.OnClick(sessionItem(i), "mouse.entered", "mouse.exited", "mouse.exited.global")
	}
	return nil
}

func (s *Sessions) Interval() time.Duration {
	return 3 * time.Second
}

func (s *Sessions) Poll(ctx context.Context) Update {
	sessions, err := s.utena.sessions(ctx)
	if err != nil {
		slog.Debug("bar: utena sessions unavailable", "error", err)
		return nil
	}
	sessions = liveSessions(sessions)
	chips := sessionChips(sessions)
	finished := s.trackAgents(sessions)

	s.mu.Lock()
	s.chips = chips
	hovered := s.hovered
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
		renderSessions(b, s.colors, chips, hovered)
		if hovered >= 0 && hovered < len(chips) {
			renderAgentPopup(b, sessionItem(hovered), chips[hovered])
		}
		for _, i := range bounce {
			b.Bounce(sessionItem(i), sessionBounce)
		}
	}
}

func (s *Sessions) trackAgents(sessions []utenaSession) []uint {
	urgent := false
	var finished []uint
	seen := map[uint]string{}
	for _, sess := range sessions {
		for _, a := range sess.ClaudeSessions {
			seen[a.ID] = a.Status
			urgent = urgent || a.Status == statusNeedsYou
			if s.agents[a.ID] == statusWorking && a.Status != statusWorking && !slices.Contains(finished, sess.ID) {
				finished = append(finished, sess.ID)
			}
		}
	}
	s.agents = seen
	s.shelf.update(urgent, len(finished) > 0)
	return finished
}

func chipProps(c Colors, chip sessionChip, hovered bool) []string {
	props := islandOf(c, largeIsland)
	switch {
	case chip.attached:
		props = hotIslandOf(c, largeIsland)
	case chip.wantsYou:
		props = append(props, P("background.color", c.IslandBorder), P("background.border_color", c.Hot))
	}
	if hovered && !chip.attached {
		props = append(props, P("background.border_color", c.Hot))
	}
	props = append(props, labelOnly(largeIsland)...)
	return append(props, "drawing=on", P("label", chip.label()))
}

func renderSessions(b *Batch, c Colors, chips []sessionChip, hovered int) {
	for i := range sessionSlots {
		if i >= len(chips) {
			b.Set(sessionItem(i), "drawing=off")
			continue
		}
		b.Set(sessionItem(i), chipProps(c, chips[i], i == hovered)...)
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
	switch ev.Sender {
	case "mouse.entered":
		s.hovered = i
	case "mouse.exited", "mouse.exited.global":
		if s.hovered == i {
			s.hovered = -1
		}
	}
	s.mu.Unlock()

	item := sessionItem(i)
	switch ev.Sender {
	case "mouse.entered":
		b.Set(item, chipProps(s.colors, chip, true)...)
		renderAgentPopup(b, item, chip)
	case "mouse.exited", "mouse.exited.global":
		b.Set(item, append(chipProps(s.colors, chip, false), "popup.drawing=off")...)
	case "mouse.clicked":
		if chip.more > 0 {
			return
		}
		if err := s.utena.do(ctx, http.MethodPut, fmt.Sprintf("/sessions/%d/activate", chip.id), nil); err != nil {
			slog.Warn("bar: activate session failed", "id", chip.id, "error", err)
		}
	}
}

func renderAgentPopup(b *Batch, host string, chip sessionChip) {
	b.Remove(sessionPopupRegex)
	if len(chip.agents) == 0 {
		b.Set(host, "popup.drawing=off")
		return
	}
	for j, status := range chip.agents[:min(len(chip.agents), maxPopupAgents)] {
		b.Add(fmt.Sprintf("%s%d", sessionPopup, j), "popup."+host,
			P("icon", pip(status)),
			P("icon.padding_left", islandPadding),
			"icon.padding_right=6",
			P("label", fmt.Sprintf("agent %d · %s", j+1, agentLabel(status))),
			P("label.padding_right", largeIsland.pad),
		)
	}
	b.Set(host, "popup.drawing=on")
}

func (s *Sessions) resetCache() {
	s.last.reset()
}

type SpaceTiles struct {
	runner  Runner
	colors  Colors
	mu      sync.Mutex
	current map[string]bool
	known   map[string]bool
	last    lastValue
}

func NewSpaceTiles(runner Runner, colors Colors) *SpaceTiles {
	return &SpaceTiles{runner: runner, colors: colors}
}

func tileItem(ws string) string { return tilePrefix + ws }

func (t *SpaceTiles) Setup(ctx context.Context, b *Batch) error {
	addWorkspaceEvents(b)
	spaces, err := listWorkspaces(ctx, t.runner)
	if err != nil {
		slog.Warn("bar: list workspaces failed, retrying on next update", "error", err)
		return nil
	}
	t.known = map[string]bool{}
	for _, s := range spaces {
		t.known[s.Workspace] = true
	}
	slices.Reverse(spaces)
	members := []string{"--add", "bracket", tilesBracket}
	for i, s := range spaces {
		left, right := tileInnerPadding, tileInnerPadding
		if i == 0 {
			right = tileOuterPadding
		}
		if i == len(spaces)-1 {
			left = tileOuterPadding
		}
		members = append(members, tileItem(s.Workspace))
		b.Add(tileItem(s.Workspace), "right",
			"drawing=off",
			P("icon", s.Workspace),
			P("icon.font", font("ExtraBold", tileSize.fontSize)),
			P("icon.padding_left", tileSize.pad),
			"icon.padding_right=5",
			P("label.font", largeAppFont),
			P("label.padding_right", tileSize.pad),
			"background.drawing=on",
			P("background.color", transparent),
			P("background.height", tileSize.height),
			P("background.corner_radius", tileSize.radius),
			P("padding_left", left),
			P("padding_right", right),
			P("click_script", "aerospace workspace "+s.Workspace),
		)
		b.Listen(tileItem(s.Workspace), "mouse.entered", "mouse.exited")
	}
	b.Raw(members...)
	b.Set(tilesBracket, bracketIsland(t.colors, largeIsland)...)
	return nil
}

func (t *SpaceTiles) Events() []string {
	return []string{WorkspaceChangeEvent, WorkspaceMonitorChangeEvent, "front_app_switched", "space_windows_change", "display_change", "system_woke"}
}

func (t *SpaceTiles) Handle(ctx context.Context, ev Event, b *Batch) {
	var windows []aerospaceWindow
	var windowsErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		windows, windowsErr = listWindows(ctx, t.runner)
	}()
	spaces, err := listWorkspaces(ctx, t.runner)
	<-done
	if err != nil {
		slog.Warn("bar: list workspaces failed", "error", err)
		return
	}
	if windowsErr != nil {
		slog.Warn("bar: list windows failed", "error", windowsErr)
		return
	}
	spaces = slices.DeleteFunc(spaces, func(s aerospaceWorkspace) bool { return !t.known[s.Workspace] })
	rendered := &Batch{}
	current := renderTiles(rendered, t.colors, spaces, windows, ev.Env["FOCUSED_WORKSPACE"])
	t.mu.Lock()
	t.current = current
	t.mu.Unlock()
	if t.last.changed(strings.Join(rendered.Args(), "\x00")) {
		b.Raw(rendered.Args()...)
	}
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
	fill := transparent
	if ev.Sender == "mouse.entered" {
		fill = t.colors.IslandBorder
	}
	b.Set(ev.Name, P("background.color", fill))
}

func (t *SpaceTiles) resetCache() {
	t.last.reset()
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
		label := strings.Join(apps[s.Workspace], " ")
		current := s.Visible || s.Workspace == focused
		currents[s.Workspace] = current
		fill, text := transparent, c.Text
		if current {
			fill, text = c.Hot, c.HotText
		}
		b.Set(tileItem(s.Workspace),
			onOff("drawing", current || label != ""),
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
		NewLayout(runner, cfg.Margin, cfg.MarginOverride, cfg.HideOnMonitor),
		NewSessions(utenaURL, c, shelf),
		NewPet(runner, c, shelf),
		NewSpaceTiles(runner, c),
	}
}
