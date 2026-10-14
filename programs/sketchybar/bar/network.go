package bar

import (
	"bufio"
	"context"
	"strings"
	"time"
)

const networkItem = "network"

type Network struct {
	runner Runner
	colors Colors
	mood   *Mood
	iface  string
	icon   string
	last   lastValue
}

func NewNetwork(runner Runner, colors Colors, mood *Mood) *Network {
	return &Network{runner: runner, colors: colors, mood: mood}
}

func (n *Network) Setup(ctx context.Context, b *Batch) error {
	b.Add(networkItem, "right",
		P("icon", iconOffline),
		P("icon.padding_left", islandPadding),
		"icon.padding_right=4",
		"label.drawing=off",
		P("padding_left", itemPadding),
		"padding_right=0",
		P("click_script", "open 'x-apple.systempreferences:com.apple.Network-Settings.extension'"),
	)
	b.Raw("--add", "bracket", connectBracket, networkItem, bluetoothItem)
	b.Set(connectBracket, bracketIsland(n.colors, smallIsland)...)
	return nil
}

func (n *Network) Interval() time.Duration {
	return 5 * time.Second
}

func (n *Network) Poll(ctx context.Context) Update {
	n.resolveInterface(ctx)
	n.mood.SetOnline(n.iface != "")
	icon := n.icon
	if n.iface == "" {
		icon = iconOffline
	}
	if !n.last.changed(icon) {
		return nil
	}
	return func(b *Batch) {
		b.Set(networkItem, P("icon", icon))
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
	n.iface, n.icon = iface, iconEthernet
	if iface == "" {
		return
	}
	if out, err := n.runner.Run(ctx, "ipconfig", "getsummary", iface); err == nil && parseFields(string(out))["InterfaceType"] == "WiFi" {
		n.icon = iconWifi
	}
}

func (n *Network) resetCache() {
	n.last.reset()
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
