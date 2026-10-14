package bar

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	commandTimeout = 10 * time.Second
	bounceFrames   = "14"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return out, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return out, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

type Batch struct {
	bin  string
	emit string
	args []string
}

func P(key string, value any) string {
	return fmt.Sprintf("%s=%v", key, value)
}

func (b *Batch) Add(name, position string, props ...string) *Batch {
	b.args = append(b.args, "--add", "item", name, position)
	return b.Set(name, props...)
}

func (b *Batch) Set(name string, props ...string) *Batch {
	if len(props) == 0 {
		return b
	}
	b.args = append(b.args, "--set", name)
	b.args = append(b.args, props...)
	return b
}

func (b *Batch) Subscribe(name string, events ...string) *Batch {
	if len(events) == 0 {
		return b
	}
	b.args = append(b.args, "--subscribe", name)
	b.args = append(b.args, events...)
	return b
}

func (b *Batch) OnClick(name string, events ...string) *Batch {
	return b.Listen(name, append([]string{"mouse.clicked"}, events...)...)
}

func (b *Batch) Listen(name string, events ...string) *Batch {
	b.Set(name, P("script", b.emit))
	return b.Subscribe(name, events...)
}

func (b *Batch) Bounce(name string, height int) *Batch {
	return b.Raw("--animate", "sin", bounceFrames, "--set", name, P("y_offset", height), "y_offset=0")
}

func (b *Batch) Remove(name string) *Batch {
	b.args = append(b.args, "--remove", name)
	return b
}

func (b *Batch) Raw(args ...string) *Batch {
	b.args = append(b.args, args...)
	return b
}

func (b *Batch) Args() []string {
	return b.args
}

func (b *Batch) apply(ctx context.Context, r Runner) error {
	if len(b.args) == 0 {
		return nil
	}
	if _, err := r.Run(ctx, b.bin, b.args...); err != nil {
		return fmt.Errorf("failed to apply sketchybar batch: %w", err)
	}
	return nil
}

func onOff(key string, on bool) string {
	if on {
		return key + "=on"
	}
	return key + "=off"
}

func highlight(on bool) []string {
	return []string{onOff("icon.highlight", on), onOff("label.highlight", on)}
}
