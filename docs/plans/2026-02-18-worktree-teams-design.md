# Worktree Teams — Design Document

**Date:** 2026-02-18
**Status:** Draft
**Goal:** Replace the subtask CLI with a system that provides the same parallel worktree orchestration while respecting Claude Code's built-in safety mechanisms (permissions, hooks, plan mode).

## Problem Statement

The subtask CLI provided valuable capabilities for parallel work — isolated git worktrees per task, structured plan/review/implement workflows, and centralized dispatch/monitoring. However, it bypassed Claude's permission system, skipped hooks, and ran workers with no safety guardrails. This led to force pushes, unauthorized PR actions, and workers operating outside their intended scope.

We need a system that provides the same orchestration capabilities using Claude Code's native mechanisms (agent teams, skills, hooks, custom agents, plan mode) combined with a lightweight CLI for git plumbing that Claude can't handle natively.

## Architecture Overview

Four components work together:

```
+------------------------------------------------------+
|  wt CLI                                              |
|  Shell script. Manages:                              |
|  - Worktree creation/cleanup                         |
|  - Per-worktree agent definition generation          |
|  - PR stack graph (stack.json)                       |
|  - Safe git remote operations (push, rebase, pr)     |
|  - PR polling                                        |
|  - .gitignore enforcement for generated agents       |
+------------------------------------------------------+
           | generates
           v
+------------------------------------------------------+
|  Generated agent definitions (.claude/agents/wt-*)   |
|  One per worktree. Contains:                         |
|  - Hardcoded worktree path in hook commands           |
|  - PreToolUse hooks for file + git scoping            |
|  - System prompt with worktree context               |
|  - Preloaded worktree-worker skill                   |
+------------------------------------------------------+
           | referenced by
           v
+------------------------------------------------------+
|  Claude native teams + skills (plugin)               |
|  - worktree-lead skill: teaches orchestration        |
|  - worktree-worker skill: teaches workflow           |
|  - worktree-retrospective skill: session analysis    |
|  - TeamCreate / SendMessage / TaskCreate             |
|  - Plan mode for review gates                        |
|  - TeammateIdle / TaskCompleted hooks for QA         |
+------------------------------------------------------+
           | enforced by
           v
+------------------------------------------------------+
|  Hook scripts (.worktrees/hooks/)                    |
|  Copied from plugin on first wt create:              |
|  - file-guard.sh: Edit/Write path enforcement        |
|  - bash-guard.sh: command deny-list + git safety     |
|  - teammate-idle.sh: enforce poll-before-idle        |
|  - task-completed.sh: quality gate on completion     |
+------------------------------------------------------+
```

### File Layout

```
.worktrees/
  hooks/
    file-guard.sh             # Edit/Write: allow in worktree, deny outside
    bash-guard.sh             # Bash: deny dangerous commands, fall through rest
    teammate-idle.sh          # Enforce poll running before idle
    task-completed.sh         # Quality gate on task completion
  stack.json                  # Stack graph: worktrees, bases, PRs
  feature-x/                  # Git worktree
  feature-x/.poll-active      # Marker: poll is running
  feature-x/.poll-state.json  # Last known PR + base state
  feature-y/                  # Git worktree (stacked on feature-x)

.claude/
  agents/
    wt-feature-x.md           # Generated agent (gitignored via wt-*.md)
    wt-feature-y.md           # Generated agent (gitignored via wt-*.md)

# Plugin (worktree-teams):
  plugin.json
  skills/
    worktree-lead/skill.md
    worktree-worker/skill.md
    worktree-retrospective/skill.md
  hooks/hooks.json            # Session-level hooks (TeammateIdle, TaskCompleted)
  scripts/                    # Canonical copies of hook scripts
    file-guard.sh
    bash-guard.sh
    teammate-idle.sh
    task-completed.sh
    session-start.sh
```

## Distribution Model

### Plugin (`worktree-teams`)

Bundled as a Claude Code plugin. Contains:

- **Skills:** worktree-lead, worktree-worker, worktree-retrospective
- **Session-level hooks:** TeammateIdle, TaskCompleted, SessionStart (in `hooks/hooks.json`)
- **Hook scripts:** Canonical copies of guard scripts (in `scripts/`)

Plugin `hooks/hooks.json`:

```json
{
  "description": "Worktree Teams - session-level hooks",
  "hooks": {
    "TeammateIdle": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/teammate-idle.sh"
          }
        ]
      }
    ],
    "TaskCompleted": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/task-completed.sh"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/session-start.sh",
            "statusMessage": "Checking worktree-teams setup..."
          }
        ]
      }
    ]
  }
}
```

The `SessionStart` hook verifies the `wt` CLI is installed and at a compatible version.

### CLI (`wt`)

Standalone shell script, installed separately. On first `wt create`, copies hook scripts from the plugin to `.worktrees/hooks/` so generated agent definitions can reference them via `${CLAUDE_PROJECT_DIR}/.worktrees/hooks/`.

### Path Resolution

Generated agents are project-level files (`.claude/agents/wt-*.md`). They cannot use `${CLAUDE_PLUGIN_ROOT}` because that variable is only available to plugin-defined hooks. Instead:

1. `wt create` locates the plugin at `~/.claude/plugins/cache/*/worktree-teams/*/scripts/`
2. Copies scripts to `.worktrees/hooks/` (one-time per project)
3. Generated agents reference `${CLAUDE_PROJECT_DIR}/.worktrees/hooks/`

## CLI: `wt` Commands

### `wt create <name> --base <branch>`

- Creates git worktree at `.worktrees/<name>/` with branch `<name>`
- `--base` can be `main` or another worktree branch (for stacking)
- Generates `.claude/agents/wt-<name>.md` with hardcoded worktree path
- Updates `.worktrees/stack.json` with dependency edge
- On first run: copies hook scripts from plugin to `.worktrees/hooks/`
- Verifies `.claude/agents/wt-*.md` is in `.gitignore` (adds if needed, or errors)

### `wt list [--stack]`

- Lists all active worktrees with status (branch, base, dirty/clean)
- `--stack` shows the dependency graph in topological order

### `wt stack`

- Prints the full stack graph showing dependency order
- Marks which branches have open PRs, which are merged

### `wt info <name>`

- Prints current state for a worktree: path, branch, base, PR, dependents, stack position
- Workers call this for dynamic state instead of caching values

### `wt diff <name> [--stat]`

- Shows diff of worktree from its merge base
- `--stat` for summary view

### `wt push <name>`

- Validates the branch matches the worktree
- Pushes to `origin/<branch>` only
- Never force pushes

### `wt pr <name>`

- Creates a **draft** PR (always draft, no flag needed)
- Sets PR base branch from `stack.json` (parent worktree's branch or main)
- Records PR number/URL in `stack.json`

### `wt rebase <name>`

- Reads `stack.json` for the current base branch
- If base branch exists: rebases onto it
- If base branch was deleted (PR merged): finds base's base, rebases onto that, updates `stack.json`
- If already up to date: no-op
- The CLI is the single source of truth for rebase targets — workers never figure this out themselves

### `wt poll <name> [--timeout <minutes>]`

- Reads PR number from `stack.json`
- Writes `.worktrees/<name>/.poll-active` marker while running
- Tracks PR state + base branch HEAD in `.poll-state.json`
- Polls at interval (e.g., 30s) for changes
- On detecting a change, removes marker, outputs structured summary, exits
- Change types:
  - `base_updated`: base branch has new commits or was merged
  - `ci_failure`: CI checks failed
  - `ci_passed`: CI checks now passing
  - `review_comments`: new review comments on PR

### `wt cleanup <name>`

- Removes the git worktree
- Deletes `.claude/agents/wt-<name>.md`
- Updates `stack.json` (removes node, repoints any dependents to the cleaned-up node's base)

### `wt cleanup --all`

- Cleans up all worktrees and generated agents

### `wt check-ignore`

- Verifies `.claude/agents/wt-*.md` is gitignored
- Called internally by `wt create`, can also be run standalone
- Exits non-zero with instructions if not gitignored

### `stack.json` Format

```json
{
  "worktrees": {
    "branch-a": {
      "path": "/abs/.worktrees/branch-a",
      "branch": "branch-a",
      "base": "main",
      "pr": null
    },
    "branch-b": {
      "path": "/abs/.worktrees/branch-b",
      "branch": "branch-b",
      "base": "branch-a",
      "pr": { "number": 1235, "url": "https://github.com/org/repo/pull/1235" }
    }
  }
}
```

The dependency graph is implicit from the `base` fields.

## Generated Agent Definitions

When `wt create feature-y --base feature-x` runs, the CLI generates `.claude/agents/wt-feature-y.md`:

```yaml
---
name: wt-feature-y
description: Worker agent scoped to worktree feature-y
tools: Read, Edit, Write, Bash, Grep, Glob
skills:
  - worktree-worker
hooks:
  PreToolUse:
    - matcher: "Edit|Write"
      hooks:
        - type: command
          command: >-
            ${CLAUDE_PROJECT_DIR}/.worktrees/hooks/file-guard.sh
            /abs/path/.worktrees/feature-y
    - matcher: "Bash"
      hooks:
        - type: command
          command: >-
            ${CLAUDE_PROJECT_DIR}/.worktrees/hooks/bash-guard.sh
            /abs/path/.worktrees/feature-y feature-y
---

You are a worker agent operating in an isolated git worktree.

## Your workspace
- **Worktree path:** /abs/path/.worktrees/feature-y
- **Branch:** feature-y

## Dynamic state
Your base branch, stack position, PR number, and dependents are
managed by the `wt` CLI and stored in .worktrees/stack.json.
Always use CLI commands for current state - never rely on
previously cached values:
- `wt info feature-y` - your current base, PR, dependents
- `wt stack` - full dependency graph
- `wt rebase feature-y` - rebase onto current base
- `wt push feature-y` - push to remote (only your branch)
- `wt pr feature-y` - create draft PR
- `wt poll feature-y` - poll for PR changes (must be background)
- `wt diff feature-y` - view changes from merge base

## Rules
- All file modifications MUST target files within your worktree path.
- Never force push. Never push to main or master.
- Never mark PRs as ready for review.
- Never reply to PR review comments.
- Do not push unless explicitly asked or your task requires it.
- Remote git operations (push, fetch, pull) must go through the wt CLI.
- GitHub write operations (pr create, comment, review) must go through the wt CLI.
- GitHub read operations (pr view, pr checks, pr diff, api GET) are allowed directly.
- `wt poll` must always run as a background task.
```

### What's static vs dynamic

| Static (in agent definition) | Dynamic (in stack.json, read via CLI) |
|---|---|
| Worktree path | Base branch |
| Branch name | Stack position |
| Hook commands + paths | PR number/URL |
| Permission mode, tools, skills | Dependents list |
| Rules | Merge state |

Agent definitions cannot be regenerated once a worker is spawned. All mutable state lives in `stack.json` and is read at runtime via `wt info`, `wt rebase`, etc.

## Permission Model

### Layers

```
Read, Grep, Glob
  -> Listed in agent tools, read-only, no hooks needed

Edit, Write
  -> file-guard.sh:
       in worktree path? -> permissionDecision: "allow" (auto-approve)
       outside?          -> permissionDecision: "deny"

Bash
  -> bash-guard.sh deny-list:
       git push/fetch/pull   -> deny ("use wt CLI")
       gh write operations   -> deny ("use wt CLI")
       wt poll (foreground)  -> deny ("must be background")
     everything else         -> exit 0 (fall through to
                                inherited project permissions)
```

### file-guard.sh

```bash
#!/bin/bash
# Allows edits only within the worktree path
WORKTREE_PATH="$1"
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

if [[ -n "$FILE_PATH" && "$FILE_PATH" == "$WORKTREE_PATH"* ]]; then
  jq -n '{hookSpecificOutput:{hookEventName:"PreToolUse",
    permissionDecision:"allow",
    permissionDecisionReason:"File is within worktree scope"}}'
else
  jq -n --arg wt "$WORKTREE_PATH" '{hookSpecificOutput:{hookEventName:"PreToolUse",
    permissionDecision:"deny",
    permissionDecisionReason:("File outside worktree: " + $wt)}}'
fi
```

### bash-guard.sh

```bash
#!/bin/bash
# Deny-list for dangerous Bash commands. Everything else falls through
# to inherited project permissions.
WORKTREE_PATH="$1"
BRANCH="$2"
INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')
BACKGROUND=$(echo "$INPUT" | jq -r '.tool_input.run_in_background // false')

deny() {
  jq -n --arg r "$1" '{hookSpecificOutput:{hookEventName:"PreToolUse",
    permissionDecision:"deny",
    permissionDecisionReason:$r}}'
  exit 0
}

# --- Deny: git remote operations (must use wt CLI) ---
echo "$COMMAND" | grep -qE '^git\s+push\b'           && deny "Use 'wt push $BRANCH' instead"
echo "$COMMAND" | grep -qE '^git\s+(fetch|pull)\b'    && deny "Use 'wt rebase $BRANCH' for updates"
echo "$COMMAND" | grep -qE 'push\s+.*(-f|--force)'    && deny "Force push is not allowed"

# --- Deny: gh write operations (must use wt CLI) ---
echo "$COMMAND" | grep -qE '^gh\s+pr\s+(create|ready|merge|close|edit|comment|review)\b' \
  && deny "GitHub write operations must go through wt CLI"
echo "$COMMAND" | grep -qE '^gh\s+api\s+.*-X\s*(POST|PUT|PATCH|DELETE)\b' \
  && deny "GitHub write API calls must go through wt CLI"

# --- Deny: wt poll must be background ---
if echo "$COMMAND" | grep -qE '^wt\s+poll\b'; then
  [ "$BACKGROUND" != "true" ] && deny "wt poll must run as a background task"
fi

# --- Everything else: fall through to project permissions ---
exit 0
```

### Permission flow summary

| Tool | Hook | Behavior |
|---|---|---|
| Read | none | Always allowed (read-only) |
| Grep | none | Always allowed (read-only) |
| Glob | none | Always allowed (read-only) |
| Edit | file-guard.sh | In worktree -> auto-approve. Outside -> deny |
| Write | file-guard.sh | In worktree -> auto-approve. Outside -> deny |
| Bash: `wt *` | bash-guard.sh | Allowed (except poll must be background) |
| Bash: `git push/fetch/pull` | bash-guard.sh | Deny -> use wt CLI |
| Bash: `gh pr view/checks/diff` | bash-guard.sh | Falls through -> allowed |
| Bash: `gh pr create/ready/...` | bash-guard.sh | Deny -> use wt CLI |
| Bash: `gh api -X POST/...` | bash-guard.sh | Deny -> use wt CLI |
| Bash: everything else | bash-guard.sh | Falls through -> inherited project permissions |

## Session-Level Hooks

### teammate-idle.sh

Enforces that workers with an open PR have an active background poll before going idle.

```bash
#!/bin/bash
INPUT=$(cat)
TEAMMATE=$(echo "$INPUT" | jq -r '.teammate_name')

# Only applies to worktree workers
[[ "$TEAMMATE" != wt-* ]] && exit 0

NAME="${TEAMMATE#wt-}"
STATE="$CLAUDE_PROJECT_DIR/.worktrees/stack.json"
PR=$(jq -r --arg n "$NAME" '.worktrees[$n].pr.number // empty' "$STATE" 2>/dev/null)

# No PR yet - ok to idle
[ -z "$PR" ] && exit 0

# Check for active poll marker
MARKER="$CLAUDE_PROJECT_DIR/.worktrees/$NAME/.poll-active"
if [ -f "$MARKER" ]; then
  exit 0  # poll running, ok to idle
fi

# Block idle - worker must start poll first
echo "You have PR #$PR but no active background poll. Run: wt poll $NAME (as a background task) before going idle." >&2
exit 2
```

### task-completed.sh

Quality gate that can verify work before a task is marked complete.

```bash
#!/bin/bash
INPUT=$(cat)
TEAMMATE=$(echo "$INPUT" | jq -r '.teammate_name // empty')

# Only applies to worktree workers
[[ "$TEAMMATE" != wt-* ]] && exit 0

# Additional quality checks can be added here:
# - Verify tests pass
# - Check for uncommitted changes
# - Validate changes are within worktree scope

exit 0
```

## Stack Lifecycle

### Creating a stack

```
User: "Implement feature A, then feature B on top of it,
       then feature C on top of that"

Lead:
  wt create branch-a --base main
  wt create branch-b --base branch-a
  wt create branch-c --base branch-b
```

`wt stack` output:
```
main
 +-- branch-a
      +-- branch-b
           +-- branch-c
```

### Spawning workers

```
TeamCreate("my-stack")

Task(name: "branch-a", subagent_type: "wt-branch-a",
     team_name: "my-stack", mode: "plan",
     prompt: "Implement feature A...")

Task(name: "branch-b", subagent_type: "wt-branch-b",
     team_name: "my-stack", mode: "plan",
     prompt: "Implement feature B...")

Task(name: "branch-c", subagent_type: "wt-branch-c",
     team_name: "my-stack", mode: "plan",
     prompt: "Implement feature C...")
```

All three start in plan mode. They explore and draft plans in parallel (isolated worktrees), then the lead reviews and approves each plan.

### Poll-driven cascade

Every worker with a PR runs `wt poll <name>` as a background task. The poll watches two things: the PR state (CI, comments, merge) and the base branch HEAD.

When any change is detected, the worker wakes up and responds:

| Poll event | Worker response |
|---|---|
| `base_updated` | `wt rebase <name>`, `wt push <name>`, new poll |
| `ci_failure` | Read failures, fix code, `wt push <name>`, new poll |
| `ci_passed` | No action, continue polling |
| `review_comments` | Read comments (gh read allowed), address in code, `wt push <name>`, new poll |

### Cascade propagation

Cascades propagate automatically through the stack via polling. No inter-worker messaging is needed:

```
branch-a worker pushes a fix
  -> branch-b's poll detects base_updated
  -> branch-b runs wt rebase, wt push
  -> branch-c's poll detects base_updated
  -> branch-c runs wt rebase, wt push
```

The cascade works identically for:
- Review comment fixes (push changes base HEAD)
- CI fixes (push changes base HEAD)
- PR merges (base branch deleted, `wt rebase` repoints to base's base)

Workers never tell each other to rebase. Each worker independently detects its base moved and asks the CLI (`wt rebase`) what to do. The CLI is the single source of truth for rebase targets.

### Merge handling

When a PR merges, the base branch is typically deleted. `wt rebase` handles this:

1. Reads stack.json: branch-b's base is `branch-a`
2. `branch-a` branch is gone (merged)
3. Finds branch-a's base was `main`
4. Rebases branch-b onto `main`
5. Updates stack.json: branch-b's base is now `main`

The worker then pushes, which triggers the next downstream worker's poll.

### Cleanup

`wt cleanup branch-a` removes:
- The git worktree `.worktrees/branch-a/`
- The agent definition `.claude/agents/wt-branch-a.md`
- The node from `stack.json` (repoints any dependents if needed)

## Skills

### worktree-lead

Loaded into the user's main session via the Skill tool. Teaches orchestration:

- How to create worktrees and stacks with `wt create`
- How to spawn workers with the correct generated agent type (`wt-<name>`)
- Always use `mode: "plan"` for the review gate
- Monitor progress via TaskList
- Workers self-manage their PR lifecycle and stack cascades
- The user controls merging — never mark PRs as ready

### worktree-worker

Preloaded into every generated agent via `skills: [worktree-worker]`. Subagents don't inherit skills from the parent, so this must be explicitly listed. Teaches:

- The full lifecycle: plan -> implement -> push -> PR -> poll
- How to respond to each poll event type
- All dynamic state comes from `wt` CLI, never cache values
- Rules enforcement (what's allowed, what's blocked, why)

### worktree-retrospective

For reviewing session logs to find plugin issues. Teaches how to:

- Find permission gaps (legitimate commands denied by guards)
- Find permission holes (dangerous commands that got through)
- Identify workflow breakdowns (missing polls, stale state, skipped rebases)
- Find CLI bugs (rebase target errors, stack.json desync)
- Produce actionable reports: commands to add/remove from allow/deny lists

Uses the session parser scripts from the retrospectives directory and teaches writing targeted parsers for worker-specific log analysis.

## Design Decisions

### Why a CLI + plugin, not pure skills?

The CLI handles git plumbing that requires persistent state (stack.json, worktree lifecycle, poll markers) and reliable execution (push validation, rebase target resolution). Skills provide the intelligence layer — teaching agents how and when to use the CLI. This separation means the CLI can be tested independently and the state survives session crashes.

### Why generated agents instead of a single worker agent?

Each worktree needs its own hardcoded path in the PreToolUse hooks. A single generic agent would need runtime path resolution (state files, session binding, etc.). Generated agents are self-contained — the path is baked in at creation time, no lookup needed.

### Why poll-driven cascades instead of inter-worker messaging?

Each worker independently watches its own base branch. When the base moves (for any reason — push, merge, rebase), the worker detects it and responds. This is simpler and more robust than having workers coordinate with each other. It also means workers don't need to know who depends on them.

### Why `wt rebase` instead of workers figuring out rebase targets?

The CLI owns `stack.json` and the git state. It can handle edge cases (deleted branches, repointing after merges, already-up-to-date checks) reliably. Workers just run `wt rebase <name>` and trust the CLI to do the right thing.

### Why PreToolUse hooks return allow/deny instead of using dontAsk mode?

`dontAsk` mode auto-denies anything not in permission allow rules, but allow rules are static and can't be parameterized per-agent at spawn time. PreToolUse hooks with `permissionDecision: "allow"` auto-approve within-worktree operations without prompting, while `"deny"` blocks dangerous operations. Everything else falls through to inherited project permissions. This gives us dynamic scoping without fighting the permission system.

### Why always draft PRs?

The user controls when PRs are marked ready for review. Workers should never mark PRs ready without authorization. By making `wt pr` always create drafts and denying `gh pr ready`, this class of problems is eliminated.
