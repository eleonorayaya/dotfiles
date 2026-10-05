package pkg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func releaseServer(t *testing.T, assets map[string][]byte) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			list := []map[string]string{}
			for name := range assets {
				list = append(list, map[string]string{"name": name, "browser_download_url": srv.URL + "/dl/" + name})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "v1.0.0", "assets": list})
			return
		}
		body, ok := assets[strings.TrimPrefix(r.URL.Path, "/dl/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestReleaseInstall_Archive(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := releaseServer(t, map[string][]byte{
		"lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz": makeTarGz(t, map[string]string{
			"lsd-v1.2.0-x86_64-unknown-linux-gnu/lsd":       "LSD",
			"lsd-v1.2.0-x86_64-unknown-linux-gnu/README.md": "readme",
		}),
	})
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "lsd", Release: &Release{Repo: "lsd-rs/lsd", Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFile(t, filepath.Join(i.BinDir(), "lsd"), "LSD")
	info, _ := os.Stat(filepath.Join(i.BinDir(), "lsd"))
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("binary not executable: %v", info.Mode())
	}
}

func TestReleaseInstall_RawBinary(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := releaseServer(t, map[string][]byte{
		"posh-linux-amd64": []byte("POSH"),
		"posh-linux-arm64": []byte("WRONG"),
	})
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "oh-my-posh", Release: &Release{
		Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`,
		Arch: map[string]string{"amd64": "amd64", "arm64": "arm64"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(i.BinDir(), "oh-my-posh"), "POSH")
}

func TestReleaseInstall_CopiesDirs(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := releaseServer(t, map[string][]byte{
		"helix-25.07.1-x86_64-linux.tar.gz": makeTarGz(t, map[string]string{
			"helix-25.07.1-x86_64-linux/hx":                        "HX",
			"helix-25.07.1-x86_64-linux/runtime/themes/x.toml":     "theme",
			"helix-25.07.1-x86_64-linux/runtime/queries/go/hl.scm": "query",
		}),
	})
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "hx", Release: &Release{
		Repo:  "helix-editor/helix",
		Asset: `helix-.*-{arch}-linux\.tar\.gz`,
		Dirs:  map[string]string{"runtime": "~/.config/helix/runtime"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(i.BinDir(), "hx"), "HX")
	assertFile(t, filepath.Join(i.Home, ".config/helix/runtime/queries/go/hl.scm"), "query")
}

func TestReleaseInstall_APIError(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer srv.Close()
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "lsd", Release: &Release{Repo: "lsd-rs/lsd", Asset: `lsd`}})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 error, got %v", err)
	}
}

func TestInstallFile_NoDebris(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("BIN"), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(dir, "bin")

	if err := installFile(src, filepath.Join(binDir, "ok")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(filepath.Join(binDir, "ok"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode = %v, want 0755", info.Mode().Perm())
	}

	blocked := filepath.Join(binDir, "blocked")
	if err := os.MkdirAll(filepath.Join(blocked, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := installFile(src, blocked); err == nil {
		t.Fatal("expected error installing over non-empty directory")
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if got := strings.Join(names, ","); got != "blocked,ok" {
		t.Errorf("bin dir contains %q, want only blocked,ok", got)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}
