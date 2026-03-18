---
description: Review all changes before committing — checks conventions, types, tests, and quality
agent: manager
subtask: true
---

Review all current changes before committing. This is a quality gate.

Workflow:
1. Run `git diff` (and `git diff --staged` if there are staged files) to see all modifications
2. Read `CONVENTIONS.md` from the project root
3. For each modified file, verify:
   a. **Conventions compliance**: naming, imports, layer boundaries, Value Objects in domain, DTO patterns
   b. **Type safety**: no `any`, proper return types, null handling
   c. **Architecture**: no cross-context imports, correct dependency direction (infra → app → domain)
   d. **Completeness**: are there files that should have been modified but weren't? (e.g., module registration, exports)
   e. **Tests**: were tests added or updated for the changed behavior? Flag if missing
   f. **Imports**: no unused imports, no duplicate imports, no circular dependencies
4. Run `npx tsc --noEmit` to verify type safety
5. Run the project linter if available
6. Run tests related to the changed modules if identifiable

Present a structured review report:

```
## Review Summary

### ✅ Passing
- [what looks good]

### ⚠️ Warnings
- [non-blocking issues, suggestions]

### ❌ Issues
- [must-fix before commit]

### 📋 Missing
- [tests not written, modules not registered, etc.]

### Recommendation
- Ready to commit / Needs fixes
```

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Do NOT fix code yourself — only report findings
- Do NOT modify any files
- Be specific: include file paths and line references
- Compare against CONVENTIONS.md, not generic best practices
- If everything looks good, say so clearly and suggest `/commit`
