# Plan: Versioned `clasing-skill` CLI installer

## Goal
Add a new centralized CLI installer named `clasing-skill` that can install the `skills` repo and other supported repos/packages such as `neurox` through one command for Claude Code, OpenCode, or both. The MVP must provide interactive and non-interactive flows, whole-package version selection and tracking, durable `skills.config.json` / `skills.lock.json` state, and blocking preflight validation before any install side effects.

## Business Context
- Primary users need one predictable installer instead of repo-specific scripts.
- The CLI must support both interactive selection and automatic mode from flags/defaults.
- Version tracking is at the package/repo level, not per-skill.
- `skills.config.json` stores user intent/defaults; `skills.lock.json` stores the resolved installed state.
- Preflight failures must stop the run before any writes.
- Neurox is mandatory where the selected package/target depends on it.
- Rollback/uninstall are explicitly out of scope for MVP.

## Technical Context
- The current installer already exists in `scripts/setup.sh` and contains reusable target-specific logic: `install_opencode()`, `install_claude()`, `require_neurox()`, backup creation, and Claude MCP configuration.
- Claude asset rendering already lives in `scripts/install_claude_assets.py`; it should remain the renderer for the `skills` package instead of being rewritten.
- There is no root `package.json`, `pyproject.toml`, or `go.mod`; the repo already depends on `python3` and shell scripts for installation, so the safest MVP CLI is a Python standard-library application with a thin executable wrapper.
- `opencode/package.json` only exists for the OpenCode config bundle, so the new installer should not depend on Node tooling for its own runtime.
- No materialized `.atl/skill-registry.md` exists; the current shared architecture is repo assets + thin adapters. The plan should preserve that boundary.

Planned command/file flow:

1. User runs `./clasing-skill`.
2. `clasing-skill` executes `python3 -m scripts.clasing_skill`.
3. CLI loads `scripts/clasing_skill/catalog.json`, `skills.config.json`, and `skills.lock.json`.
4. CLI resolves packages, targets, and versions from flags or interactive prompts.
5. CLI resolves the exact repo ref/commit for each selected package.
6. CLI runs global + package + target preflight checks.
7. For package `skills`, CLI clones/checks out the selected ref into a temp worktree and runs that checkout's `scripts/setup.sh --claude/--opencode`.
8. For package `neurox`, CLI clones/checks out the selected ref, builds the binary, installs it into a user-writable bin directory, and verifies `neurox status`.
9. After successful install, CLI writes requested intent to `skills.config.json` and resolved install results to `skills.lock.json`.

## Implementation Steps

### Step 1: Scaffold the centralized CLI and package catalog
- **What**: Add the new `clasing-skill` entrypoint, Python package structure, and explicit supported-package catalog.
- **Why**: The CLI needs one stable command surface and one centralized registry of supported packages/adapters before version resolution or install logic can be added.
- **Where**: `clasing-skill`, `scripts/__init__.py`, `scripts/clasing_skill/__init__.py`, `scripts/clasing_skill/__main__.py`, `scripts/clasing_skill/cli.py`, `scripts/clasing_skill/models.py`, `scripts/clasing_skill/catalog.py`, `scripts/clasing_skill/catalog.json`
- **How**:
  1. Create a repo-root executable file `clasing-skill`:
     ```bash
     #!/usr/bin/env bash
     set -euo pipefail
     REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
     exec python3 -m scripts.clasing_skill "$@"
     ```
  2. Make `scripts/` importable by adding `scripts/__init__.py` and expose the package via `scripts/clasing_skill/__main__.py`:
     ```python
     from .cli import main

     if __name__ == "__main__":
         raise SystemExit(main())
     ```
  3. In `scripts/clasing_skill/models.py`, define the core dataclasses used everywhere else:
     ```python
     @dataclass(slots=True)
     class PackageDefinition:
         id: str
         display_name: str
         repo_url: str
         adapter: str
         supported_targets: tuple[str, ...]
         default_version: str
         requires_neurox: bool
         install_strategy: str

     @dataclass(slots=True)
     class InstallRequest:
         packages: list[str]
         targets: list[str]
         versions: dict[str, str]
         interactive: bool
         state_dir: Path
     ```
  4. In `scripts/clasing_skill/catalog.json`, declare the initial package catalog as data, not hard-coded branches. Use this exact MVP shape:
     ```json
     {
       "version": 1,
       "packages": {
         "skills": {
           "displayName": "Skills",
           "repoUrl": "https://github.com/joeldevz/skills.git",
           "adapter": "skills_repo",
           "supportedTargets": ["claude", "opencode"],
           "defaultVersion": "latest",
           "requiresNeurox": true,
           "installStrategy": "git_checkout_setup_script"
         },
         "neurox": {
           "displayName": "Neurox",
           "repoUrl": "https://github.com/joeldevz/neurox.git",
           "adapter": "neurox_binary",
           "supportedTargets": ["claude", "opencode"],
           "defaultVersion": "latest",
           "requiresNeurox": false,
           "installStrategy": "git_checkout_go_build"
         }
       }
     }
     ```
  5. In `scripts/clasing_skill/catalog.py`, implement `def load_catalog(path: Path) -> dict[str, PackageDefinition]:` that validates required keys and raises `ValueError` on malformed catalog data.
  6. In `scripts/clasing_skill/cli.py`, implement `def build_parser() -> argparse.ArgumentParser` with these flags only for MVP: `--package`, `--target`, `--version`, `--non-interactive`, `--yes`, `--state-dir`, `--list-packages`, `--list-versions`, `--help`.
  7. Keep runtime dependencies to Python standard library only (`argparse`, `dataclasses`, `json`, `pathlib`, `subprocess`, `tempfile`, `shutil`, `textwrap`).
- **Acceptance**:
  - `./clasing-skill --help` prints the new CLI usage.
  - `python3 -m scripts.clasing_skill --list-packages` prints `skills` and `neurox` from `catalog.json`.
  - The CLI imports cleanly with no third-party Python dependency.
  - `catalog.json` is the single source of truth for supported packages in MVP.
- **Status**: [x] done

### Step 2: Define durable installer state and whole-package version resolution
- **What**: Add the config/lock schemas, JSON load-save layer, and deterministic git-based version resolution for local and external packages.
- **Why**: Version selection/tracking is the core MVP promise, and it must be durable and reproducible across repeated installs.
- **Where**: `schemas/skills.config.schema.json`, `schemas/skills.lock.schema.json`, `scripts/clasing_skill/state.py`, `scripts/clasing_skill/resolver.py`, `tests/clasing_skill/test_state.py`, `tests/clasing_skill/test_resolver.py`
- **How**:
  1. Create `schemas/skills.config.schema.json` with this exact persisted shape:
     ```json
     {
       "$schema": "https://json-schema.org/draft/2020-12/schema",
       "type": "object",
       "required": ["version", "defaults", "packages"],
       "properties": {
         "version": { "const": 1 },
         "defaults": {
           "type": "object",
           "required": ["interactive", "targets"],
           "properties": {
             "interactive": { "type": "boolean" },
             "targets": {
               "type": "array",
               "items": { "enum": ["claude", "opencode"] },
               "uniqueItems": true
             }
           }
         },
         "packages": {
           "type": "object",
           "additionalProperties": {
             "type": "object",
             "required": ["version", "targets"],
             "properties": {
               "version": { "type": "string" },
               "targets": {
                 "type": "array",
                 "items": { "enum": ["claude", "opencode"] },
                 "uniqueItems": true
               }
             }
           }
         }
       }
     }
     ```
  2. Create `schemas/skills.lock.schema.json` with the resolved install shape:
     ```json
     {
       "$schema": "https://json-schema.org/draft/2020-12/schema",
       "type": "object",
       "required": ["version", "generatedAt", "packages"],
       "properties": {
         "version": { "const": 1 },
         "generatedAt": { "type": "string" },
         "packages": {
           "type": "object",
           "additionalProperties": {
             "type": "object",
             "required": ["requestedVersion", "resolvedVersion", "commit", "repoUrl", "targets"],
             "properties": {
               "requestedVersion": { "type": "string" },
               "resolvedVersion": { "type": "string" },
               "resolvedRef": { "type": "string" },
               "commit": { "type": "string" },
               "repoUrl": { "type": "string" },
               "targets": {
                 "type": "object",
                 "patternProperties": {
                   "^(claude|opencode)$": {
                     "type": "object",
                     "required": ["status", "installedAt", "artifacts"],
                     "properties": {
                       "status": { "enum": ["installed", "unchanged"] },
                       "installedAt": { "type": "string" },
                       "artifacts": {
                         "type": "array",
                         "items": { "type": "string" }
                       }
                     }
                   }
                 }
               }
             }
           }
         }
       }
     }
     ```
  3. In `scripts/clasing_skill/state.py`, implement:
     - `def load_config(path: Path) -> dict:`
     - `def load_lock(path: Path) -> dict:`
     - `def write_config(path: Path, data: dict) -> None`
     - `def write_lock(path: Path, data: dict) -> None`
     - `def initialize_missing_state(state_dir: Path) -> tuple[Path, Path]`
     Behavior rules:
     - Missing files bootstrap to empty valid documents.
     - Malformed JSON raises a typed `StateFileError` with the file path.
     - Writes must be atomic: write `*.tmp`, `fsync`, then `replace()`.
  4. In `scripts/clasing_skill/resolver.py`, implement the version API:
     - `def list_versions(package: PackageDefinition, workdir: Path | None = None) -> list[str]`
     - `def resolve_version(package: PackageDefinition, selector: str) -> ResolvedVersion`
     - `def checkout_package(package: PackageDefinition, resolved: ResolvedVersion, temp_root: Path) -> Path`
  5. Use git as the version authority for both this repo and external repos:
     - `latest` = newest semver-like tag returned by `git tag --sort=-version:refname`
     - explicit tag like `v1.2.3` = must exist as `refs/tags/v1.2.3`
     - explicit branch/ref = verify with `git ls-remote <repo> <ref>` and record the resolved commit SHA
     - special selector `workspace` is allowed only for package `skills` when the current working directory is this repo; record `resolvedVersion: "workspace"`, the current commit SHA, and a `dirty` boolean in lock metadata
  6. Resolve versions before preflight so the lock file always records an exact `resolvedVersion`, `resolvedRef`, and `commit`.
- **Acceptance**:
  - Missing `skills.config.json` and `skills.lock.json` are created automatically as valid version-1 documents.
  - Corrupted state files fail with a clear error naming the bad file.
  - `python3 -m scripts.clasing_skill --list-versions --package skills` returns a sorted tag list or `workspace` when running inside the repo.
  - Lock entries contain exact commit SHAs, not only loose selectors like `latest`.
- **Status**: [x] done

### Step 3: Implement request resolution, interactive UX, and non-interactive defaults
- **What**: Build the MVP prompt flow that can collect packages, targets, and versions interactively, while reusing config defaults in non-interactive mode.
- **Why**: The product requires one command for both guided users and automation.
- **Where**: `scripts/clasing_skill/cli.py`, `scripts/clasing_skill/prompts.py`, `tests/clasing_skill/test_cli.py`
- **How**:
  1. In `scripts/clasing_skill/prompts.py`, implement simple standard-input prompts; do not add `questionary`, `click`, or `typer` in MVP.
  2. Add these functions:
     ```python
     def prompt_for_packages(available: dict[str, PackageDefinition]) -> list[str]
     def prompt_for_targets(default_targets: list[str]) -> list[str]
     def prompt_for_version(package_id: str, versions: list[str], default_version: str) -> str
     def confirm_plan(summary_lines: list[str], assume_yes: bool) -> bool
     ```
  3. Use this exact interactive order:
     - package selection (`skills`, `neurox`; allow comma-separated multiple values)
     - target selection (`claude`, `opencode`, `both`)
     - version selection for each chosen package (`latest`, explicit tag, `workspace` for `skills` only)
     - confirmation screen showing package → requested version → resolved target list
  4. In `scripts/clasing_skill/cli.py`, implement `def resolve_request(args, catalog, config) -> InstallRequest` with these rules:
     - If `--non-interactive` is set, require every missing value to come from flags or `skills.config.json`; otherwise exit with code `2` and message `Missing package/target/version for non-interactive mode`.
     - `--package` may be repeated; `--version` may be repeated as `package=selector`; `--target` accepts `claude`, `opencode`, or `both`.
     - If the user selects `both`, normalize to `['claude', 'opencode']` before validation.
     - Persist the user's requested defaults back to `skills.config.json` only after a successful install.
  5. Keep the UX predictable: every prompt must show the stored default when available, e.g. `Target [claude,opencode]:`.
  6. Print a final summary block before execution:
     ```text
     Install plan
     - skills  -> latest    -> claude, opencode
     - neurox  -> v0.9.0    -> claude
     ```
- **Acceptance**:
  - Running `./clasing-skill` with no flags enters the prompt flow.
  - Running `./clasing-skill --non-interactive --package skills --target claude --version skills=workspace` skips prompts entirely.
  - Running non-interactive mode without enough inputs exits before validation/install.
  - Stored defaults from `skills.config.json` are used when flags are absent.
- **Status**: [x] done

### Step 4: Implement blocking preflight validation before any install side effects
- **What**: Add a validator layer that checks global prerequisites, package compatibility, target compatibility, and Neurox constraints before any install command runs.
- **Why**: Preflight validation is an explicit MVP requirement and must prevent partial installs.
- **Where**: `scripts/clasing_skill/preflight.py`, `scripts/clasing_skill/cli.py`, `tests/clasing_skill/test_preflight.py`
- **How**:
  1. In `scripts/clasing_skill/preflight.py`, define:
     ```python
     @dataclass(slots=True)
     class ValidationIssue:
         level: str  # "error" | "warning"
         package_id: str | None
         target: str | None
         message: str
         fix_hint: str | None

     def run_preflight(request: InstallRequest, catalog: dict[str, PackageDefinition]) -> list[ValidationIssue]
     ```
  2. Split validation into these functions and call them in order:
     - `validate_state_files(...)`
     - `validate_package_target_compatibility(...)`
     - `validate_global_dependencies(...)`
     - `validate_target_dependencies(...)`
     - `validate_neurox_requirements(...)`
     - `validate_install_destinations(...)`
  3. Enforce these blocking rules for MVP:
     - `git` and `python3` must exist for every install.
     - Package id must exist in `catalog.json`.
     - Selected targets must be in `package.supported_targets`.
     - `skills` installation requires `neurox` in `PATH` because `scripts/setup.sh` calls `require_neurox()` for both `install_opencode()` and `install_claude()`.
     - `skills` + `opencode` requires `bun` or `npm` in `PATH` because `scripts/setup.sh` installs dependencies.
     - `skills` + `claude` requires write access to `~/.claude` and `~/.claude.json` parent directory.
     - `skills` + `opencode` requires write access to `~/.config/opencode` parent directory.
     - `neurox` install requires `go` in `PATH`; if building with SQLite FTS5 remains mandatory, also check `CGO_ENABLED` is not explicitly `0` and show the same guidance currently documented in `docs/installation.md`.
     - If any selected package-target pair fails validation, abort the entire run before any install step.
  4. Print grouped validation output like:
     ```text
     Preflight failed
     [skills][opencode] Error: bun or npm not found
       Fix: install bun or npm, then rerun clasing-skill
     [skills][claude] Error: neurox not found in PATH
       Fix: install neurox first or run clasing-skill --package neurox ...
     ```
  5. Do not allow `--yes` to bypass validation; it only skips the final confirmation prompt.
- **Acceptance**:
  - The CLI exits before any filesystem mutation when preflight returns an error.
  - Missing `neurox` blocks `skills` installs for both Claude and OpenCode.
  - Unsupported package-target combinations fail clearly.
  - Validation output tells the user exactly how to fix the issue.
- **Status**: [x] done

### Step 5: Implement package adapters and install execution
- **What**: Add the concrete install adapters for `skills` and `neurox`, reusing existing repo installers instead of duplicating target logic.
- **Why**: This is the execution core of the new CLI and the step that realizes the shared-assets + thin-adapters architecture.
- **Where**: `scripts/clasing_skill/installer.py`, `scripts/clasing_skill/adapters/__init__.py`, `scripts/clasing_skill/adapters/skills_repo.py`, `scripts/clasing_skill/adapters/neurox.py`, `tests/clasing_skill/test_adapters.py`
- **How**:
  1. In `scripts/clasing_skill/installer.py`, define the adapter contract:
     ```python
     @dataclass(slots=True)
     class InstallResult:
         package_id: str
         resolved_version: str
         commit: str
         targets: dict[str, dict]

     class PackageInstaller(Protocol):
         def install(self, checkout_dir: Path, request: InstallRequest, package: PackageDefinition) -> InstallResult: ...
     ```
  2. Implement `scripts/clasing_skill/adapters/skills_repo.py` with:
     - `class SkillsRepoInstaller:`
     - `def install(self, checkout_dir: Path, request: InstallRequest, package: PackageDefinition) -> InstallResult`
     - For target `claude`, run `python3` only indirectly via the checked-out repo's `scripts/setup.sh --claude`.
     - For target `opencode`, run the checked-out repo's `scripts/setup.sh --opencode`.
     - For `both`, run `--claude` then `--opencode` sequentially so failures are attributable by target.
     - Capture artifact paths exactly as they are today: `~/.claude`, `~/.claude/agents`, `~/.claude/skills`, `~/.claude/CLAUDE.md`, `~/.config/opencode`.
  3. Do not rewrite `scripts/setup.sh` or `scripts/install_claude_assets.py` in MVP. The adapter must treat them as the authoritative implementation for the `skills` package.
  4. Implement `scripts/clasing_skill/adapters/neurox.py` with:
     - `class NeuroxInstaller:`
     - `def install(self, checkout_dir: Path, request: InstallRequest, package: PackageDefinition) -> InstallResult`
     - Run `go build -tags fts5 -o neurox .` in the checked-out repo.
     - Install the binary into `~/.local/bin/neurox` by default; create `~/.local/bin` if needed.
     - Verify with `~/.local/bin/neurox status` after install.
     - Record the installed binary path as the package artifact for every selected target.
  5. In `scripts/clasing_skill/installer.py`, add `def install_packages(request: InstallRequest, catalog: dict[str, PackageDefinition]) -> list[InstallResult]` that:
     - resolves and checks out each package version
     - dispatches by `package.adapter`
     - aborts immediately on the first subprocess failure
     - never writes `skills.config.json` or `skills.lock.json` until all packages succeed
  6. Keep target-specific behavior thin: package adapters decide what to run; `scripts/setup.sh` continues to own per-target `skills` install behavior.
- **Acceptance**:
  - `skills` installs are executed by calling the checked-out repo's existing `scripts/setup.sh` rather than copied/reimplemented shell logic.
  - `neurox` installs build and verify the binary from the selected repo version.
  - A failure in any package/target aborts lock/config writes.
  - Install results contain artifact paths per target.
- **Status**: [x] done

### Step 6: Persist successful state, update docs, and add verification coverage
- **What**: Write config/lock updates only on success, document the new CLI, and add unit/smoke tests that lock the MVP behavior.
- **Why**: The feature is incomplete without durable tracking, reproducible docs, and regression coverage.
- **Where**: `scripts/clasing_skill/state.py`, `README.md`, `docs/installation.md`, `tests/clasing_skill/test_cli.py`, `tests/clasing_skill/test_state.py`, `tests/clasing_skill/test_resolver.py`, `tests/clasing_skill/test_preflight.py`, `tests/clasing_skill/test_adapters.py`
- **How**:
  1. After `install_packages(...)` succeeds, update `skills.config.json` with requested user intent:
     ```json
     {
       "version": 1,
       "defaults": {
         "interactive": true,
         "targets": ["claude", "opencode"]
       },
       "packages": {
         "skills": {
           "version": "latest",
           "targets": ["claude", "opencode"]
         },
         "neurox": {
           "version": "v0.9.0",
           "targets": ["claude"]
         }
       }
     }
     ```
  2. Update `skills.lock.json` with resolved state from the actual install results:
     ```json
     {
       "version": 1,
       "generatedAt": "2026-04-08T12:00:00Z",
       "packages": {
         "skills": {
           "requestedVersion": "latest",
           "resolvedVersion": "v1.4.2",
           "resolvedRef": "refs/tags/v1.4.2",
           "commit": "abc123...",
           "repoUrl": "https://github.com/joeldevz/skills.git",
           "targets": {
             "claude": {
               "status": "installed",
               "installedAt": "2026-04-08T12:00:00Z",
               "artifacts": ["~/.claude", "~/.claude/agents", "~/.claude/skills", "~/.claude/CLAUDE.md"]
             },
             "opencode": {
               "status": "installed",
               "installedAt": "2026-04-08T12:00:00Z",
               "artifacts": ["~/.config/opencode"]
             }
           }
         }
       }
     }
     ```
  3. If a package is already installed at the same resolved commit for the same targets, record `status: "unchanged"` in the lock entry and print `Already installed at requested version` instead of pretending a fresh install occurred.
  4. Update `README.md` Quick Start and `docs/installation.md` to make `./clasing-skill` the preferred entry point while documenting `scripts/setup.sh` as the underlying `skills` package installer used internally by the CLI.
  5. Add unit tests with `unittest` + `unittest.mock` for:
     - config/lock bootstrap and malformed JSON handling
     - version resolution rules (`latest`, tag, branch, `workspace`)
     - non-interactive missing-input failure
     - preflight failure aggregation
     - adapter subprocess dispatch for `skills` and `neurox`
  6. Add one smoke test that patches `subprocess.run` and verifies the end-to-end order: request resolution → version resolution → preflight → adapter dispatch → state writes.
- **Acceptance**:
  - Successful installs write both `skills.config.json` and `skills.lock.json` with the correct requested vs resolved values.
  - Reinstalling the same resolved version can be reported as `unchanged`.
  - README and installation docs point users at `./clasing-skill`.
  - Test suite covers interactive resolution, version resolution, preflight, and adapter dispatch.
- **Status**: [x] done

## Verification
```bash
# CLI entrypoint
./clasing-skill --help

# Unit tests
python3 -m unittest discover -s tests/clasing_skill -p 'test_*.py'

# Catalog / version discovery
python3 -m scripts.clasing_skill --list-packages
python3 -m scripts.clasing_skill --list-versions --package skills

# Non-interactive smoke using current checkout
./clasing-skill --non-interactive --package skills --target claude --version skills=workspace --state-dir /tmp/clasing-skill-smoke

# Preflight failure check (expected non-zero if neurox missing)
./clasing-skill --non-interactive --package skills --target opencode --version skills=workspace --state-dir /tmp/clasing-skill-smoke
```

Manual checks:
- Run `./clasing-skill` with no flags and verify the prompt order is package → target → version → confirmation.
- Install `skills` for `claude` and verify the CLI actually invokes the checked-out repo's `scripts/setup.sh --claude`.
- Install `skills` for `opencode` and verify `~/.config/opencode` is recorded in `skills.lock.json`.
- Install `neurox` and verify `~/.local/bin/neurox status` succeeds and the binary path appears in lock artifacts.
- Re-run the same command and verify the lock entry can report `unchanged` when the resolved commit matches.

## Risks / Notes
- The repo currently has no packaging metadata for a distributable CLI, so MVP should ship as a repo-root executable plus Python package, not as a published pip/npm package.
- `workspace` mode is necessary for local development before tags exist, but it must be clearly marked in `skills.lock.json` so it is not mistaken for a tagged release.
- `neurox` install behavior may require OS-specific handling beyond `~/.local/bin`; keep MVP scoped to user-writable Unix-like environments unless the user asks for broader platform support.
- `scripts/setup.sh` currently performs backups internally; the new CLI should not add a second conflicting backup layer around `skills` installs in MVP.
