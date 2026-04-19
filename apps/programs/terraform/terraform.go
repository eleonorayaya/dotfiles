package terraform

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terraform"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetLanguageEnabled(config.LanguageTerraform)
}

func (a *App) Install(cfg *config.Config) error {
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

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Aliases: []app.Alias{
			{Name: "tf", Command: "tofu"},
			{Name: "tfmt", Command: "tofu fmt -recursive"},
		},
	}, nil
}
