package util

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path"
)

func GenerateTemplateFromFS(srcFS fs.FS, name string, data map[string]any, outFile string) error {
	slog.Debug("generating template", "templatePath", name, "outFile", outFile)

	tmpl, err := template.New(path.Base(name)).ParseFS(srcFS, name)
	if err != nil {
		return fmt.Errorf("GenerateFile: failed to parse template %s: %w", name, err)
	}

	file, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("GenerateFile: error creating temp file %s: %w", outFile, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("GenerateFile: failed to execute template %s: %w", name, err)
	}

	slog.Debug("generated template", "templatePath", name, "outFile", outFile)

	return nil
}
