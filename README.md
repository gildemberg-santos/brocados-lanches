# Brocados Lanches — Projeto

**Status:** Fase 1 — CI, seed e harness E2E entregues; Fase 2 com catálogo navegável, personalização e carrinho entregues; checkout, pedido e painel ainda não implementados.  
**Data de entrada:** 18/11/2026  
**Origem:** `~/Downloads/brocados_lanches_design_system_responsive_prototype_studio.html`  
**Classificação:** material de cliente — uso interno.

---

## 1. Objetivo

Reunir a presença digital de pedidos da Brocados Lanches:

- o protótipo e o design system recebidos;
- o esqueleto de produção em Go, Vue e MongoDB;
- a documentação completa de produto, arquitetura, dados, API, design, backlog, testes e operação;
- o snapshot funcional do site público `https://cardpiobrocadoslanchess.ola.click/` como referência validável;
- o inventário do que já está pronto e do que ainda é simulação;
- o briefing e as pendências para a V1.

O arquivo `design-system/brocados-lanches-design-system-v2.5.html` é uma **referência visual e funcional**, não a base do produto. Ele não tem backend, persistência, autenticação nem pagamento. O site público da OlaClick é referência funcional e de conteúdo, não fonte visual nem importação automática.

**Stack:** Go + Chi (backend), Vue 3 + Vite + TypeScript + Pinia (frontend), MongoDB 8+ e Docker Compose para o ambiente integrado. As regras de arquitetura, testes e commit estão em `AGENTS.md`.

---

## 2. O que já existe

### Protótipo de design

O arquivo `design-system/brocados-lanches-design-system-v2.5.html` documenta a identidade visual e as telas de referência:

| Bloco | Conteúdo |
|---|---|
| Design System | Identidade e paleta, escala tipográfica, botões, cards, inputs, responsividade e componentes admin |
| Protótipo do cliente | Home, Cardápio, Personalizar, Carrinho, Checkout e Créditos |
| Painel administrativo | Dashboard, Produtos, Categorias, Locais e Taxas, Conta e Contato |
| Responsividade | Mobile em 375 px e desktop em 1280 px |
| Dados de demonstração | Categorias, produtos, zonas e pedidos fictícios em `design-system/seed.json` |

### Esqueleto de produção

O esqueleto da Fase 1 já inclui:

- API Go + Chi com `/healthz`, `/readyz` e raiz versionada em `/api/v1`;
- conexão com MongoDB e verificação real no readiness;
- domínio e repositórios de catálogo com validação, índices idempotentes e testes de integração com MongoDB real;
- API pública de catálogo com `GET /api/v1/menu` e `GET /api/v1/delivery-zones`, filtros de ativos/disponibilidade e testes de service e handler;
- servidor HTTP com timeouts, logs JSON sem PII e shutdown gracioso;
- frontend Vue 3 com Vite, TypeScript strict, Pinia, Vue Router e Tailwind compilado;
- home, cardápio com busca e filtro por categoria, detalhe de produto e personalização com grupos de opções, adicionais, observação e total estimado;
- carrinho persistente com schema versionado, validação contra o catálogo, quantidade, remoção, limpeza, estado vazio e totais estimados pela menor taxa ativa;
- estados de carregando, vazio, erro, esgotado e sucesso nas telas do catálogo e do carrinho;
- componentes reutilizáveis (`BaseButton`, `ProductCard`, `OptionGroup`, `AddonRow`, `PriceText` e outros) e services/stores tipados;
- tokens do design system importados diretamente de `design-system/tokens.css`;
- testes de backend com `testing`/`httptest` e teste de integração opcional com MongoDB real;
- testes de frontend com Vitest + Vue Test Utils;
- seed de desenvolvimento e teste a partir de `design-system/seed.json`, com valores em centavos e marcação de dado fictício;
- harness E2E com Playwright, smoke da home e navegação para o cardápio em desktop e mobile, com API isolada em `brocados_test`;
- CI no GitHub Actions com jobs de documentação/design, backend, frontend e E2E;
- Dockerfiles de API e web, Nginx same-origin e `compose.yaml` com MongoDB;
- validação de documentação e design system no gate de commit.

### Planejamento da V1

A documentação completa está em `docs/README.md`:

- `docs/03-prd.md` — escopo, personas, requisitos e critérios de lançamento;
- `docs/04-arquitetura.md` — camadas, fluxos, segurança e deploy;
- `docs/05-modelo-de-dados.md` — coleções, campos, índices e retenção;
- `docs/06-contrato-api.md` — endpoints, erros e idempotência;
- `docs/07-design-system-e-experiencia.md` — tokens, componentes, telas e acessibilidade;
- `docs/08-backlog.md` — épicos, histórias, tarefas e aceite;
- `docs/09-qualidade-e-testes.md` — matriz de testes e gates;
- `docs/10-operacao-e-lancamento.md` — ambientes, backup, monitoramento e LGPD;
- `docs/11-roadmap.md` — fases, marcos e caminho crítico;
- `docs/12-referencia-ola-click.md` — snapshot do site público;
- `docs/13-decisoes-e-pendencias.md` — decisões e perguntas abertas.

> **Atenção:** telefone, endereço, preços e taxas do seed são **placeholders**. O snapshot do site público também precisa de validação. Nada entra em produção sem confirmação do cliente.

---

## 3. Estrutura

```text
brocados-lanches/
├── AGENTS.md
├── README.md
├── Makefile
├── compose.yaml
├── compose.offline.yaml
├── .dockerignore
├── .env.example
├── backend/
│   ├── cmd/api/              # composição, configuração e shutdown
│   ├── internal/
│   │   ├── config/           # ambiente e validação
│   │   ├── httpapi/          # router, handlers e middleware
│   │   ├── repository/       # conexão MongoDB e repositórios de catálogo
│   │   ├── domain/           # entidades, enums e erros de domínio
│   │   ├── service/          # regras de negócio e casos de uso
│   │   └── platform/         # clock, IDs, log e integrações
│   ├── scripts/seed/         # seed de desenvolvimento e teste
│   ├── vendor/               # dependências Go versionadas para build offline
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── router/
│   │   ├── services/         # cliente HTTP e endpoints
│   │   ├── stores/           # Pinia
│   │   ├── views/
│   │   └── styles/           # Tailwind e import dos tokens
│   ├── e2e/                  # Playwright (smoke da base)
│   ├── Dockerfile
│   ├── Dockerfile.runtime
│   ├── nginx.conf
│   ├── index.html
│   ├── package.json
│   ├── playwright.config.ts
│   ├── pnpm-lock.yaml
│   ├── vite.config.ts
│   └── tsconfig.json
├── scripts/                  # validadores de documentação e design system
├── .github/workflows/ci.yml  # CI
├── .githooks/pre-commit
├── design-system/
├── docs/
├── briefing/
├── assets/
└── referencias/
```

**Nomes:** pastas em `kebab-case`; componentes Vue em `PascalCase.vue`; stores `useXStore.ts`; composables `useX.ts`; arquivos Go em `snake_case.go`.

---

## 4. Como rodar

### Pré-requisitos

| Ferramenta | Versão mínima |
|---|---|
| Go | 1.27 |
| Node.js | 22.12 |
| pnpm | 10.2 |
| Docker + Compose | Docker 29 / Compose v2.20+ |

### Com Docker Compose

```bash
cp .env.example .env
make up
```

O `make up` constrói as imagens, sobe MongoDB, API e web, e espera os health checks. Depois:

| Serviço | URL |
|---|---|
| Aplicação web | http://localhost:8081 |
| API | http://localhost:8080 |
| Liveness | http://localhost:8080/healthz |
| Readiness | http://localhost:8080/readyz |
| Readiness via web | http://localhost:8081/readyz |
| MongoDB | `localhost:27017` |

Comandos úteis:

```bash
make logs     # acompanha os logs
make ps       # mostra o estado dos containers
make down     # derruba os containers sem apagar o volume do MongoDB
```

O backend é construído com as dependências de `backend/vendor/`, sem acesso à rede. O frontend baixa as dependências do npm no build da imagem; o primeiro build precisa de acesso ao registry.

Se o Docker não conseguir acessar o registry npm, use o caminho offline:

```bash
make up-offline
```

Ele compila o frontend no host (`make build-web`) e monta a imagem web com `frontend/Dockerfile.runtime` e `compose.offline.yaml`, sem instalar pacotes npm dentro do Docker.

### Desenvolvimento local

Com um MongoDB disponível em `localhost:27017`:

```bash
cp .env.example .env
make setup
make dev
```

- API: `http://localhost:8080`
- Frontend com Vite: `http://localhost:5173`
- O Vite faz proxy de `/api`, `/healthz` e `/readyz` para a API.
- Para subir apenas o MongoDB via Docker: `docker compose up -d mongo`.

Também é possível rodar os lados separadamente:

```bash
make dev-api
make dev-web
```

### Variáveis de ambiente

`cp .env.example .env` cria a configuração local. O arquivo `.env` não é versionado e é carregado automaticamente pelo `Makefile`.

| Variável | Padrão | Uso |
|---|---|---|
| `APP_ENV` | `development` | `development`, `test` ou `production` |
| `PORT` | `8080` | Porta interna da API |
| `MONGO_URI` | `mongodb://localhost:27017` | Conexão usada fora do Compose |
| `MONGO_DB` | `brocados_dev` | Banco da API |
| `API_PORT` | `8080` | Porta publicada da API |
| `WEB_PORT` | `8081` | Porta publicada da web |
| `MONGO_PORT` | `27017` | Porta publicada do MongoDB |

### Dados de desenvolvimento

Os dados de `design-system/seed.json` são **fictícios** e servem apenas para desenvolvimento e teste.

```bash
make seed         # aplica em brocados_dev
make seed-reset   # esvazia categorias, produtos, zonas e loja, e reaplica
make seed-check   # valida o JSON sem conectar no MongoDB
```

- Reais são convertidos para centavos em Go, sem ponto flutuante (`23` → `2300`).
- Cada documento recebe `seed.fictitious: true` e `seed.source: "design-system/seed.json"`.
- O comando recusa `APP_ENV=production`, o banco `brocados` e nomes sem sufixo `_dev` ou `_test`.
- O reset não toca em usuários, sessões, pedidos nem auditoria.
- Nenhum usuário admin é criado pelo seed.

> Nunca rode seed ou testes contra um banco de produção. Use `brocados_dev` ou `brocados_test`.

---

## 5. Testes e validação

```bash
make fmt          # gofmt + Prettier
make fmt-check    # verifica formatação sem alterar arquivos
make lint         # go vet/golangci-lint + ESLint + vue-tsc
make test         # docs, design system, seed, Go e Vitest
make docs-check   # links relativos da documentação
make design-check # tokens, cores literais e CDNs no frontend
make build        # binário Go em backend/bin/api e dist do frontend
```

O teste de integração do repositório e o seed usam um MongoDB real:

```bash
cd backend && MONGO_TEST_URI=mongodb://localhost:27017 \
  go test -race ./internal/seed ./internal/repository -v

MONGO_URI=mongodb://localhost:27017 MONGO_DB=brocados_test make seed
```

O E2E usa Playwright, sobe a API com `brocados_test` e não faz parte de `make test`:

```bash
cd frontend && pnpm exec playwright install chromium
docker compose up -d mongo
make test-e2e
```

- `make test-e2e` sobe Vite e a API nas portas 5173 e 8080.
- Se a API já estiver rodando em 8080, pare antes: o servidor do E2E não é reaproveitado para não usar o banco de desenvolvimento.
- Para apontar para outro MongoDB, defina `MONGO_URI` e `MONGO_DB` no ambiente.

### CI

O workflow `.github/workflows/ci.yml` roda em pull request, push para `main` e execução manual, com quatro jobs:

| Job | Verifica |
|---|---|
| Documentação e design system | `docs-check`, `design-check` e testes dos validadores |
| Backend | `gofmt`, `go vet`, `go test -race` com MongoDB 8, seed e build |
| Frontend | Prettier, ESLint, `vue-tsc`, Vitest e build |
| E2E | Chromium + Playwright com MongoDB 8 como serviço |

O pipeline não usa secrets: as versões vêm de `backend/go.mod`, `frontend/package.json` e `frontend/pnpm-lock.yaml`. O relatório do Playwright fica como artefato quando o job E2E falha.

O gate de commit executa `make lint` e `make test`. Ative o hook uma vez por clone:

```bash
git config core.hooksPath .githooks
```

Se alguma verificação não puder ser executada, registre no PR o que ficou sem validação.

---

## 6. Limites do protótipo e do esqueleto

O protótipo **não** entrega:

- persistência: recarregar a página zera carrinho, produtos e configurações;
- pedido real: o botão de WhatsApp apenas simula o envio;
- backend, banco, autenticação ou controle de acesso;
- pagamento;
- imagens reais dos produtos;
- SEO, acessibilidade auditada ou política de privacidade;
- Tailwind compilado: o protótipo usa CDN e precisa de internet.

O esqueleto **ainda não** entrega:

- checkout e cotação server-side;
- criação e acompanhamento de pedidos;
- login e painel administrativo;
- deploy de produção.

Esses pontos estão priorizados em `docs/08-backlog.md` e `docs/11-roadmap.md`. O site público de referência está analisado em `docs/12-referencia-ola-click.md`.

---

## 7. Próximos passos

1. Fechar o Épico 0: confirmar catálogo, taxas, horários, contatos, pagamentos, fotos e textos legais.
2. Estender o smoke E2E para o fluxo de catálogo e carrinho.
3. Implementar cotação, checkout e pedido persistido (Épicos 3 e 4).
4. Implementar autenticação, catálogo admin, fila e transições (Épicos 6 e 7).
5. Fechar acessibilidade, SEO, testes, deploy, backup e go-live (Épicos 8 a 10).

---

## 8. Documentos relacionados

- `AGENTS.md` — regras de arquitetura, testes, commit e operação.
- `briefing/chamado.md` — escopo, pendências e dados a confirmar.
- `docs/README.md` — índice da documentação.
- `docs/01-inventario-prototipo.md` — inventário funcional e técnico do HTML.
- `docs/02-plano-de-producao.md` — plano histórico da Fase 1.
- `docs/03-prd.md` — PRD da V1.
- `docs/04-arquitetura.md` — arquitetura técnica.
- `docs/05-modelo-de-dados.md` — modelo MongoDB.
- `docs/06-contrato-api.md` — contrato HTTP.
- `docs/07-design-system-e-experiencia.md` — design system e UX.
- `docs/08-backlog.md` — backlog executável.
- `docs/09-qualidade-e-testes.md` — qualidade e testes.
- `docs/10-operacao-e-lancamento.md` — operação e go-live.
- `docs/11-roadmap.md` — roadmap.
- `docs/12-referencia-ola-click.md` — análise do site público.
- `docs/13-decisoes-e-pendencias.md` — decisões e pendências.
- `design-system/README.md` — regra visual e validação obrigatória.
- `design-system/tokens.css` — fonte dos tokens de marca.
- `design-system/tokens.json` — os mesmos tokens para ferramentas.
- `design-system/seed.json` — dados fictícios para desenvolvimento.
- `design-system/brocados-lanches-design-system-v2.5.html` — protótipo visual; somente leitura.
- `compose.yaml` — MongoDB, API e web no ambiente integrado.
