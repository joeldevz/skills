---
description: Apply human feedback to the current step
agent: execution-orchestrator
---

Apply the following human feedback to the current step in PLAN.md:

"{argument}"

Workflow:
1. Read PLAN.md and locate the current step under review
2. Interpret the feedback and transform it into precise implementation instructions
3. Delegate the requested fixes to `ts-expert-coder`
4. Present the updated files, changes made, and verification results
5. Stop and ask the human to review again

Context:
- Working directory: {workdir}
- Current project: {project}
- Human feedback: {argument}

Important:
- Do not move to the next step yet
- Do not implement code yourself
- Mark the step as needing fixes before delegation when appropriate

When finished, explicitly ask the user to review the revised changes.
