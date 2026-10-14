package bar

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	petItem     = "petsprite"
	petZzzItem  = "petzzz"
	petBracket  = "petroom"
	petPixel    = 5
	petIdleNap  = 5 * time.Minute
	petHopFor   = 4 * time.Second
	petTick     = time.Second
	petBlinkGap = 6
)

var petBase = []string{
	"P..........P",
	"PP........PP",
	"PcP......PcP",
	".PcccccccccP",
	".cccccccccc.",
	"ccbbccccbbcc",
	"ccbwccccbwcc",
	"ckcccmcccckc",
	".cccccccccc.",
	"..R.cccc.R..",
	".RR.cccc.RR.",
	"....c..c....",
}

var petColors = map[byte]color.NRGBA{
	'P': {0xf2, 0x9b, 0xbd, 0xff},
	'c': {0xff, 0xf6, 0xfa, 0xff},
	'b': {0x4a, 0x8e, 0xd8, 0xff},
	'w': {0xff, 0xff, 0xff, 0xff},
	'k': {0xf7, 0xb9, 0xd0, 0xff},
	'm': {0xa8, 0x08, 0x48, 0xff},
	'R': {0x7c, 0xc4, 0xef, 0xff},
	'o': {0x5a, 0x2a, 0x3a, 0xff},
}

func petFrames() map[string][]string {
	closed := withRows(petBase, map[int]string{5: "cccccccccccc", 6: "ccooccccoocc"})
	return map[string][]string{
		"awake": petBase,
		"blink": closed,
		"sleep": withRows(closed, map[int]string{7: "ckcccccccckc"}),
		"wave":  withRows(petBase, map[int]string{8: "RccccccccccR", 9: "R.R.cccc.R.R", 10: "....cccc...."}),
	}
}

func withRows(base []string, rows map[int]string) []string {
	out := append([]string(nil), base...)
	for i, row := range rows {
		out[i] = row
	}
	return out
}

func renderSprite(rows []string, scale int) *image.NRGBA {
	n := len(rows)
	size := (n + 2) * scale
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	filled := func(x, y int) bool {
		return y >= 0 && y < n && x >= 0 && x < len(rows[y]) && rows[y][x] != '.'
	}
	for y := -1; y <= n; y++ {
		for x := -1; x <= n; x++ {
			var c color.NRGBA
			switch {
			case filled(x, y):
				c = petColors[rows[y][x]]
			case filled(x-1, y) || filled(x+1, y) || filled(x, y-1) || filled(x, y+1):
				c = petColors['o']
			default:
				continue
			}
			for dy := range scale {
				for dx := range scale {
					img.SetNRGBA((x+1)*scale+dx, (y+1)*scale+dy, c)
				}
			}
		}
	}
	return img
}

func writePetFrames(dir string) (map[string]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create pet dir: %w", err)
	}
	paths := map[string]string{}
	for name, rows := range petFrames() {
		path := filepath.Join(dir, name+".png")
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s: %w", path, err)
		}
		err = png.Encode(f, renderSprite(rows, petPixel))
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to encode %s: %w", path, err)
		}
		paths[name] = path
	}
	return paths, nil
}

type Shelf struct {
	mu             sync.Mutex
	needsAttention bool
	finishedAt     time.Time
}

func (s *Shelf) update(needsAttention, finished bool, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.needsAttention = needsAttention
	if finished {
		s.finishedAt = now
	}
}

func (s *Shelf) snapshot() (bool, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.needsAttention, s.finishedAt
}

type Pet struct {
	runner Runner
	colors Colors
	shelf  *Shelf
	frames map[string]string
	now    func() time.Time
	idle   time.Duration
}

func NewPet(runner Runner, colors Colors, shelf *Shelf) *Pet {
	return &Pet{runner: runner, colors: colors, shelf: shelf, now: time.Now}
}

func (p *Pet) Setup(ctx context.Context, b *Batch) error {
	cache, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to find cache dir: %w", err)
	}
	frames, err := writePetFrames(filepath.Join(cache, "shizuku", "pet"))
	if err != nil {
		return err
	}
	p.frames = frames
	side := (len(petBase) + 2) * petPixel / 2
	b.Add(petItem, "center",
		"background.drawing=on",
		"background.color=0x00000000",
		P("background.image", frames["awake"]),
		"background.image.scale=0.5",
		P("width", side),
		"icon.drawing=off",
		"label.drawing=off",
		P("padding_left", itemPadding+islandPadding),
		"padding_right=4",
	)
	b.Add(petZzzItem, "center",
		"icon.drawing=off",
		"label=z z ₊",
		P("label.font", font("ExtraBold", 11)),
		P("label.color", p.colors.Hot),
		"drawing=off",
		"y_offset=4",
		"padding_left=0",
		P("padding_right", itemPadding+islandPadding),
	)
	b.Raw("--add", "bracket", petBracket, petItem, petZzzItem)
	b.Set(petBracket, append(bracketIsland(p.colors), P("background.height", islandHeight+8))...)
	return nil
}

func (p *Pet) Stream(ctx context.Context, updates chan<- Update) {
	ticks, lastFrame, hopped := 0, "", time.Time{}
	for ctx.Err() == nil {
		ticks++
		needsYou, finishedAt := p.shelf.snapshot()
		now := p.now()
		frame := p.frameFor(ctx, ticks, needsYou, now)
		hop := !finishedAt.IsZero() && now.Sub(finishedAt) < petHopFor && finishedAt.After(hopped)
		if hop {
			hopped = finishedAt
		}
		if frame != lastFrame || hop {
			lastFrame = frame
			send(ctx, updates, func(b *Batch) {
				b.Set(petItem, P("background.image", p.frames[frame]))
				b.Set(petZzzItem, onOff("drawing", frame == "sleep"))
				if hop {
					b.Raw("--animate", "sin", "12", "--set", petItem, "y_offset=10", "y_offset=0")
				}
			})
		}
		sleep(ctx, petTick)
	}
}

func (p *Pet) frameFor(ctx context.Context, tick int, needsYou bool, now time.Time) string {
	switch {
	case needsYou:
		if tick%2 == 0 {
			return "wave"
		}
		return "awake"
	case now.Hour() < 6 || p.userIdle(ctx, tick) >= petIdleNap:
		return "sleep"
	case tick%petBlinkGap == 0:
		return "blink"
	default:
		return "awake"
	}
}

func (p *Pet) userIdle(ctx context.Context, tick int) time.Duration {
	if tick%10 != 1 {
		return p.idle
	}
	if out, err := p.runner.Run(ctx, "ioreg", "-c", "IOHIDSystem", "-d", "4", "-k", "HIDIdleTime"); err == nil {
		p.idle = parseHIDIdle(string(out))
	}
	return p.idle
}

func parseHIDIdle(out string) time.Duration {
	for line := range strings.Lines(out) {
		_, v, ok := strings.Cut(line, `"HIDIdleTime" = `)
		if !ok {
			continue
		}
		ns, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err == nil {
			return time.Duration(ns)
		}
	}
	return 0
}
