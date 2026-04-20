package shizuku

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func (b *Builder) Command() *cobra.Command {
	root := &cobra.Command{
		Use: "shizuku",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if b.opts.Verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
			if envProfile := os.Getenv("SHIZUKU_PROFILE"); envProfile != "" && b.opts.Profile == "" {
				b.opts.Profile = envProfile
			}
			return nil
		},
	}
	root.PersistentFlags().BoolVarP(&b.opts.Verbose, "verbose", "v", false, "Enable verbose output")
	root.PersistentFlags().StringVarP(&b.opts.Profile, "profile", "p", b.opts.Profile, "Profile to use (overlays on base)")

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

			fmt.Printf("\n%d file(s) with differences. Diff files written to %s/\n\n", report.TotalChanged, report.OutDir)

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

			return nil
		},
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install application dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Install(context.Background())
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List apps active in the current profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses := b.List()

			if b.opts.Profile != "" {
				fmt.Printf("Profile: %s\n\n", b.opts.Profile)
			} else {
				fmt.Println("Profile: (base)")
				fmt.Println()
			}
			for _, s := range statuses {
				fmt.Printf("  %-12s %s\n", s.Category, s.Name)
			}
			return nil
		},
	}

	root.AddCommand(syncCmd, diffCmd, installCmd, listCmd)
	return root
}
