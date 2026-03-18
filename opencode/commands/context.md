---
description: Save project discoveries and decisions to Engram persistent memory
agent: planner
subtask: true
---

Save important project discoveries and decisions to Engram persistent memory for future sessions.

If an argument is provided, save that specific observation:
"{argument}"

If no argument is provided, scan recent work and extract learnings:

Workflow:
1. Read PLAN.md if it exists — extract architectural decisions made during planning
2. Run `git log --oneline -10` to see recent work
3. Read CONVENTIONS.md if it exists — note any conventions that were tricky or non-obvious
4. Identify the most valuable things to persist:
   - Architectural decisions (why X instead of Y)
   - Non-obvious patterns discovered in the codebase
   - Gotchas or bugs found and fixed
   - Dependencies or configurations that were important
   - User preferences or constraints expressed during planning
5. For each discovery, save to Engram with:
   - A searchable title
   - Structured content (What/Why/Where/Learned)
   - Appropriate type (decision, architecture, discovery, pattern, bugfix, config)
   - The correct project name

Present a summary of what was saved:
```
## Saved to Memory

1. [title] — [type] — [brief description]
2. [title] — [type] — [brief description]
...

These observations will be available in future sessions via `mem_search`.
```

Context:
- Working directory: {workdir}
- Current project: {project}
- Observation: {argument}

Important:
- Do NOT save trivial information (file renames, minor edits)
- DO save decisions, gotchas, architecture patterns, and user preferences
- Use `mem_search` first to avoid duplicating existing observations
- Use `topic_key` for evolving topics so they update instead of duplicating
