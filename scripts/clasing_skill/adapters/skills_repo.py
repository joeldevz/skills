"""Skills repo adapter for clasing-skill.

Installs the skills package by calling the checked-out repo's scripts/setup.sh.
"""

from __future__ import annotations

import subprocess
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path

from ..installer import InstallResult, TargetResult
from ..models import InstallRequest, PackageDefinition


class SkillsRepoInstaller:
    """Installer for the skills repository.

    Delegates to the checked-out repo's scripts/setup.sh for actual installation,
    avoiding duplication of target-specific logic.
    """

    def install(
        self,
        checkout_dir: Path,
        request: InstallRequest,
        package: PackageDefinition,
    ) -> InstallResult:
        """Install skills package by running setup.sh for each target.

        Runs setup.sh --claude then --opencode sequentially so failures
        are attributable by target.

        SECURITY WARNING: This executes setup.sh from the checked-out repo.
        The script has full access to your system. Only install from trusted
        sources (e.g., official joeldevz/skills repository).

        Args:
            checkout_dir: Path to the checked-out skills repo
            request: Original install request
            package: Package definition from catalog

        Returns:
            InstallResult with artifact paths per target

        Raises:
            subprocess.CalledProcessError: If setup.sh fails
        """
        setup_script = checkout_dir / "scripts" / "setup.sh"

        # Security: Verify the script is from a trusted source
        # In MVP, we only warn - more strict verification would require
        # signature verification or checksum validation
        print(f"\n⚠️  Security Notice: About to execute {setup_script}")
        print(f"   Source: {package.repo_url}")
        print(f"   This script will have full system access.")
        print(f"   Only proceed if you trust this source.\n")

        if not setup_script.exists():
            raise FileNotFoundError(f"Setup script not found at {setup_script}")

        # Get commit from the checked-out repo
        commit = self._get_commit(checkout_dir)

        targets_result: dict[str, TargetResult] = {}
        timestamp = self._get_iso_timestamp()

        # Install for each target sequentially
        for target in request.targets:
            if target == "claude":
                self._install_claude(setup_script)
                targets_result["claude"] = TargetResult(
                    status="installed",
                    installed_at=timestamp,
                    artifacts=[
                        "~/.claude",
                        "~/.claude/agents",
                        "~/.claude/skills",
                        "~/.claude/CLAUDE.md",
                    ],
                )
            elif target == "opencode":
                self._install_opencode(setup_script)
                targets_result["opencode"] = TargetResult(
                    status="installed",
                    installed_at=timestamp,
                    artifacts=["~/.config/opencode"],
                )

        # Get version info
        requested_version = request.versions.get(package.id, package.default_version)
        # resolved_version comes from the checked-out commit - we'll use commit short hash
        # as a fallback since we don't have the resolved version object here
        resolved_version = (
            requested_version  # The installer will update this with actual resolved
        )

        return InstallResult(
            package_id=package.id,
            requested_version=requested_version,
            resolved_version=resolved_version,
            resolved_ref="",  # Will be updated by caller with actual resolved ref
            commit=commit,
            dirty=False,  # Will be updated by caller for workspace mode
            targets=targets_result,
        )

    def _install_claude(self, setup_script: Path) -> None:
        """Run setup.sh --claude.

        Args:
            setup_script: Path to setup.sh in the checked-out repo

        Raises:
            subprocess.CalledProcessError: If setup.sh fails
        """
        subprocess.run(
            ["bash", str(setup_script), "--claude"],
            check=True,
            capture_output=True,
            text=True,
        )

    def _install_opencode(self, setup_script: Path) -> None:
        """Run setup.sh --opencode.

        Args:
            setup_script: Path to setup.sh in the checked-out repo

        Raises:
            subprocess.CalledProcessError: If setup.sh fails
        """
        subprocess.run(
            ["bash", str(setup_script), "--opencode"],
            check=True,
            capture_output=True,
            text=True,
        )

    def _get_commit(self, checkout_dir: Path) -> str:
        """Get the current commit SHA from the checked-out repo.

        Args:
            checkout_dir: Path to the checked-out repo

        Returns:
            Full commit SHA
        """
        result = subprocess.run(
            ["git", "rev-parse", "HEAD"],
            cwd=checkout_dir,
            capture_output=True,
            text=True,
            check=True,
        )
        return result.stdout.strip()

    def _get_iso_timestamp(self) -> str:
        """Get current timestamp in ISO format."""
        return datetime.now(timezone.utc).isoformat()
