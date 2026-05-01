package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Profile string `yaml:"profile" env:"SHIZUKU_PROFILE"`
}

// Load reads the YAML config at path (silent no-op if missing), then overrides
// any field whose "env" tag names a non-empty environment variable.
func Load(path string) (*Config, error) {
	cfg, err := loadRaw(path)
	if err != nil {
		return nil, err
	}
	applyEnv(cfg)
	return cfg, nil
}

// Set loads the raw file (no env override), updates the field named by its yaml
// tag, and writes the result back. Env-derived values never bleed into the file.
func Set(path, key, value string) error {
	cfg, err := loadRaw(path)
	if err != nil {
		return err
	}
	if err := setField(cfg, key, value); err != nil {
		return err
	}
	return write(path, cfg)
}

// Get returns the value of the field whose yaml tag equals key.
func Get(cfg *Config, key string) (string, error) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if t.Field(i).Tag.Get("yaml") == key {
			return v.Field(i).String(), nil
		}
	}
	return "", fmt.Errorf("unknown config key %q", key)
}

// YAML marshals the config to a YAML string.
func (c *Config) YAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

func loadRaw(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func applyEnv(cfg *Config) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if envKey := t.Field(i).Tag.Get("env"); envKey != "" {
			if val := os.Getenv(envKey); val != "" {
				v.Field(i).SetString(val)
			}
		}
	}
}

func setField(cfg *Config, key, value string) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if t.Field(i).Tag.Get("yaml") == key {
			v.Field(i).SetString(value)
			return nil
		}
	}
	return fmt.Errorf("unknown config key %q", key)
}

func write(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}
