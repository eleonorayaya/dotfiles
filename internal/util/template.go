package util

import (
	"fmt"
	"html/template"
	"log/slog"
	"os"
)

func GenerateTemplateFile(templatePath string, data map[string]any, outFile string) error {
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
