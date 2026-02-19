---
name: checkpoint
description: Create a WIP (Work In Progress) commit for the current changes. Use during iterative development to mark safe restore points between significant steps. Helps track progress and enables easy rollback if needed.
---

# Checkpoint Skill

Creates a WIP (Work In Progress) commit to save the current state of work during iterative development. This provides safe restore points and helps track progress through multi-step changes.

## When to Use

Invoke this skill when:
- Making progress on a multi-step feature
- Completing one piece of user feedback in a series
- Finishing a logical unit of work but not ready for final commit
- Want to save state before trying a risky change
- Implementing iterative improvements based on user feedback
- Need a rollback point during complex refactoring

**User phrases:**
- "checkpoint this"
- "save this progress"
- "make a WIP commit"
- "commit this for now"

**Proactive usage:**
- After completing each distinct improvement during iterative feedback
- Before starting an architectural change
- After each successful refactoring step

## What This Skill Does

1. **Check git status** to see what changes exist
2. **Stage all changes** (modified, new, and deleted files)
3. **Create descriptive WIP commit** based on what changed
4. **Verify commit** was created successfully

## Process

### Step 1: Check Current State

Run `git status` to see what has changed:
```bash
git status
```

### Step 2: Review Changes

Optionally run `git diff` to review what will be committed:
```bash
git diff
```

### Step 3: Stage All Changes

Stage all modifications, additions, and deletions:
```bash
git add -A
```

### Step 4: Create WIP Commit

Create a commit with a descriptive message that starts with "WIP:":

```bash
git commit -m "WIP: <brief description of changes>"
```

**Commit message format:**
- Start with "WIP: " prefix
- Use conventional commit type if clear (feat/fix/refactor)
- Briefly describe what was changed
- Include scope if applicable

**Examples:**
```bash
# Generic WIP
git commit -m "WIP: add manga search functionality"

# With conventional commit type
git commit -m "WIP(auth): refactor repositories to generic pattern"

# After implementing user feedback
git commit -m "WIP(api): move transformation to client layer"

# During multi-step refactoring
git commit -m "WIP: extract shared transformation logic"
```

### Step 5: Verify Success

Run `git log -1` to confirm the commit was created:
```bash
git log -1 --oneline
```

## Integration with Development Workflow

### Iterative Development Pattern

When user provides multiple pieces of feedback:
```
1. Implement first improvement
2. /checkpoint (WIP commit)
3. Implement second improvement
4. /checkpoint (WIP commit)
5. Implement third improvement
6. /commit (final, squashed commit with comprehensive message)
```

### Refactoring Workflow

During complex refactoring:
```
1. Add new generic methods alongside old specific methods
2. /checkpoint "add generic repository methods"
3. Migrate first caller to use new methods
4. /checkpoint "migrate MediaService to generic methods"
5. Migrate remaining callers
6. /checkpoint "migrate all callers to generic methods"
7. Remove old specific methods
8. /commit (final commit)
```

### Before Risky Changes

Before attempting something that might not work:
```
1. Implement working version
2. /checkpoint "working state before optimization"
3. Try optimization
4. If it fails: git reset --hard HEAD (rollback to checkpoint)
5. If it works: continue and /commit when done
```

## WIP vs Final Commits

**WIP Commits:**
- Mark progress during development
- May not pass all tests/linting
- May have incomplete features
- Safe to rewrite history (squash, amend)
- Stay on feature branch
- Not pushed to main/master

**Final Commits:**
- Complete, working features
- Pass all tests and linting
- Have comprehensive commit messages
- Include all related changes
- Ready for pull request
- Should not be rewritten after push

## Squashing WIP Commits

Before creating a pull request, squash WIP commits into a comprehensive final commit:

```bash
# Interactive rebase to squash last N commits
git rebase -i HEAD~N

# In the editor, mark commits to squash:
pick abc1234 WIP: add manga search
squash def5678 WIP: move transformation to client
squash ghi9012 WIP: add bulk lookup optimization

# Write comprehensive final commit message
# Save and exit
```

Or use git reset to create a single new commit:
```bash
# Soft reset to N commits ago (keeps changes staged)
git reset --soft HEAD~N

# Create new comprehensive commit
git commit -m "feat(scope): add feature with integration"
```

## Best Practices

### DO:
- Checkpoint after each logical step
- Use descriptive WIP messages
- Checkpoint before risky changes
- Squash WIPs into final commit before PR
- Use checkpoints during iterative feedback

### DON'T:
- Push WIP commits to main/master
- Let WIP commits accumulate without eventual squash
- Use "WIP" for production-ready changes (use proper /commit instead)
- Skip checkpoints during complex multi-step work

## Relationship to Other Skills

- **`/checkpoint`**: Quick WIP commit during development
- **`/commit`**: Final, comprehensive commit ready for PR
- **`/create-pr`**: Create pull request (requires final commits, not WIPs)

**Workflow example:**
```
Development:
/checkpoint -> /checkpoint -> /checkpoint

Before PR:
git rebase -i (squash WIPs)
/create-pr (with clean history)
```

## Examples

### Example 1: Iterative Improvements

User: "Move transformation to client, consolidate status mapping, and remove custom DTOs"

```
Agent implements first change...
Agent: /checkpoint

WIP commit created:
"WIP(api): move data transformation to client"

Agent implements second change...
Agent: /checkpoint

WIP commit created:
"WIP(api): consolidate status mapping in client"

Agent implements third change...
Agent: /checkpoint

WIP commit created:
"WIP(api): remove custom response DTOs"

Agent squashes and creates final commit:
"refactor(api): consolidate transformation logic in client layer"
```

### Example 2: Refactoring Safety

```
Agent: I'm about to refactor the repository pattern. Let me checkpoint the current working state first.
Agent: /checkpoint

WIP commit created:
"WIP: checkpoint before repository refactoring"

Agent implements refactoring...
Agent encounters issue...
Agent: git reset --hard HEAD  (rollback to checkpoint)
Agent tries different approach...
Agent: /checkpoint

WIP commit created:
"WIP: implement generic repository pattern"
```

### Example 3: Multi-Step Feature

```
Agent implements database schema changes...
Agent: /checkpoint

WIP commit created:
"WIP(db): add new field to entity"

Agent implements backend API...
Agent: /checkpoint

WIP commit created:
"WIP(api): add search endpoint"

Agent implements frontend UI...
Agent: /checkpoint

WIP commit created:
"WIP(ui): add search component"

Agent squashes all WIPs and creates final commit:
"feat: add search functionality with API integration"
```

## Error Handling

If git operations fail:

**Merge conflicts:**
- Resolve conflicts manually
- Run `/checkpoint` again after resolution

**Nothing to commit:**
- Verify changes exist with `git status`
- Check if changes were already committed
- Skip checkpoint if no new changes

**Pre-commit hooks fail:**
- Fix linting/test failures
- Run `/checkpoint` again after fixes
- Or use `--no-verify` flag if truly WIP

## Summary

The `/checkpoint` skill helps you:
- Mark progress during development
- Enable easy rollback to working states
- Track iterative improvements
- Create safe restore points before risky changes
- Structure work into logical steps

Use it liberally during development, then squash WIP commits into comprehensive final commits before creating pull requests.
