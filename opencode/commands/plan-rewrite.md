---
description: Rewrite or improve the current PLAN.md
agent: step-builder-agent
---

Review and improve the existing PLAN.md for the current project.

Workflow:
1. Read the current PLAN.md
2. Scan the codebase to validate whether the plan still matches reality
3. Identify ambiguity, missing dependencies, missing business rules, or poor step boundaries
4. Ask only the missing questions needed to fix those gaps
5. Rewrite PLAN.md so it is actionable for the execution-orchestrator and ts-expert-coder

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Do not implement code
- Preserve valid parts of the plan when possible
- Improve step ordering, acceptance criteria, and verification commands

When finished, tell the user the updated plan is ready and that they can use `/execute` to continue.
