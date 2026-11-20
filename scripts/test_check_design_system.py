"""Testes do validador de design system."""

from __future__ import annotations

import json
import shutil
import tempfile
import unittest
from pathlib import Path

import check_design_system

REPO_ROOT = Path(__file__).resolve().parent.parent


class CheckDesignSystemTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        tokens_dir = self.root / "design-system"
        tokens_dir.mkdir()
        shutil.copy(
            REPO_ROOT / "design-system" / "tokens.css",
            tokens_dir / "tokens.css",
        )
        shutil.copy(
            REPO_ROOT / "design-system" / "tokens.json",
            tokens_dir / "tokens.json",
        )

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def write(self, name: str, content: str) -> Path:
        path = self.root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
        return path

    def write_tokens_json(self, data: dict) -> None:
        self.write("design-system/tokens.json", json.dumps(data, ensure_ascii=False))

    def read_tokens_json(self) -> dict:
        return json.loads(
            (self.root / "design-system" / "tokens.json").read_text(encoding="utf-8")
        )

    def test_repository_tokens_are_consistent(self) -> None:
        self.assertEqual(check_design_system.collect_errors(self.root), [])

    def test_detects_color_mismatch_between_css_and_json(self) -> None:
        data = self.read_tokens_json()
        data["colors"]["brand"]["primary"] = "#000000"
        self.write_tokens_json(data)

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("colors.brand.primary" in error for error in errors))

    def test_detects_missing_css_variable(self) -> None:
        css_path = self.root / "design-system" / "tokens.css"
        css = css_path.read_text(encoding="utf-8").replace(
            "--brand-primary: #FF8C00;", ""
        )
        css_path.write_text(css, encoding="utf-8")

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("--brand-primary" in error for error in errors))

    def test_detects_missing_json_field(self) -> None:
        data = self.read_tokens_json()
        del data["colors"]["brand"]["primary"]
        self.write_tokens_json(data)

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("colors.brand.primary" in error for error in errors))

    def test_detects_hardcoded_hex_color(self) -> None:
        self.write(
            "frontend/src/components/Bad.vue",
            "<style>.bad { color: #FF8C00; }</style>\n",
        )

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("cor literal #FF8C00" in error for error in errors))

    def test_detects_rgb_and_hsl_colors(self) -> None:
        self.write(
            "frontend/src/components/Bad.vue",
            "<style>.bad { color: rgb(255, 140, 0); background: hsl(0, 0%, 0%); }</style>\n",
        )

        errors = check_design_system.collect_errors(self.root)

        self.assertEqual(
            sum("cor literal rgb/hsl" in error for error in errors),
            2,
        )

    def test_accepts_token_reference(self) -> None:
        self.write(
            "frontend/src/components/Good.vue",
            "<style>.good { color: var(--brand-primary); }</style>\n",
        )

        self.assertEqual(check_design_system.collect_errors(self.root), [])

    def test_detects_tailwind_cdn(self) -> None:
        self.write(
            "frontend/index.html",
            '<script src="https://cdn.tailwindcss.com"></script>\n',
        )

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("Tailwind CDN" in error for error in errors))

    def test_detects_google_fonts_cdn(self) -> None:
        self.write(
            "frontend/index.html",
            '<link href="https://fonts.googleapis.com/css2?family=Inter">\n',
        )

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("Google Fonts CDN" in error for error in errors))

    def test_ignores_node_modules(self) -> None:
        self.write(
            "frontend/node_modules/pacote/Bad.vue",
            "<style>.bad { color: #000000; }</style>\n",
        )

        self.assertEqual(check_design_system.collect_errors(self.root), [])

    def test_ignores_playwright_artifacts(self) -> None:
        self.write(
            "frontend/playwright-report/index.html",
            "<style>.bad { color: #FF8C00; }</style>\n",
        )
        self.write(
            "frontend/test-results/smoke/error-context.html",
            "<style>.bad { color: #000000; }</style>\n",
        )

        self.assertEqual(check_design_system.collect_errors(self.root), [])

    def test_detects_missing_tokens_files(self) -> None:
        (self.root / "design-system" / "tokens.css").unlink()

        errors = check_design_system.collect_errors(self.root)

        self.assertTrue(any("tokens.css" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
