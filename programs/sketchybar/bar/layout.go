package bar

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
)

const itemPadding = 6

type Layout struct {
	runner        Runner
	margin        int
	override      *MarginOverride
	hideOnMonitor string
}

func NewLayout(runner Runner, margin int, override *MarginOverride, hideOnMonitor string) *Layout {
	return &Layout{runner: runner, margin: margin, override: override, hideOnMonitor: hideOnMonitor}
}

func (l *Layout) Setup(ctx context.Context, b *Batch) error {
	return nil
}

func (l *Layout) Events() []string {
	return []string{"display_change", "system_woke"}
}

func (l *Layout) Handle(ctx context.Context, ev Event, b *Batch) {
	margin, monitors := l.margin, ""
	if l.override != nil || l.hideOnMonitor != "" {
		out, err := l.runner.Run(ctx, "aerospace", "list-monitors", "--format", "%{monitor-name}")
		if err != nil {
			slog.Warn("bar: list monitors failed", "error", err)
		}
		monitors = string(out)
	}
	if l.override != nil && allMonitorsMatch(monitors, l.override.Monitor) {
		margin = l.override.Margin
	}
	props := []string{P("margin", max(margin-itemPadding, 0))}
	if l.hideOnMonitor != "" {
		props = append(props, onOff("hidden", allMonitorsMatch(monitors, l.hideOnMonitor)))
	}
	b.Raw(append([]string{"--bar"}, props...)...)
}

func allMonitorsMatch(out, pattern string) bool {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return false
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return false
	}
	for name := range strings.Lines(trimmed) {
		if !re.MatchString(strings.TrimSpace(name)) {
			return false
		}
	}
	return true
}
