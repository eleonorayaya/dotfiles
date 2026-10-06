package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "shizuku_linux_amd64.tar.gz")
	if err := os.WriteFile(file, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	sums := filepath.Join(dir, "checksums.txt")

	write := func(content string) {
		if err := os.WriteFile(sums, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(sha([]byte("other")) + "  shizuku_darwin_arm64.tar.gz\n" + sha([]byte("payload")) + "  shizuku_linux_amd64.tar.gz\n")
	if err := verifyChecksum(sums, "shizuku_linux_amd64.tar.gz", file); err != nil {
		t.Errorf("expected match: %v", err)
	}

	write(sha([]byte("tampered")) + "  shizuku_linux_amd64.tar.gz\n")
	if err := verifyChecksum(sums, "shizuku_linux_amd64.tar.gz", file); err == nil {
		t.Error("expected mismatch error")
	}

	write(sha([]byte("payload")) + "  something_else.tar.gz\n")
	if err := verifyChecksum(sums, "shizuku_linux_amd64.tar.gz", file); err == nil {
		t.Error("expected missing-entry error")
	}
}

func TestUpdateBinary(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	archive := makeTarGz(t, map[string]string{"shizuku": "NEW", "README.md": "x"})
	srv := releaseServer(t, map[string][]byte{
		"shizuku_linux_amd64.tar.gz": archive,
		"checksums.txt":              []byte(sha(archive) + "  shizuku_linux_amd64.tar.gz\n"),
	})
	i.APIBase = srv.URL

	dest := filepath.Join(t.TempDir(), "shizuku")
	if err := os.WriteFile(dest, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := i.updateBinary("eleonorayaya/dotfiles", "shizuku", dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, dest, "NEW")
}

func TestUpdateBinary_BadChecksumLeavesBinary(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	archive := makeTarGz(t, map[string]string{"shizuku": "NEW"})
	srv := releaseServer(t, map[string][]byte{
		"shizuku_linux_amd64.tar.gz": archive,
		"checksums.txt":              []byte(sha([]byte("nope")) + "  shizuku_linux_amd64.tar.gz\n"),
	})
	i.APIBase = srv.URL

	dest := filepath.Join(t.TempDir(), "shizuku")
	if err := os.WriteFile(dest, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := i.updateBinary("eleonorayaya/dotfiles", "shizuku", dest); err == nil {
		t.Fatal("expected checksum error")
	}
	assertFile(t, dest, "OLD")
}
