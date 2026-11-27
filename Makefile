# Makefile raiz — contrato de entrada do projeto.
# Backend Go + frontend Vue + MongoDB via Docker Compose.

# Carrega .env no ambiente dos comandos quando o arquivo existir.
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help setup vendor dev dev-api dev-web test test-api test-web test-docs test-design test-config test-e2e \
	lint lint-api lint-web fmt fmt-api fmt-web fmt-check docs-check design-check config-check build build-api build-web \
	seed seed-reset seed-check migrate-reference-check migrate-reference up up-offline down logs ps

help:
	@printf '%s\n' \
		'Comandos disponíveis:' \
		'  make setup       Instala dependências de backend e frontend' \
		'  make vendor      Atualiza backend/vendor após mudar dependências Go' \
		'  make dev         Sobe API e frontend em modo desenvolvimento' \
		'  make dev-api     Sobe apenas a API' \
		'  make dev-web     Sobe apenas o frontend' \
		'  make test        Roda documentação, design, configuração de proxy, seed, backend e frontend' \
		'  make test-e2e    Semena brocados_test e roda o Playwright (exige MongoDB)' \
		'  make lint        Roda go vet/golangci-lint, ESLint e vue-tsc' \
		'  make fmt         Formata Go e frontend' \
		'  make fmt-check   Verifica formatação sem alterar arquivos' \
		'  make config-check  Valida os proxies de desenvolvimento e produção do WebSocket' \
		'  make seed        Aplica os dados fictícios no banco de desenvolvimento' \
		'  make seed-reset  Esvazia as coleções semeadas e reaplica os dados fictícios' \
		'  make seed-check  Valida o seed.json sem acessar o MongoDB' \
		'  make migrate-reference-check  Valida o plano da migração sem acessar o MongoDB' \
		'  make migrate-reference  Aplica a migração versionada (exige MIGRATION_CONFIRM)' \
		'  make build       Compila backend e frontend' \
		'  make up          Sobe MongoDB, API e web com Docker Compose' \
		'  make up-offline  Compila o frontend no host e sobe sem baixar npm no Docker' \
		'  make down        Derruba os containers' \
		'  make logs        Acompanha os logs do Compose'

setup:
	cd backend && go mod download
	cd frontend && pnpm install

vendor:
	cd backend && go mod vendor

dev:
	$(MAKE) -j2 dev-api dev-web

dev-api:
	cd backend && go run ./cmd/api

dev-web:
	cd frontend && pnpm dev

test: test-docs test-design test-config seed-check test-api test-web

test-api:
	cd backend && go test -race ./...

test-web:
	cd frontend && pnpm test

test-e2e:
	cd backend && APP_ENV=test MONGO_DB=brocados_test MONGO_URI=mongodb://127.0.0.1:27017 go run ./scripts/seed
	cd frontend && pnpm test:e2e

test-docs:
	python3 -m unittest discover -s scripts -p 'test_check_docs.py'

test-design:
	python3 -m unittest discover -s scripts -p 'test_check_design_system.py'

test-config:
	python3 -m unittest discover -s scripts -p 'test_check_realtime_proxy.py'

seed:
	cd backend && go run ./scripts/seed

seed-reset:
	cd backend && go run ./scripts/seed -reset

seed-check:
	cd backend && go run ./scripts/seed -dry-run

# A migração é somente leitura por padrão. Para gravar, o chamador precisa
# exportar MIGRATION_CONFIRM=APPLY_VALIDATED_REFERENCE_DATA_20261125.
migrate-reference-check:
	cd backend && go run ./scripts/migrate/20261125-dados-referencia

migrate-reference:
	cd backend && go run ./scripts/migrate/20261125-dados-referencia -apply -confirm "$$MIGRATION_CONFIRM"

lint: lint-api lint-web

lint-api:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd backend && golangci-lint run; \
	else \
		printf '%s\n' 'golangci-lint não encontrado; usando go vet.'; \
		cd backend && go vet ./...; \
	fi

lint-web:
	cd frontend && pnpm lint
	cd frontend && pnpm typecheck

fmt: fmt-api fmt-web

fmt-api:
	cd backend && gofmt -w ./cmd ./internal ./scripts

fmt-web:
	cd frontend && pnpm format

fmt-check: fmt-check-api fmt-check-web

fmt-check-api:
	@cd backend && test -z "$$(gofmt -l ./cmd ./internal ./scripts)" || \
		(printf '%s\n' 'Go precisa de gofmt:'; gofmt -l ./cmd ./internal ./scripts; exit 1)

fmt-check-web:
	cd frontend && pnpm format:check

docs-check:
	python3 scripts/check_docs.py

design-check:
	python3 scripts/check_design_system.py

config-check:
	python3 scripts/check_realtime_proxy.py

build: build-api build-web

build-api:
	cd backend && mkdir -p bin && CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd/api

build-web:
	cd frontend && pnpm build

up:
	docker compose up --build -d --wait

up-offline: build-web
	docker compose -f compose.yaml -f compose.offline.yaml up --build -d --wait

down:
	docker compose down --remove-orphans

logs:
	docker compose logs -f --tail=100

ps:
	docker compose ps
