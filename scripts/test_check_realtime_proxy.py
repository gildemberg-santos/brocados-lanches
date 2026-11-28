"""Testes do validador de proxy WebSocket."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

import check_realtime_proxy


VALID_NGINX = """
map $http_upgrade $brocados_connection_upgrade {
    default upgrade;
    ''      close;
}
server {
    location /api/ {
        proxy_pass http://api:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $brocados_connection_upgrade;
        proxy_read_timeout 3600s;
    }
}
"""

VALID_NGINX_PRODUCTION = VALID_NGINX.replace(
    "http://api:8080", "http://127.0.0.1:${API_PORT}"
)

# Reproduz o template de produção antes da correção: HTTP/1.1 presente,
# mas nenhum cabeçalho de upgrade e nenhum timeout longo.
INVALID_NGINX_PRODUCTION = """
server {
    location /api/ {
        proxy_pass http://127.0.0.1:${API_PORT};
        proxy_http_version 1.1;
        proxy_set_header Host $http_host;
    }
}
"""

VALID_VITE = """
const apiProxy = {
  target: 'http://127.0.0.1:8080',
  ws: true,
}
export default defineConfig({
  server: {
    allowedHosts: ['overhelpful-readily-lynell.ngrok-free.dev'],
  },
})
"""

NGINX_ERROR = "nginx.conf"
PRODUCTION_ERROR = "deploy/nginx.conf.template"
VITE_ERROR = "vite.config.ts"


class CheckRealtimeProxyTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / "frontend").mkdir()
        (self.root / "deploy").mkdir()
        self.nginx = self.root / "frontend" / "nginx.conf"
        self.nginx_production = self.root / "deploy" / "nginx.conf.template"
        self.vite = self.root / "frontend" / "vite.config.ts"

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def configure(
        self,
        nginx: str = VALID_NGINX,
        vite: str = VALID_VITE,
        production: str = VALID_NGINX_PRODUCTION,
    ) -> None:
        self.nginx.write_text(nginx, encoding="utf-8")
        self.nginx_production.write_text(production, encoding="utf-8")
        self.vite.write_text(vite, encoding="utf-8")

    def test_configuracao_valida(self) -> None:
        self.configure()

        self.assertEqual(check_realtime_proxy.collect_errors(self.root), [])

    def test_upgrade_sem_mapa_e_rejeitado(self) -> None:
        self.configure(nginx=VALID_NGINX.replace("$brocados_connection_upgrade {", "{"))

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(any(NGINX_ERROR in error for error in errors))

    def test_location_sem_cabecalho_upgrade_e_rejeitado(self) -> None:
        self.configure(nginx=VALID_NGINX.replace("proxy_set_header Upgrade $http_upgrade;\n", ""))

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(any("Upgrade" in error for error in errors))

    def test_location_sem_timeout_e_rejeitado(self) -> None:
        self.configure(nginx=VALID_NGINX.replace("proxy_read_timeout 3600s;\n", ""))

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(any("proxy_read_timeout" in error for error in errors))

    def test_template_producao_sem_upgrade_e_rejeitado(self) -> None:
        self.configure(production=INVALID_NGINX_PRODUCTION)

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(
            any(PRODUCTION_ERROR in error and "mapa" in error for error in errors)
        )
        self.assertTrue(
            any(PRODUCTION_ERROR in error and "Upgrade" in error for error in errors)
        )
        self.assertTrue(
            any(PRODUCTION_ERROR in error and "Connection" in error for error in errors)
        )
        self.assertTrue(
            any(
                PRODUCTION_ERROR in error and "proxy_read_timeout" in error
                for error in errors
            )
        )

    def test_template_producao_ausente_e_reportado(self) -> None:
        self.configure()
        self.nginx_production.unlink()

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertIn(f"{PRODUCTION_ERROR}: arquivo não encontrado", errors)

    def test_proxy_vite_sem_ws_e_rejeitado(self) -> None:
        self.configure(vite=VALID_VITE.replace("ws: true,", ""))

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(any(VITE_ERROR in error for error in errors))

    def test_tunnel_host_ausente_e_rejeitado(self) -> None:
        self.configure(vite=VALID_VITE.replace(
            "allowedHosts: ['overhelpful-readily-lynell.ngrok-free.dev'],",
            "allowedHosts: [],",
        ))

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(any(VITE_ERROR in error and "ngrok-free.dev" in error for error in errors))

    def test_arquivo_ausente_e_reportado(self) -> None:
        self.configure()
        self.vite.unlink()

        errors = check_realtime_proxy.collect_errors(self.root)

        self.assertTrue(any("vite.config.ts: arquivo não encontrado" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
