package bar

import (
	"bufio"
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
	cpuItem    = "cpu"
	memoryItem = "memory"
)

var pageSizeRe = regexp.MustCompile(`page size of (\d+) bytes`)

type System struct {
	runner   Runner
	ncpu     int
	memBytes uint64
}

func NewSystem(runner Runner) *System {
	return &System{runner: runner, ncpu: runtime.NumCPU()}
}

func (s *System) Setup(ctx context.Context, b *Batch) error {
	b.Add("item", memoryItem, "right", P("icon", iconMemory), "click_script=open -a 'Activity Monitor'")
	b.Add("item", cpuItem, "right", P("icon", iconCPU), "click_script=open -a 'Activity Monitor'")
	return nil
}

func (s *System) Interval() time.Duration {
	return 5 * time.Second
}

func (s *System) Poll(ctx context.Context) Update {
	cpu, cpuErr := s.cpuPercent(ctx)
	if cpuErr != nil {
		slog.Debug("bar: cpu lookup failed", "error", cpuErr)
	}
	mem, memErr := s.memoryPercent(ctx)
	if memErr != nil {
		slog.Debug("bar: memory lookup failed", "error", memErr)
	}
	return func(b *Batch) {
		if cpuErr == nil {
			b.Set(cpuItem, P("label", fmt.Sprintf("%.0f%%", cpu)))
		}
		if memErr == nil {
			b.Set(memoryItem, P("label", fmt.Sprintf("%.0f%%", mem)))
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
	if ncpu <= 0 {
		return 0, fmt.Errorf("invalid cpu count %d", ncpu)
	}
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
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		k, v, ok := strings.Cut(scanner.Text(), ":")
		if !ok {
			continue
		}
		n, err := strconv.ParseUint(strings.TrimSuffix(strings.TrimSpace(v), "."), 10, 64)
		if err != nil {
			continue
		}
		pages[strings.TrimSpace(k)] = n
	}

	app := pages["Anonymous pages"]
	if purgeable := pages["Pages purgeable"]; purgeable < app {
		app -= purgeable
	}
	used := app + pages["Pages wired down"] + pages["Pages occupied by compressor"]
	return used * pageSize, nil
}
