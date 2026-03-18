---
description: Show an annotated diff of changes from the current step
agent: execution-orchestrator
subtask: true
---

Show the changes made in the current PLAN.md step with context and annotations.

Workflow:
1. Read PLAN.md and identify the current step (`[~]` in progress or last `[x]` done)
2. Run `git diff` to get all unstaged changes
3. Run `git diff --staged` to get staged changes
4. For each changed file:
   a. Show the file path and what step it belongs to
   b. Show the diff with surrounding context
   c. Add a brief annotation explaining what changed and why
5. Summarize: files added, files modified, files deleted, total lines changed

Present the output as:

```
## Changes for Step N: <title>

### <file path> (modified/added/deleted)
> <brief explanation of the change>

<relevant diff hunks>

---

### Summary
- Files added: N
- Files modified: N
- Files deleted: N
- Lines: +N / -N
```

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Do NOT modify any files
- Group changes by their purpose, not just alphabetically
- If changes span multiple steps, flag that clearly
- Keep annotations concise — one sentence per file
