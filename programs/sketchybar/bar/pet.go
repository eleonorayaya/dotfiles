package bar

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	petItem           = "petsprite"
	petZzzItem        = "petzzz"
	petBracket        = "petroom"
	petPixel          = 5
	petIdleNap        = 5 * time.Minute
	petHopFor         = 4 * time.Second
	petTick           = time.Second
	petBlinkGap       = 6
	petRoam           = 72
	petStroll         = 8
	petHome           = petRoam / 2
	petWalkSteps      = 3
	petStepGap        = 250 * time.Millisecond
	petWalkFrames     = 45
	petHopHeight      = 10
	petNightUntilHour = 6
	petIdleCheckTicks = 10
	petImageScale     = "0.5"
)

var (
	petEdge = itemPadding + petRoomSize.pad
	petSide = (len(petBase) + 2) * petPixel / 2
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
	lookRight := withRows(petBase, map[int]string{5: "cccbbccccbbc", 6: "cccwbccccwbc"})
	stepA := withRows(lookRight, map[int]string{9: ".R..cccc..R.", 10: "RR..cccc..RR", 11: "...c....c..."})
	stepB := withRows(lookRight, map[int]string{9: "..R.cccc.R..", 10: ".RR.cccc.RR.", 11: "....c..c...."})
	return map[string][]string{
		"awake":  petBase,
		"blink":  closed,
		"sleep":  withRows(closed, map[int]string{7: "ckcccccccckc"}),
		"wave":   withRows(petBase, map[int]string{8: "RccccccccccR", 9: "R.R.cccc.R.R", 10: "....cccc...."}),
		"walk1r": stepA,
		"walk2r": stepB,
		"walk1l": mirrored(stepA),
		"walk2l": mirrored(stepB),
	}
}

func mirrored(rows []string) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		b := []byte(row)
		slices.Reverse(b)
		out[i] = string(b)
	}
	return out
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

type Pet struct {
	runner Runner
	colors Colors
	shelf  *Shelf
	frames map[string]string
	idle   time.Duration
}

func NewPet(runner Runner, colors Colors, shelf *Shelf) *Pet {
	return &Pet{runner: runner, colors: colors, shelf: shelf}
}

func petPadding(pos int) []string {
	return []string{P("padding_left", petEdge+pos), P("padding_right", petEdge+petRoam-pos)}
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
	b.Add(petItem, "center", append([]string{
		"background.drawing=on",
		P("background.color", transparent),
		P("background.image", frames["awake"]),
		P("background.image.scale", petImageScale),
		P("width", petSide),
		"icon.drawing=off",
		"label.drawing=off",
	}, petPadding(petHome)...)...)
	b.Add(petZzzItem, "center",
		"icon.drawing=off",
		"label=z z ₊",
		P("label.font", font("ExtraBold", 11)),
		P("label.color", p.colors.Hot),
		"drawing=off",
		"y_offset=4",
		"label.padding_left=0",
		P("label.padding_right", petRoomSize.pad),
		"padding_left=0",
		"padding_right=0",
	)
	b.Raw("--add", "bracket", petBracket, petItem, petZzzItem)
	b.Set(petBracket, bracketIsland(p.colors, petRoomSize)...)
	return nil
}

func (p *Pet) Stream(ctx context.Context, updates chan<- Update) {
	ticks, lastFrame, hopped, pos := 0, "", time.Time{}, petHome
	for ctx.Err() == nil {
		ticks++
		urgent, finishedAt := p.shelf.snapshot()
		frame := p.frameFor(ctx, ticks, urgent, time.Now())
		hop := time.Since(finishedAt) < petHopFor && finishedAt.After(hopped)
		if hop {
			hopped = finishedAt
		}
		if frame != "sleep" && !urgent && ticks%petStroll == 0 {
			if target := rand.IntN(petRoam + 1); target != pos {
				p.stroll(ctx, updates, pos, target, frame)
				pos, lastFrame = target, frame
			}
		}
		if frame != lastFrame || hop {
			lastFrame = frame
			send(ctx, updates, func(b *Batch) {
				b.Set(petItem, P("background.image", p.frames[frame]))
				b.Set(petZzzItem, onOff("drawing", frame == "sleep"))
				if hop {
					b.Bounce(petItem, petHopHeight)
				}
			})
		}
		sleep(ctx, petTick)
	}
}

func (p *Pet) stroll(ctx context.Context, updates chan<- Update, from, to int, rest string) {
	dir := "r"
	if to < from {
		dir = "l"
	}
	for step := range petWalkSteps {
		frame := fmt.Sprintf("walk%d%s", step%2+1, dir)
		send(ctx, updates, func(b *Batch) {
			b.Set(petItem, P("background.image", p.frames[frame]))
			if step == 0 {
				b.Raw(append([]string{"--animate", "linear", strconv.Itoa(petWalkFrames), "--set", petItem}, petPadding(to)...)...)
			}
		})
		sleep(ctx, petStepGap)
	}
	send(ctx, updates, func(b *Batch) {
		b.Set(petItem, P("background.image", p.frames[rest]))
	})
}

func (p *Pet) frameFor(ctx context.Context, tick int, urgent bool, now time.Time) string {
	switch {
	case urgent:
		if tick%2 == 0 {
			return "wave"
		}
		return "awake"
	case now.Hour() < petNightUntilHour || p.userIdle(ctx, tick) >= petIdleNap:
		return "sleep"
	case tick%petBlinkGap == 0:
		return "blink"
	default:
		return "awake"
	}
}

func (p *Pet) userIdle(ctx context.Context, tick int) time.Duration {
	if tick%petIdleCheckTicks != 1 {
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
