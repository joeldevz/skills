---
description: Show current progress from PLAN.md
agent: manager
subtask: true
---

Read PLAN.md and report the current implementation status.

Return:
1. Completed steps
2. Current step in progress or under review
3. Remaining pending steps
4. Next recommended action for the human

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Do not implement code
- Do not modify application files unless you need to update plan state for consistency
- Keep the status report concise and actionable
