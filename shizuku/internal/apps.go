package internal

import (
	"fmt"
	"os"
	"path"
)

func listAppFiles(appDir string, relativePath string) ([]string, error) {
	appFiles := make([]string, 0)

	listDir := path.Join(appDir, relativePath)
	entries, err := os.ReadDir(listDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s app files: %w", appDir, err)
	}

	for _, entry := range entries {
		entryPath := path.Join(relativePath, entry.Name())

		if entry.IsDir() {

			subDirFiles, err := listAppFiles(appDir, entryPath)
			if err != nil {
				return nil, err
			}

			appFiles = append(appFiles, subDirFiles...)
		} else {
			appFiles = append(appFiles, entryPath)
		}
	}

	return appFiles, nil
}

func GenerateAppFiles(appName string, data map[string]any, outDir string) (map[string]string, error) {
	appFileDir := path.Join("apps", appName, "contents")

	appOutDir := path.Join(outDir, appName)
	if err := os.Mkdir(appOutDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create out dir: %w", err)
	}

	generatedFiles := make(map[string]string)
	appFiles, err := listAppFiles(appFileDir, "")
	if err != nil {
		return nil, err
	}

	for _, file := range appFiles {
		fileName, outFile, err := GenerateAppFile(appFileDir, file, data, appOutDir)
		if err != nil {
			return nil, fmt.Errorf("failed to generate sketchybarrc: %w", err)
		}

		generatedFiles[fileName] = outFile
	}

	return generatedFiles, nil
}

func SyncAppFile(fileName, filePath string, outDir string) error {
	outFilePath := path.Join(outDir, fileName)
	if err := EnsureDirExists(path.Dir(outFilePath)); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", fileName, err)
	}

	if err := CopyFile(filePath, outFilePath); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", fileName, err)
	}

	return nil
}

func SyncAppFiles(fileMap map[string]string, outDir string) error {
	for fileName, filePath := range fileMap {
		if err := SyncAppFile(fileName, filePath, outDir); err != nil {
			return fmt.Errorf("failed to sync file %s: %w", fileName, err)
		}
	}

	return nil
}

