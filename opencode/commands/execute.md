---
description: Execute the next pending step from PLAN.md
agent: manager
---

Execute the next pending step from PLAN.md.

Workflow:
1. Read PLAN.md from the project root
2. Identify the next pending step
3. Delegate implementation of that single step to `coder`
4. Present the modified files, what changed, and verification results
5. Stop and request human review before any further step

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Do not skip the human review loop
- Do not implement code yourself
- Update PLAN.md step status based on progress and approval state

When finished, clearly tell the user whether the step is awaiting approval or additional fixes.
