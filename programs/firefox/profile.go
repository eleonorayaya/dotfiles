package firefox

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func defaultProfileDir(firefoxDir string) (string, error) {
	rel, err := installDefault(filepath.Join(firefoxDir, "installs.ini"))
	if err != nil {
		return "", err
	}
	if rel == "" {
		rel, err = profilesDefault(filepath.Join(firefoxDir, "profiles.ini"))
		if err != nil {
			return "", err
		}
	}
	if rel == "" {
		return "", nil
	}
	if filepath.IsAbs(rel) {
		return rel, nil
	}
	return filepath.Join(firefoxDir, rel), nil
}

func installDefault(path string) (string, error) {
	cfg, err := loadINI(path)
	if err != nil || cfg == nil {
		return "", err
	}
	for _, section := range cfg.Sections() {
		if p := section.Key("Default").String(); p != "" {
			return p, nil
		}
	}
	return "", nil
}

func profilesDefault(path string) (string, error) {
	cfg, err := loadINI(path)
	if err != nil || cfg == nil {
		return "", err
	}
	var fallback string
	for _, section := range cfg.Sections() {
		p := section.Key("Path").String()
		if p == "" {
			continue
		}
		if section.Key("Default").String() == "1" {
			return p, nil
		}
		if fallback == "" {
			fallback = p
		}
	}
	return fallback, nil
}

func loadINI(path string) (*ini.File, error) {
	cfg, err := ini.Load(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return cfg, nil
}
