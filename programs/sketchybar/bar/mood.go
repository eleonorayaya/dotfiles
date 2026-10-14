package bar

import (
	"context"
	"sync"
	"time"
)

const (
	kaomojiItem     = "kaomoji"
	cpuHighPercent  = 80
	cpuHighSamples  = 2
	sparklyDuration = 5 * time.Minute

	faceHappy     = "(˶^ω^˶)"
	faceSparkly   = "(ﾉ´ヮ`)ﾉ*:･ﾟ☆"
	faceFlustered = "(＞_＜)"
	faceLost      = "(´；ω；`)"
	faceSleepy    = "(˘ω˘) zzZ"
)

type Mood struct {
	mu           sync.Mutex
	now          func() time.Time
	cpuHigh      int
	offline      bool
	unread       int
	sparklyUntil time.Time
}

func NewMood() *Mood {
	return &Mood{now: time.Now}
}

func (m *Mood) SetCPU(percent float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if percent >= cpuHighPercent {
		m.cpuHigh++
	} else {
		m.cpuHigh = 0
	}
}

func (m *Mood) SetOnline(online bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.offline = !online
}

func (m *Mood) SetUnread(total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if total > m.unread {
		m.sparklyUntil = m.now().Add(sparklyDuration)
	}
	m.unread = total
}

func (m *Mood) Face() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	switch {
	case now.Hour() < 6:
		return faceSleepy
	case m.offline:
		return faceLost
	case m.cpuHigh >= cpuHighSamples:
		return faceFlustered
	case now.Before(m.sparklyUntil):
		return faceSparkly
	default:
		return faceHappy
	}
}

type Kaomoji struct {
	mood   *Mood
	colors Colors
	last   lastValue
}

func NewKaomoji(mood *Mood, colors Colors) *Kaomoji {
	return &Kaomoji{mood: mood, colors: colors}
}

func (k *Kaomoji) Setup(ctx context.Context, b *Batch) error {
	props := append(island(k.colors), "icon.drawing=off", P("label.padding_left", islandPadding), P("label", faceHappy))
	b.Add(kaomojiItem, "center", props...)
	return nil
}

func (k *Kaomoji) Interval() time.Duration {
	return 5 * time.Second
}

func (k *Kaomoji) Poll(ctx context.Context) Update {
	face := k.mood.Face()
	if !k.last.changed(face) {
		return nil
	}
	return func(b *Batch) {
		b.Set(kaomojiItem, P("label", face))
	}
}

func (k *Kaomoji) resetCache() {
	k.last.reset()
}
