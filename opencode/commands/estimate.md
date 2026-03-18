---
description: Estimate implementation effort for the current PLAN.md
agent: planner
subtask: true
---

Estimate the implementation effort for the current PLAN.md.

Workflow:
1. Read PLAN.md completely
2. For each step, estimate effort based on:
   - **Complexity**: How many files, how much logic, how many decisions
   - **Risk**: Is this touching existing code? New patterns? External dependencies?
   - **Dependencies**: Does it need other steps first? Does it need manual work (migrations, config)?
3. Assign each step a T-shirt size:
   - **XS**: ~5 min — trivial change, single file, no decisions
   - **S**: ~15 min — straightforward, 1-3 files, clear pattern to follow
   - **M**: ~30 min — moderate logic, 3-5 files, some decisions needed
   - **L**: ~1 hour — complex logic, multiple files, new patterns or integrations
   - **XL**: ~2+ hours — significant complexity, many files, high risk or unknowns

Present the estimate as:

```
## Estimation: <Plan Title>

| Step | Title | Size | Est. | Risk |
|------|-------|------|------|------|
| 1 | ... | S | ~15m | Low |
| 2 | ... | M | ~30m | Medium |
| ... | ... | ... | ... | ... |

### Total estimate: ~X hours Y minutes

### Risk factors
- [anything that could blow up the estimate]

### Recommendations
- [which steps to tackle first]
- [which steps can be parallelized]
- [where to spend extra review time]
```

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Be realistic, not optimistic — include time for verification and iteration
- These estimates assume the coder agent, not a human typing manually
- Flag steps that are likely to need multiple iterations
- If some steps are already done, exclude them from the total
