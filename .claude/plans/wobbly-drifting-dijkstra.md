# Centralize Claude Code Allowed Commands via Shizuku

## Context

Claude Code allowed commands (`permissions.allow`) are currently scattered across 4 project-specific `.claude/settings.local.json` files with heavy duplication. Common shell commands like `Bash(grep:*)`, `Bash(find:*)`, `Bash(ls:*)` appear in nearly every project.

The claude Shizuku app (already implemented) manages `~/.claude/settings.json` with an additive merge strategy for `enabledPlugins`. This plan extends that same merge to also manage `permissions.allow` — centralizing common shell and task commands while leaving project-specific ones (npx nx, cargo, scripts, Edit scopes, WebFetch domains) local.

## Changes

### `apps/claude/claude.go`

Add a `desiredAllowedCommands` variable and extend `mergeSettings` to merge them into `permissions.allow`.

**New variable:**
```go
var desiredAllowedCommands = []string{
    "Bash(grep:*)",
    "Bash(find:*)",
    "Bash(ls:*)",
    "Bash(tree:*)",
    "Bash(cat:*)",
    "Bash(wc:*)",
    "Bash(xargs:*)",
    "Bash(bash:*)",
    "Bash(task:*)",
    "Bash(git add:*)",
    "Bash(git commit:*)",
    "Bash(git --version:*)",
    "Bash(brew --prefix:*)",
    "Skill(task)",
}
```

**Merge logic addition in `mergeSettings`** (after the existing `enabledPlugins` merge):
1. Get or create `settings["permissions"]` as `map[string]any`
2. Get or create `permissions["allow"]` as `[]any`
3. Build a set of existing entries for dedup
4. Append each desired command not already present
5. Assign back to `settings["permissions"]["allow"]`

Same additive strategy as plugins — existing allowed commands are preserved.

## Verification

1. `/task build`
2. `/task run diff --print` — confirm `permissions.allow` appears in settings.json with all desired commands, existing fields preserved
3. `/task run sync --verbose`
