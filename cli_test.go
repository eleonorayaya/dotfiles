package shizuku_test

import (
	"os"
	"path/filepath"
	"testing"

	shizuku "github.com/eleonorayaya/shizuku"
)

func writeProfileFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, "profile")
	if err := os.WriteFile(path, []byte(name+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolveProfile_FlagWins(t *testing.T) {
	dir := t.TempDir()
	profileFile := writeProfileFile(t, dir, "personal")

	got, err := shizuku.ResolveProfile("work", "personal", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "work" {
		t.Errorf("expected %q, got %q", "work", got)
	}
}

func TestResolveProfile_EnvWinsOverFile(t *testing.T) {
	dir := t.TempDir()
	profileFile := writeProfileFile(t, dir, "personal")

	got, err := shizuku.ResolveProfile("", "work", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "work" {
		t.Errorf("expected %q, got %q", "work", got)
	}
}

func TestResolveProfile_FileUsedWhenNoFlagOrEnv(t *testing.T) {
	dir := t.TempDir()
	profileFile := writeProfileFile(t, dir, "personal")

	got, err := shizuku.ResolveProfile("", "", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "personal" {
		t.Errorf("expected %q, got %q", "personal", got)
	}
}

func TestResolveProfile_EmptyWhenNothingSet(t *testing.T) {
	dir := t.TempDir()
	profileFile := filepath.Join(dir, "profile")

	got, err := shizuku.ResolveProfile("", "", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestResolveProfile_FileWhitespaceStripped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile")
	if err := os.WriteFile(path, []byte("  work  \n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := shizuku.ResolveProfile("", "", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "work" {
		t.Errorf("expected %q, got %q", "work", got)
	}
}
