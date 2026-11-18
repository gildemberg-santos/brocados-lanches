"""Testes do validador de documentação."""

from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

import check_docs


class CheckDocsTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / "docs").mkdir()

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def write(self, name: str, content: str) -> Path:
        path = self.root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
        return path

    def test_valid_relative_link(self) -> None:
        self.write("docs/guia.md", "# Guia\n")
        self.write("README.md", "[Guia](docs/guia.md)\n")

        self.assertEqual(check_docs.collect_errors(self.root), [])

    def test_broken_relative_link(self) -> None:
        self.write("README.md", "[Falta](docs/nao-existe.md)\n")

        errors = check_docs.collect_errors(self.root)

        self.assertEqual(len(errors), 1)
        self.assertIn("link quebrado", errors[0])
        self.assertIn("docs/nao-existe.md", errors[0])

    def test_external_anchor_and_protocol_links_are_ignored(self) -> None:
        self.write(
            "README.md",
            "[Site](https://example.com) [Topo](#topo) [Email](mailto:dev@example.com)\n",
        )

        self.assertEqual(check_docs.collect_errors(self.root), [])

    def test_anchor_and_query_in_relative_link_are_ignored(self) -> None:
        self.write("docs/guia.md", "# Guia\n")
        self.write("README.md", "[Guia](docs/guia.md#secao?x=1)\n")

        self.assertEqual(check_docs.collect_errors(self.root), [])

    def test_image_link_is_validated(self) -> None:
        self.write("README.md", "![Logo](assets/logo.png)\n")

        errors = check_docs.collect_errors(self.root)

        self.assertEqual(len(errors), 1)
        self.assertIn("assets/logo.png", errors[0])

    def test_unbalanced_code_fence(self) -> None:
        self.write("README.md", "```bash\necho oi\n")

        errors = check_docs.collect_errors(self.root)

        self.assertEqual(len(errors), 1)
        self.assertIn("cerca de código", errors[0])

    def test_balanced_code_fence(self) -> None:
        self.write("README.md", "```bash\necho oi\n```\n")

        self.assertEqual(check_docs.collect_errors(self.root), [])

    def test_ignored_directory_is_not_scanned(self) -> None:
        self.write("node_modules/pacote/README.md", "[Falta](nao-existe.md)\n")

        self.assertEqual(check_docs.collect_errors(self.root), [])

    def test_links_with_spaces_use_first_token(self) -> None:
        self.write("docs/guia.md", "# Guia\n")
        self.write("README.md", '[Guia](docs/guia.md "Título do link")\n')

        self.assertEqual(check_docs.collect_errors(self.root), [])


if __name__ == "__main__":
    unittest.main()
