# Claude Code cleanup: migrating another machine

`task run -- sync` adds settings but never removes them, so a machine that synced before this cleanup keeps the removed plugins, mods, marketplaces and allow rules until you remove them. Run the steps below once on each such machine. A new machine needs only the sync.

## 1. Update the dotfiles checkout and the plugin marketplace

```bash
git -C <dotfiles checkout> pull --ff-only
claude plugin marketplace update eleonorayaya-claude-code
```

The update installs better-claude 2.0.0, which no longer blocks `find` and `sed -i`, and git-workflows 1.10.0, which no longer has the `checkpoint` and `git-workflow` skills.

## 2. Remove plugins and marketplaces that are no longer used

```bash
for p in superpowers@superpowers-marketplace prismantis@prismantis \
  go-patterns@eleonorayaya-claude-code cc-toolkit@eleonorayaya-claude-code \
  wails@go-skills go@go-skills cobra-viper@go-skills fileflow-pathologize@go-skills \
  go-release@go-skills go-spec-reviewer@go-skills charm-dev@charm-dev-skills \
  notion-workspace-plugin@notion-plugin-marketplace \
  skill-creator@claude-plugins-official ruby-lsp@claude-plugins-official; do
  claude plugin uninstall "$p" || true
done
for m in superpowers-marketplace prismantis go-skills charm-dev-skills notion-plugin-marketplace; do
  claude plugin marketplace remove "$m" || true
done
```

`|| true` skips plugins and marketplaces the machine never had.

## 3. Sync

From the dotfiles checkout:

```bash
task run -- sync
```

This moves `defaultMode: auto` under `permissions` and turns off the unused synced Anthropic skills (`skillOverrides`).

## 4. Remove stale settings by hand

```bash
f=~/.claude/settings.json; t=$(mktemp)
jq '.permissions.allow -= [
  "Bash(bash:*)", "Bash(xargs:*)", "Bash(find:*)",
  "Bash(git add:*)", "Bash(git commit:*)", "Bash(git fetch:*)", "Bash(git push:*)",
  "Bash(git rebase:*)", "Bash(git cherry-pick:*)",
  "Bash(grep:*)", "Bash(ls:*)", "Bash(tree:*)", "Bash(cat:*)", "Bash(wc:*)",
  "Bash(echo:*)", "Bash(head:*)", "Bash(tail:*)", "Skill(task)", "Bash(gh pr create*)"
] | del(.pluginConfigs["prismantis@prismantis"])' "$f" > "$t" && jq empty "$t" && cat "$t" > "$f"; rm -f "$t"
: > ~/.claude/CLAUDE.md
[ -L ~/.claude/skills/terminal-browser ] && [ ! -e ~/.claude/skills/terminal-browser ] && rm ~/.claude/skills/terminal-browser
```

- Sync doesn't write an empty global CLAUDE.md, so the `: >` line empties the old one.
- The last line deletes the old `terminal-browser` skill link only if it is broken. The terminal-browser plugin provides the skill.

## 5. Check

```bash
jq '{defaultMode: .permissions.defaultMode, skillOverrides: (.skillOverrides | length), plugins: (.enabledPlugins | keys)}' ~/.claude/settings.json
```

Expect `"auto"`, `13`, and no superpowers, prismantis, cc-toolkit, go-skills, charm-dev or notion entries. Start a new session to pick up the changes.
