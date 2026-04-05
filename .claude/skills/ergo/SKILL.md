---
name: ergo
description: >-
  ALWAYS use this skill when the user mentions ergo, .ergo/, or any ergo command (plan, claim, set, new, list, show, sequence, prune, compact, init, where). ALWAYS use when a `.ergo/` directory exists and the user asks what to work on next, what is ready/blocked, or status of tasks/epics. ALWAYS use to: break a feature into tasks with dependencies (3+ commits or API + UI + tests + docs); claim or mark a task done/blocked/error by ID; coordinate parallel work across agent teammates; or wrap up a session ("land the plane", "finish up and push", commit .ergo/). The `ergo` CLI is a local task/epic planner for multi-agent coordination, storing plans in `.ergo/`. Do NOT trigger for Linear, Jira, Asana, GitHub Projects, Taskfile/Make, strategic roadmaps, personal todo lists, single-file edits, or Claude Code native TaskCreate/TodoWrite.
---

# Ergo Feature Planning

Turn a feature request into a repo-local plan tracked by `ergo`: a backlog of well-scoped tasks, with dependency edges, grouped into epics.

## When to use

**Use ergo when** the work would span 3+ commits, touches multiple concerns (API + UI + tests + docs), or has ambiguity that needs resolving before implementation.

**Skip ergo when** the change is straightforward enough to just implement — a bug fix, a single-concern feature, routine refactoring. Don't plan what you can just do.

**Unsure?** Ask: "Want an ergo plan first, or should I just implement?"

## Bootstrap

1. Expect `ergo` to be globally installed. If missing, install a precompiled binary from the [releases page](https://github.com/jamestjat/ergo/releases) — pick the archive matching the host OS/arch (`darwin_arm64`, `darwin_amd64`, `linux_arm64`, `linux_amd64`, `windows_arm64`, `windows_amd64`), extract, and move `ergo` onto `PATH`. On macOS/Linux:
   ```bash
   # replace VERSION, OS, ARCH
   curl -sSL https://github.com/jamestjat/ergo/releases/download/vVERSION/ergo_VERSION_OS_ARCH.tar.gz \
     | tar -xz -C /usr/local/bin ergo
   ```
2. Run `ergo --help` and `ergo quickstart` to learn the CLI before creating plans.

## Planning

Planning naturally surfaces unknowns, ambiguities, and decisions. **Resolve them now, during planning, by asking the user.** Present options clearly with tradeoffs, get an answer, and write the decision into the epic body or task AC. Do not write "Consult Me" or "TBD" into task bodies as a way to defer a conversation you could have right now — that just creates a mid-implementation block for a future agent that has less context than you do.

The test: if you can describe the options and tradeoffs to the user today, ask today. The only decisions that belong in task bodies as checkpoints are ones that literally require an implementation artifact to evaluate (e.g., "produce a UI mockup, then get approval before proceeding"). For those, write a **Checkpoint** section with the specific artifact to produce and the specific question to answer — not a vague "consult me."

If a task's shape depends on an unresolved decision and the user can't or won't decide yet, make it a spike instead.

Critique continuously as you build the plan — when writing one task reveals that earlier tasks should be merged or split, fix it immediately rather than deferring to a review pass.

### Epics

One per coherent feature area. Body includes scope, non-goals, constraints, and if it makes sense, key decisions and assumptions. Tasks are grouped into epics, but dependencies can cross epic boundaries when needed.

Tasks that don't fit a larger feature area can be left ungrouped. Use judgment.

### Tasks

The unit of execution. Each task should be:
- **One atomic, reviewable change** — completable in a single session. Not trivial, not sprawling.
- **Ideally, automatically verifiable** via acceptance criteria and runnable gates. When human verification is truly needed, include exact instructions so implementing agents know precisely how to verify.
- **Split on real boundaries** only (API surface, data model, migration, tests, docs).
- **Friendly to smaller models** — the implementing model might not be as smart as you. Before finalizing a task, consider it from the perspective of a less capable model without the context of the full user conversation or the whole backlog. Will they succeed?

**Spikes** produce knowledge, not code. Prefix with `spike:`. Dependent tasks should note what they're waiting to learn from the spike.

Task body template (trim to fit — omit empty sections):

```md
## Goal
- <1–3 bullets: concrete outcome and why it matters>

## Context
- <Background, links to docs/designs — only when not obvious from the goal>

## Acceptance Criteria
- <Observable behavior, edge cases, definition of done>

## Checkpoint (rare — only when a decision requires an implementation artifact)
- Produce: <specific artifact, e.g., "ASCII mockup of the banner layout">
- Then ask: <specific question, e.g., "Does this information hierarchy work?">
- Do not proceed past this point without user approval.

## Validation Gates
- <Exact commands to prove it works — tests, lint, format>
```

### Dependencies

Add edges only for true ordering constraints — maximize parallelism and task independence.

## Plan review

Before presenting to the user, do a final confirmation pass:

- **Coverage** — API, tests, docs, migrations, edge cases all accounted for?
- **Sizing** — anything too small (fold in) or too large (split)?
- **Dependencies** — missing edges that'll cause churn? unnecessary edges blocking parallelism?
- **Validation** — every task has runnable gates?
- **Risk** — 1–3 highest-risk tasks identified; spikes or mitigation added?
- **Open calls** — every judgment call resolved with the user, not deferred to task bodies? If a task still says "Consult Me" or "TBD," that's a planning failure — go ask now.
- **Cruft** — will the planned work leave behind unowned debt? Consider cleanup tasks.

Fix what you find, then **present an executive summary to the user for approval**. Concise overview of epics, key tasks, and major decisions, expressed in high-level language, minimizing jargon. Get buy-in before investing time in implementation.

## Executing ergo plans

1. Claim a ready task.
2. Implement it. Stop and consult the user if important questions arise.
3. Commit using repo conventions. **Do not** include `.ergo/` files in per-task commits.
4. Mark task done. On completion:
   - **Completion note** — update the task body with a brief note on what was done: decisions made, approach taken, anything non-obvious. Think PR description, not essay.
   - **Result link** — if the task produced a concrete deliverable (a new file, component, doc), attach it with `result_path`. **Do not** create standalone result files just to have a link.
   - **After completing a spike** — update dependent tasks with what was learned so the knowledge flows forward.
   - If a task can't be completed, mark it blocked or error — never leave tasks in progress.
5. **If the plan needs to change** — a task is unnecessary, scope shifted, new work emerged — update the plan and note why. Plans are living documents, not contracts.
6. When an epic is done, commit the `.ergo/` state with a message like `plan: complete <epic name>`. Otherwise, go to 1.

## Agent rules

| Rule | Why |
|------|-----|
| **Always use `--json`** | Structured output; never parse human-mode text. |
| **Always pass `--agent`** | Required for claims. Format: `<model>@<hostname>`. |
| **Never leave tasks `doing`** | Resolve to `done`, `blocked`, `error`, or `canceled`. |
| **`doing`/`error` require a claim** | Shows who's working or who failed. |
| **`todo`/`done`/`canceled` clear the claim** | Ownership only while active. |
| **Prefer flags over JSON piping** | Flags work on every OS; piping requires shell-aware quoting. |
| **Don't commit `.ergo/` per task** | Commit `.ergo/` once, when an epic completes. |

## Multi-agent coordination

Ergo is the single source of truth for task state across agent teammates. Do not duplicate in Claude Code's TaskCreate — dual bookkeeping drifts.

- **Race-safe claiming** — multiple teammates can call `ergo claim` simultaneously; exactly one wins per task, losers get the next ready task. Dependencies are enforced automatically, so tasks with unmet deps won't appear in claim results.
- **Unique agent identity** — each teammate passes a distinct `--agent` value (e.g., `sonnet-1@claude-code`, `sonnet-2@claude-code`) so claims are attributable.
- **Lead role** — creates the plan, spawns teammates with the claim/implement/mark-done loop, monitors via `ergo --json list`, redirects off-track teammates.

## Landing the plane

When the user says **"land the plane"** (or "wrap up", "finish the session"), the session is not done until `git push` succeeds.

1. **File remaining work** — create ergo tasks for any follow-up discovered this session.
2. **Run quality gates** — tests, lint, build. If gates fail, file tasks capturing the failures.
3. **Update ergo state** — mark completed tasks `done` (with results), move blocked/error tasks, and verify nothing remains in `doing`.
4. **Commit and push — non-negotiable** — `git add .ergo/`, commit with `plan: land session — <summary>`, `git pull --rebase`, `git push`, confirm `git status` shows up-to-date with origin. Never say "ready to push when you are" — push it. Unpushed work breaks multi-agent coordination.
5. **Clean up** — `ergo prune --yes` if the plan is large.
6. **Hand off** — summary of what shipped, tasks filed, gate status, and the recommended next task from `ergo --json list --ready`.
