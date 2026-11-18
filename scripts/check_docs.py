#!/usr/bin/env python3
"""Valida links relativos e cercas de código da documentação Markdown."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
IGNORED_DIRS = {
    ".git",
    ".githooks",
    ".vite",
    "__pycache__",
    "coverage",
    "dist",
    "node_modules",
    "vendor",
}
LINK_RE = re.compile(r"!?\[[^\]]*\]\(([^)]+)\)")
FENCE = "```"


def markdown_files(root: Path) -> list[Path]:
    """Lista arquivos Markdown, ignorando diretórios de build e dependências."""
    return sorted(
        path
        for path in root.rglob("*.md")
        if not any(part in IGNORED_DIRS for part in path.relative_to(root).parts)
    )


def is_external(target: str) -> bool:
    """Links externos, âncoras e protocolos especiais não são resolvidos no disco."""
    return target.startswith(("http://", "https://", "mailto:", "tel:", "#"))


def normalize_target(target: str) -> str:
    """Remove título, query e fragmento de um alvo Markdown."""
    target = target.strip()
    if target.startswith("<") and target.endswith(">"):
        target = target[1:-1]
    if " " in target:
        target = target.split(" ", 1)[0]
    target = target.split("#", 1)[0].split("?", 1)[0]
    return target.replace("%20", " ")


def resolve_link(markdown_file: Path, target: str) -> Path | None:
    """Resolve um link relativo; retorna None para âncoras e links vazios."""
    normalized = normalize_target(target)
    if not normalized:
        return None
    return (markdown_file.parent / normalized).resolve()


def collect_errors(root: Path = ROOT) -> list[str]:
    """Retorna todos os problemas encontrados na documentação."""
    errors: list[str] = []
    for path in markdown_files(root):
        relative = path.relative_to(root)
        try:
            text = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            errors.append(f"{relative}: não é UTF-8 válido")
            continue

        if text.count(FENCE) % 2:
            errors.append(f"{relative}: cerca de código ``` sem fechamento")

        for match in LINK_RE.finditer(text):
            target = match.group(1).strip()
            if is_external(target):
                continue
            resolved = resolve_link(path, target)
            if resolved is not None and not resolved.exists():
                errors.append(f"{relative}: link quebrado -> {target}")

    return errors


def main() -> int:
    errors = collect_errors()
    if errors:
        print("Documentação inválida:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    total = len(markdown_files(ROOT))
    print(f"Documentação OK: {total} arquivo(s) Markdown validado(s).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
