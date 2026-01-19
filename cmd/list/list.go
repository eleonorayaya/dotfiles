package list

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/apps"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var ListCommand = &cobra.Command{
	Use:   "list",
	Short: "List all available apps and their enabled status",
	RunE:  list,
}

func list(cmd *cobra.Command, args []string) error {
	appConfig, err := shizukuconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	allApps := apps.GetApps()

	fmt.Println("Available apps:")
	fmt.Println()

	for _, app := range allApps {
		enabled := app.Enabled(appConfig)
		status := "disabled"
		if enabled {
			status = "enabled"
		}

		fmt.Printf("  %-20s %s\n", app.Name(), status)
	}

	return nil
}
