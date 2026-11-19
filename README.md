# Brocados Lanches — Projeto

**Status:** Fase 1 — esqueleto de produção disponível; catálogo, checkout e painel ainda não implementados.  
**Data de entrada:** 18/11/2026  
**Origem:** `~/Downloads/brocados_lanches_design_system_responsive_prototype_studio.html`  
**Classificação:** material de cliente — uso interno.

---

## 1. Objetivo

Reunir a presença digital de pedidos da Brocados Lanches:

- o protótipo e o design system recebidos;
- o esqueleto de produção em Go, Vue e MongoDB;
- o inventário do que já está pronto e do que ainda é simulação;
- o briefing e as pendências para a V1.

O arquivo `design-system/brocados-lanches-design-system-v2.5.html` é uma **referência visual e funcional**, não a base do produto. Ele não tem backend, persistência, autenticação nem pagamento.

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
- servidor HTTP com timeouts, logs JSON sem PII e shutdown gracioso;
- frontend Vue 3 com Vite, TypeScript strict, Pinia, Vue Router e Tailwind compilado;
- tela inicial que consulta o readiness da API;
- tokens do design system importados diretamente de `design-system/tokens.css`;
- testes de backend com `testing`/`httptest` e teste de integração opcional com MongoDB real;
- testes de frontend com Vitest + Vue Test Utils;
- Dockerfiles de API e web, Nginx same-origin e `compose.yaml` com MongoDB;
- validação de documentação e design system no gate de commit.

> **Atenção:** telefone, endereço, preços e taxas do seed são **placeholders**. Precisam ser confirmados com o cliente antes de qualquer uso público.

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
│   │   ├── repository/       # conexão MongoDB
│   │   ├── domain/           # entidades e erros (próximas fases)
│   │   ├── service/          # regras de negócio (próximas fases)
│   │   └── platform/         # clock, IDs, log e integrações
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
│   ├── Dockerfile
│   ├── Dockerfile.runtime
│   ├── nginx.conf
│   ├── index.html
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── vite.config.ts
│   └── tsconfig.json
├── scripts/                  # validadores de documentação e design system
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

> Nunca rode seed ou testes contra um banco de produção. Use `brocados_dev` ou `brocados_test`.

---

## 5. Testes e validação

```bash
make fmt          # gofmt + Prettier
make lint         # go vet/golangci-lint + ESLint + vue-tsc
make test         # docs, design system, Go e Vitest
make docs-check   # links relativos da documentação
make design-check # tokens, cores literais e CDNs no frontend
make build        # binário Go em backend/bin/api e dist do frontend
```

O teste de integração do repositório é opcional e usa um MongoDB real:

```bash
cd backend && MONGO_TEST_URI=mongodb://localhost:27017 \
  go test -race ./internal/repository -run TestConnectIntegration -v
```

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

- catálogo, cardápio e personalização;
- carrinho e checkout;
- criação e acompanhamento de pedidos;
- login e painel administrativo;
- seed de desenvolvimento, E2E e deploy de produção.

Esses pontos estão priorizados em `docs/02-plano-de-producao.md`.

---

## 7. Próximos passos

1. Confirmar dados reais da loja: nome, endereço, WhatsApp, Instagram, horários e tempo médio de entrega.
2. Validar cardápio, categorias, preços, adicionais e zonas de entrega.
3. Definir o fluxo de fechamento do pedido e o papel do WhatsApp.
4. Implementar o modelo de dados e o catálogo no backend e no frontend.
5. Implementar carrinho, checkout, persistência do pedido e painel administrativo.
6. Adicionar seed de desenvolvimento, E2E e pipeline de CI.
7. Definir hospedagem, domínio e política de retenção de dados.

---

## 8. Documentos relacionados

- `AGENTS.md` — regras de arquitetura, testes, commit e operação.
- `briefing/chamado.md` — escopo, pendências e dados a confirmar.
- `docs/01-inventario-prototipo.md` — inventário funcional e técnico do HTML.
- `docs/02-plano-de-producao.md` — fases, riscos e definição de pronto da V1.
- `design-system/README.md` — regra visual e validação obrigatória.
- `design-system/tokens.css` — fonte dos tokens de marca.
- `design-system/tokens.json` — os mesmos tokens para ferramentas.
- `design-system/seed.json` — dados fictícios para desenvolvimento.
- `design-system/brocados-lanches-design-system-v2.5.html` — protótipo visual; somente leitura.
- `compose.yaml` — MongoDB, API e web no ambiente integrado.
