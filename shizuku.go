package shizuku

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sort"
	"time"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
)

type Options struct {
	OutDir  string
	Verbose bool
}

type Builder struct {
	opts      Options
	languages []app.App
	programs  []app.App
	agents    []app.App
}

func New(opts Options) *Builder {
	return &Builder{opts: opts}
}

func (b *Builder) AddLanguage(a app.App) *Builder {
	b.languages = append(b.languages, a)
	return b
}

func (b *Builder) AddLanguages(apps ...app.App) *Builder {
	b.languages = append(b.languages, apps...)
	return b
}

func (b *Builder) AddProgram(a app.App) *Builder {
	b.programs = append(b.programs, a)
	return b
}

func (b *Builder) AddPrograms(apps ...app.App) *Builder {
	b.programs = append(b.programs, apps...)
	return b
}

func (b *Builder) AddAgent(a app.App) *Builder {
	b.agents = append(b.agents, a)
	return b
}

func (b *Builder) AddAgents(apps ...app.App) *Builder {
	b.agents = append(b.agents, apps...)
	return b
}

func (b *Builder) AllApps() []app.App {
	all := make([]app.App, 0, len(b.languages)+len(b.programs)+len(b.agents))
	all = append(all, b.languages...)
	all = append(all, b.programs...)
	all = append(all, b.agents...)
	return all
}

func (b *Builder) resolveOutDir() (string, error) {
	outDir := b.opts.OutDir
	if outDir == "" {
		outDir = path.Join("out", fmt.Sprintf("%v", time.Now().Unix()))
	}
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating output dir: %w", err)
	}
	return outDir, nil
}

func (b *Builder) Init() error {
	created, configPath, err := config.InitConfig()
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

func (b *Builder) Sync(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	outDir, err := b.resolveOutDir()
	if err != nil {
		return err
	}

	enabledLanguages := app.FilterEnabledApps(b.languages, cfg)
	enabledPrograms := app.FilterEnabledApps(b.programs, cfg)
	enabledAgents := app.FilterEnabledApps(b.agents, cfg)

	if err := syncApps(enabledLanguages, outDir, cfg); err != nil {
		return err
	}
	if err := syncApps(enabledPrograms, outDir, cfg); err != nil {
		return err
	}

	syncCtx := app.CollectAgentConfigs(append(enabledLanguages, enabledPrograms...))

	for _, a := range enabledAgents {
		slog.Info("app syncing", "appName", a.Name())

		if syncer, ok := a.(app.ContextualSyncer); ok {
			if err := syncer.SyncWithContext(outDir, cfg, syncCtx); err != nil {
				return fmt.Errorf("could not sync %s: %w", a.Name(), err)
			}
		} else if syncer, ok := a.(app.FileSyncer); ok {
			if err := syncer.Sync(outDir, cfg); err != nil {
				return fmt.Errorf("could not sync %s: %w", a.Name(), err)
			}
		}

		slog.Info("app synced", "appName", a.Name())
	}

	allEnabled := append(append(enabledLanguages, enabledPrograms...), enabledAgents...)
	return b.syncEnv(allEnabled, outDir)
}

func syncApps(apps []app.App, outDir string, cfg *config.Config) error {
	for _, a := range apps {
		slog.Info("app syncing", "appName", a.Name())

		if syncer, ok := a.(app.FileSyncer); ok {
			if err := syncer.Sync(outDir, cfg); err != nil {
				return fmt.Errorf("could not sync %s: %w", a.Name(), err)
			}

			slog.Info("app synced", "appName", a.Name())
		}
	}
	return nil
}

func (b *Builder) syncEnv(enabled []app.App, outDir string) error {
	envSetups := []*app.EnvSetup{}
	for _, a := range enabled {
		if provider, ok := a.(app.EnvProvider); ok {
			envSetup, err := provider.Env()
			if err != nil {
				return fmt.Errorf("failed to get env setup for %s: %w", a.Name(), err)
			}
			envSetups = append(envSetups, envSetup)
		}
	}

	envFileMap, err := app.GenerateEnvFiles(envSetups, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate env files: %w", err)
	}

	if err := app.SyncAppFiles(envFileMap, "~/.config/shizuku/"); err != nil {
		return fmt.Errorf("failed to sync env files: %w", err)
	}

	return nil
}

type DiffResult struct {
	Name    string
	Changed []string
	FileMap map[string]string
}

type DiffReport struct {
	Results      []DiffResult
	TotalChanged int
	OutDir       string
}

func (b *Builder) Diff(ctx context.Context) (*DiffReport, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	outDir, err := b.resolveOutDir()
	if err != nil {
		return nil, err
	}

	enabledLanguages := app.FilterEnabledApps(b.languages, cfg)
	enabledPrograms := app.FilterEnabledApps(b.programs, cfg)
	enabledAgents := app.FilterEnabledApps(b.agents, cfg)

	var results []DiffResult

	languageResults, err := diffApps(enabledLanguages, outDir, cfg)
	if err != nil {
		return nil, err
	}
	results = append(results, languageResults...)

	programResults, err := diffApps(enabledPrograms, outDir, cfg)
	if err != nil {
		return nil, err
	}
	results = append(results, programResults...)

	syncCtx := app.CollectAgentConfigs(append(enabledLanguages, enabledPrograms...))

	for _, a := range enabledAgents {
		slog.Debug("generating files for diff", "appName", a.Name())

		var result *app.GenerateResult
		if generator, ok := a.(app.ContextualGenerator); ok {
			result, err = generator.GenerateWithContext(outDir, cfg, syncCtx)
		} else if generator, ok := a.(app.FileGenerator); ok {
			result, err = generator.Generate(outDir, cfg)
		} else {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("could not generate %s: %w", a.Name(), err)
		}

		changed, err := app.DiffAppFiles(result)
		if err != nil {
			return nil, fmt.Errorf("could not diff %s: %w", a.Name(), err)
		}

		if len(changed) > 0 {
			sort.Strings(changed)
			results = append(results, DiffResult{Name: a.Name(), Changed: changed, FileMap: result.FileMap})
		}
	}

	allEnabled := append(append(enabledLanguages, enabledPrograms...), enabledAgents...)
	envResult, envChanged, err := b.diffEnv(allEnabled, outDir)
	if err != nil {
		return nil, err
	}
	if len(envChanged) > 0 {
		results = append(results, DiffResult{Name: "shizuku (env)", Changed: envChanged, FileMap: envResult.FileMap})
	}

	total := 0
	for _, r := range results {
		total += len(r.Changed)
	}

	return &DiffReport{Results: results, TotalChanged: total, OutDir: outDir}, nil
}

func diffApps(apps []app.App, outDir string, cfg *config.Config) ([]DiffResult, error) {
	var results []DiffResult

	for _, a := range apps {
		generator, ok := a.(app.FileGenerator)
		if !ok {
			continue
		}

		slog.Debug("generating files for diff", "appName", a.Name())

		result, err := generator.Generate(outDir, cfg)
		if err != nil {
			return nil, fmt.Errorf("could not generate %s: %w", a.Name(), err)
		}

		changed, err := app.DiffAppFiles(result)
		if err != nil {
			return nil, fmt.Errorf("could not diff %s: %w", a.Name(), err)
		}

		if len(changed) > 0 {
			sort.Strings(changed)
			results = append(results, DiffResult{Name: a.Name(), Changed: changed, FileMap: result.FileMap})
		}
	}

	return results, nil
}

func (b *Builder) diffEnv(enabled []app.App, outDir string) (*app.GenerateResult, []string, error) {
	envSetups := []*app.EnvSetup{}
	for _, a := range enabled {
		if provider, ok := a.(app.EnvProvider); ok {
			envSetup, err := provider.Env()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get env setup for %s: %w", a.Name(), err)
			}
			envSetups = append(envSetups, envSetup)
		}
	}

	envFileMap, err := app.GenerateEnvFiles(envSetups, outDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate env files: %w", err)
	}

	envResult := &app.GenerateResult{
		FileMap: envFileMap,
		DestDir: "~/.config/shizuku/",
	}
	envChanged, err := app.DiffAppFiles(envResult)
	if err != nil {
		return nil, nil, fmt.Errorf("could not diff env file: %w", err)
	}

	return envResult, envChanged, nil
}

func (b *Builder) Install(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	enabledApps := app.FilterEnabledApps(b.AllApps(), cfg)

	for _, a := range enabledApps {
		if installer, ok := a.(app.Installer); ok {
			slog.Info("installing app dependencies", "appName", a.Name())

			if err := installer.Install(cfg); err != nil {
				return fmt.Errorf("failed to install %s: %w", a.Name(), err)
			}

			slog.Info("app dependencies installed", "appName", a.Name())
		}
	}

	return nil
}

type AppStatus struct {
	Name    string
	Enabled bool
}

func (b *Builder) List() ([]AppStatus, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	all := b.AllApps()
	statuses := make([]AppStatus, 0, len(all))
	for _, a := range all {
		statuses = append(statuses, AppStatus{Name: a.Name(), Enabled: a.Enabled(cfg)})
	}
	return statuses, nil
}
