"""Testes do validador estático da integração com o Google Tag Manager."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

import check_analytics

FILES = (
    "frontend/index.html",
    "frontend/csp.ts",
    "frontend/vite.config.ts",
    "frontend/nginx.conf",
    "deploy/nginx.conf.template",
    "frontend/src/views/PrivacyView.vue",
)


class CheckAnalyticsTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        for relative in FILES:
            destination = self.root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes((check_analytics.ROOT / relative).read_bytes())

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def replace(self, relative: str, old: str, new: str) -> None:
        path = self.root / relative
        path.write_text(path.read_text(encoding="utf-8").replace(old, new), encoding="utf-8")

    def test_integracao_valida(self) -> None:
        self.assertEqual(check_analytics.collect_errors(self.root), [])

    def test_container_e_snippet_sao_obrigatorios(self) -> None:
        self.replace("frontend/index.html", check_analytics.GTM_ID, "GTM-ERRADO")
        self.replace("frontend/index.html", "gtm.start", "gtag('js'")

        errors = check_analytics.collect_errors(self.root)
        self.assertTrue(any("container GTM" in error for error in errors))
        self.assertTrue(any("snippet padrão" in error for error in errors))

    def test_nonce_do_dev_e_do_prod_sao_obrigatorios(self) -> None:
        self.replace("frontend/csp.ts", "randomBytes(16)", "randomBytes(0)")
        for relative in check_analytics.NGINX_FILES:
            self.replace(relative, "'nonce-$request_id'", "")
            self.replace(relative, "sub_filter_once off;", "")

        errors = check_analytics.collect_errors(self.root)
        self.assertTrue(any("nonce aleatório" in error for error in errors))
        self.assertEqual(sum(1 for error in errors if "nonce por requisição" in error), 2)
        self.assertEqual(sum(1 for error in errors if "todas as ocorrências" in error), 2)

    def test_script_src_sem_unsafe_inline(self) -> None:
        for relative in check_analytics.NGINX_FILES:
            self.replace(relative, "'nonce-$request_id'", "'unsafe-inline'")

        errors = check_analytics.collect_errors(self.root)
        self.assertEqual(sum(1 for error in errors if "unsafe-inline" in error), 2)

    def test_privacidade_precisa_declarar_o_gtm(self) -> None:
        self.replace("frontend/src/views/PrivacyView.vue", "Google Tag Manager", "Ferramenta externa")

        errors = check_analytics.collect_errors(self.root)
        self.assertTrue(any("falta declarar o GTM" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
