# ergo Public CLI Spec (Contracts)

Stable contracts for ergo's CLI behavior. Read this when you need precise guarantees about output format, exit codes, or validation rules.

## Output Contracts

### General

- Success: exit code `0`
- Failure: non-zero exit, informative error on stderr
- `--json`: exactly one JSON value on stdout (object or array), no non-JSON noise

### `--json` shapes

| Command | Shape |
|---------|-------|
| `list --json` | Array of items |
| `show --json` | Object (epics include `children` with `body`) |
| `new`, `set`, `plan`, `sequence`, `claim`, `prune`, `compact` | JSON object |

### `claim` with no ready tasks

- Exit code `0` (not an error)
- Human: prints "No ready ergo tasks."
- `--json`: JSON object explicitly indicating "no ready"

### `claim` success with `--json`

- Includes `reminder`: `"When you have completed this claimed task, you MUST mark it done."`

## Input Mode Contracts

For `new epic`, `new task`, `set`:
- Default: single JSON object from stdin (entire stdin; trailing newlines OK)
- TTY (not piped): flags-only mode (`--title`, `--body`, `--state`, etc.)
- `--body-stdin`: stdin is literal body text (non-empty), not JSON. Other fields via flags.
  - `--body` and `--body-stdin` are mutually exclusive
  - `new * --body-stdin` requires `--title`

For `plan`:
- Single JSON object on stdin
- No `--body-stdin` or flags-only support
- Parse failures: `parse_error`
- Semantic validation failures: `validation_failed`

## State Machine

States: `todo | doing | done | blocked | canceled | error`

### Allowed Transitions

| From | To |
|------|----|
| `todo` | `doing`, `blocked`, `canceled` |
| `doing` | `done`, `blocked`, `error`, `canceled` |
| `blocked` | `todo`, `doing`, `canceled` |
| `error` | `doing` (retry), `todo` (reassign), `canceled` |
| `done` | `todo` (reopen) |
| `canceled` | `todo` (reopen) |

### Claim Invariants

- `doing`/`error` require claim
- `todo`/`done`/`canceled` clear claim
- `blocked` may have claim or not
- `claim` in JSON implies `state=doing` unless `state` explicitly set

## Dependency Rules

- Task-to-task: allowed
- Epic-to-epic: allowed
- Task-to-epic or epic-to-task: **forbidden**
- Self-dependencies: **forbidden**
- Cycles: **rejected** at creation time

## Prune & Compact

Two-phase deletion:

1. `prune` = logical deletion (tombstones)
   - Eligible: tasks in `done` or `canceled`
   - Preserved: tasks in `todo`, `doing`, `blocked`, `error`
   - Epics with no remaining children also pruned
2. `compact` = physical deletion (rewrites log)

Pruned IDs:
- Don't appear in `list`
- Can't be used as dependency endpoints
- Can't be claimed or updated
- Dependencies to/from pruned IDs are dropped (won't block other work)

## Concurrency

- Mutations serialized by advisory lock on `.ergo/lock`
- Lock acquisition is fail-fast (non-blocking)
- "Lock busy" = exit quickly, caller retries
- `claim` oldest-ready is race-safe: exactly one process wins

## IDs

- 6-character uppercase identifiers
- Entities: Task (has state, claimable) and Epic (structural, no state)

## Validation

- Unknown JSON fields rejected with "did you mean?" suggestions
- `result_summary`: single-line, max 120 characters
- `result_path` and `result_summary` are required together
