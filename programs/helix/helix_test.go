package helix

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSteelInstallSkipsWhenInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	bin := filepath.Join(home, ".cargo", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "hx"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := New(Options{Steel: true}).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "src", "helix")); !os.IsNotExist(err) {
		t.Errorf("expected no clone, got err %v", err)
	}
}
