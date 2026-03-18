---
description: Create a project plan from a task request
agent: step-builder-agent
---

Plan the task "{argument}" for the current project.

Workflow:
1. Scan the codebase first to gather technical context
2. Ask the user the minimum necessary business and technical questions in thematic blocks
3. Confirm your understanding before writing the final plan
4. Generate or replace PLAN.md in the project root

Context:
- Working directory: {workdir}
- Current project: {project}
- Requested task: {argument}

Important:
- Do not implement code
- Do not stop at a draft outline; produce a full PLAN.md
- Make each step small enough to be reviewed independently
- Include acceptance criteria and verification steps

When finished, tell the user the plan is ready and that they can use `/execute` to start implementation.
