package sketchybar

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/eleonorayaya/shizuku/programs/sketchybar/bar"
	"github.com/eleonorayaya/shizuku/util"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var socketPath, which string

	cmd := &cobra.Command{
		Use:   "bar",
		Short: "Run the sketchybar daemon",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if root := cmd.Root(); root.PersistentPreRun != nil {
				root.PersistentPreRun(cmd, args)
			}
			spec, err := bar.SpecFor(which)
			if err != nil {
				return err
			}
			if socketPath == "" {
				socketPath = spec.SocketPath()
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&socketPath, "socket", "", "Unix socket the daemon listens on (default depends on --bar)")
	cmd.PersistentFlags().StringVar(&which, "bar", bar.BarTop, "Which bar to drive: top or bottom")

	configPath := configDir + configName
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Configure sketchybar and serve its events",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon(configPath, socketPath, which)
		},
	}
	runCmd.Flags().StringVar(&configPath, "config", configPath, "Path to the generated bar config")

	emitCmd := &cobra.Command{
		Use:    "emit",
		Short:  "Forward a sketchybar event to the daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bar.Emit(socketPath, bar.EventFromEnv())
		},
	}

	cmd.AddCommand(runCmd, emitCmd)
	return cmd
}

func runDaemon(configPath, socketPath, which string) error {
	path, err := util.NormalizeFilePath(configPath)
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}
	cfg, err := bar.LoadConfig(path)
	if err != nil {
		return err
	}
	spec, err := bar.SpecFor(which)
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable: %w", err)
	}
	emit := strconv.Quote(exe) + " bar emit --socket " + strconv.Quote(socketPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server, err := bar.Listen(socketPath)
	if err != nil {
		return err
	}
	defer server.Close()

	events := make(chan bar.Event, 64)
	go bar.Serve(ctx, server, events)

	runner := bar.ExecRunner{}
	return bar.NewDaemon(cfg, runner, emit, spec.Placement(cfg), spec.Items(cfg, runner)).Run(ctx, events)
}
