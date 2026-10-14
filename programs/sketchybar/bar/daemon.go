package bar

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
)

type Update func(b *Batch)

type Item interface {
	Setup(ctx context.Context, b *Batch) error
}

type EventHandler interface {
	Events() []string
	Handle(ctx context.Context, ev Event, b *Batch)
}

type ClickHandler interface {
	Prefix() string
	Click(ctx context.Context, ev Event, b *Batch)
}

type Poller interface {
	Interval() time.Duration
	Poll(ctx context.Context) Update
}

type Streamer interface {
	Stream(ctx context.Context, updates chan<- Update)
}

type cacher interface {
	resetCache()
}

const (
	sinkItem       = "shizuku.events"
	healthInterval = 10 * time.Second
	maxMissedPings = 3
	resetupBackoff = 5 * time.Second
)

type Daemon struct {
	runner      Runner
	cfg         Config
	emit        string
	items       []Item
	updates     chan Update
	lastApplied time.Time
	lastSetup   time.Time
}

func NewDaemon(cfg Config, runner Runner, emit string, items []Item) *Daemon {
	return &Daemon{runner: runner, cfg: cfg, emit: emit, items: items, updates: make(chan Update, 32)}
}

func (d *Daemon) newBatch() *Batch {
	return &Batch{emit: d.emit}
}

func (d *Daemon) Setup(ctx context.Context) error {
	d.lastSetup = time.Now()
	b := d.newBatch()
	b.Remove("/.*/")
	b.Raw("--bar")
	b.Raw(barProps(d.cfg.Colors)...)
	b.Raw("--default")
	b.Raw(defaultProps(d.cfg.Colors)...)

	for _, item := range d.items {
		if err := item.Setup(ctx, b); err != nil {
			return fmt.Errorf("failed to set up bar item: %w", err)
		}
	}

	b.Add(sinkItem, "left", "drawing=off", "updates=on", P("script", d.emit))
	b.Subscribe(sinkItem, d.subscribedEvents()...)
	if err := b.apply(ctx, d.runner); err != nil {
		return err
	}
	for _, item := range d.items {
		if c, ok := item.(cacher); ok {
			c.resetCache()
		}
	}
	return nil
}

func (d *Daemon) subscribedEvents() []string {
	var out []string
	for _, h := range d.eventHandlers() {
		for _, ev := range h.Events() {
			if !slices.Contains(out, ev) {
				out = append(out, ev)
			}
		}
	}
	return out
}

func (d *Daemon) eventHandlers() []EventHandler {
	var out []EventHandler
	for _, item := range d.items {
		if h, ok := item.(EventHandler); ok {
			out = append(out, h)
		}
	}
	return out
}

func (d *Daemon) Run(ctx context.Context, events <-chan Event) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := d.Setup(ctx); err != nil {
		return err
	}

	for _, item := range d.items {
		if p, ok := item.(Poller); ok {
			go poll(ctx, p, d.updates)
		}
		if s, ok := item.(Streamer); ok {
			go s.Stream(ctx, d.updates)
		}
	}

	d.render(ctx)

	health := time.NewTicker(healthInterval)
	defer health.Stop()
	missed := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-events:
			d.dispatch(ctx, append([]Event{ev}, drain(events)...))
		case u := <-d.updates:
			b := d.newBatch()
			u(b)
			for _, more := range drain(d.updates) {
				more(b)
			}
			d.apply(ctx, b)
		case <-health.C:
			if time.Since(d.lastApplied) < healthInterval {
				missed = 0
				continue
			}
			if _, err := d.runner.Run(ctx, "sketchybar", "--query", "bar"); err != nil {
				missed++
				if missed >= maxMissedPings {
					return fmt.Errorf("sketchybar stopped responding: %w", err)
				}
				continue
			}
			missed = 0
		}
	}
}

func (d *Daemon) render(ctx context.Context) {
	b := d.newBatch()
	for _, h := range d.eventHandlers() {
		h.Handle(ctx, Event{}, b)
	}
	d.apply(ctx, b)
}

func (d *Daemon) dispatch(ctx context.Context, events []Event) {
	var subscribed []Event
	for _, ev := range events {
		if ev.Name == sinkItem {
			subscribed = append(subscribed, ev)
			continue
		}
		d.click(ctx, ev)
	}
	if len(subscribed) == 0 {
		return
	}

	b := d.newBatch()
	for _, h := range d.eventHandlers() {
		if ev, ok := lastMatching(subscribed, h.Events()); ok {
			h.Handle(ctx, ev, b)
		}
	}
	d.apply(ctx, b)
}

func (d *Daemon) click(ctx context.Context, ev Event) {
	for _, item := range d.items {
		c, ok := item.(ClickHandler)
		if !ok || !strings.HasPrefix(ev.Name, c.Prefix()) {
			continue
		}
		go func() {
			b := d.newBatch()
			c.Click(ctx, ev, b)
			send(ctx, d.updates, func(out *Batch) { out.Raw(b.Args()...) })
		}()
	}
}

func lastMatching(events []Event, senders []string) (Event, bool) {
	for i := len(events) - 1; i >= 0; i-- {
		if slices.Contains(senders, events[i].Sender) {
			return events[i], true
		}
	}
	return Event{}, false
}

func (d *Daemon) apply(ctx context.Context, b *Batch) {
	err := b.apply(ctx, d.runner)
	if err == nil {
		d.lastApplied = time.Now()
		return
	}
	slog.Warn("bar: update failed", "error", err)
	if strings.Contains(err.Error(), "Item not found") && time.Since(d.lastSetup) > resetupBackoff {
		slog.Info("bar: items missing, configuring sketchybar again")
		if err := d.Setup(ctx); err != nil {
			slog.Warn("bar: setup failed", "error", err)
			return
		}
		d.render(ctx)
	}
}

func drain[T any](ch <-chan T) []T {
	var out []T
	for {
		select {
		case v := <-ch:
			out = append(out, v)
		default:
			return out
		}
	}
}

func poll(ctx context.Context, p Poller, updates chan<- Update) {
	ticker := time.NewTicker(p.Interval())
	defer ticker.Stop()
	for {
		if u := p.Poll(ctx); u != nil {
			send(ctx, updates, u)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func send(ctx context.Context, updates chan<- Update, u Update) {
	select {
	case updates <- u:
	case <-ctx.Done():
	}
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func barProps(c Colors) []string {
	return []string{
		"height=32",
		"position=top",
		"padding_left=16",
		"padding_right=16",
		"margin=12",
		"y_offset=12",
		P("color", c.Bar),
		P("border_color", c.BarBorder),
		"border_width=2",
		"corner_radius=8",
		"blur_radius=48",
		"sticky=off",
		"font_smoothing=on",
	}
}

func defaultProps(c Colors) []string {
	return []string{
		P("icon.font", symbolFont+":Semibold:15.0"),
		P("icon.color", c.Icon),
		P("icon.highlight_color", c.IconHighlight),
		"icon.padding_left=3",
		"icon.padding_right=3",
		"label.font=SF Pro Display:Thin:16.0",
		P("label.color", c.Label),
		P("label.highlight_color", c.LabelHighlight),
		"label.padding_left=3",
		"label.padding_right=3",
		"padding_left=3",
		"padding_right=3",
		"background.height=30",
		"background.corner_radius=9",
		"popup.background.border_width=2",
		"popup.background.corner_radius=9",
		P("popup.background.border_color", c.PopupBorder),
		P("popup.background.color", c.PopupBackground),
		"popup.blur_radius=20",
		"popup.background.shadow.drawing=on",
		"updates=when_shown",
	}
}
