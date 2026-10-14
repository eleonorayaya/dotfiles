package bar

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	networkItem      = "network"
	routeRefreshPoll = 10
)

type ifaceCounters struct {
	in, out uint64
}

type Network struct {
	runner   Runner
	colors   Colors
	mood     *Mood
	now      func() time.Time
	polls    int
	iface    string
	icon     string
	last     ifaceCounters
	lastSeen time.Time
}

func NewNetwork(runner Runner, colors Colors, mood *Mood) *Network {
	return &Network{runner: runner, colors: colors, mood: mood, now: time.Now}
}

func (n *Network) Setup(ctx context.Context, b *Batch) error {
	b.Add(networkItem, "right", append(island(n.colors),
		P("click_script", "open 'x-apple.systempreferences:com.apple.Network-Settings.extension'"))...)
	return nil
}

func (n *Network) Interval() time.Duration {
	return 3 * time.Second
}

func (n *Network) Poll(ctx context.Context) Update {
	if n.iface == "" || n.polls%routeRefreshPoll == 0 {
		n.resolveInterface(ctx)
	}
	n.polls++
	n.mood.SetOnline(n.iface != "")
	if n.iface == "" {
		return func(b *Batch) {
			b.Set(networkItem, P("icon", iconOffline), "label=offline")
		}
	}

	counters, err := n.counters(ctx)
	if err != nil {
		slog.Debug("bar: network counters failed", "error", err)
		n.iface = ""
		return nil
	}

	now := n.now()
	label := ""
	if !n.lastSeen.IsZero() {
		secs := now.Sub(n.lastSeen).Seconds()
		label = fmt.Sprintf("↓%s ↑%s", formatRate(delta(counters.in, n.last.in), secs), formatRate(delta(counters.out, n.last.out), secs))
	}
	n.last, n.lastSeen = counters, now

	icon := n.icon
	return func(b *Batch) {
		b.Set(networkItem, P("icon", icon), P("label", label))
	}
}

func (n *Network) resolveInterface(ctx context.Context) {
	iface := ""
	if out, err := n.runner.Run(ctx, "route", "-n", "get", "default"); err == nil {
		iface = parseFields(string(out))["interface"]
	}
	if iface == n.iface {
		return
	}
	n.iface, n.icon, n.lastSeen = iface, iconEthernet, time.Time{}
	if iface == "" {
		return
	}
	if out, err := n.runner.Run(ctx, "ipconfig", "getsummary", iface); err == nil && parseFields(string(out))["InterfaceType"] == "WiFi" {
		n.icon = iconWifi
	}
}

func (n *Network) counters(ctx context.Context) (ifaceCounters, error) {
	out, err := n.runner.Run(ctx, "netstat", "-ibn", "-I", n.iface)
	if err != nil {
		return ifaceCounters{}, err
	}
	return parseNetstat(string(out), n.iface)
}

func parseFields(out string) map[string]string {
	fields := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		if k, v, ok := strings.Cut(scanner.Text(), ":"); ok {
			fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return fields
}

func parseNetstat(out, iface string) (ifaceCounters, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return ifaceCounters{}, fmt.Errorf("no netstat rows for %s", iface)
	}
	header := strings.Fields(lines[0])
	inCol, outCol := slices.Index(header, "Ibytes"), slices.Index(header, "Obytes")
	if inCol < 0 || outCol < 0 {
		return ifaceCounters{}, fmt.Errorf("unexpected netstat header: %q", lines[0])
	}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < len(header)-1 || fields[0] != iface || !strings.HasPrefix(fields[2], "<Link#") {
			continue
		}
		inField := fields[len(fields)-(len(header)-inCol)]
		outField := fields[len(fields)-(len(header)-outCol)]
		in, err := strconv.ParseUint(inField, 10, 64)
		if err != nil {
			return ifaceCounters{}, fmt.Errorf("bad Ibytes %q: %w", inField, err)
		}
		out, err := strconv.ParseUint(outField, 10, 64)
		if err != nil {
			return ifaceCounters{}, fmt.Errorf("bad Obytes %q: %w", outField, err)
		}
		return ifaceCounters{in: in, out: out}, nil
	}
	return ifaceCounters{}, fmt.Errorf("no link row for %s", iface)
}

func delta(cur, prev uint64) uint64 {
	if cur < prev {
		return 0
	}
	return cur - prev
}

func formatRate(bytes uint64, secs float64) string {
	if secs <= 0 {
		return "0B"
	}
	rate := float64(bytes) / secs
	units := []string{"B", "K", "M", "G"}
	i := 0
	for rate >= 1000 && i < len(units)-1 {
		rate /= 1000
		i++
	}
	if i == 0 || rate >= 10 {
		return fmt.Sprintf("%.0f%s", rate, units[i])
	}
	return fmt.Sprintf("%.1f%s", rate, units[i])
}
