# Plugin Marketplace: Create 4 Reusable Plugins

## Context

We have a new plugin marketplace at `claude-code/` (name: `eleonorayaya-claude-code`). The goal is to consolidate reusable skills scattered across ~/dev projects into shareable plugins that work across machines. Marin-specific skills stay local.

## Plugins to Create

### 1. `git-workflows`
**Source:** `~/dev/marin/.claude/skills/checkpoint/` and `~/dev/marin/.claude/skills/git-workflow/`
**Adapt:** Strip Nx-specific content from git-workflow; keep conventional commits, atomic commits, safety checks, pre-commit hook handling. Keep checkpoint as-is (already generic).

```
claude-code/plugins/git-workflows/
├── .claude-plugin/
│   └── plugin.json
└── skills/
    ├── checkpoint/
    │   └── SKILL.md
    └── git-workflow/
        └── SKILL.md
```

### 2. `skill-authoring`
**Source:** `~/dev/utena/.claude/skills/skill-authoring/`
**Adapt:** This is a modular skill with reference files. Copy the main SKILL.md and all reference docs.

```
claude-code/plugins/skill-authoring/
├── .claude-plugin/
│   └── plugin.json
└── skills/
    └── skill-authoring/
        ├── SKILL.md
        └── references/
            ├── skill-md-format.md
            ├── reference-tables.md
            └── skill-placement.md
```

### 3. `go-patterns`
**Source:** `~/dev/utena/.claude/skills/go-patterns/`
**Adapt:** Modular skill with reference files. Copy main SKILL.md and all reference docs for layered Go architecture.

```
claude-code/plugins/go-patterns/
├── .claude-plugin/
│   └── plugin.json
└── skills/
    └── go-patterns/
        ├── SKILL.md
        └── references/
            ├── store-patterns.md
            ├── error-handling.md
            ├── testing.md
            ├── event-bus.md
            ├── chi-routing.md
            └── module-structure.md
```

### 4. `bubbletea-tui`
**Source:** `~/dev/utena/.claude/skills/bubbletea-tui/`
**Adapt:** Modular skill with reference files. Copy main SKILL.md and all reference docs.

```
claude-code/plugins/bubbletea-tui/
├── .claude-plugin/
│   └── plugin.json
└── skills/
    └── bubbletea-tui/
        ├── SKILL.md
        └── references/
            ├── key-bindings.md
            ├── list-views.md
            ├── view-routing.md
            └── window-sizing.md
```

## Steps

1. **Read all source skills** — Read full content of each source SKILL.md and reference files
2. **Create git-workflows plugin** — Directory structure, plugin.json, adapt and write skill files
3. **Create skill-authoring plugin** — Directory structure, plugin.json, copy skill files
4. **Create go-patterns plugin** — Directory structure, plugin.json, copy skill files
5. **Create bubbletea-tui plugin** — Directory structure, plugin.json, copy skill files
6. **Register all 4 plugins** in `claude-code/.claude-plugin/marketplace.json`

## Key Details

- All `plugin.json` files start at version `1.0.0`
- `marketplace.json` uses short source names (e.g., `"git-workflows"`) since `pluginRoot` is `./plugins`
- For git-workflow: strip references to Nx, marin-specific paths, and project-specific commands; keep the universal git practices
- For modular skills with references: preserve the reference file structure as-is

## Verification

```
/plugin marketplace add ./claude-code
/plugin install git-workflows@eleonorayaya-claude-code
# Verify /checkpoint and /git-workflow skills are available
```
