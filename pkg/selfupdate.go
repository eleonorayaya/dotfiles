package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func SelfUpdate(repo, project string) error {
	i, err := defaultInstaller()
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate running binary: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to resolve running binary: %w", err)
	}

	return i.updateBinary(repo, project, exe)
}

func (i *Installer) updateBinary(repo, project, dest string) error {
	rel, err := i.latestRelease(repo)
	if err != nil {
		return err
	}

	archiveName := fmt.Sprintf("%s_%s_%s.tar.gz", project, i.GOOS, i.GOARCH)
	archive, ok := rel.asset(archiveName)
	if !ok {
		return fmt.Errorf("release %s has no asset %s", rel.TagName, archiveName)
	}
	sums, ok := rel.asset("checksums.txt")
	if !ok {
		return fmt.Errorf("release %s has no checksums.txt", rel.TagName)
	}

	tmp, err := os.MkdirTemp("", "shizuku-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	archivePath := filepath.Join(tmp, archiveName)
	sumsPath := filepath.Join(tmp, "checksums.txt")
	if err := i.download(archive.URL, archivePath); err != nil {
		return err
	}
	if err := i.download(sums.URL, sumsPath); err != nil {
		return err
	}
	if err := verifyChecksum(sumsPath, archiveName, archivePath); err != nil {
		return err
	}

	root := filepath.Join(tmp, "extract")
	if err := extract(archivePath, root); err != nil {
		return err
	}
	bin, err := findInTree(root, project, false)
	if err != nil {
		return err
	}
	if err := installFile(bin, dest); err != nil {
		return err
	}

	slog.Info("updated from github release", "tag", rel.TagName, "path", dest)
	return nil
}

func verifyChecksum(sumsPath, name, filePath string) error {
	data, err := os.ReadFile(sumsPath)
	if err != nil {
		return fmt.Errorf("failed to read checksums: %w", err)
	}

	want := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("no checksum listed for %s", name)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to hash %s: %w", filePath, err)
	}

	if got := hex.EncodeToString(h.Sum(nil)); got != want {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", name, got, want)
	}
	return nil
}
