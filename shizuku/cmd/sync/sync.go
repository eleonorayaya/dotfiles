package sync

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/eleonorayaya/shizuku/apps/sketchybar"
	"github.com/spf13/cobra"
)

var SyncCommand = &cobra.Command{
	Use:   "sync [flags] configs_path",
	Short: "",
	RunE:  sync,
}

func sync(cmd *cobra.Command, args []string) error {

	buildId := fmt.Sprintf("%v", time.Now().Unix())

	outDir := path.Join("out", buildId)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("error created output dir: %w", err)
	}

	if err := sketchybar.Sync(outDir); err != nil {
		return fmt.Errorf("could not sync sketchybar: %w", err)
	}

	fmt.Printf("%s", outDir)
	return nil
}

