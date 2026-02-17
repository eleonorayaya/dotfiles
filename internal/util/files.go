package util

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

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

	if _, err := os.Lstat(normalizedOutFile); err == nil {
		backupPath := normalizedOutFile + ".bak"
		if err := os.Rename(normalizedOutFile, backupPath); err != nil {
			return fmt.Errorf("failed to back up existing destination file: %w", err)
		}
	}

	sourceFile, err := os.Open(normalizedInFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(normalizedOutFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	err = os.Chmod(normalizedOutFile, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to set destination file perms: %w", err)
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

func ReadJSONMap(path string) (map[string]any, error) {
	normalized, err := NormalizeFilePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize path: %w", err)
	}

	data, err := os.ReadFile(normalized)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return result, nil
}

func WriteJSONMap(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	merged, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	merged = append(merged, '\n')

	if err := os.WriteFile(path, merged, os.ModePerm); err != nil {
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
