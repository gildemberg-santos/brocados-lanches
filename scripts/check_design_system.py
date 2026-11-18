#!/usr/bin/env python3
"""Valida o design system e bloqueia padrões visuais fora dos tokens."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
TOKENS_CSS = Path("design-system/tokens.css")
TOKENS_JSON = Path("design-system/tokens.json")
SCAN_SUFFIXES = {".css", ".html", ".js", ".mjs", ".scss", ".ts", ".vue"}
IGNORED_PARTS = {"node_modules", "dist", ".vite", "coverage", "__pycache__"}
HEX_RE = re.compile(r"#[0-9a-fA-F]{3,8}\b")
COLOR_FUNCTION_RE = re.compile(r"\b(?:rgb|rgba|hsl|hsla)\s*\(", re.IGNORECASE)
CDN_PATTERNS = {
    "Tailwind CDN": "cdn.tailwindcss.com",
    "Google Fonts CDN": "fonts.googleapis.com",
    "Google Fonts static": "fonts.gstatic.com",
}

COLOR_TOKENS = {
    "--bg-primary": ("colors", "background", "primary"),
    "--bg-surface": ("colors", "background", "surface"),
    "--bg-input": ("colors", "background", "input"),
    "--bg-border": ("colors", "background", "border"),
    "--brand-primary": ("colors", "brand", "primary"),
    "--brand-hover": ("colors", "brand", "hover"),
    "--brand-neon": ("colors", "brand", "neon"),
    "--text-main": ("colors", "text", "main"),
    "--text-sub": ("colors", "text", "sub"),
    "--status-open": ("colors", "status", "open"),
    "--status-error": ("colors", "status", "error"),
}
FONT_TOKENS = {
    "--font-display": ("typography", "families", "display"),
    "--font-sans": ("typography", "families", "sans"),
}
SIZE_TOKENS = {
    "--font-size-display": ("typography", "scale", "display", "size"),
    "--line-height-display": ("typography", "scale", "display", "lineHeight"),
    "--font-size-h1": ("typography", "scale", "h1", "size"),
    "--font-size-h2": ("typography", "scale", "h2", "size"),
    "--font-size-price": ("typography", "scale", "price", "size"),
    "--font-size-body": ("typography", "scale", "body", "size"),
    "--font-size-caption": ("typography", "scale", "caption", "size"),
}


def read_css_variables(path: Path) -> dict[str, str]:
    """Extrai as variáveis CSS declaradas no arquivo."""
    if not path.exists():
        return {}
    text = path.read_text(encoding="utf-8")
    return dict(re.findall(r"(--[a-z0-9-]+)\s*:\s*([^;]+);", text))


def read_json(path: Path) -> dict:
    """Lê o JSON de tokens; retorna vazio quando ausente ou inválido."""
    if not path.exists():
        return {}
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        return {}


def json_value(data: dict, path: tuple[str, ...]):
    """Navega no JSON de tokens sem lançar exceção."""
    current = data
    for key in path:
        if not isinstance(current, dict) or key not in current:
            return None
        current = current[key]
    return current


def normalize_font(value: str) -> str:
    """Extrai a primeira família de uma declaração font-family."""
    match = re.search(r"""["']?([^,"']+)""", value.strip())
    return match.group(1).strip() if match else value.strip()


def validate_tokens(root: Path) -> list[str]:
    """Compara tokens.css com tokens.json."""
    errors: list[str] = []
    css_path = root / TOKENS_CSS
    json_path = root / TOKENS_JSON

    if not css_path.exists():
        errors.append(f"{TOKENS_CSS}: arquivo ausente")
    if not json_path.exists():
        errors.append(f"{TOKENS_JSON}: arquivo ausente")
    if errors:
        return errors

    css = read_css_variables(css_path)
    data = read_json(json_path)

    if not css:
        errors.append(f"{TOKENS_CSS}: nenhuma variável CSS encontrada")
    if not data:
        errors.append(f"{TOKENS_JSON}: JSON inválido ou vazio")
    if errors:
        return errors

    for css_name, json_path_tuple in COLOR_TOKENS.items():
        css_value = css.get(css_name)
        json_value_raw = json_value(data, json_path_tuple)
        dotted = ".".join(json_path_tuple)

        if css_value is None:
            errors.append(f"{TOKENS_CSS}: variável {css_name} ausente")
            continue
        if json_value_raw is None:
            errors.append(f"{TOKENS_JSON}: campo {dotted} ausente")
            continue
        if css_value.strip().upper() != str(json_value_raw).strip().upper():
            errors.append(
                f"{TOKENS_JSON}: {dotted} ({json_value_raw}) diverge de "
                f"{TOKENS_CSS} ({css_value.strip()})"
            )

    for css_name, json_path_tuple in FONT_TOKENS.items():
        css_value = css.get(css_name)
        json_value_raw = json_value(data, json_path_tuple)
        dotted = ".".join(json_path_tuple)

        if css_value is None:
            errors.append(f"{TOKENS_CSS}: variável {css_name} ausente")
            continue
        if json_value_raw is None:
            errors.append(f"{TOKENS_JSON}: campo {dotted} ausente")
            continue
        if normalize_font(css_value) != str(json_value_raw).strip():
            errors.append(
                f"{TOKENS_JSON}: {dotted} ({json_value_raw}) diverge de "
                f"{TOKENS_CSS} ({normalize_font(css_value)})"
            )

    for css_name, json_path_tuple in SIZE_TOKENS.items():
        css_value = css.get(css_name)
        json_value_raw = json_value(data, json_path_tuple)
        dotted = ".".join(json_path_tuple)

        if css_value is None:
            errors.append(f"{TOKENS_CSS}: variável {css_name} ausente")
            continue
        if json_value_raw is None:
            errors.append(f"{TOKENS_JSON}: campo {dotted} ausente")
            continue
        if css_value.strip() != str(json_value_raw).strip():
            errors.append(
                f"{TOKENS_JSON}: {dotted} ({json_value_raw}) diverge de "
                f"{TOKENS_CSS} ({css_value.strip()})"
            )

    return errors


def frontend_files(root: Path) -> list[Path]:
    """Lista apenas os arquivos do frontend que precisam seguir o design system."""
    frontend = root / "frontend"
    if not frontend.exists():
        return []

    files: list[Path] = []
    for path in frontend.rglob("*"):
        if not path.is_file() or path.suffix not in SCAN_SUFFIXES:
            continue
        if any(part in IGNORED_PARTS for part in path.relative_to(frontend).parts):
            continue
        files.append(path)
    return sorted(files)


def scan_frontend(root: Path) -> list[str]:
    """Procura cores literais e dependências visuais proibidas no frontend."""
    errors: list[str] = []

    for path in frontend_files(root):
        relative = path.relative_to(root)
        text = path.read_text(encoding="utf-8")

        lowered = text.lower()
        for label, pattern in CDN_PATTERNS.items():
            if pattern in lowered:
                errors.append(f"{relative}: {label} não é permitido; use build local")

        for line_number, line in enumerate(text.splitlines(), 1):
            for match in HEX_RE.finditer(line):
                errors.append(
                    f"{relative}:{line_number}: cor literal {match.group(0)}; "
                    "use um token do design system"
                )
            for _ in COLOR_FUNCTION_RE.finditer(line):
                errors.append(
                    f"{relative}:{line_number}: cor literal rgb/hsl; "
                    "use um token do design system"
                )

    return errors


def collect_errors(root: Path = ROOT) -> list[str]:
    """Retorna inconsistências de token e violações visuais no frontend."""
    return validate_tokens(root) + scan_frontend(root)


def main() -> int:
    errors = collect_errors()
    if errors:
        print("Design system inválido:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    total = len(frontend_files(ROOT))
    print(
        "Design system OK: tokens consistentes; "
        f"{total} arquivo(s) de frontend validado(s)."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
