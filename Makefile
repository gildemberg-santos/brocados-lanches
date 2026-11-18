# Makefile raiz — contrato de entrada do projeto.
# Os alvos de backend e frontend entram na Fase 1, junto com o esqueleto.

.PHONY: docs-check test-docs test

docs-check:
	python3 scripts/check_docs.py

test-docs:
	python3 -m unittest discover -s scripts -p 'test_*.py'

test: test-docs
