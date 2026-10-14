package bar

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const (
	frontAppItem = "front_app"
	clockItem    = "clock"
)

type FrontApp struct {
	runner Runner
	colors Colors
}

func NewFrontApp(runner Runner, colors Colors) *FrontApp {
	return &FrontApp{runner: runner, colors: colors}
}

func (f *FrontApp) Setup(ctx context.Context, b *Batch) error {
	b.Add(frontAppItem, "left", append(island(f.colors), P("icon.font", appFont))...)
	return nil
}

func (f *FrontApp) Events() []string {
	return []string{"front_app_switched"}
}

func (f *FrontApp) Handle(ctx context.Context, ev Event, b *Batch) {
	name := ev.Info
	if name == "" {
		out, err := f.runner.Run(ctx, "aerospace", "list-windows", "--focused", "--format", "%{app-name}")
		if err != nil {
			slog.Debug("bar: focused app lookup failed", "error", err)
		}
		name = strings.TrimSpace(string(out))
	}
	b.Set(frontAppItem, P("icon", appIcon(name)), P("label", strings.ToLower(name)))
}

type Clock struct {
	colors Colors
}

func NewClock(colors Colors) *Clock {
	return &Clock{colors: colors}
}

func (c *Clock) Setup(ctx context.Context, b *Batch) error {
	b.Add(clockItem, "right", append(hotIsland(c.colors),
		P("icon.font", font("ExtraBold", 13)),
		"icon.padding_right=6",
		P("label.font", font("Medium", 12)),
		"click_script=open -a Calendar",
	)...)
	return nil
}

func (c *Clock) Stream(ctx context.Context, updates chan<- Update) {
	for ctx.Err() == nil {
		now := time.Now()
		send(ctx, updates, clockUpdate(now))
		sleep(ctx, now.Truncate(time.Minute).Add(time.Minute).Sub(now))
	}
}

func (c *Clock) Events() []string {
	return []string{"system_woke"}
}

func (c *Clock) Handle(ctx context.Context, ev Event, b *Batch) {
	clockUpdate(time.Now())(b)
}

func clockUpdate(t time.Time) Update {
	return func(b *Batch) {
		b.Set(clockItem, P("icon", strings.ToLower(t.Format("3:04 pm"))), P("label", strings.ToLower(t.Format("Mon 2"))))
	}
}

func parseStatusLabel(out string) string {
	const marker = `"label"="`
	_, rest, ok := strings.Cut(out, marker)
	if !ok {
		return ""
	}
	label, _, ok := strings.Cut(rest, `"`)
	if !ok {
		return ""
	}
	return label
}

type lastValue struct {
	mu    sync.Mutex
	set   bool
	value string
}

func (l *lastValue) changed(v string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.set && l.value == v {
		return false
	}
	l.set, l.value = true, v
	return true
}

func (l *lastValue) reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.set = false
}
