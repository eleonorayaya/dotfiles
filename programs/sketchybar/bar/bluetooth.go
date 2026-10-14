package bar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

const (
	bluetoothItem         = "bluetooth"
	bluetoothDevicePrefix = "bluetooth.device."
)

type bluetoothDevice struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
}

type Bluetooth struct {
	runner  Runner
	emitCmd string
	devices []bluetoothDevice
}

func NewBluetooth(runner Runner, emitCmd string) *Bluetooth {
	return &Bluetooth{runner: runner, emitCmd: emitCmd}
}

func (bt *Bluetooth) Setup(ctx context.Context, b *Batch) error {
	b.Add("item", bluetoothItem, "right", P("icon", iconBluetooth), "label.drawing=off", "popup.align=right")
	OnClick(b, bluetoothItem, bt.emitCmd)
	b.Subscribe(bluetoothItem, "mouse.exited.global")
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
	icon := iconBluetoothOff
	if strings.TrimSpace(string(out)) == "1" {
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
	switch {
	case ev.Sender == "mouse.exited.global":
		b.Set(bluetoothItem, "popup.drawing=off")
	case ev.Name == bluetoothItem:
		if bt.popupOpen(ctx) {
			b.Set(bluetoothItem, "popup.drawing=off")
			return
		}
		bt.renderDevices(ctx, b)
		b.Set(bluetoothItem, "popup.drawing=on")
	case strings.HasPrefix(ev.Name, bluetoothDevicePrefix):
		bt.toggleDevice(ctx, strings.TrimPrefix(ev.Name, bluetoothDevicePrefix))
		bt.renderDevices(ctx, b)
	}
}

func (bt *Bluetooth) popupOpen(ctx context.Context) bool {
	out, err := bt.runner.Run(ctx, "sketchybar", "--query", bluetoothItem)
	if err != nil {
		return false
	}
	var item struct {
		Popup struct {
			Drawing string `json:"drawing"`
		} `json:"popup"`
	}
	if err := json.Unmarshal(out, &item); err != nil {
		return false
	}
	return item.Popup.Drawing == "on"
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
	devices, err := parseBluetoothDevices(out)
	if err != nil {
		slog.Warn("bar: bluetooth device parse failed", "error", err)
		return
	}
	bt.devices = devices

	b.Remove("/" + strings.ReplaceAll(bluetoothDevicePrefix, ".", `\.`) + ".*/")
	for i, d := range devices {
		name := fmt.Sprintf("%s%d", bluetoothDevicePrefix, i)
		icon := iconBluetooth
		if d.Connected {
			icon = iconBluetoothConn
		}
		b.Add("item", name, "popup."+bluetoothItem,
			P("icon", icon),
			P("label", d.Name),
			onOff("icon.highlight", d.Connected),
			onOff("label.highlight", d.Connected),
		)
		OnClick(b, name, bt.emitCmd)
	}
}

func parseBluetoothDevices(data []byte) ([]bluetoothDevice, error) {
	var devices []bluetoothDevice
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("failed to parse blueutil output: %w", err)
	}
	return devices, nil
}
