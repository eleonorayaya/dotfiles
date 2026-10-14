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

const (
	sinkItem       = "shizuku.events"
	healthInterval = 10 * time.Second
	maxMissedPings = 3
)

var customEvents = []string{"aerospace_workspace_change", "aerospace_workspace_monitor_change"}

type Daemon struct {
	runner  Runner
	cfg     Config
	emitCmd string
	items   []Item
}

func NewDaemon(cfg Config, runner Runner, emitCmd string, items []Item) *Daemon {
	return &Daemon{runner: runner, cfg: cfg, emitCmd: emitCmd, items: items}
}

func (d *Daemon) Setup(ctx context.Context) error {
	b := &Batch{}
	b.Remove("/.*/")
	b.Raw("--bar")
	b.Raw(barProps(d.cfg)...)
	b.Raw("--default")
	b.Raw(defaultProps(d.cfg)...)
	for _, ev := range customEvents {
		b.Raw("--add", "event", ev)
	}

	b.Add("item", sinkItem, "left", "drawing=off", "updates=on", P("script", d.emitCmd))
	b.Subscribe(sinkItem, d.subscribedEvents()...)

	for _, item := range d.items {
		if err := item.Setup(ctx, b); err != nil {
			return fmt.Errorf("failed to set up bar item: %w", err)
		}
	}
	return b.Apply(ctx, d.runner)
}

func (d *Daemon) subscribedEvents() []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range d.items {
		h, ok := item.(EventHandler)
		if !ok {
			continue
		}
		for _, ev := range h.Events() {
			if !seen[ev] {
				seen[ev] = true
				out = append(out, ev)
			}
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

	updates := make(chan Update, 32)
	for _, item := range d.items {
		if p, ok := item.(Poller); ok {
			go poll(ctx, p, updates)
		}
		if s, ok := item.(Streamer); ok {
			go s.Stream(ctx, updates)
		}
	}

	d.dispatch(ctx, Event{Sender: "forced"})

	health := time.NewTicker(healthInterval)
	defer health.Stop()
	missed := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-events:
			if ev.Sender == quitSender {
				slog.Info("bar: replaced by a new daemon, exiting")
				return nil
			}
			d.dispatch(ctx, ev)
		case u := <-updates:
			b := &Batch{}
			u(b)
			d.apply(ctx, b)
		case <-health.C:
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

func (d *Daemon) dispatch(ctx context.Context, ev Event) {
	if ev.Sender == "forced" && ev.Name != "" && ev.Name != sinkItem {
		return
	}
	b := &Batch{}
	if strings.HasPrefix(ev.Sender, "mouse.") && ev.Name != sinkItem {
		for _, item := range d.items {
			if c, ok := item.(ClickHandler); ok && strings.HasPrefix(ev.Name, c.Prefix()) {
				c.Click(ctx, ev, b)
			}
		}
	} else {
		for _, item := range d.items {
			if h, ok := item.(EventHandler); ok && handles(h, ev.Sender) {
				h.Handle(ctx, ev, b)
			}
		}
	}
	d.apply(ctx, b)
}

func handles(h EventHandler, sender string) bool {
	return sender == "forced" || slices.Contains(h.Events(), sender)
}

func (d *Daemon) apply(ctx context.Context, b *Batch) {
	if err := b.Apply(ctx, d.runner); err != nil {
		slog.Warn("bar: update failed", "error", err)
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

func OnClick(b *Batch, name, emitCmd string) {
	b.Set(name, P("script", emitCmd))
	b.Subscribe(name, "mouse.clicked")
}

func barProps(cfg Config) []string {
	s := cfg.Bar
	return []string{
		P("height", s.Height),
		"position=top",
		"padding_left=16",
		"padding_right=16",
		P("margin", s.Margin),
		P("y_offset", s.YOffset),
		P("color", s.Color),
		P("border_color", s.BorderColor),
		P("border_width", s.BorderWidth),
		P("corner_radius", s.CornerRadius),
		P("blur_radius", s.BlurRadius),
		"sticky=off",
		"font_smoothing=on",
	}
}

func defaultProps(cfg Config) []string {
	c := cfg.Colors
	return []string{
		P("icon.font", cfg.Fonts.Icon+":Regular:16.0"),
		P("icon.color", c.Icon),
		P("icon.highlight_color", c.IconHighlight),
		"icon.padding_left=3",
		"icon.padding_right=3",
		P("label.font", cfg.Fonts.Text+":Regular:14.0"),
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
