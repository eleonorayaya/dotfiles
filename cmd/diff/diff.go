package diff

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"sort"
	"time"

	shizuku "github.com/eleonorayaya/shizuku"
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var showContent bool

var DiffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Show what would change on next sync",
	RunE:  runDiff,
}

func init() {
	DiffCommand.Flags().BoolVarP(&showContent, "print", "p", false, "Print diff contents to stdout")
}

type diffResult struct {
	name    string
	changed []string
	fileMap map[string]string
}

func runDiff(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	slog.Debug("using source directory", "cwd", cwd)

	appConfig, err := shizukuconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	buildId := fmt.Sprintf("%v", time.Now().Unix())

	outDir := path.Join("out", buildId)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating output dir: %w", err)
	}

	enabledLanguages := shizukuapp.FilterEnabledApps(shizuku.GetLanguages(), appConfig)
	enabledPrograms := shizukuapp.FilterEnabledApps(shizuku.GetPrograms(), appConfig)
	enabledAgents := shizukuapp.FilterEnabledApps(shizuku.GetAgents(), appConfig)

	var results []diffResult
	totalChanged := 0

	languageResults, err := diffApps(enabledLanguages, outDir, appConfig)
	if err != nil {
		return err
	}
	results = append(results, languageResults...)

	programResults, err := diffApps(enabledPrograms, outDir, appConfig)
	if err != nil {
		return err
	}
	results = append(results, programResults...)

	ctx := shizukuapp.CollectAgentConfigs(append(enabledLanguages, enabledPrograms...))

	for _, app := range enabledAgents {
		slog.Debug("generating files for diff", "appName", app.Name())

		var result *shizukuapp.GenerateResult
		if generator, ok := app.(shizukuapp.ContextualGenerator); ok {
			result, err = generator.GenerateWithContext(outDir, appConfig, ctx)
		} else if generator, ok := app.(shizukuapp.FileGenerator); ok {
			result, err = generator.Generate(outDir, appConfig)
		} else {
			continue
		}

		if err != nil {
			return fmt.Errorf("could not generate %s: %w", app.Name(), err)
		}

		changed, err := shizukuapp.DiffAppFiles(result)
		if err != nil {
			return fmt.Errorf("could not diff %s: %w", app.Name(), err)
		}

		if len(changed) > 0 {
			sort.Strings(changed)
			results = append(results, diffResult{name: app.Name(), changed: changed, fileMap: result.FileMap})
		}
	}

	for _, r := range results {
		totalChanged += len(r.changed)
	}

	allEnabled := append(append(enabledLanguages, enabledPrograms...), enabledAgents...)

	envSetups := []*shizukuapp.EnvSetup{}
	for _, app := range allEnabled {
		if provider, ok := app.(shizukuapp.EnvProvider); ok {
			envSetup, err := provider.Env()
			if err != nil {
				return fmt.Errorf("failed to get env setup for %s: %w", app.Name(), err)
			}
			envSetups = append(envSetups, envSetup)
		}
	}

	envFileMap, err := shizukuapp.GenerateEnvFiles(envSetups, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate env files: %w", err)
	}

	envResult := &shizukuapp.GenerateResult{
		FileMap: envFileMap,
		DestDir: "~/.config/shizuku/",
	}
	envChanged, err := shizukuapp.DiffAppFiles(envResult)
	if err != nil {
		return fmt.Errorf("could not diff env file: %w", err)
	}
	if len(envChanged) > 0 {
		results = append(results, diffResult{name: "shizuku (env)", changed: envChanged, fileMap: envResult.FileMap})
		totalChanged += len(envChanged)
	}

	if totalChanged == 0 {
		fmt.Println("No differences found.")
		return nil
	}

	for _, r := range results {
		fmt.Printf("%s:\n", r.name)
		for _, f := range r.changed {
			fmt.Printf("  M %s\n", f)
		}
	}

	fmt.Printf("\n%d file(s) with differences. Diff files written to %s/\n", totalChanged, outDir)

	if showContent {
		fmt.Println()
		for _, r := range results {
			for _, f := range r.changed {
				diffPath := r.fileMap[f] + ".diff"
				content, err := os.ReadFile(diffPath)
				if err != nil {
					return fmt.Errorf("failed to read diff file %s: %w", diffPath, err)
				}
				fmt.Println(string(content))
			}
		}
	}

	return nil
}

func diffApps(apps []shizukuapp.App, outDir string, config *shizukuconfig.Config) ([]diffResult, error) {
	var results []diffResult

	for _, app := range apps {
		generator, ok := app.(shizukuapp.FileGenerator)
		if !ok {
			continue
		}

		slog.Debug("generating files for diff", "appName", app.Name())

		result, err := generator.Generate(outDir, config)
		if err != nil {
			return nil, fmt.Errorf("could not generate %s: %w", app.Name(), err)
		}

		changed, err := shizukuapp.DiffAppFiles(result)
		if err != nil {
			return nil, fmt.Errorf("could not diff %s: %w", app.Name(), err)
		}

		if len(changed) > 0 {
			sort.Strings(changed)
			results = append(results, diffResult{name: app.Name(), changed: changed, fileMap: result.FileMap})
		}
	}

	return results, nil
}
