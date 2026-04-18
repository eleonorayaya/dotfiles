package terraform

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terraform"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetLanguageEnabled(shizukuconfig.LanguageTerraform)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.AddTap("hashicorp/tap"); err != nil {
		return fmt.Errorf("failed to add hashicorp tap: %w", err)
	}

	if err := util.InstallBrewPackage("hashicorp/tap/terraform-ls", false); err != nil {
		return fmt.Errorf("failed to install terraform-ls: %w", err)
	}

	if err := util.InstallBrewPackage("opentofu", false); err != nil {
		return fmt.Errorf("failed to install opentofu: %w", err)
	}

	if err := util.AddTap("spacelift-io/spacelift"); err != nil {
		return fmt.Errorf("failed to add spacelift tap: %w", err)
	}

	if err := util.InstallBrewPackage("spacelift-io/spacelift/spacectl", false); err != nil {
		return fmt.Errorf("failed to install spacectl: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Aliases: []shizukuapp.Alias{
			{Name: "tf", Command: "tofu"},
			{Name: "tfmt", Command: "tofu fmt -recursive"},
		},
	}, nil
}
