package util

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func fileMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

func TestCopyFile_PreservesExecutableBit(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name    string
		content string
		mode    os.FileMode
		want    os.FileMode
	}{
		{"config", "key = 1\n", 0o644, 0o644},
		{"open-config", "key = 1\n", 0o777 &^ 0o111, 0o644},
		{"exec", "echo hi\n", 0o755, 0o755},
		{"shebang", "#!/bin/sh\necho hi\n", 0o644, 0o755},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := filepath.Join(dir, tc.name+".src")
			dest := filepath.Join(dir, tc.name+".dest")
			writeFile(t, src, tc.content, tc.mode)

			if err := CopyFile(src, dest); err != nil {
				t.Fatalf("CopyFile: %v", err)
			}

			if got := fileMode(t, dest); got != tc.want {
				t.Errorf("mode = %o, want %o", got, tc.want)
			}
		})
	}
}

func TestCopyFile_FixesModeOfExistingIdenticalFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")
	writeFile(t, src, "same\n", 0o644)
	writeFile(t, dest, "same\n", 0o777)

	if err := CopyFile(src, dest); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	if got := fileMode(t, dest); got != 0o644 {
		t.Errorf("mode = %o, want 644", got)
	}
	if _, err := os.Lstat(dest + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no backup for identical content, got err %v", err)
	}
}

func TestCopyFileFromFS_ModeFromShebang(t *testing.T) {
	dir := t.TempDir()
	srcFS := fstest.MapFS{
		"script": {Data: []byte("#!/bin/bash\n"), Mode: 0o444},
		"config": {Data: []byte("a=b\n"), Mode: 0o444},
	}

	for name, want := range map[string]os.FileMode{"script": 0o755, "config": 0o644} {
		dest := filepath.Join(dir, name)
		if err := CopyFileFromFS(srcFS, name, dest); err != nil {
			t.Fatalf("CopyFileFromFS %s: %v", name, err)
		}
		if got := fileMode(t, dest); got != want {
			t.Errorf("%s mode = %o, want %o", name, got, want)
		}
	}
}

func TestCopyFile_BackupKeepsFirstDifferingVersion(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")
	writeFile(t, dest, "original\n", 0o644)

	for _, content := range []string{"v1\n", "v1\n", "v2\n"} {
		writeFile(t, src, content, 0o644)
		if err := CopyFile(src, dest); err != nil {
			t.Fatalf("CopyFile: %v", err)
		}
		if got := readFile(t, dest); got != content {
			t.Errorf("dest = %q, want %q", got, content)
		}
		if got := readFile(t, dest+".bak"); got != "original\n" {
			t.Errorf("backup = %q, want original", got)
		}
	}
}

func TestCopyFile_NoBackupForNewOrUnchangedFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")
	writeFile(t, src, "content\n", 0o644)

	for range 2 {
		if err := CopyFile(src, dest); err != nil {
			t.Fatalf("CopyFile: %v", err)
		}
	}

	if _, err := os.Lstat(dest + ".bak"); !os.IsNotExist(err) {
		t.Errorf("expected no backup, got err %v", err)
	}
}

func TestCopyFile_ReplacesSymlinkWithoutWritingThroughIt(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	target := filepath.Join(dir, "target")
	dest := filepath.Join(dir, "dest")
	writeFile(t, src, "new\n", 0o644)
	writeFile(t, target, "target\n", 0o644)
	writeFile(t, dest+".bak", "older backup\n", 0o644)
	if err := os.Symlink(target, dest); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(src, dest); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() {
		t.Errorf("expected dest to be a regular file, got %v", info.Mode())
	}
	if got := readFile(t, target); got != "target\n" {
		t.Errorf("symlink target was modified: %q", got)
	}
	if got := readFile(t, dest+".bak"); got != "older backup\n" {
		t.Errorf("existing backup was overwritten: %q", got)
	}
}

func TestWriteJSON_IsNotExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	if err := WriteJSON(path, map[string]any{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if got := fileMode(t, path); got&0o111 != 0 {
		t.Errorf("mode = %o, want no exec bits", got)
	}
}
