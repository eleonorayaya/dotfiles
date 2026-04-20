package app

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/eleonorayaya/shizuku/util"
)

type GenerateResult struct {
	FileMap map[string]string
	DestDir string
}

var binaryExtensions = map[string]bool{
	".wasm": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".ico":  true,
	".webp": true,
	".pdf":  true,
	".zip":  true,
	".tar":  true,
	".gz":   true,
}

func isBinaryFile(fileName string) bool {
	return binaryExtensions[path.Ext(fileName)]
}

func DiffAppFiles(result *GenerateResult) ([]string, error) {
	var changed []string

	for fileName, generatedPath := range result.FileMap {
		if isBinaryFile(fileName) {
			continue
		}

		destPath, err := util.NormalizeFilePath(path.Join(result.DestDir, fileName))
		if err != nil {
			return nil, fmt.Errorf("failed to normalize dest path for %s: %w", fileName, err)
		}

		diffSrc := destPath
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			diffSrc = "/dev/null"
		}

		cmd := exec.Command("diff", "-u", "-b", "-I", "# Generated at:", diffSrc, generatedPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// exit code 1 means files differ, which is expected
			} else {
				return nil, fmt.Errorf("failed to diff %s: %w", fileName, err)
			}
		}

		if len(output) > 0 {
			diffPath := generatedPath + ".diff"
			if err := os.WriteFile(diffPath, output, 0644); err != nil {
				return nil, fmt.Errorf("failed to write diff for %s: %w", fileName, err)
			}
			changed = append(changed, fileName)
		}
	}

	return changed, nil
}

func listFSFiles(srcFS fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk app contents: %w", err)
	}
	return files, nil
}

func GenerateAppFiles(name string, contents fs.FS, data map[string]any, outDir string) (map[string]string, error) {
	appOutDir := path.Join(outDir, name)
	if err := os.Mkdir(appOutDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create out dir: %w", err)
	}

	contentsFS, err := fs.Sub(contents, "contents")
	if err != nil {
		return nil, fmt.Errorf("failed to open contents subdir: %w", err)
	}

	generatedFiles := make(map[string]string)
	files, err := listFSFiles(contentsFS)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName, outFile, err := generateAppFile(contentsFS, file, data, appOutDir)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s: %w", file, err)
		}

		generatedFiles[fileName] = outFile
	}

	return generatedFiles, nil
}

func generateAppFile(contentsFS fs.FS, inFile string, data map[string]any, appOutDir string) (string, string, error) {
	fileName := strings.ReplaceAll(inFile, ".tmpl", "")

	outFilePath := path.Join(appOutDir, fileName)
	outDirPath := path.Dir(outFilePath)

	if err := util.EnsureDirExists(outDirPath); err != nil {
		return "", "", fmt.Errorf("failed out make out dir %s for config file %s: %w", outDirPath, inFile, err)
	}

	if path.Ext(inFile) == ".tmpl" {
		if err := util.GenerateTemplateFromFS(contentsFS, inFile, data, outFilePath); err != nil {
			return "", "", fmt.Errorf("failed to generate file: %w", err)
		}
	} else {
		if err := util.CopyFileFromFS(contentsFS, inFile, outFilePath); err != nil {
			return "", "", fmt.Errorf("failed to generate file: %w", err)
		}
	}

	return fileName, outFilePath, nil
}

func SyncAppFile(fileName, filePath string, outDir string) error {
	outFilePath := path.Join(outDir, fileName)
	if err := util.EnsureDirExists(path.Dir(outFilePath)); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", fileName, err)
	}

	if err := util.CopyFile(filePath, outFilePath); err != nil {
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

func FetchRemoteAppFiles(outDir string, appName string, remoteFiles map[string]string) (map[string]string, error) {
	appOutDir := path.Join(outDir, appName)
	fileMap := make(map[string]string)

	for relPath, url := range remoteFiles {
		buildPath := path.Join(appOutDir, relPath)

		if err := util.EnsureDirExists(path.Dir(buildPath)); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to download %s: %w", relPath, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download %s: status %d", relPath, resp.StatusCode)
		}

		out, err := os.Create(buildPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s: %w", relPath, err)
		}
		defer out.Close()

		if _, err := io.Copy(out, resp.Body); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", relPath, err)
		}

		fileMap[relPath] = buildPath
	}

	return fileMap, nil
}
