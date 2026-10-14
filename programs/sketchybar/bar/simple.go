package bar

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

const (
	frontAppItem = "front_app"
	clockItem    = "clock"
	mailItem     = "mail"
)

type FrontApp struct {
	runner Runner
}

func NewFrontApp(runner Runner) *FrontApp {
	return &FrontApp{runner: runner}
}

func (f *FrontApp) Setup(ctx context.Context, b *Batch) error {
	b.Add(frontAppItem, "left", "icon.drawing=off", "label.padding_left=12")
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
	b.Set(frontAppItem, P("label", name))
}

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Setup(ctx context.Context, b *Batch) error {
	b.Add(clockItem, "right", "click_script=open -a Calendar", "label.padding_left=9")
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
		b.Set(clockItem, P("icon", t.Format("Mon Jan 2")), P("label", t.Format("3:04 PM")))
	}
}

type Mail struct {
	runner Runner
	app    string
	last   lastValue
}

func NewMail(runner Runner, app string) *Mail {
	return &Mail{runner: runner, app: app}
}

func (m *Mail) Setup(ctx context.Context, b *Batch) error {
	b.Add(mailItem, "right", P("icon", iconMail), "drawing=off", P("click_script", "open -a '"+m.app+"'"))
	return nil
}

func (m *Mail) Interval() time.Duration {
	return 30 * time.Second
}

func (m *Mail) Poll(ctx context.Context) Update {
	out, err := m.runner.Run(ctx, "lsappinfo", "info", "-only", "StatusLabel", m.app)
	if err != nil {
		slog.Debug("bar: mail badge lookup failed", "error", err)
	}
	count := parseStatusLabel(string(out))
	if !m.last.changed(count) {
		return nil
	}
	return func(b *Batch) {
		b.Set(mailItem, onOff("drawing", count != "" && count != "0"), P("label", count))
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
	set   bool
	value string
}

func (l *lastValue) changed(v string) bool {
	if l.set && l.value == v {
		return false
	}
	l.set, l.value = true, v
	return true
}
