package shizuku

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/config"
	"github.com/spf13/cobra"
)

func defaultConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "shizuku", "shizuku.yml")
}

func (b *Builder) Command() *cobra.Command {
	root := &cobra.Command{
		Use: "shizuku",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if b.opts.Verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
			cfg, err := config.Load(defaultConfigPath())
			if err != nil {
				return err
			}
			if b.opts.Profile != "" {
				slog.Info("using profile", "profile", b.opts.Profile)
			} else if cfg.Profile != "" {
				b.opts.Profile = cfg.Profile
				slog.Info("using profile", "profile", b.opts.Profile)
			} else {
				slog.Warn("no profile set, using base profile")
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

	root.AddCommand(syncCmd, diffCmd, installCmd, listCmd, configCmd())
	return root
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage shizuku configuration",
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Set(defaultConfigPath(), args[0], args[1]); err != nil {
				return err
			}
			slog.Info("config updated", "key", args[0], "value", args[1])
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a config value, or print all config if no key given",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(defaultConfigPath())
			if err != nil {
				return err
			}
			if len(args) == 0 {
				out, err := cfg.YAML()
				if err != nil {
					return err
				}
				fmt.Print(out)
				return nil
			}
			val, err := config.Get(cfg, args[0])
			if err != nil {
				return err
			}
			fmt.Println(val)
			return nil
		},
	}

	cmd.AddCommand(setCmd, getCmd)
	return cmd
}
