---
description: Generate or run tests for the current step or specified module
agent: coder
---

Generate or run tests for: "{argument}"

If no argument is provided, generate tests for the files changed in the current PLAN.md step.

Workflow:
1. Determine what needs testing:
   - If argument is a file/module path → test that specific module
   - If argument is "run" → just run existing tests, don't generate new ones
   - If no argument → read PLAN.md, find the current `[~]` or last `[x]` step, test those files
2. Read existing test files nearby to understand the testing patterns:
   - Test framework (Jest, Vitest, etc.)
   - File naming convention (`*.spec.ts`, `*.test.ts`)
   - Mocking patterns (jest.mock, custom factories, test fixtures)
   - Assertion style
3. Read `CONVENTIONS.md` for testing conventions if it exists
4. For each file that needs tests:
   a. Read the source file completely
   b. Identify public API, edge cases, error paths, and business rules
   c. Generate tests following the existing patterns
   d. Place the test file following the project convention (co-located or `__tests__/`)
5. Run the tests: `npm test -- --testPathPattern=<pattern>` or equivalent
6. If tests fail, fix them (up to 3 attempts)

Test quality rules:
- Test behavior, not implementation details
- Cover happy path, edge cases, and error cases
- Use descriptive test names: `should <expected behavior> when <condition>`
- Mock external dependencies (DB, HTTP, queues) — never test infra in unit tests
- For command/query handlers: test the handler with mocked repository
- For entities: test business rules and invariants
- For controllers: test only routing and DTO validation if applicable

Context:
- Working directory: {workdir}
- Current project: {project}
- Target: {argument}

Important:
- Follow existing test patterns exactly — do not introduce new test utilities
- If the project uses test factories or fixtures, reuse them
- Run verification after generating tests
- Report: tests generated, tests run, pass/fail results
