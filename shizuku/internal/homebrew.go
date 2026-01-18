package internal

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func GetBrewAppPrefix(appName string) (string, error) {
	prefix, err := runBrewCommand("--prefix", appName)
	if err != nil {
		return "", fmt.Errorf("brew --prefix failed: %w", err)
	}

	return prefix, nil
}

func runBrewCommand(args ...string) (string, error) {
	out, err := exec.Command("brew", args...).Output()
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(string(out)), nil
}
