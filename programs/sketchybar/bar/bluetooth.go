package bar

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	bluetoothItem         = "bluetooth"
	bluetoothDevicePrefix = "bluetooth.device."
	bluetoothDevicesRegex = `/bluetooth\.device\..*/`
)

type bluetoothDevice struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
}

type Bluetooth struct {
	runner  Runner
	power   lastValue
	mu      sync.Mutex
	open    bool
	devices []bluetoothDevice
}

func NewBluetooth(runner Runner) *Bluetooth {
	return &Bluetooth{runner: runner}
}

func (bt *Bluetooth) Setup(ctx context.Context, b *Batch) error {
	b.Add(bluetoothItem, "right", P("icon", iconBluetooth), "label.drawing=off", "popup.align=right")
	b.OnClick(bluetoothItem, "mouse.exited.global")
	return nil
}

func (bt *Bluetooth) Interval() time.Duration {
	return 30 * time.Second
}

func (bt *Bluetooth) Poll(ctx context.Context) Update {
	out, err := bt.runner.Run(ctx, "blueutil", "--power")
	if err != nil {
		slog.Debug("bar: bluetooth power lookup failed", "error", err)
		return nil
	}
	power := strings.TrimSpace(string(out))
	if !bt.power.changed(power) {
		return nil
	}
	icon := iconBluetoothOff
	if power == "1" {
		icon = iconBluetooth
	}
	return func(b *Batch) {
		b.Set(bluetoothItem, P("icon", icon))
	}
}

func (bt *Bluetooth) Prefix() string {
	return bluetoothItem
}

func (bt *Bluetooth) Click(ctx context.Context, ev Event, b *Batch) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	switch {
	case ev.Name == bluetoothItem && ev.Sender == "mouse.exited.global":
		if bt.open {
			bt.setOpen(b, false)
		}
	case ev.Name == bluetoothItem && ev.Sender == "mouse.clicked":
		if !bt.open {
			bt.renderDevices(ctx, b)
		}
		bt.setOpen(b, !bt.open)
	case strings.HasPrefix(ev.Name, bluetoothDevicePrefix) && ev.Sender == "mouse.clicked":
		bt.toggleDevice(ctx, strings.TrimPrefix(ev.Name, bluetoothDevicePrefix))
		bt.renderDevices(ctx, b)
	}
}

func (bt *Bluetooth) setOpen(b *Batch, open bool) {
	bt.open = open
	b.Set(bluetoothItem, onOff("popup.drawing", open))
}

func (bt *Bluetooth) toggleDevice(ctx context.Context, index string) {
	i, err := strconv.Atoi(index)
	if err != nil || i < 0 || i >= len(bt.devices) {
		return
	}
	d := bt.devices[i]
	action := "--connect"
	if d.Connected {
		action = "--disconnect"
	}
	if _, err := bt.runner.Run(ctx, "blueutil", action, d.Address); err != nil {
		slog.Warn("bar: bluetooth toggle failed", "device", d.Name, "error", err)
	}
}

func (bt *Bluetooth) renderDevices(ctx context.Context, b *Batch) {
	out, err := bt.runner.Run(ctx, "blueutil", "--paired", "--format", "json")
	if err != nil {
		slog.Warn("bar: bluetooth device lookup failed", "error", err)
		return
	}
	devices, err := decodeJSON[[]bluetoothDevice](out, "blueutil output")
	if err != nil {
		slog.Warn("bar: bluetooth device lookup failed", "error", err)
		return
	}
	bt.devices = devices

	b.Remove(bluetoothDevicesRegex)
	for i, d := range devices {
		name := fmt.Sprintf("%s%d", bluetoothDevicePrefix, i)
		icon := iconBluetooth
		if d.Connected {
			icon = iconBluetoothConn
		}
		props := []string{P("icon", icon), P("label", d.Name)}
		b.Add(name, "popup."+bluetoothItem, append(props, highlight(d.Connected)...)...)
		b.OnClick(name)
	}
}
