package bar

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

const (
	badgesBracket  = "badges"
	badgeDot       = "•"
	badgeIconInset = 10
	bubbleReach    = 5
	bubbleOverlap  = 18
	bubbleDigit    = 6
	bubblePadding  = 8
)

type Badges struct {
	runner Runner
	colors Colors
	apps   []string
	mood   *Mood
	last   lastValue
}

func NewBadges(runner Runner, colors Colors, apps []string, mood *Mood) *Badges {
	return &Badges{runner: runner, colors: colors, apps: apps, mood: mood}
}

func badgeIconItem(i int) string  { return fmt.Sprintf("badge.%d", i) }
func badgeCountItem(i int) string { return fmt.Sprintf("badge.%d.count", i) }

func (bg *Badges) Setup(ctx context.Context, b *Batch) error {
	for i, app := range bg.apps {
		b.Add(badgeCountItem(i), "right",
			"drawing=off",
			"icon.drawing=off",
			P("label.font", font("ExtraBold", 9)),
			P("label.color", bg.colors.HotText),
			"label.padding_left=4",
			"label.padding_right=4",
			"background.drawing=on",
			P("background.color", bg.colors.Hot),
			"background.height=14",
			"background.corner_radius=7",
			"y_offset=7",
		)
		b.Add(badgeIconItem(i), "right",
			"drawing=off",
			P("icon", appIcon(app)),
			P("icon.font", appFont),
			P("icon.padding_left", badgeIconInset-bubbleReach),
			P("icon.padding_right", badgeIconInset),
			P("padding_left", itemPadding),
			"padding_right=0",
			"label.drawing=off",
			P("click_script", fmt.Sprintf("open -a %q", app)),
		)
	}
	b.Raw("--add", "bracket", badgesBracket, `/badge\..*/`)
	b.Set(badgesBracket, append(bracketIsland(bg.colors, smallIsland), "drawing=off")...)
	return nil
}

func (bg *Badges) Interval() time.Duration {
	return 15 * time.Second
}

func (bg *Badges) Poll(ctx context.Context) Update {
	labels := make([]string, len(bg.apps))
	for i, app := range bg.apps {
		out, err := bg.runner.Run(ctx, "lsappinfo", "info", "-only", "StatusLabel", app)
		if err != nil {
			slog.Debug("bar: badge lookup failed", "app", app, "error", err)
		}
		labels[i] = parseStatusLabel(string(out))
	}
	bg.mood.SetUnread(unreadTotal(labels))
	if !bg.last.changed(strings.Join(labels, "\x00")) {
		return nil
	}
	return func(b *Batch) {
		renderBadges(b, labels)
	}
}

func renderBadges(b *Batch, labels []string) {
	shown := false
	for i, label := range labels {
		visible := label != "" && label != "0"
		shown = shown || visible
		b.Set(badgeIconItem(i), onOff("drawing", visible))
		props := []string{onOff("drawing", visible), P("label", label), "background.height=14", "background.corner_radius=7", "y_offset=7"}
		if label == badgeDot {
			props = []string{onOff("drawing", visible), "label=", "background.height=8", "background.corner_radius=4", "y_offset=9"}
		}
		b.Set(badgeCountItem(i), append(props, bubbleOffsets(label)...)...)
	}
	b.Set(badgesBracket, onOff("drawing", shown))
}

func bubbleOffsets(label string) []string {
	width := bubblePadding
	if label != badgeDot {
		width += bubbleDigit * len([]rune(label))
	}
	return []string{P("padding_left", -bubbleOverlap), P("padding_right", bubbleOverlap-width)}
}

func unreadTotal(labels []string) int {
	total := 0
	for _, label := range labels {
		if n, err := strconv.Atoi(label); err == nil {
			total += n
		} else if label == badgeDot {
			total++
		}
	}
	return total
}

func (bg *Badges) resetCache() {
	bg.last.reset()
}
