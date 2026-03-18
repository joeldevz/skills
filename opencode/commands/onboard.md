---
description: Onboard to a project - scan codebase and learn conventions before working
agent: step-builder-agent
subtask: true
---

Onboard to this project before doing any work.

Workflow:
1. Read `CONVENTIONS.md` if it exists in the project root
2. Read `AGENTS.md` if it exists
3. Read `package.json` to understand the stack, scripts, and dependencies
4. Read `tsconfig.json` for compiler settings and path aliases
5. Scan `src/` structure to understand the architecture (modules, contexts, layers)
6. Check for existing tests to understand the testing patterns
7. Check for linter/formatter config (`.eslintrc`, `.prettierrc`, `eslint.config.mjs`)
8. Check for `.env.example` or config files to understand environment setup

Return a concise summary with:
- Stack detected (runtime, framework, ORM, test framework, etc.)
- Architecture pattern (DDD, MVC, modular, etc.)
- Key conventions found
- Folder structure overview
- Build/test/lint commands
- Any important gotchas or non-obvious patterns

Context:
- Working directory: {workdir}
- Current project: {project}

Important:
- Do not implement code
- Do not modify any files
- This is read-only exploration
- Keep the summary actionable and concise
