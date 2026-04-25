# Deprecation Notice

The Python installer (`scripts/clasing_skill/`) has been superseded by the Go binary.
These files are kept for reference only and will be removed in v2.0.

## Migration

Use the Go binary instead:

- **macOS/Linux**: 
  ```bash
  curl -fsSL https://raw.githubusercontent.com/joeldevz/skills/main/scripts/install.sh | bash
  ```

- **Windows (PowerShell)**:
  ```powershell
  irm https://raw.githubusercontent.com/joeldevz/skills/main/scripts/install.ps1 | iex
  ```

- **Homebrew** (when available):
  ```bash
  brew install joeldevz/tap/clasing-skill
  ```

## What Changed

The new Go binary:
- ✅ No Python runtime required
- ✅ No bash scripts at runtime
- ✅ Native Windows, macOS, Linux support
- ✅ Cross-platform path handling (APPDATA, LOCALAPPDATA, HOME)
- ✅ Distributed via GitHub Releases with checksums
- ✅ All logic ported to Go packages (paths, adapters/claude, adapters/opencode)

## Legacy Code

The old Python code in this directory is provided for reference:
- `clasing_skill/` — Python package (deprecated)
- `../../../scripts/install_claude_assets.py` — Claude assets installer (superseded by `adapters/claude.go`)
