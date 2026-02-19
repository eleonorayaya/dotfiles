# Claude Code Plugin Marketplace

Personal plugin marketplace for reusable Claude Code skills and configuration.

## Marketplace Structure

```
claude-code/
├── .claude-plugin/
│   └── marketplace.json    # marketplace catalog
├── plugins/
│   └── <plugin-name>/
│       ├── .claude-plugin/
│       │   └── plugin.json # plugin manifest
│       ├── skills/
│       │   └── <skill-name>/
│       │       └── SKILL.md
│       ├── agents/         # optional
│       └── hooks/          # optional
└── CLAUDE.md
```

## Adding a Plugin

1. Create the plugin directory and manifest:

```bash
mkdir -p plugins/<plugin-name>/.claude-plugin
mkdir -p plugins/<plugin-name>/skills
```

2. Create `plugins/<plugin-name>/.claude-plugin/plugin.json`:

```json
{
  "name": "<plugin-name>",
  "description": "What this plugin does",
  "version": "1.0.0"
}
```

3. Add the plugin to `.claude-plugin/marketplace.json` in the `plugins` array:

```json
{
  "name": "<plugin-name>",
  "source": "<plugin-name>",
  "description": "What this plugin does"
}
```

Sources use short names because `metadata.pluginRoot` is set to `./plugins`.

## Adding a Skill to a Plugin

Create `plugins/<plugin-name>/skills/<skill-name>/SKILL.md`:

```markdown
---
description: When to invoke this skill
---

Skill instructions here.
```

The skill becomes available as `/<skill-name>` after installation.

## Versioning

Bump the `version` in the plugin's `plugin.json` when making changes. Claude Code uses the version to detect updates and manage its cache.

## Installing Locally

```
/plugin marketplace add ./claude-code
/plugin install <plugin-name>@claude-code
```

## Updating After Changes

```
/plugin marketplace update claude-code
```

## Key Constraints

- Plugins are copied to a cache on install. Files outside a plugin's directory (like `../shared-utils`) won't be available. Use symlinks if sharing files between plugins.
- Use `${CLAUDE_PLUGIN_ROOT}` in hooks and MCP server configs to reference files within the installed plugin directory.
- Skill names must be unique across all installed plugins.
