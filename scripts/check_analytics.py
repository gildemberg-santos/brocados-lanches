#!/usr/bin/env python3
"""Valida a instalação do GTM, o nonce do CSP e a política de privacidade."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

GTM_ID = "GTM-NLTLDV4N"
NONCE_PLACEHOLDER = "__CSP_NONCE__"
NGINX_FILES = ("frontend/nginx.conf", "deploy/nginx.conf.template")


def collect_errors(root: Path = ROOT) -> list[str]:
    """Retorna os problemas estáticos da integração com o Google Tag Manager."""
    errors: list[str] = []

    def read(relative: str) -> str:
        path = root / relative
        if not path.is_file():
            errors.append(f"{relative}: arquivo não encontrado")
            return ""
        return path.read_text(encoding="utf-8")

    index = read("frontend/index.html")
    csp_ts = read("frontend/csp.ts")
    vite_config = read("frontend/vite.config.ts")
    privacy = read("frontend/src/views/PrivacyView.vue")

    if GTM_ID not in index:
        errors.append("frontend/index.html: container GTM não configurado")
    if f'nonce="{NONCE_PLACEHOLDER}"' not in index:
        errors.append("frontend/index.html: snippet do GTM sem placeholder de nonce")
    if "gtm.start" not in index or "gtm.js?id=" not in index:
        errors.append("frontend/index.html: use o snippet padrão do container GTM")
    if "gtag(" in index:
        errors.append("frontend/index.html: snippet de GA4 não faz parte desta integração")

    if "AsyncLocalStorage" not in csp_ts or "randomBytes(16)" not in csp_ts:
        errors.append("frontend/csp.ts: falta nonce aleatório por resposta HTML")
    if "buildCsp" not in csp_ts or NONCE_PLACEHOLDER not in csp_ts:
        errors.append("frontend/csp.ts: falta a política e o placeholder do nonce")
    if "cspNoncePlugin()" not in vite_config:
        errors.append("frontend/vite.config.ts: plugin do CSP não está ativo")

    for relative in NGINX_FILES:
        nginx = read(relative)
        if "'nonce-$request_id'" not in nginx:
            errors.append(f"{relative}: script-src sem nonce por requisição")
        for origin in (
            "https://www.googletagmanager.com",
            "https://www.google-analytics.com",
            "https://*.sentry.io",
        ):
            if origin not in nginx:
                errors.append(f"{relative}: CSP sem {origin}")
        script_src = re.search(r"script-src ([^;]+);", nginx)
        if script_src and "'unsafe-inline'" in script_src.group(1):
            errors.append(f"{relative}: script-src não pode usar 'unsafe-inline'")
        if f"sub_filter '{NONCE_PLACEHOLDER}' '$request_id';" not in nginx:
            errors.append(f"{relative}: falta substituir o placeholder pelo nonce")
        if "sub_filter_once off;" not in nginx:
            errors.append(f"{relative}: sub_filter precisa substituir todas as ocorrências")

    if "Não usamos analytics" in privacy or "nem serviços de analytics" in privacy:
        errors.append("frontend/src/views/PrivacyView.vue: a política ainda nega o uso de analytics")
    if "Google Tag Manager" not in privacy or "em todas as visitas" not in privacy:
        errors.append("frontend/src/views/PrivacyView.vue: falta declarar o GTM e seu carregamento")

    return errors


def main() -> int:
    errors = collect_errors()
    if errors:
        print("Integração do GTM inválida:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print("Integração do GTM OK: snippet padrão, nonce por resposta e política atualizada.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
