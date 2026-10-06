package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func (r *ghRelease) asset(name string) (ghAsset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return ghAsset{}, false
}

func (i *Installer) latestRelease(repo string) (*ghRelease, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/releases/latest", i.APIBase, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if i.Token != "" {
		req.Header.Set("Authorization", "Bearer "+i.Token)
	}

	resp, err := i.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release for %s: %w", repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release for %s: status %d", repo, resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to decode release for %s: %w", repo, err)
	}
	return &rel, nil
}

func (i *Installer) download(url, dest string) error {
	resp, err := i.HTTP.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}
	return nil
}

func (i *Installer) releaseInstall(s Spec) error {
	r := s.Release
	if s.Bin == "" {
		return fmt.Errorf("release install for %s requires Bin", r.Repo)
	}

	rel, err := i.latestRelease(r.Repo)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(rel.Assets))
	for _, a := range rel.Assets {
		names = append(names, a.Name)
	}
	name, err := r.matchAsset(i.GOARCH, names)
	if err != nil {
		return err
	}
	asset, _ := rel.asset(name)

	tmp, err := os.MkdirTemp("", "shizuku-pkg-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	downloaded := filepath.Join(tmp, name)
	if err := i.download(asset.URL, downloaded); err != nil {
		return err
	}

	binPath := downloaded
	root := tmp
	if isArchive(name) {
		root = filepath.Join(tmp, "extract")
		if err := extract(downloaded, root); err != nil {
			return err
		}
		binPath, err = findInTree(root, s.Bin, false)
		if err != nil {
			return err
		}
	}

	if err := installFile(binPath, filepath.Join(i.BinDir(), s.Bin)); err != nil {
		return err
	}

	for src, dst := range r.Dirs {
		dir, err := findInTree(root, src, true)
		if err != nil {
			return err
		}
		if err := replaceDir(dir, i.expandHome(dst)); err != nil {
			return err
		}
	}

	slog.Info("installed from github release", "repo", r.Repo, "tag", rel.TagName, "bin", s.Bin)
	return nil
}

func (i *Installer) expandHome(p string) string {
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(i.Home, rest)
	}
	return p
}

func isArchive(name string) bool {
	for _, ext := range []string{".tar.gz", ".tgz", ".tar.xz"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func extract(archive, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	out, err := exec.Command("tar", "-xf", archive, "-C", dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract %s: %w: %s", archive, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func findInTree(root, name string, wantDir bool) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() != name {
			return nil
		}
		if wantDir && d.IsDir() || !wantDir && d.Type().IsRegular() {
			found = p
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to search %s: %w", root, err)
	}
	if found == "" {
		return "", fmt.Errorf("%s not found in release archive", name)
	}
	return found, nil
}

func installFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(dest), err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.CreateTemp(filepath.Dir(dest), "."+filepath.Base(dest)+".*")
	if err != nil {
		return fmt.Errorf("failed to create temp file for %s: %w", dest, err)
	}
	tmp := out.Name()
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return fmt.Errorf("failed to write %s: %w", tmp, err)
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("failed to close %s: %w", tmp, err)
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("failed to chmod %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("failed to move %s into place: %w", dest, err)
	}
	return nil
}

func replaceDir(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to remove %s: %w", dest, err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(dest), err)
	}
	if err := os.CopyFS(dest, os.DirFS(src)); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", src, dest, err)
	}
	return nil
}
