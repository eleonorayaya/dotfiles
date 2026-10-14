package bar

import (
	"context"
	"fmt"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return out, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

type Batch struct {
	args []string
}

func P(key string, value any) string {
	return fmt.Sprintf("%s=%v", key, value)
}

func (b *Batch) Add(kind, name, position string, props ...string) *Batch {
	b.args = append(b.args, "--add", kind, name, position)
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

func (b *Batch) Remove(name string) *Batch {
	b.args = append(b.args, "--remove", name)
	return b
}

func (b *Batch) Raw(args ...string) *Batch {
	b.args = append(b.args, args...)
	return b
}

func (b *Batch) Empty() bool {
	return len(b.args) == 0
}

func (b *Batch) Args() []string {
	return b.args
}

func (b *Batch) Apply(ctx context.Context, r Runner) error {
	if b.Empty() {
		return nil
	}
	if _, err := r.Run(ctx, "sketchybar", b.args...); err != nil {
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
