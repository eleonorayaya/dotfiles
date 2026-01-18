package internal

import (
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/eleonorayaya/shizuku/internal/util"
)

const AppsRoot = "apps"

func GenerateAppFile(appDir string, inFile string, data map[string]any, appOutDir string) (string, string, error) {
	fileName := strings.ReplaceAll(inFile, ".tmpl", "")

	inFilePath := path.Join(appDir, inFile)

	outFilePath := path.Join(appOutDir, fileName)
	outDirPath := path.Dir(outFilePath)

	if err := util.EnsureDirExists(outDirPath); err != nil {
		return "", "", fmt.Errorf("failed out make out dir %s for config file %s: %w", outDirPath, inFile, err)
	}

	if path.Ext(inFilePath) == ".tmpl" {
		if err := generateTemplateFile(inFilePath, data, outFilePath); err != nil {
			return "", "", fmt.Errorf("failed to generate file: %w", err)
		}
	} else {
		if err := util.CopyFile(inFilePath, outFilePath); err != nil {
			return "", "", fmt.Errorf("failed to generate file: %w", err)
		}
	}

	return fileName, outFilePath, nil
}

func generateTemplateFile(templatePath string, data map[string]any, outFile string) error {
	slog.Debug("generating template", "templatePath", templatePath, "outFile", outFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("GenerateFile: failed to parse template %s: %w", templatePath, err)
	}

	file, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("GenerateFile: error creating temp file %s: %w", outFile, err)
	}

	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("GenerateFile: failed to execute template %s: %w", templatePath, err)
	}

	slog.Debug("generated template", "templatePath", templatePath, "outFile", outFile)

	return nil
}
