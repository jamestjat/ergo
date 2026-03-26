---
name: ergo
description: >-
  Official skill for the `ergo` CLI tool — a local-first, concurrency-safe
  task/epic planner that stores plans in `.ergo/` JSONL logs. Use this skill
  whenever the user mentions ergo, .ergo, ergo plan, ergo claim, ergo set, ergo
  list, ergo sequence, or ergo prune. Also trigger when the user wants to break
  multi-commit work into a dependency-ordered task backlog with epics (3+
  commits, multiple concerns like API + UI + tests), scope implementation tasks
  into claimable units, or manage agent work queues — even without naming ergo
  explicitly. Do NOT trigger for other task trackers (Linear, Jira, Asana),
  strategic roadmap planning, or single-file bug fixes and refactors.
license: MIT
domain: project-management
role: specialist
scope: operations
output-format: commands
triggers:
  - ergo
  - task planning
  - epic
  - feature plan
  - claim task
  - ready work
  - work queue
  - sequence
  - dependencies
  - break down work
  - backlog
  - scope tasks
metadata:
  author: ergo
  version: 0.11.1
---

<!-- TOC: When to Use | Critical Rules | Planning | Quick Workflow | Essential Commands | Dependencies | Results | State Machine | Execution | Troubleshooting | References -->

# ergo -- Fast Task/Epic Planning for Agents

> **Append-only JSONL storage** in `.ergo/plans.jsonl`. Git-friendly, auditable, concurrency-safe.

## When to Use

**Use ergo when** work spans 3+ commits, touches multiple concerns (API + UI + tests + docs), or has ambiguity that needs resolving before implementation.

**Skip ergo when** the change is straightforward enough to just implement — a bug fix, a single-concern feature, routine refactoring. Don't plan what you can just do.

**Unsure?** Ask: "Want an ergo plan first, or should I just implement?"

## Critical Rules for Agents

| Rule | Why |
|------|-----|
| **ALWAYS use `--json`** | Structured output; agents must never parse human-mode text |
| **ALWAYS pass `--agent`** | Required for claims; format: `<model>@<hostname>` |
| **Never leave tasks `doing`** | Always resolve to `done`, `blocked`, `error`, or `canceled` |
| **`doing`/`error` require claim** | Shows who's working / who failed |
| **`todo`/`done`/`canceled` clear claim** | Ownership only while active |
| **No cross-kind deps** | task-to-task or epic-to-epic only; no task-to-epic |
| **No cycles** | Dependency cycles are rejected |
| **Use `printf '%s'`** | Reliable JSON delivery; heredocs can be flaky in some tooling |
| **Don't commit `.ergo/` per-task** | Commit `.ergo/` once when epic completes |

## Planning Methodology

Planning surfaces unknowns and decisions. **Resolve them now by asking the user.** Present options with tradeoffs, get an answer, write the decision into the epic body or task acceptance criteria. Don't write "TBD" or "Consult Me" — that creates a mid-implementation block for a future agent with less context than you have now.

The only decisions that belong as deferred checkpoints are ones that literally require an implementation artifact to evaluate (e.g., "produce a UI mockup, then get approval").

### Epics

One per coherent feature area. Body includes scope, non-goals, constraints, and key decisions. Tasks that don't fit a larger area can be left ungrouped.

### Tasks

Each task should be:
- **One atomic, reviewable change** — completable in a single session
- **Ideally auto-verifiable** via acceptance criteria and runnable gates
- **Split on real boundaries** only (API surface, data model, tests, docs)
- **Friendly to smaller models** — the implementing agent might have less context than you

**Spikes** produce knowledge, not code. Prefix with `spike:`. Dependent tasks should note what they're waiting to learn.

Task body template (trim to fit — omit empty sections):

```md
## Goal
- <1–3 bullets: concrete outcome>

## Acceptance Criteria
- <Observable behavior, edge cases, definition of done>

## Validation Gates
- <Exact commands to prove it works — tests, lint, format>
```

### Dependencies

Add edges only for true ordering constraints — maximize parallelism.

### Plan Review Checklist

Before presenting to the user:
- **Coverage** — API, tests, docs, migrations, edge cases?
- **Sizing** — anything too small (fold in) or too large (split)?
- **Dependencies** — missing edges causing churn? unnecessary edges blocking parallelism?
- **Validation** — every task has runnable gates?
- **Risk** — 1–3 highest-risk tasks identified; spikes added?
- **Open calls** — every judgment call resolved, not deferred?

Present an executive summary to the user for approval before implementation.

## Quick Workflow

```bash
AGENT="sonnet@agent-host"

# 1. Claim oldest ready task
ergo --json claim --agent "$AGENT"

# 2. Do the work described in the task body

# 3. Attach results and mark done
printf '%s' '{"state":"done","result_path":"src/auth.go","result_summary":"Auth module"}' | ergo set ABCDEF

# 4. Repeat -- claim next ready task
ergo --json claim --agent "$AGENT"
```

## Essential Commands

### Initialize

```bash
ergo init                    # Create .ergo/ in current directory
ergo where                   # Print active .ergo/ path
```

### Create Work

```bash
# Epic (grouping node, no state)
printf '%s' '{"title":"Auth","body":"Signup and login"}' | ergo new epic

# Task (unit of work with state)
printf '%s' '{"title":"Hash passwords","body":"Use bcrypt","epic":"OFKSTE"}' | ergo new task

# Atomic create-and-claim (claim implies state=doing)
printf '%s' '{"title":"Fix CVE","claim":"sonnet@agent-host"}' | ergo new task

# Multi-line body via --body-stdin (stdin is literal text, not JSON)
printf '%s\n' '## Goal' '- Implement login' | ergo new task --body-stdin --title "Login flow" --epic OFKSTE

# Flags-only (when stdin is TTY)
ergo new task --title "Login" --body "Implement signup" --epic OFKSTE
```

### Plan a Feature (atomic)

Create epic + tasks + dependencies in one call. `after` references task titles (exact, case-sensitive).

```bash
printf '%s' '{"title":"Auth","body":"User auth system","tasks":[{"title":"Middleware","body":"JWT validation"},{"title":"Login","body":"POST /login","after":["Middleware"]},{"title":"Signup","body":"POST /signup","after":["Middleware"]}]}' | ergo --json plan
```

### Claim Work

```bash
# Oldest ready task (deps satisfied, state=todo)
ergo --json claim --agent sonnet@agent-host

# Specific task by ID
ergo --json claim ABCDEF --agent sonnet@agent-host

# Oldest ready within an epic
ergo --json claim --agent sonnet@agent-host --epic OFKSTE
```

If no tasks are ready: prints message, exits 0 (not an error).

### Update State

```bash
# Mark done
printf '%s' '{"state":"done"}' | ergo set ABCDEF

# Mark done with results
printf '%s' '{"state":"done","result_path":"docs/spec.md","result_summary":"Spec v1"}' | ergo set ABCDEF

# Mark blocked
printf '%s' '{"state":"blocked"}' | ergo set ABCDEF

# Mark error (requires claim -- pass --agent if unclaimed)
printf '%s' '{"state":"error"}' | ergo set ABCDEF

# Cancel
printf '%s' '{"state":"canceled"}' | ergo set ABCDEF

# Reopen (done/canceled -> todo)
printf '%s' '{"state":"todo"}' | ergo set ABCDEF

# Update body via --body-stdin
printf '%s\n' '## Status' '- Blocked on review' | ergo set ABCDEF --body-stdin --state blocked

# Flags-only (TTY)
ergo set ABCDEF --state done
```

### View Work

```bash
ergo --json list                       # All active work
ergo --json list --ready               # Only ready tasks (deps satisfied)
ergo --json list --epic OFKSTE         # Tasks within epic (includes done for context)
ergo --json list --epics               # Only epics
ergo --json list --all                 # Include done and canceled

ergo --json show ABCDEF                # Task/epic detail (epics include children with bodies)
ergo show ABCDEF                       # Human-readable Markdown
```

### Dependencies

```bash
ergo sequence TASK_A TASK_B            # A before B
ergo sequence TASK_A TASK_B TASK_C     # A then B then C (chain)
ergo sequence rm TASK_A TASK_B         # Remove: B no longer depends on A
```

### Results

Results are pointers to project files. Multiple calls accumulate (newest first). Only attach when the task produced a concrete deliverable — don't create standalone files just to have a link.

```bash
printf '%s' '{"result_path":"docs/report.md","result_summary":"Final report"}' | ergo set GHIJKL
```

Both `result_path` and `result_summary` are required together. Summary: single-line, max 120 chars.

### Maintenance

```bash
ergo prune                   # Dry-run: preview what would be removed
ergo prune --yes             # Apply: remove done/canceled tasks + empty epics
ergo compact                 # Collapse log to current state (physical deletion)
```

## Executing Ergo Plans

1. `ergo --json claim --agent <identity>` — claim a ready task
2. Implement it. Stop and consult the user if important questions arise.
3. Commit using repo conventions. **Do not** include `.ergo/` files in per-task commits.
4. Mark task done:
   - **Completion note** — update body with brief note on what was done (decisions, approach, anything non-obvious)
   - **Result link** — attach with `result_path` if the task produced a concrete deliverable
   - **After a spike** — update dependent tasks with what was learned
   - If a task can't be completed, mark `blocked` or `error` — never leave `doing`
5. **If the plan needs to change** — update it and note why. Plans are living documents.
6. When epic is done, commit `.ergo/` state: `plan: complete <epic name>`

## State Machine

```
States: todo | doing | done | blocked | canceled | error

todo ──→ doing ──→ done
 │        │ │       │
 │        │ │       └──→ todo (reopen)
 │        │ └──→ blocked
 │        └──→ error ──→ doing (retry) / todo (reassign) / canceled
 └──→ canceled
```

| State | Claim required? |
|-------|----------------|
| `todo` | No claim (cleared) |
| `doing` | **Required** |
| `done` | No claim (cleared) |
| `blocked` | Optional |
| `error` | **Required** (shows who failed) |
| `canceled` | No claim (cleared) |

## Global Flags

```
--dir <path>       Discovery start (or .ergo path directly)
--agent <identity> Agent ID for claims (e.g. sonnet@agent-host)
--json             JSON output (required for agents)
--quiet, -q        Suppress non-essential output
--verbose, -v      Verbose debug output
-h, --help         Show help
-V, --version      Print version
```

## JSON Input Fields (all strings)

Unknown keys are rejected with "did you mean?" suggestions.

| Field | new | set | Notes |
|-------|-----|-----|-------|
| `title` | required | optional | |
| `body` | optional | optional | |
| `epic` | optional | optional | Epic ID; `""` to unassign (JSON only) |
| `state` | optional | optional | todo/doing/done/blocked/canceled/error |
| `claim` | optional | optional | Agent ID; `""` to unclaim (JSON only) |
| `result_path` | - | optional | File path (requires result_summary) |
| `result_summary` | - | optional | One-liner (requires result_path) |

Epics only support: `title`, `body`.

## Anti-Patterns

- Parsing human-mode output instead of using `--json`
- Leaving tasks in `doing` after finishing work
- Forgetting `--agent` when claiming
- Creating task-to-epic dependencies (forbidden)
- Using heredocs for JSON input (fragile; use `printf '%s'`)
- Assuming epics have state (they don't -- they're structural)
- Writing "TBD"/"Consult Me" in task bodies instead of resolving during planning
- Committing `.ergo/` in every per-task commit

## Troubleshooting

```bash
ergo where                   # Verify .ergo/ location
ergo --json list             # Check current state
ergo quickstart              # Full reference manual
```

**"Lock busy"**: Another ergo process holds the lock. Retry after a moment.

## References

| Topic | File |
|-------|------|
| CLI spec (contracts) | [references/SPEC.md](references/SPEC.md) |
| Full quickstart manual | [references/QUICKSTART.md](references/QUICKSTART.md) |
