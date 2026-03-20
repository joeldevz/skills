# Skills Team Workflow

This repository installs a portable planning-and-execution workflow for Claude Code.

Available agents:
- `planner` for discovery, requirements, and `PLAN.md`
- `manager` for step selection, review-loop orchestration, and coder handoff generation
- `coder` for bounded implementation, tests, and verification

Working rules:
- Keep `PLAN.md` as the visible source of truth
- Execute one approved step at a time
- Require human review after every implementation pass
- Delegate bounded code changes to `coder`
- Prefer existing project conventions over inventing new patterns
- Use Neurox as the durable memory system via `neurox_session_start`, `neurox_context`, `neurox_recall`, and `neurox_save`

Claude Code constraint:
- Claude subagents cannot spawn other subagents
- Because of that, the main conversation is the top-level orchestrator
- Use `planner` and `coder` as subagents when helpful
- Use `manager` as a scoping/review helper, but keep any multi-agent orchestration in the main conversation

Installed slash skills include `/onboard`, `/plan`, `/plan-rewrite`, `/estimate`, `/execute`, `/apply-feedback`, `/diff`, `/status`, `/rollback`, `/test`, `/review`, `/docs`, `/context`, `/commit`, and `/pr`.

Installer note:
- `scripts/setup.sh --claude` also writes a user-level Neurox MCP entry to `~/.claude.json` so Claude Code can call the Neurox tools directly.

For planning-heavy work, start with `/onboard` or `/plan`.
