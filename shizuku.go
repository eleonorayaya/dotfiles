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
	"github.com/eleonorayaya/shizuku/styles"
)

type Options struct {
	OutDir  string
	Profile string
	Verbose bool
}

type Profile struct {
	Languages []app.Language
	Programs  []app.Program
	Agents    []app.Agent
}

type Builder struct {
	opts     Options
	styles   styles.Styles
	base     Profile
	profiles map[string]*Profile
	target   *Profile
}

type Option func(*Builder)

func WithOutDir(dir string) Option {
	return func(b *Builder) { b.opts.OutDir = dir }
}

func WithVerbose(v bool) Option {
	return func(b *Builder) { b.opts.Verbose = v }
}

func WithProfileName(name string) Option {
	return func(b *Builder) { b.opts.Profile = name }
}

func WithStyles(s styles.Styles) Option {
	return func(b *Builder) { b.styles = s }
}

func WithLanguages(langs ...app.Language) Option {
	return func(b *Builder) {
		b.target.Languages = append(b.target.Languages, langs...)
	}
}

func WithPrograms(progs ...app.Program) Option {
	return func(b *Builder) {
		b.target.Programs = append(b.target.Programs, progs...)
	}
}

func WithAgents(agents ...app.Agent) Option {
	return func(b *Builder) {
		b.target.Agents = append(b.target.Agents, agents...)
	}
}

func WithProfile(name string, opts ...Option) Option {
	return func(b *Builder) {
		p, ok := b.profiles[name]
		if !ok {
			p = &Profile{}
			b.profiles[name] = p
		}
		prev := b.target
		b.target = p
		for _, opt := range opts {
			opt(b)
		}
		b.target = prev
	}
}

func New(opts ...Option) *Builder {
	b := &Builder{
		styles:   styles.New(),
		profiles: map[string]*Profile{},
	}
	b.target = &b.base
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func mergeNamed[T app.Named](base, overlay []T) []T {
	idx := map[string]int{}
	out := make([]T, 0, len(base)+len(overlay))
	for _, a := range base {
		idx[a.Name()] = len(out)
		out = append(out, a)
	}
	for _, a := range overlay {
		if i, ok := idx[a.Name()]; ok {
			out[i] = a
		} else {
			idx[a.Name()] = len(out)
			out = append(out, a)
		}
	}
	return out
}

func (b *Builder) activeProfile() Profile {
	if b.opts.Profile == "" {
		return b.base
	}
	p, ok := b.profiles[b.opts.Profile]
	if !ok {
		return b.base
	}
	return Profile{
		Languages: mergeNamed(b.base.Languages, p.Languages),
		Programs:  mergeNamed(b.base.Programs, p.Programs),
		Agents:    mergeNamed(b.base.Agents, p.Agents),
	}
}

func (b *Builder) makeContext(outDir string) *app.Context {
	return &app.Context{
		OutDir:  outDir,
		Profile: b.opts.Profile,
		Styles:  b.styles,
	}
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

func collectAgentContext(profile Profile) app.AgentContext {
	providers := []app.AgentConfigProvider{}
	for _, l := range profile.Languages {
		if p, ok := l.(app.AgentConfigProvider); ok {
			providers = append(providers, p)
		}
	}
	for _, p := range profile.Programs {
		if pv, ok := p.(app.AgentConfigProvider); ok {
			providers = append(providers, pv)
		}
	}
	return app.CollectAgentConfigs(providers)
}

func collectEnvSetups(profile Profile) ([]*app.EnvSetup, error) {
	envSetups := []*app.EnvSetup{}
	collect := func(named app.Named) error {
		provider, ok := named.(app.EnvProvider)
		if !ok {
			return nil
		}
		envSetup, err := provider.Env()
		if err != nil {
			return fmt.Errorf("failed to get env setup for %s: %w", named.Name(), err)
		}
		envSetups = append(envSetups, envSetup)
		return nil
	}
	for _, l := range profile.Languages {
		if err := collect(l); err != nil {
			return nil, err
		}
	}
	for _, p := range profile.Programs {
		if err := collect(p); err != nil {
			return nil, err
		}
	}
	for _, a := range profile.Agents {
		if err := collect(a); err != nil {
			return nil, err
		}
	}
	return envSetups, nil
}

func (b *Builder) Sync(ctx context.Context) error {
	outDir, err := b.resolveOutDir()
	if err != nil {
		return err
	}
	appCtx := b.makeContext(outDir)
	profile := b.activeProfile()

	for _, p := range profile.Programs {
		if err := syncProgram(p, appCtx); err != nil {
			return err
		}
	}

	agentCtx := collectAgentContext(profile)
	for _, a := range profile.Agents {
		slog.Info("app syncing", "appName", a.Name())
		if err := a.Sync(appCtx, agentCtx); err != nil {
			return fmt.Errorf("could not sync %s: %w", a.Name(), err)
		}
		slog.Info("app synced", "appName", a.Name())
	}

	return b.syncEnv(profile, outDir)
}

func syncProgram(p app.Program, ctx *app.Context) error {
	syncer, ok := p.(app.FileSyncer)
	if !ok {
		return nil
	}
	slog.Info("app syncing", "appName", p.Name())
	if err := syncer.Sync(ctx); err != nil {
		return fmt.Errorf("could not sync %s: %w", p.Name(), err)
	}
	slog.Info("app synced", "appName", p.Name())
	return nil
}

func (b *Builder) syncEnv(profile Profile, outDir string) error {
	envSetups, err := collectEnvSetups(profile)
	if err != nil {
		return err
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
	outDir, err := b.resolveOutDir()
	if err != nil {
		return nil, err
	}
	appCtx := b.makeContext(outDir)
	profile := b.activeProfile()

	var results []DiffResult

	for _, p := range profile.Programs {
		generator, ok := p.(app.FileGenerator)
		if !ok {
			continue
		}
		slog.Debug("generating files for diff", "appName", p.Name())
		result, err := generator.Generate(appCtx)
		if err != nil {
			return nil, fmt.Errorf("could not generate %s: %w", p.Name(), err)
		}
		changed, err := app.DiffAppFiles(result)
		if err != nil {
			return nil, fmt.Errorf("could not diff %s: %w", p.Name(), err)
		}
		if len(changed) > 0 {
			sort.Strings(changed)
			results = append(results, DiffResult{Name: p.Name(), Changed: changed, FileMap: result.FileMap})
		}
	}

	agentCtx := collectAgentContext(profile)
	for _, a := range profile.Agents {
		slog.Debug("generating files for diff", "appName", a.Name())
		result, err := a.Generate(appCtx, agentCtx)
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

	envResult, envChanged, err := b.diffEnv(profile, outDir)
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

func (b *Builder) diffEnv(profile Profile, outDir string) (*app.GenerateResult, []string, error) {
	envSetups, err := collectEnvSetups(profile)
	if err != nil {
		return nil, nil, err
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
	outDir, err := b.resolveOutDir()
	if err != nil {
		return err
	}
	appCtx := b.makeContext(outDir)
	profile := b.activeProfile()

	install := func(named app.Named) error {
		installer, ok := named.(app.Installer)
		if !ok {
			return nil
		}
		slog.Info("installing app dependencies", "appName", named.Name())
		if err := installer.Install(appCtx); err != nil {
			return fmt.Errorf("failed to install %s: %w", named.Name(), err)
		}
		slog.Info("app dependencies installed", "appName", named.Name())
		return nil
	}

	for _, l := range profile.Languages {
		if err := install(l); err != nil {
			return err
		}
	}
	for _, p := range profile.Programs {
		if err := install(p); err != nil {
			return err
		}
	}
	for _, a := range profile.Agents {
		if err := install(a); err != nil {
			return err
		}
	}
	return nil
}

type AppStatus struct {
	Name     string
	Category string
}

func (b *Builder) List() []AppStatus {
	profile := b.activeProfile()
	statuses := make([]AppStatus, 0, len(profile.Languages)+len(profile.Programs)+len(profile.Agents))
	for _, l := range profile.Languages {
		statuses = append(statuses, AppStatus{Name: l.Name(), Category: "language"})
	}
	for _, p := range profile.Programs {
		statuses = append(statuses, AppStatus{Name: p.Name(), Category: "program"})
	}
	for _, a := range profile.Agents {
		statuses = append(statuses, AppStatus{Name: a.Name(), Category: "agent"})
	}
	return statuses
}
