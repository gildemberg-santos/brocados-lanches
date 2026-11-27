"""Testes do validador estático da integração do Sentry."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

import check_sentry

FILES = (
    "backend/cmd/api/main.go",
    "backend/internal/httpapi/reporter.go",
    "backend/internal/httpapi/sentry.go",
    "frontend/src/features/monitoring/sentry.ts",
    "frontend/src/features/monitoring/sentryScrub.ts",
    "frontend/src/services/http.ts",
    "frontend/nginx.conf",
    "deploy/nginx.conf.template",
    ".env.example",
    "frontend/Dockerfile",
)


class CheckSentryTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        for relative in FILES:
            destination = self.root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes((check_sentry.ROOT / relative).read_bytes())

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def replace(self, relative: str, old: str, new: str) -> None:
        path = self.root / relative
        path.write_text(path.read_text(encoding="utf-8").replace(old, new), encoding="utf-8")

    def test_repositorio_valido(self) -> None:
        self.assertEqual(check_sentry.collect_errors(self.root), [])

    def test_dsn_literal_e_rejeitado(self) -> None:
        relative = "frontend/src/features/monitoring/sentry.ts"
        self.replace(
            relative,
            "const dsn = import.meta.env.VITE_SENTRY_DSN",
            "const dsn = 'https://0123456789abcdef0123456789abcdef@o1.ingest.us.sentry.io/1'",
        )

        errors = check_sentry.collect_errors(self.root)
        self.assertTrue(any("DSN literal" in error for error in errors))

    def test_tracing_e_replay_sao_rejeitados(self) -> None:
        relative = "frontend/src/features/monitoring/sentry.ts"
        self.replace(relative, "tracesSampleRate: 0", "tracesSampleRate: 1")
        self.replace(relative, "attachProps: false", "replayIntegration()")

        errors = check_sentry.collect_errors(self.root)
        self.assertTrue(any("tracesSampleRate" in error for error in errors))
        self.assertTrue(any("Replay" in error for error in errors))

    def test_sem_before_send_e_flush_e_rejeitado(self) -> None:
        relative = "backend/cmd/api/main.go"
        self.replace(relative, "SanitizeSentryEvent", "unsafePassthrough")
        self.replace(relative, "sentry.Flush", "sentry.NoFlush")

        errors = check_sentry.collect_errors(self.root)
        self.assertTrue(any("BeforeSend" in error for error in errors))
        self.assertTrue(any("Flush" in error for error in errors))

    def test_csp_sem_sentry_e_rejeitado(self) -> None:
        for relative in ("frontend/nginx.conf", "deploy/nginx.conf.template"):
            self.replace(relative, "connect-src 'self' https://*.sentry.io ", "connect-src 'self' ")

        errors = check_sentry.collect_errors(self.root)
        self.assertEqual(sum(1 for error in errors if "connect-src" in error), 2)

    def test_variaveis_de_build_documentadas(self) -> None:
        self.replace(".env.example", "VITE_SENTRY_DSN", "VITE_SENTRY_REMOVIDO")

        errors = check_sentry.collect_errors(self.root)
        self.assertTrue(any("VITE_SENTRY_DSN" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
