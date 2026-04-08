"""Interactive prompts for clasing-skill CLI.

Standard-library only prompts for collecting user input.
"""

from __future__ import annotations

import sys
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .models import PackageDefinition


def _input_with_default(prompt: str, default: str | None = None) -> str:
    """Get input from user with optional default value.

    Args:
        prompt: Base prompt text.
        default: Default value to show in brackets.

    Returns:
        User input or default if empty.
    """
    if default:
        full_prompt = f"{prompt} [{default}]: "
    else:
        full_prompt = f"{prompt}: "

    try:
        value = input(full_prompt).strip()
    except EOFError:
        print()
        return default or ""

    return value if value else (default or "")


def prompt_for_packages(available: dict[str, "PackageDefinition"]) -> list[str]:
    """Prompt user to select one or more packages.

    Shows available packages and allows comma-separated selection.

    Args:
        available: Dictionary of package_id -> PackageDefinition.

    Returns:
        List of selected package IDs.
    """
    print("\nAvailable packages:")
    for pkg_id in sorted(available.keys()):
        pkg = available[pkg_id]
        print(f"  {pkg_id} - {pkg.display_name}")
        print(f"    Targets: {', '.join(pkg.supported_targets)}")
        print(f"    Default version: {pkg.default_version}")

    available_ids = set(available.keys())

    while True:
        value = _input_with_default("\nSelect packages (comma-separated)", "skills")
        if not value:
            print("Please select at least one package.")
            continue

        selected = [p.strip() for p in value.split(",")]
        selected = [p for p in selected if p]  # Remove empty strings

        invalid = [p for p in selected if p not in available_ids]
        if invalid:
            print(f"Invalid package(s): {', '.join(invalid)}")
            print(f"Available: {', '.join(sorted(available_ids))}")
            continue

        return selected


def prompt_for_targets(default_targets: list[str]) -> list[str]:
    """Prompt user to select target environments.

    Args:
        default_targets: Default targets from config.

    Returns:
        List of selected targets (claude, opencode).
    """
    default_str = ",".join(default_targets) if default_targets else "claude,opencode"

    print("\nTarget environments:")
    print("  claude    - Claude Code")
    print("  opencode  - OpenCode")
    print("  both      - Both Claude Code and OpenCode")

    while True:
        value = _input_with_default("Select target(s)", default_str)
        if not value:
            targets = default_targets if default_targets else ["claude", "opencode"]
            return targets

        targets: list[str] = []
        for t in value.split(","):
            t = t.strip().lower()
            if t == "both":
                targets.extend(["claude", "opencode"])
            elif t in ("claude", "opencode"):
                targets.append(t)
            else:
                print(f"Invalid target: {t}")
                print("Valid: claude, opencode, both")
                targets = []
                break

        if targets:
            # Remove duplicates while preserving order
            seen: set[str] = set()
            unique: list[str] = []
            for t in targets:
                if t not in seen:
                    seen.add(t)
                    unique.append(t)
            return unique


def prompt_for_version(
    package_id: str,
    versions: list[str],
    default_version: str,
) -> str:
    """Prompt user to select a version for a package.

    Args:
        package_id: ID of the package being configured.
        versions: List of available versions.
        default_version: Default version to suggest.

    Returns:
        Selected version string.
    """
    print(f"\nAvailable versions for {package_id}:")

    # Show versions, marking default
    display_versions = versions if versions else [default_version]
    for v in display_versions:
        if v == default_version:
            print(f"  {v} (default)")
        elif v == "workspace":
            print(f"  {v} (current checkout)")
        else:
            print(f"  {v}")

    while True:
        value = _input_with_default(f"Select version for {package_id}", default_version)
        if not value:
            return default_version

        # Allow any version input (will be validated later by resolver)
        return value


def confirm_plan(summary_lines: list[str], assume_yes: bool) -> bool:
    """Show install plan and ask for confirmation.

    Args:
        summary_lines: Lines of the plan summary.
        assume_yes: If True, skip prompt and return True.

    Returns:
        True if user confirms or assume_yes is set.
    """
    print("\n" + "=" * 50)
    print("Install plan")
    print("=" * 50)
    for line in summary_lines:
        print(line)
    print("=" * 50)

    if assume_yes:
        print("Auto-confirmed (--yes)")
        return True

    while True:
        try:
            response = input("\nProceed with installation? [Y/n]: ").strip().lower()
        except EOFError:
            print()
            return False

        if response in ("", "y", "yes"):
            return True
        if response in ("n", "no"):
            return False
        print("Please enter 'y' or 'n'")


def prompt_missing_packages(
    available: dict[str, "PackageDefinition"],
    default_packages: list[str] | None = None,
) -> list[str]:
    """Prompt for packages when none were specified via CLI.

    Args:
        available: Dictionary of available packages.
        default_packages: Default packages from config.

    Returns:
        List of selected package IDs.
    """
    if default_packages:
        # Confirm defaults
        default_str = ",".join(default_packages)
        print(f"\nDefault packages from config: {default_str}")
        return prompt_for_packages(available)

    return prompt_for_packages(available)


def prompt_missing_targets(default_targets: list[str] | None = None) -> list[str]:
    """Prompt for targets when none were specified via CLI.

    Args:
        default_targets: Default targets from config.

    Returns:
        List of selected targets.
    """
    return prompt_for_targets(default_targets or ["claude", "opencode"])


def prompt_missing_version(
    package_id: str,
    versions: list[str],
    default_version: str,
) -> str:
    """Prompt for version when not specified via CLI.

    Args:
        package_id: ID of the package.
        versions: Available versions.
        default_version: Default version from package definition.

    Returns:
        Selected version string.
    """
    return prompt_for_version(package_id, versions, default_version)


def exit_interactive_required(
    what: str,
    flag: str | None = None,
    config_hint: str | None = None,
) -> None:
    """Exit with code 2 for missing required input in non-interactive mode.

    Args:
        what: Description of what's missing (e.g., "package").
        flag: CLI flag that would provide this value (e.g., "--package").
        config_hint: Config path hint for resolving via config file.

    Raises:
        SystemExit: With code 2.
    """
    if flag:
        msg = f"Error: {flag} required in non-interactive mode"
        if config_hint:
            msg += f" (or set {config_hint})"
    else:
        msg = f"Error: Missing {what} for non-interactive mode"
    print(msg, file=sys.stderr)
    sys.exit(2)


def confirm_trust_setup_scripts(
    catalog: dict[str, "PackageDefinition"],
    package_ids: list[str],
) -> bool:
    """Prompt for trust confirmation before executing external setup.sh scripts.

    Args:
        catalog: Package catalog.
        package_ids: List of package IDs being installed.

    Returns:
        True if user confirms trust, False otherwise.
    """
    print("\n" + "=" * 60)
    print("SECURITY WARNING: External Script Execution")
    print("=" * 60)
    print("\nThe following packages will execute external setup.sh scripts:")

    for pkg_id in package_ids:
        pkg = catalog.get(pkg_id)
        if pkg and pkg.adapter == "skills_repo":
            print(f"  - {pkg_id}: {pkg.repo_url}")

    print("\nThese scripts will have FULL SYSTEM ACCESS.")
    print("Only proceed if you TRUST these sources.")
    print("=" * 60)

    while True:
        try:
            response = (
                input("\nDo you trust these sources and want to proceed? [y/N]: ")
                .strip()
                .lower()
            )
        except EOFError:
            print()
            return False

        if response in ("y", "yes"):
            return True
        if response in ("", "n", "no"):
            return False
        print("Please enter 'y' or 'n'")
