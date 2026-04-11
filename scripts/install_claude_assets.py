#!/usr/bin/env python3

from __future__ import annotations

import json
import shutil
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
OPENCODE_CONFIG = ROOT / "opencode" / "opencode.json"
OPENCODE_SKILLS = ROOT / "opencode" / "skills"
OPENCODE_COMMANDS = ROOT / "opencode" / "commands"
OPENCODE_TEMPLATES = ROOT / "opencode" / "templates"
CLAUDE_OVERLAY = ROOT / "claude-code" / "CLAUDE.md"
OBSOLETE_COMMANDS = {"plan", "execute", "test", "review", "status"}


def parse_frontmatter(text: str) -> tuple[dict[str, str], str]:
    if not text.startswith("---\n"):
        return {}, text

    end = text.find("\n---\n", 4)
    if end == -1:
        return {}, text

    raw = text[4:end].splitlines()
    body = text[end + 5 :]
    data: dict[str, str] = {}
    for line in raw:
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        data[key.strip()] = value.strip()
    return data, body.lstrip()


def dump_yaml_list(items: list[str], indent: int = 0) -> str:
    prefix = " " * indent
    return "\n".join(f"{prefix}- {item}" for item in items)


NEUROX_SKILL_BLOCK = """
## Neurox Memory (obligatorio)

Esta skill DEBE usar Neurox para memoria persistente:
- **Al iniciar**: `neurox_recall(query="{tema relevante}")` — buscar contexto previo
- **Cross-namespace**: `neurox_recall(query="{tema}")` sin namespace — inteligencia de otros proyectos
- **Al descubrir algo**: `neurox_save(...)` inmediatamente — no esperar al final
- Si no tienes acceso a Neurox tools, documenta en tu output qué información guardar.
"""


def normalize_command_body(body: str) -> str:
    replacements = {
        '"{argument}"': '"$ARGUMENTS"',
        "{argument}": "$ARGUMENTS",
        "{workdir}": "the current working directory",
        "{project}": "the current project",
        "Engram memory (`mem_search`)": "Neurox memory (`neurox_recall`)",
        "Engram persistent memory": "Neurox persistent memory",
        "Engram": "Neurox",
        "`mem_search`": "`neurox_recall`",
        "~/.config/opencode/templates/": "~/.claude/templates/",
        "Use `topic_key` for evolving topics so they update instead of duplicating": "Prefer updating an existing memory note when the topic already exists",
    }
    for old, new in replacements.items():
        body = body.replace(old, new)
    return body.rstrip() + "\n"


def write_text(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def render_agents(target: Path) -> None:
    config = json.loads(OPENCODE_CONFIG.read_text(encoding="utf-8"))
    agents = config["agent"]

    skill_map: dict[str, list[str]] = {
        "orchestrator": ["security"],
        "product-planner": ["prd"],
        "tech-planner": ["prd", "nestjs-patterns", "typescript-advanced-types"],
        "coder": ["nestjs-patterns", "typescript-advanced-types"],
        "verifier": [],
        "test-reviewer": [],
        "security": ["security"],
        "skill-validator": [],
        "advisor": [],
        "manager": [],
    }

    for name, agent in agents.items():
        prompt = agent["prompt"]
        description = agent["description"]
        skills = skill_map.get(name, [])

        frontmatter = [
            "---",
            f"name: {name}",
            f"description: {description}",
            "model: inherit",
            "memory: local",
        ]

        if skills:
            frontmatter.append("skills:")
            frontmatter.append(dump_yaml_list(skills, indent=2))

        frontmatter.append("---")
        content = "\n".join(frontmatter) + "\n\n" + prompt.strip() + "\n"
        write_text(target / "agents" / f"{name}.md", content)


def render_shared_skills(target: Path) -> None:
    for skill_dir in OPENCODE_SKILLS.iterdir():
        if not skill_dir.is_dir():
            continue
        destination = target / "skills" / skill_dir.name
        if destination.exists():
            shutil.rmtree(destination)
        shutil.copytree(skill_dir, destination)


def render_shared_conventions(target: Path) -> None:
    """Copy _shared/ protocol files to ~/.claude/skills/_shared/."""
    src = OPENCODE_SKILLS / "_shared"
    if not src.exists():
        return
    destination = target / "skills" / "_shared"
    if destination.exists():
        shutil.rmtree(destination)
    shutil.copytree(src, destination)


def render_templates(target: Path) -> None:
    destination = target / "templates"
    if destination.exists():
        shutil.rmtree(destination)
    shutil.copytree(OPENCODE_TEMPLATES, destination)


def command_intro(command_name: str, agent_name: str) -> str:
    if agent_name in ("planner", "tech-planner"):
        return (
            f"Use the `tech-planner` subagent for `/{command_name}` unless the task is too small to justify delegation.\n"
            "Keep the final answer concise and action-oriented.\n"
        )
    if agent_name == "coder":
        return (
            f"Use the `coder` subagent for `/{command_name}` whenever code or tests must be written or updated.\n"
            "Keep the work bounded to the requested scope.\n"
        )
    return (
        f"Run `/{command_name}` from the main conversation following the orchestrator workflow.\n"
        "Important: Claude subagents cannot spawn other subagents, so keep orchestration in the main thread.\n"
        "Delegate bounded code changes to `coder`, planning to `tech-planner`, and reviews to specialized agents.\n"
    )


def render_command_skills(target: Path) -> None:
    # Remove command skills that were deleted from the repo so Claude stays in sync.
    command_root = target / "skills"
    for obsolete in OBSOLETE_COMMANDS:
        obsolete_path = command_root / obsolete
        if obsolete_path.exists():
            shutil.rmtree(obsolete_path)

    for command_file in sorted(OPENCODE_COMMANDS.glob("*.md")):
        metadata, body = parse_frontmatter(command_file.read_text(encoding="utf-8"))
        name = command_file.stem
        description = metadata.get("description", f"Run /{name}")
        description = description.replace(
            "Engram persistent memory", "Neurox persistent memory"
        )
        agent_name = metadata.get("agent", "manager")
        transformed = normalize_command_body(body)

        frontmatter = [
            "---",
            f"name: {name}",
            f"description: {description}",
            "disable-model-invocation: true",
            "---",
        ]

        content = "\n".join(frontmatter)
        content += "\n\n"
        content += command_intro(name, agent_name)
        content += "\n"
        content += transformed
        content += NEUROX_SKILL_BLOCK
        write_text(target / "skills" / name / "SKILL.md", content)


def main() -> None:
    target = Path.home() / ".claude"
    target.mkdir(parents=True, exist_ok=True)
    render_agents(target)
    render_shared_skills(target)
    render_shared_conventions(target)
    render_templates(target)
    render_command_skills(target)
    print(f"Rendered Claude assets in {target}")
    print(f"Overlay file available at {CLAUDE_OVERLAY}")


if __name__ == "__main__":
    main()
