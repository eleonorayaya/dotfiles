package shizuku

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
)

type fakeProgram struct {
	name string
	tag  string
}

func (f fakeProgram) Name() string { return f.name }

func programNames(p Profile) []string {
	out := []string{}
	for _, prog := range p.Programs {
		out = append(out, prog.Name())
	}
	return out
}

func newTestBuilder(profile string) *Builder {
	return New(
		WithProfileName(profile),
		WithPrograms(fakeProgram{name: "git"}, fakeProgram{name: "notion", tag: "base"}),
		WithProfile("desktop",
			WithPrograms(fakeProgram{name: "kitty"}),
		),
		WithProfile("work",
			Extends("desktop"),
			WithPrograms(fakeProgram{name: "k9s"}, fakeProgram{name: "notion", tag: "work"}),
		),
		WithProfile("loop-a", Extends("loop-b")),
		WithProfile("loop-b", Extends("loop-a")),
		WithProfile("orphan", Extends("missing")),
	)
}

func TestActiveProfile_BaseWhenUnset(t *testing.T) {
	p, err := newTestBuilder("").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(programNames(p), ","); got != "git,notion" {
		t.Errorf("got %q", got)
	}
}

func TestActiveProfile_SingleLayer(t *testing.T) {
	p, err := newTestBuilder("desktop").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(programNames(p), ","); got != "git,notion,kitty" {
		t.Errorf("got %q", got)
	}
}

func TestActiveProfile_ExtendsChain(t *testing.T) {
	p, err := newTestBuilder("work").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(programNames(p), ","); got != "git,notion,kitty,k9s" {
		t.Errorf("got %q", got)
	}
}

func TestActiveProfile_LaterLayerOverridesByName(t *testing.T) {
	p, err := newTestBuilder("work").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, prog := range p.Programs {
		if prog.Name() == "notion" && prog.(fakeProgram).tag != "work" {
			t.Errorf("notion not overridden, tag=%q", prog.(fakeProgram).tag)
		}
	}
}

func TestActiveProfile_UnknownProfile(t *testing.T) {
	if _, err := newTestBuilder("nope").activeProfile(); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestActiveProfile_UnknownParent(t *testing.T) {
	if _, err := newTestBuilder("orphan").activeProfile(); err == nil {
		t.Fatal("expected error for unknown parent profile")
	}
}

func TestActiveProfile_Cycle(t *testing.T) {
	_, err := newTestBuilder("loop-a").activeProfile()
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

var _ app.Program = fakeProgram{}

type failingInstaller struct {
	name  string
	calls *[]string
}

func (f failingInstaller) Name() string { return f.name }

func (f failingInstaller) Install(ctx *app.Context) error {
	*f.calls = append(*f.calls, f.name)
	return fmt.Errorf("%s broke", f.name)
}

func TestInstall_ContinuesAndJoinsErrors(t *testing.T) {
	calls := []string{}
	b := New(
		WithOutDir(t.TempDir()),
		WithPrograms(failingInstaller{name: "a", calls: &calls}, failingInstaller{name: "b", calls: &calls}),
	)
	err := b.Install(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Join(calls, ",") != "a,b" {
		t.Errorf("expected both installers to run, got %v", calls)
	}
	if !strings.Contains(err.Error(), "a broke") || !strings.Contains(err.Error(), "b broke") {
		t.Errorf("expected both errors, got %v", err)
	}
}
