package shizuku

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func (b *Builder) Command() *cobra.Command {
	var showContent bool

	root := &cobra.Command{
		Use: "shizuku",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if b.opts.Verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
			return nil
		},
	}
	root.PersistentFlags().BoolVarP(&b.opts.Verbose, "verbose", "v", false, "Enable verbose output")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize shizuku configuration directory and create default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Init()
		},
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all application configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Sync(context.Background())
		},
	}

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Show what would change on next sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := b.Diff(context.Background())
			if err != nil {
				return err
			}

			if report.TotalChanged == 0 {
				fmt.Println("No differences found.")
				return nil
			}

			for _, r := range report.Results {
				fmt.Printf("%s:\n", r.Name)
				for _, f := range r.Changed {
					fmt.Printf("  M %s\n", f)
				}
			}

			fmt.Printf("\n%d file(s) with differences. Diff files written to %s/\n", report.TotalChanged, report.OutDir)

			if showContent {
				fmt.Println()
				for _, r := range report.Results {
					for _, f := range r.Changed {
						diffPath := r.FileMap[f] + ".diff"
						content, err := os.ReadFile(diffPath)
						if err != nil {
							return fmt.Errorf("failed to read diff file %s: %w", diffPath, err)
						}
						fmt.Println(string(content))
					}
				}
			}

			return nil
		},
	}
	diffCmd.Flags().BoolVarP(&showContent, "print", "p", false, "Print diff contents to stdout")

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install application dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Install(context.Background())
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available apps and their enabled status",
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses, err := b.List()
			if err != nil {
				return err
			}

			fmt.Println("Available apps:")
			fmt.Println()
			for _, s := range statuses {
				status := "disabled"
				if s.Enabled {
					status = "enabled"
				}
				fmt.Printf("  %-20s %s\n", s.Name, status)
			}

			return nil
		},
	}

	root.AddCommand(initCmd, syncCmd, diffCmd, installCmd, listCmd)
	return root
}
