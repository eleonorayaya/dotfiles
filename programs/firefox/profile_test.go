package firefox

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultProfileDirPrefersInstallDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "installs.ini"), "[2656FF1E876E9973]\nDefault=Profiles/abc.Main\nLocked=1\n")
	writeFile(t, filepath.Join(dir, "profiles.ini"), "[Profile0]\nName=default\nIsRelative=1\nPath=Profiles/xyz.default\nDefault=1\n")

	got, err := defaultProfileDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "Profiles/abc.Main"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultProfileDirFallsBackToProfilesDefault(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "profiles.ini"), "[General]\nVersion=2\n\n[Profile1]\nPath=Profiles/one\n\n[Profile0]\nPath=Profiles/two\nDefault=1\n")

	got, err := defaultProfileDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "Profiles/two"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultProfileDirWithoutFirefox(t *testing.T) {
	got, err := defaultProfileDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected no profile, got %q", got)
	}
}
