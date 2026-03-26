# ergo Quickstart -- Complete Reference Manual

## 1. Initialize

```bash
ergo init                    # Create .ergo/ in current directory
ergo where                   # Print active .ergo/ path
```

## 2. Agent Workflow

Identity format: `<model>@<hostname>` (e.g. `sonnet@agent-host`).

### Core Loop

```bash
# 1. Claim oldest ready work
ergo --json claim --agent sonnet@agent-host
# Output includes reminder: "When you have completed this claimed task, you MUST mark it done."

# 2. Do the work described in the task body

# 3. Attach results and set state
printf '%s' '{"result_path":"docs/x.md","result_summary":"Spec v1","state":"done"}' | ergo set ABCDEF
# Never leave tasks in doing. Always set done, blocked, error, or canceled.

# 4. Pull full epic context
ergo --json show EPICID
# Returns {"epic": ..., "children": ...} with body for each child
```

### Common Variants

```bash
# Claim specific task by ID
ergo claim ABCDEF --agent sonnet@agent-host

# Claim oldest ready within epic
ergo claim --agent sonnet@agent-host --epic ABCDEF

# Change state (uses existing claim if present)
printf '%s' '{"state":"blocked"}' | ergo set ABCDEF

# If no tasks ready: prints message, exit 0
```

## 3. Create Work (JSON stdin)

Use `printf '%s'` for reliable JSON delivery. All fields are strings. Unknown keys rejected with suggestions.

```bash
# Task
printf '%s' '{"title":"Login","body":"Implement signup","epic":"OFKSTE"}' | ergo new task

# Epic
printf '%s' '{"title":"Auth","body":"Signup/Login epic"}' | ergo new epic

# Atomic create-and-claim
printf '%s' '{"title":"Fix CVE","claim":"sonnet@agent-host"}' | ergo new task
# claim implies state=doing unless state explicitly set
```

### Body from stdin (--body-stdin)

When body is multi-line markdown. Stdin is literal text, metadata via flags.

```bash
# Create
printf '%s\n' '## Goal' '- Do the thing' | ergo new task --body-stdin --title "Do the thing"

# Update
printf '%s\n' '## Status' '- Waiting on review' | ergo set ABCDEF --body-stdin --state blocked
```

### Flags only (TTY)

```bash
ergo new task --title "Login" --body "Implement signup" --epic OFKSTE
ergo set ABCDEF --state done
```

## 3d. Plan a Feature (single JSON document)

Create epic + tasks + dependencies atomically. `after` references task titles (exact, case-sensitive).

```bash
printf '%s' '{"title":"Auth","tasks":[{"title":"Middleware"},{"title":"Login","after":["Middleware"]}]}' | ergo --json plan
```

Edge semantics: when `B.after=["A"]`, B depends on A.

## 4. Dependencies

```bash
ergo sequence TASK_A TASK_B              # A before B
ergo sequence TASK_A TASK_B TASK_C       # Chain: A then B then C
ergo sequence rm TASK_A TASK_B           # Remove: B no longer depends on A
ergo list --ready                        # Tasks with all deps done/canceled
```

## 5. Update & Failures

```bash
# Blocked (waiting on human/external)
printf '%s' '{"state":"blocked"}' | ergo set ABCDEF

# Error (requires claim)
printf '%s' '{"state":"error"}' | ergo set ABCDEF
# If unclaimed, pass --agent or set claim in JSON
```

## 6. Prune + Compact

`prune` logically deletes completed work. Default is dry-run.

Pruned: tasks in `done`/`canceled`, epics with no remaining children.
Preserved: tasks in `todo`, `doing`, `blocked`, `error`; epics with active children.

```bash
ergo prune                   # Preview
ergo prune --yes             # Apply
ergo compact                 # Rewrite log (physical deletion)
```

## 7. Attach Results

Results are pointers to project files (relative paths). Multiple calls accumulate, newest first.

```bash
printf '%s' '{"result_path":"docs/report.md","result_summary":"Final report"}' | ergo set GHIJKL
ergo show GHIJKL --json | jq '.results[]'
```

## 8. Epic Hierarchies

Epics complete when all children are done or canceled. They have no own state.
Epic-to-epic deps block all tasks in the dependent epic.

```bash
ergo sequence EPIC_A EPIC_B
```

## Reference: State Machine

States: `todo`, `doing`, `blocked`, `error`, `done`, `canceled`

Invariants:
- `doing`/`error` require claim; pass `--agent` or set claim in JSON
- `todo`/`done`/`canceled` clear claim

## Reference: Rules

- Dependencies: task-to-task or epic-to-epic only; cycles forbidden
- Agent ID format: `<model>@<hostname>` (recommended, not enforced)
