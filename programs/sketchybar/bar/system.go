package bar

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	cpuItem      = "cpu"
	cpuGraphItem = "cpu.graph"
	memoryItem   = "memory"
	memoryBar    = "memory.bar"
	graphWidth   = 44
	meterWidth   = 44
	meterHeight  = 8
)

var pageSizeRe = regexp.MustCompile(`page size of (\d+) bytes`)

const systemBracket = "system"

type System struct {
	runner   Runner
	colors   Colors
	mood     *Mood
	ncpu     int
	memBytes uint64
}

func NewSystem(runner Runner, colors Colors, mood *Mood) *System {
	return &System{runner: runner, colors: colors, mood: mood, ncpu: runtime.NumCPU()}
}

func (s *System) Setup(ctx context.Context, b *Batch) error {
	c := s.colors
	click := "click_script=open -a 'Activity Monitor'"
	b.Raw("--add", "slider", memoryBar, "right", strconv.Itoa(meterWidth))
	b.Set(memoryBar,
		P("slider.width", meterWidth),
		P("slider.highlight_color", c.Hot),
		P("slider.background.color", c.IslandBorder),
		P("slider.background.height", meterHeight),
		P("slider.background.corner_radius", meterHeight/2),
		"slider.knob.drawing=off",
		"icon.drawing=off",
		"label.drawing=off",
		P("padding_left", 0),
		P("padding_right", itemPadding+islandPadding),
		click,
	)
	b.Add(memoryItem, "right", P("icon", iconMemory), "label.drawing=off", "icon.padding_left=0", "icon.padding_right=6", "padding_left=0", "padding_right=0", click)
	b.Raw("--add", "graph", cpuGraphItem, "right", strconv.Itoa(graphWidth))
	b.Set(cpuGraphItem,
		P("graph.color", c.Hot),
		P("graph.fill_color", faded(c.Hot)),
		"graph.line_width=1.5",
		"background.drawing=on",
		"background.color=0x01000000",
		"background.height=16",
		"icon.drawing=off",
		"label.drawing=off",
		"padding_left=0",
		"padding_right=12",
		click,
	)
	b.Add(cpuItem, "right", P("icon", iconCPU), "label.drawing=off", P("icon.padding_left", islandPadding), "icon.padding_right=6", P("padding_left", itemPadding), "padding_right=0", click)
	b.Raw("--add", "bracket", systemBracket, cpuItem, cpuGraphItem, memoryItem, memoryBar)
	b.Set(systemBracket, bracketIsland(s.colors)...)
	return nil
}

func (s *System) Interval() time.Duration {
	return 3 * time.Second
}

func (s *System) Poll(ctx context.Context) Update {
	cpu, cpuErr := s.cpuPercent(ctx)
	if cpuErr != nil {
		slog.Debug("bar: cpu lookup failed", "error", cpuErr)
	} else {
		s.mood.SetCPU(cpu)
	}
	mem, memErr := s.memoryPercent(ctx)
	if memErr != nil {
		slog.Debug("bar: memory lookup failed", "error", memErr)
	}
	return func(b *Batch) {
		if cpuErr == nil {
			b.Raw("--push", cpuGraphItem, fmt.Sprintf("%.3f", cpu/100))
		}
		if memErr == nil {
			b.Set(memoryBar, P("slider.percentage", fmt.Sprintf("%.0f", mem)))
		}
	}
}

func (s *System) cpuPercent(ctx context.Context) (float64, error) {
	out, err := s.runner.Run(ctx, "ps", "-A", "-o", "%cpu=")
	if err != nil {
		return 0, err
	}
	return parseCPU(string(out), s.ncpu)
}

func (s *System) memoryPercent(ctx context.Context) (float64, error) {
	if s.memBytes == 0 {
		out, err := s.runner.Run(ctx, "sysctl", "-n", "hw.memsize")
		if err != nil {
			return 0, err
		}
		total, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("bad hw.memsize: %w", err)
		}
		s.memBytes = total
	}
	out, err := s.runner.Run(ctx, "vm_stat")
	if err != nil {
		return 0, err
	}
	used, err := parseVMStatUsed(string(out))
	if err != nil {
		return 0, err
	}
	return float64(used) / float64(s.memBytes) * 100, nil
}

func parseCPU(out string, ncpu int) (float64, error) {
	total := 0.0
	for field := range strings.FieldsSeq(out) {
		v, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return 0, fmt.Errorf("bad cpu value %q: %w", field, err)
		}
		total += v
	}
	return min(total/float64(ncpu), 100), nil
}

func parseVMStatUsed(out string) (uint64, error) {
	m := pageSizeRe.FindStringSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("vm_stat: page size not found")
	}
	pageSize, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("vm_stat: bad page size: %w", err)
	}

	pages := map[string]uint64{}
	for k, v := range parseFields(out) {
		if n, err := strconv.ParseUint(strings.TrimSuffix(v, "."), 10, 64); err == nil {
			pages[k] = n
		}
	}

	app := pages["Anonymous pages"]
	if purgeable := pages["Pages purgeable"]; purgeable < app {
		app -= purgeable
	}
	used := app + pages["Pages wired down"] + pages["Pages occupied by compressor"]
	return used * pageSize, nil
}
