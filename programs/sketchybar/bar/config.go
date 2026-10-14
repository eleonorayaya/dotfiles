package bar

import (
	"encoding/json"
	"fmt"
	"os"
)

type Colors struct {
	Island          string `json:"island"`
	IslandBorder    string `json:"islandBorder"`
	Text            string `json:"text"`
	Hot             string `json:"hot"`
	HotText         string `json:"hotText"`
	PopupBackground string `json:"popupBackground"`
	PopupBorder     string `json:"popupBorder"`
}

type MarginOverride struct {
	Monitor string `json:"monitor"`
	Margin  int    `json:"margin"`
}

type Config struct {
	Colors         Colors          `json:"colors"`
	Badges         []string        `json:"badges,omitempty"`
	Margin         int             `json:"margin"`
	TopGap         int             `json:"topGap"`
	BottomGap      int             `json:"bottomGap"`
	HideOnMonitor  string          `json:"hideOnMonitor,omitempty"`
	MarginOverride *MarginOverride `json:"marginOverride,omitempty"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read bar config: %w", err)
	}
	return decodeJSON[Config](data, "bar config")
}

func decodeJSON[T any](data []byte, what string) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("failed to parse %s: %w", what, err)
	}
	return v, nil
}
