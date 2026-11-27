#!/usr/bin/env python3
"""Valida se os proxies de desenvolvimento e produção repassam WebSocket."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
CONNECTION_MAP = "brocados_connection_upgrade"


def check_nginx(text: str) -> list[str]:
    """Confere o mapa de upgrade e os cabeçalhos do location /api/."""
    errors: list[str] = []
    if not re.search(r"map\s+\$http_upgrade\s+\$" + CONNECTION_MAP + r"\s*\{", text):
        errors.append("nginx.conf: falta o mapa $http_upgrade -> $" + CONNECTION_MAP)
    if not re.search(r"default\s+upgrade\s*;", text):
        errors.append("nginx.conf: o mapa de upgrade precisa de 'default upgrade;'")
    if not re.search(r"''\s+close\s*;", text):
        errors.append("nginx.conf: o mapa de upgrade precisa de \"'' close;\"")

    location = re.search(r"location\s+/api/\s*\{(?P<body>.*?)\n\s*\}", text, re.DOTALL)
    if location is None:
        errors.append("nginx.conf: location /api/ não encontrado")
        return errors

    body = location.group("body")
    required = (
        "proxy_http_version 1.1;",
        "proxy_set_header Upgrade $http_upgrade;",
        f"proxy_set_header Connection ${CONNECTION_MAP};",
        "proxy_read_timeout 3600s;",
    )
    for directive in required:
        if directive not in body:
            errors.append(f"nginx.conf: location /api/ precisa de '{directive}'")
    return errors


def check_vite(text: str) -> list[str]:
    """Confere se o proxy do Vite aceita upgrade WebSocket."""
    if re.search(r"ws:\s*true", text):
        return []
    return ["vite.config.ts: o proxy da API precisa de 'ws: true'"]


def collect_errors(root: Path = ROOT) -> list[str]:
    """Retorna os problemas de configuração encontrados."""
    errors: list[str] = []
    nginx = root / "frontend" / "nginx.conf"
    vite = root / "frontend" / "vite.config.ts"

    for path, checker in ((nginx, check_nginx), (vite, check_vite)):
        try:
            text = path.read_text(encoding="utf-8")
        except FileNotFoundError:
            errors.append(f"{path.relative_to(root)}: arquivo não encontrado")
            continue
        errors.extend(checker(text))
    return errors


def main() -> int:
    errors = collect_errors()
    if errors:
        print("Configuração de tempo real inválida:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print("Configuração de tempo real OK: proxy Vite e nginx repassam WebSocket.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
