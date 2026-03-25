package main

import (
	"fmt"
	"log/slog"
	"os"

	diffcmd "github.com/eleonorayaya/shizuku/cmd/diff"
	initcmd "github.com/eleonorayaya/shizuku/cmd/init"
	installcmd "github.com/eleonorayaya/shizuku/cmd/install"
	listcmd "github.com/eleonorayaya/shizuku/cmd/list"
	"github.com/eleonorayaya/shizuku/cmd/sync"
	upgradecmd "github.com/eleonorayaya/shizuku/cmd/upgrade"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "shizuku",
	Short: "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		if _, err := os.Stat("apps"); os.IsNotExist(err) {
			sourceDir, err := util.NormalizeFilePath(shizukuconfig.SourceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve source directory: %w", err)
			}

			if err := os.Chdir(sourceDir); err != nil {
				return fmt.Errorf("failed to change to source directory %s: %w", sourceDir, err)
			}
		}

		return nil
	},
	Long: ``,
}

func init() {
	rootCmd.AddCommand(diffcmd.DiffCommand)
	rootCmd.AddCommand(initcmd.InitCommand)
	rootCmd.AddCommand(installcmd.InstallCommand)
	rootCmd.AddCommand(listcmd.ListCommand)
	rootCmd.AddCommand(sync.SyncCommand)
	rootCmd.AddCommand(upgradecmd.UpgradeCommand)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
