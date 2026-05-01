package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/config"
)

func TestLoad_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "" {
		t.Errorf("expected empty profile, got %q", cfg.Profile)
	}
}

func TestLoad_ProfileFromFile(t *testing.T) {
	path := writeConfig(t, "profile: work\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("expected %q, got %q", "work", cfg.Profile)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	t.Setenv("SHIZUKU_PROFILE", "personal")
	path := writeConfig(t, "profile: work\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "personal" {
		t.Errorf("expected %q, got %q", "personal", cfg.Profile)
	}
}

func TestLoad_EnvWithNoFile(t *testing.T) {
	t.Setenv("SHIZUKU_PROFILE", "work")
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("expected %q, got %q", "work", cfg.Profile)
	}
}

func TestGet_KnownKey(t *testing.T) {
	cfg := &config.Config{Profile: "work"}
	val, err := config.Get(cfg, "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "work" {
		t.Errorf("expected %q, got %q", "work", val)
	}
}

func TestGet_UnknownKey(t *testing.T) {
	cfg := &config.Config{}
	_, err := config.Get(cfg, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestSet_WritesValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	if err := config.Set(path, "profile", "work"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("expected %q, got %q", "work", cfg.Profile)
	}
}

func TestSet_UnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	err := config.Set(path, "nonexistent", "val")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestSet_DoesNotWriteEnvValues(t *testing.T) {
	t.Setenv("SHIZUKU_PROFILE", "env-value")
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	if err := config.Set(path, "profile", "file-value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "profile: file-value\n" {
		t.Errorf("unexpected file contents: %q", string(data))
	}
}

func TestYAML_AllFields(t *testing.T) {
	cfg := &config.Config{Profile: "work"}
	out, err := cfg.YAML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "profile: work\n" {
		t.Errorf("unexpected yaml: %q", out)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
