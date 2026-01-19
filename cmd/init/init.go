package initcmd

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize shizuku configuration directory and create default config file",
	RunE:  initConfig,
}

func initConfig(cmd *cobra.Command, args []string) error {
	created, configPath, err := shizukuconfig.InitConfig()
	if err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	if created {
		fmt.Printf("Created default shizuku configuration at %s\n", configPath)
	} else {
		fmt.Printf("Merged existing configuration with defaults at %s\n", configPath)
	}

	return nil
}
