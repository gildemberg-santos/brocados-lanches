# Makefile raiz — contrato de entrada do projeto.
# Os alvos de backend e frontend entram na Fase 1, junto com o esqueleto.

.PHONY: docs-check design-check test-docs test-design test

docs-check:
	python3 scripts/check_docs.py

design-check:
	python3 scripts/check_design_system.py

test-docs:
	python3 -m unittest discover -s scripts -p 'test_check_docs.py'

test-design:
	python3 -m unittest discover -s scripts -p 'test_check_design_system.py'

test: test-docs test-design
