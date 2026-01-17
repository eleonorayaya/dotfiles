package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/eleonorayaya/shizuku/cmd/sync"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "shizuku",
	Short: "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}
	},
	Long: ``,
}

func init() {
	rootCmd.AddCommand(sync.SyncCommand)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

