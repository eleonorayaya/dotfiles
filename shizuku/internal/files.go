package internal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func CopyFiles(inFiles []string, outDir string) error {
	return nil
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

func NormalizeFilePath(filePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	return strings.ReplaceAll(filePath, "~", homeDir), nil
}
