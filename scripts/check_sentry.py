#!/usr/bin/env python3
"""Valida a integração de monitoramento de erros sem rede e sem credenciais."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

DSN_PATTERN = re.compile(r"https://[0-9a-f]{16,}@[A-Za-z0-9.-]*sentry\.io/\d+", re.IGNORECASE)


def collect_errors(root: Path = ROOT) -> list[str]:
    """Retorna os problemas estáticos encontrados na integração do Sentry."""
    errors: list[str] = []

    def read(relative: str) -> str:
        path = root / relative
        if not path.is_file():
            errors.append(f"{relative}: arquivo não encontrado")
            return ""
        return path.read_text(encoding="utf-8")

    main_go = read("backend/cmd/api/main.go")
    reporter_go = read("backend/internal/httpapi/reporter.go")
    sentry_go = read("backend/internal/httpapi/sentry.go")
    sentry_ts = read("frontend/src/features/monitoring/sentry.ts")
    scrub_ts = read("frontend/src/features/monitoring/sentryScrub.ts")
    http_ts = read("frontend/src/services/http.ts")
    env_example = read(".env.example")
    dockerfile = read("frontend/Dockerfile")

    for relative, text in (
        ("backend/cmd/api/main.go", main_go),
        ("backend/internal/httpapi/reporter.go", reporter_go),
        ("backend/internal/httpapi/sentry.go", sentry_go),
        ("frontend/src/features/monitoring/sentry.ts", sentry_ts),
        ("frontend/src/features/monitoring/sentryScrub.ts", scrub_ts),
        ("frontend/src/services/http.ts", http_ts),
    ):
        if DSN_PATTERN.search(text):
            errors.append(f"{relative}: DSN literal no código; use variável de ambiente")
        if "VITE_SENTRY_AUTH_TOKEN" in text:
            errors.append(f"{relative}: token de build não pode ter prefixo VITE_")

    if "SendDefaultPii" in main_go:
        errors.append("backend/cmd/api/main.go: SendDefaultPii não é usado no SDK Go")
    if "BeforeSend:" not in main_go or "SanitizeSentryEvent" not in main_go:
        errors.append("backend/cmd/api/main.go: falta BeforeSend com SanitizeSentryEvent")
    if "sentry.Flush" not in main_go:
        errors.append("backend/cmd/api/main.go: falta sentry.Flush no shutdown")
    if "sentry.User{}" not in sentry_go or "event.Request = nil" not in sentry_go:
        errors.append("backend/internal/httpapi/sentry.go: o BeforeSend precisa limpar user e request")
    if "Recover" not in reporter_go:
        errors.append("backend/internal/httpapi/reporter.go: falta captura de panic")

    for required, relative in (
        ("tracesSampleRate: 0", "tracesSampleRate: 0 (sem tracing)"),
        ("attachProps: false", "attachProps: false"),
        ("userInfo: false", "userInfo: false"),
        ("httpBodies: []", "httpBodies: []"),
        ("status >= 500", "captura somente de 5xx/timeout"),
    ):
        if required not in sentry_ts:
            errors.append(f"frontend/src/features/monitoring/sentry.ts: falta {relative}")
    if "replayIntegration" in sentry_ts:
        errors.append("frontend/src/features/monitoring/sentry.ts: Session Replay não faz parte do escopo")
    if "reportHttpError" not in http_ts:
        errors.append("frontend/src/services/http.ts: falta captura central via reportHttpError")
    if "sanitizeSentryEvent" not in scrub_ts:
        errors.append("frontend/src/features/monitoring/sentryScrub.ts: falta sanitização do evento")

    for relative in ("frontend/nginx.conf", "deploy/nginx.conf.template"):
        nginx = read(relative)
        if "https://*.sentry.io" not in nginx:
            errors.append(f"{relative}: connect-src precisa liberar a ingestão do Sentry")

    for variable in ("SENTRY_DSN", "VITE_SENTRY_DSN", "SENTRY_ENVIRONMENT"):
        if variable not in env_example:
            errors.append(f".env.example: falta documentar {variable}")
    if "ARG VITE_SENTRY_DSN" not in dockerfile:
        errors.append("frontend/Dockerfile: falta ARG VITE_SENTRY_DSN no build do web")

    return errors


def main() -> int:
    errors = collect_errors()
    if errors:
        print("Integração do Sentry inválida:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print("Integração do Sentry OK: erros sem PII, sem tracing e CSP compatível.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
