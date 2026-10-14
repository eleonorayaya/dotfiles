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
	b.Add("item", frontAppItem, "left", "icon.drawing=off", "label.padding_left=12")
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

type Clock struct {
	now func() time.Time
}

func NewClock() *Clock {
	return &Clock{now: time.Now}
}

func (c *Clock) Setup(ctx context.Context, b *Batch) error {
	b.Add("item", clockItem, "right", "click_script=open -a Calendar", "label.padding_left=9")
	return nil
}

func (c *Clock) Interval() time.Duration {
	return 15 * time.Second
}

func (c *Clock) Poll(ctx context.Context) Update {
	t := c.now()
	return func(b *Batch) {
		b.Set(clockItem, P("icon", t.Format("Mon Jan 2")), P("label", t.Format("3:04 PM")))
	}
}

type Mail struct {
	runner Runner
	app    string
}

func NewMail(runner Runner, app string) *Mail {
	return &Mail{runner: runner, app: app}
}

func (m *Mail) Setup(ctx context.Context, b *Batch) error {
	b.Add("item", mailItem, "right", P("icon", iconMail), "drawing=off", P("click_script", "open -a '"+m.app+"'"))
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
	return func(b *Batch) {
		b.Set(mailItem, onOff("drawing", count != "" && count != "0"), P("label", count))
	}
}

func parseStatusLabel(out string) string {
	const marker = `"label"="`
	i := strings.Index(out, marker)
	if i < 0 {
		return ""
	}
	rest := out[i+len(marker):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}
