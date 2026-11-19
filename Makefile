# Makefile raiz — contrato de entrada do projeto.
# Backend Go + frontend Vue + MongoDB via Docker Compose.

# Carrega .env no ambiente dos comandos quando o arquivo existir.
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help setup vendor dev dev-api dev-web test test-api test-web test-docs test-design \
	lint lint-api lint-web fmt fmt-api fmt-web docs-check design-check build build-api build-web \
	up up-offline down logs ps

help:
	@printf '%s\n' \
		'Comandos disponíveis:' \
		'  make setup       Instala dependências de backend e frontend' \
		'  make vendor      Atualiza backend/vendor após mudar dependências Go' \
		'  make dev         Sobe API e frontend em modo desenvolvimento' \
		'  make dev-api     Sobe apenas a API' \
		'  make dev-web     Sobe apenas o frontend' \
		'  make test        Roda documentação, design system, backend e frontend' \
		'  make lint        Roda go vet/golangci-lint, ESLint e vue-tsc' \
		'  make fmt         Formata Go e frontend' \
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

test: test-docs test-design test-api test-web

test-api:
	cd backend && go test -race ./...

test-web:
	cd frontend && pnpm test

test-docs:
	python3 -m unittest discover -s scripts -p 'test_check_docs.py'

test-design:
	python3 -m unittest discover -s scripts -p 'test_check_design_system.py'

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
	cd backend && gofmt -w ./cmd ./internal

fmt-web:
	cd frontend && pnpm format

docs-check:
	python3 scripts/check_docs.py

design-check:
	python3 scripts/check_design_system.py

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
