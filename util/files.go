package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func CopyFileFromFS(srcFS fs.FS, srcName, outFile string) error {
	slog.Debug("copying file", "srcName", srcName, "outFile", outFile)

	normalizedOutFile, err := NormalizeFilePath(outFile)
	if err != nil {
		return fmt.Errorf("failed to normalize out file path: %w", err)
	}

	content, err := fs.ReadFile(srcFS, srcName)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	return writeSyncedFile(normalizedOutFile, content, 0)
}

func CopyFile(inFile, outFile string) error {
	slog.Debug("copying file", "inFile", inFile, "outFile", outFile)

	normalizedInFile, err := NormalizeFilePath(inFile)
	if err != nil {
		return fmt.Errorf("failed to normalize input file path: %w", err)
	}

	normalizedOutFile, err := NormalizeFilePath(outFile)
	if err != nil {
		return fmt.Errorf("failed to normalize out file path: %w", err)
	}

	info, err := os.Stat(normalizedInFile)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	content, err := os.ReadFile(normalizedInFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	return writeSyncedFile(normalizedOutFile, content, info.Mode())
}

func syncedFileMode(content []byte, srcMode fs.FileMode) fs.FileMode {
	if srcMode.Perm()&0o111 != 0 || bytes.HasPrefix(content, []byte("#!")) {
		return 0o755
	}
	return 0o644
}

func writeSyncedFile(dest string, content []byte, srcMode fs.FileMode) error {
	if err := replaceDestination(dest, content); err != nil {
		return err
	}

	mode := syncedFileMode(content, srcMode)
	if err := os.WriteFile(dest, content, mode); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	if err := os.Chmod(dest, mode); err != nil {
		return fmt.Errorf("failed to set destination file perms: %w", err)
	}

	return nil
}

func replaceDestination(dest string, content []byte) error {
	info, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat destination file: %w", err)
	}

	if info.Mode().IsRegular() {
		existing, err := os.ReadFile(dest)
		if err != nil {
			return fmt.Errorf("failed to read destination file: %w", err)
		}
		if bytes.Equal(existing, content) {
			return nil
		}
	}

	backupPath := dest + ".bak"
	if _, err := os.Lstat(backupPath); os.IsNotExist(err) {
		if err := os.Rename(dest, backupPath); err != nil {
			return fmt.Errorf("failed to back up existing destination file: %w", err)
		}
		return nil
	}

	if err := os.Remove(dest); err != nil {
		return fmt.Errorf("failed to remove existing destination file: %w", err)
	}
	return nil
}

func EnsureDirExists(dirPath string) error {
	normalized, err := NormalizeFilePath(dirPath)
	if err != nil {
		return fmt.Errorf("error normalizing dir path: %w", err)
	}

	if err := os.MkdirAll(normalized, os.ModePerm); err != nil {
		return fmt.Errorf("failed to ensure dir %s: %w", dirPath, err)
	}

	return nil
}

func ReadFileIfExists(path string) ([]byte, error) {
	normalized, err := NormalizeFilePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize path: %w", err)
	}

	data, err := os.ReadFile(normalized)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	return data, nil
}

func ReadJSONMap(path string) (map[string]any, error) {
	data, err := ReadFileIfExists(path)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return map[string]any{}, nil
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return result, nil
}

func WriteJSONMap(path string, data map[string]any) error {
	return WriteJSON(path, data)
}

func WriteJSON(path string, data any) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	merged, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	merged = append(merged, '\n')

	if err := os.WriteFile(path, merged, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

func NormalizeFilePath(filePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	return strings.ReplaceAll(filePath, "~", homeDir), nil
}
