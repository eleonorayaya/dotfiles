package terminal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureLine_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	if err := ensureLine(path, "source ~/.config/shizuku/shizuku.sh", "shizuku/shizuku.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "source ~/.config/shizuku/shizuku.sh\n" {
		t.Errorf("got %q", got)
	}
}

func TestEnsureLine_AppendsWithNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	if err := os.WriteFile(path, []byte("export FOO=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureLine(path, "source ~/.config/shizuku/shizuku.sh", "shizuku/shizuku.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "export FOO=1\nsource ~/.config/shizuku/shizuku.sh\n" {
		t.Errorf("got %q", got)
	}
}

func TestEnsureLine_RespectsExistingVariant(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	existing := `source "$HOME/.config/shizuku/shizuku.sh"` + "\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureLine(path, "source ~/.config/shizuku/shizuku.sh", "shizuku/shizuku.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != existing {
		t.Errorf("file changed: %q", got)
	}
}

func TestEnsureLine_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshenv")
	for range 3 {
		if err := ensureLine(path, "source ~/.config/shizuku/shizuku.zshenv", "shizuku/shizuku.zshenv"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	got, _ := os.ReadFile(path)
	if strings.Count(string(got), "shizuku.zshenv") != 1 {
		t.Errorf("got %q", got)
	}
}
