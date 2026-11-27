# AGENTS.md — Brocados Lanches

**Versão:** 1.2  
**Atualizado em:** 27/11/2026  
**Stack:** Go + Chi · Vue 3 + Vite + TypeScript · MongoDB  
**Aplica-se a:** todo o repositório

> Este arquivo é a fonte de verdade para agentes de código e para quem trabalhar no projeto. Leia inteiro antes de implementar. Em caso de conflito, prevaleça nesta ordem: **1)** este arquivo; **2)** `briefing/chamado.md`; **3)** `docs/`; **4)** o protótipo em `design-system/`.

---

## 1. Contexto e objetivo

Brocados Lanches é uma lanchonete em Aquiraz–CE. O projeto é a presença digital de pedidos: cardápio, personalização, carrinho, checkout, acompanhamento e painel administrativo.

O arquivo `design-system/brocados-lanches-design-system-v2.5.html` é um **protótipo de referência visual e funcional**, não uma base de código. Ele não tem backend, persistência, autenticação nem pagamento.

**Objetivo da V1:** o cliente monta o pedido e ele é **registrado no servidor antes de qualquer envio**; a operação gerencia catálogo, zonas e pedidos pelo painel; preços, taxas e disponibilidade são sempre validados no backend.

**Estado em 26/11/2026:** documentação de produto, arquitetura, dados, API, design, backlog, testes e operação concluída; protótipo, tokens, esqueleto de produção e fundação da V1 disponíveis. A fundação já tem CI no GitHub Actions, seed de desenvolvimento e harness E2E com Playwright. O esqueleto sobe API Go, frontend Vue e MongoDB via Compose; o catálogo já tem domínio, validação, repositórios, índices idempotentes, API pública de leitura (`GET /menu` e `GET /delivery-zones`), home, cardápio com busca/filtros e detalhe com personalização. O carrinho já persiste IDs/escolhas com schema versionado, valida itens contra o catálogo e tem tela com quantidade, remoção, limpeza, estado vazio e subtotal sem taxa; taxa e total são exibidos no checkout após a escolha de entrega/retirada e zona; o botão de adicionar do detalhe alimenta o carrinho. A cotação `POST /api/v1/orders/quote` valida itens, opções, adicionais, zona e status efetivo da loja, calcula subtotal/taxa/total no servidor com `domain.BuildQuote` e não persiste nada. O checkout (`/checkout`) coleta entrega/retirada, zona, endereço, pagamento e troco, exibe totais cotados após a escolha necessária e envia apenas IDs/escolhas. Após o pedido gravado, salva no navegador forma de entrega, zona, nome, telefone, endereço e pagamento para o próximo pedido; observação e troco ficam de fora e as preferências podem ser apagadas em `/privacidade` (D-037). `POST /api/v1/orders` recalcula os valores, persiste o snapshot com número sequencial, `publicId` opaco e chave idempotente, e devolve o link `wa.me` após a gravação; `GET /api/v1/orders/{publicId}` oferece acompanhamento público sem PII. A confirmação limpa o carrinho somente após `201` e permite reenviar o link de handoff na mesma sessão. A autenticação administrativa já está implementada: `POST /api/v1/admin/auth/login`, `POST /api/v1/admin/auth/logout` e `GET /api/v1/admin/auth/me` usam Argon2id, sessão opaca com hash SHA-256 em cookie `HttpOnly`, rotação no login, expiração ociosa/absoluta, CSRF por `Origin`/`Referer` e rate limit de 5/min por IP+e-mail. O bootstrap do primeiro `owner` é feito pelas variáveis `ADMIN_BOOTSTRAP_*` na inicialização e só cria usuário quando a coleção está vazia. O shell/guard do painel (US-061) já entrega `/admin/login`, rotas lazy, menu por papel e a tela de acesso negado; o cliente HTTP usa `credentials: 'include'` e o proxy de desenvolvimento alinha `Origin`/`Host` para não falsificar o CSRF. A API administrativa de catálogo (US-022) já cobre `GET/POST/PATCH` de categorias, produtos, adicionais e zonas com paginação, allowlist de ordenação, validação e auditoria; `GET/PATCH /api/v1/admin/store` (US-065) controla kill switch, horários e contato com validação em `America/Fortaleza` e auditoria sem PII. As APIs administrativas de pedidos (`GET /api/v1/admin/orders`, `GET /api/v1/admin/orders/{id}` e `PATCH /api/v1/admin/orders/{id}/status`) já entregam fila com PII mascarada, detalhe completo e matriz de transições com `409`, `statusHistory`, timestamps e auditoria sem PII via compare-and-set. As telas de fila, detalhe e ações contextuais de status já estão implementadas no painel, assim como o CRUD de produtos, categorias, adicionais e zonas e a configuração da loja com kill switch e horários; `GET /api/v1/store` devolve o status efetivo calculado no servidor (kill switch + horário em `America/Fortaleza`) e a próxima abertura; o painel exibe esse preview na configuração da loja. As APIs `GET/POST/PATCH /admin/users` e `GET /admin/audit-logs` e as telas `/admin/usuarios`, `/admin/auditoria` e `/admin/conta` fecham a gestão do painel, com proteção do último owner e auditoria sem PII. Os rate limits do contrato (pedido, cotação, acompanhamento, cardápio e sessão administrativa), o cabeçalho `X-Request-Id`, o scrub de PII em logs, os testes de concorrência/idempotência e os metadados por rota também estão implementados; o E2E cobre o fluxo completo de compra em desktop e mobile, mas sua execução depende de MongoDB local. O seed de desenvolvimento agora replica o catálogo público (5 categorias, 37 produtos) com as imagens originais servidas localmente em `frontend/public/products/`; contato, tempo de entrega e horários da loja passaram a seguir o status publicado em `/contato` em 25/11/2026, e as 30 zonas e taxas foram fornecidas para o desenvolvimento, mantidas como referência até validação da loja. A migração versionada `20261125-dados-referencia` já está preparada para replicar essas quatro coleções em desenvolvimento ou produção sem apagar documentos, desativando apenas o que ficou fora do plano; em produção exige `APP_ENV=production`, `MONGO_DB=brocados`, `-apply` e confirmação explícita, preserva o kill switch atual e não foi executada. Desde 26/11/2026, após o pedido gravado o servidor também registra um perfil de conveniência por telefone normalizado, sem observação nem troco, com opt-out no checkout, retenção de 24 meses e resumo apenas no detalhe administrativo autenticado (D-038); a escrita é best-effort e nunca impede a criação do pedido.

Em 26/11/2026, o checkout passou a identificar o cliente no primeiro passo. Com a opção de conveniência habilitada, ele aplica preferências locais quando o telefone coincide; caso contrário, consulta `POST /api/v1/customers/profile` com o telefone no corpo e preenche entrega, endereço, zona e pagamento a partir de uma resposta mínima. A consulta não devolve histórico, observação, troco ou taxa, é limitada por IP e telefone, e o opt-out responde `found: false` (D-039). Desde 26/11/2026, criações e transições de status também são publicadas por WebSocket: `WS /orders/{publicId}/events` para o acompanhamento público e `WS /admin/orders/events` para a fila e o detalhe administrativos, com eventos mínimos sem PII e refetch HTTP no cliente (D-040). O hub é em memória e por instância; o proxy precisa repassar o upgrade. Desde 26/11/2026, o cliente pode rever seus pedidos em `/meus-pedidos`: `POST /orders` devolve um token opaco de histórico, guardado no navegador e reenviado em `Authorization` em `GET /orders/history`, paginado e sem PII (D-041). O token é derivado por HMAC-SHA256 da chave de idempotência com `ORDER_ACCESS_SECRET`; só o hash é gravado no pedido. Limpar dados do navegador ou trocar dispositivo remove o acesso naquele navegador; a limpeza também está em `/privacidade`. Em 26/11/2026, os testes de performance passaram a incluir benchmarks do hub WebSocket e do rate limiter (`make bench`) e um teste de carga HTTP com 50 leituras concorrentes de `/menu`, pedidos com chaves distintas e duplicatas (`make test-load`), conforme `docs/09-qualidade-e-testes.md` §11. Desde 26/11/2026, o painel tem integração de webhook em Configurações da Loja (`/admin/integracoes/webhook`, somente `owner`): `GET/PATCH /api/v1/admin/integrations/webhook` guarda URL e ativação em `store_settings.webhook`, aceita qualquer URL absoluta, sem exigir HTTPS, e audita apenas `active`/`urlConfigured` (D-042). Toda criação ou transição persistida dispara um POST assíncrono e best-effort com `{number, storeNumber, message}` — telefone do cliente e WhatsApp da loja lidos do mesmo `store_settings` —, timeout de 3 s, sem retry e sem duplicata em replay; falha externa nunca altera o pedido (D-046). A URL e o corpo contêm PII, então não entram em log, auditoria nem no WebSocket (D-040); o risco de SSRF e de envio a terceiros foi aceito explicitamente. Também em 26/11/2026, o telefone passou a ser validado no checkout e na API como número brasileiro possível (DDD em uso, celular com 9 dígitos iniciados em `9` ou fixo com 8 dígitos iniciados de `2` a `5`), para que um número inválido não seja usado como destinatário do webhook (D-043). Em 27/11/2026, o formulário de produtos do painel passou a aceitar upload de imagem: `POST /api/v1/admin/uploads` grava o arquivo na coleção `uploads` do MongoDB (JPEG/PNG/WebP/AVIF, até 2 MiB, tipo detectado pela assinatura) e `GET /api/v1/uploads/{id}` serve a imagem sem PII; o campo de imagem aceita tanto URL absoluta quanto os caminhos relativos já usados no seed (D-045).

Em 27/11/2026, o monitoramento de erros passou a usar Sentry somente para falhas inesperadas: `5xx`/panics na API e `5xx`/timeout/erro de carregamento de rota no navegador; `4xx` de regra de negócio, health checks e reconexões WebSocket não viram evento (D-047). `SENTRY_DSN` e `VITE_SENTRY_DSN` ficam por ambiente, sem valor no repositório, e DSN vazio desliga o SDK. O `BeforeSend` remove `user`/`request`, descarta chaves sensíveis e máscara telefone, e-mail, CEP, query, token e link `wa.me` nos dois lados; não há tracing, Session Replay, captura de props, source maps nem `SENTRY_AUTH_TOKEN`. O CSP dos proxies abre `connect-src` para `https://*.sentry.io`, e `scripts/check_sentry.py` valida essas invariantes no CI.

Em 25/11/2026, uma auditoria de conformidade comparou as telas com o protótipo e com `docs/07-design-system-e-experiencia.md`. Os desvios, a severidade e as tarefas de correção estão registrados em `docs/14-auditoria-design-system.md` e no Épico 12 de `docs/08-backlog.md`. O Épico 12 foi concluído em 25/11/2026, com as decisões de design registradas em `docs/13-decisoes-e-pendencias.md` (D-029 a D-031); privacidade, termos e créditos foram publicados em 25/11/2026 (D-033), com o canal do titular nos canais já existentes em `/contato` e os créditos liberados para indexação.

> A seção 5 de `docs/02-plano-de-producao.md` sugeria Next.js + Postgres. Esta decisão foi substituída pela stack definida neste arquivo.

---

## 2. Como um agente deve trabalhar

1. Leia este arquivo e o documento da área que vai mexer antes de editar.
2. Verifique o estado atual (`git status`, arquivos existentes). Não sobrescreva trabalho em andamento.
3. Faça a menor mudança que resolve o problema. Não misture refatoração com funcionalidade.
4. Não invente dados de negócio (preços, bairros, horários, WhatsApp, textos legais). Use `design-system/seed.json` apenas em desenvolvimento e marque como fictício.
5. **Nenhum arquivo de código nasce sem teste.** Endpoint, componente, store, script, migração ou correção de bug exige teste automatizado correspondente. Teste não é opcional.
6. Rode formatação, lint, typecheck, testes **e validação de documentação** antes de concluir. Nenhum commit sai sem testes e docs atualizados. Se algo não puder ser executado, diga explicitamente o que ficou sem verificação.
7. Atualize a documentação quando a decisão ou o contrato mudar.
8. Peça confirmação antes de adicionar dependência estrutural, trocar a stack ou alterar o design system.
9. Responda e documente em **português**. Código, nomes de API e identificadores em **inglês**.

---

## 3. Regras de negócio inegociáveis

1. **O servidor é a única fonte de verdade** para preço, adicional, taxa de entrega, total, disponibilidade e status da loja. O cliente envia apenas IDs, quantidades e escolhas.
2. **Nenhum pedido existe só no navegador.** Persista no MongoDB e só depois libere o handoff para o WhatsApp. Se o handoff falhar, o pedido continua salvo e pode ser reenviado.
3. **Não aceitar pedido** com loja fechada, zona inativa, produto indisponível ou adicional inválido para o produto.
4. **Retirada zera a taxa de entrega.** Entrega usa a taxa da zona selecionada.
5. **Congele os valores no pedido** (nome, preço unitário, adicionais, taxa e total). Mudanças futuras no catálogo não podem alterar pedidos antigos.
6. **Pedido público só por identificador opaco.** Nunca exponha ID sequencial ou listagem de pedidos de terceiros.
7. **Painel administrativo exige autenticação.** Nenhuma escrita anônima. Toda mutação administrativa gera auditoria.
8. **Dados pessoais são minimizados.** Nome, telefone e endereço existem apenas para executar o pedido. Não vão para logs, analytics ou mensagens de erro.
9. **Dinheiro é inteiro em centavos.** Proibido `float`/`float64`/`parseFloat` para valores monetários.
10. **Fuso da operação:** `America/Fortaleza`. Persistência e API em UTC; exibição no fuso local.
11. **Dados do seed são apenas de desenvolvimento.** Telefone, endereço, tempo de entrega e horários seguem o status publicado em `/contato`, e as zonas e taxas foram fornecidas em 25/11/2026; pedidos permanecem fictícios, e categorias, produtos, preços, disponibilidade e imagens foram copiados do catálogo público em 23/11/2026 como referência visual. Tudo precisa de validação do cliente antes de produção.
12. **A loja só aceita pedido** se o horário permitir **e** o interruptor manual estiver ligado. O interruptor é o kill switch da operação.
13. **Nenhum código sem teste.** Todo comportamento novo ou corrigido tem teste automatizado, e todos os testes precisam passar antes do commit.
14. **O design system é a fonte de verdade visual.** Nenhuma cor, fonte, escala, espaçamento, raio, sombra ou componente fora do padrão documentado. Toda UI passa por `make design-check` e revisão visual antes do commit; mudança de padrão visual exige aprovação e atualização conjunta de `tokens.css` e `tokens.json`.

---

## 4. Stack decidida

| Camada | Decisão | Observações |
|---|---|---|
| Backend | Go + Chi (`github.com/go-chi/chi/v5`) | Versão fixada em `backend/go.mod` |
| Banco | MongoDB 8+ | Driver oficial `go.mongodb.org/mongo-driver/v2` |
| Frontend | Vue 3 + Vite + TypeScript | Composition API, `<script setup lang="ts">` |
| Estado | Pinia | Stores no formato setup |
| Rotas | Vue Router | Views com lazy loading |
| Estilo | Tailwind CSS compilado | Tokens de `design-system/tokens.css`; nunca CDN |
| Validação | Zod no frontend; validação explícita no backend | Servidor continua autoritativo |
| Testes backend | `testing` + `httptest` + testcontainers | Sem mock de MongoDB nos testes de repositório |
| Testes frontend | Vitest + Vue Test Utils + Playwright | Fluxo crítico ponta a ponta |
| Pacotes frontend | pnpm | Lockfile commitado |
| Auth admin | Sessão em cookie `HttpOnly` | Token opaco com hash no MongoDB |
| Logs | `log/slog` em JSON | Sem PII |
| Tempo real | WebSocket (`github.com/coder/websocket`) | Hub em memória; eventos mínimos sem PII; canal admin autenticado |
| Ambiente integrado | Docker Compose | MongoDB, API e web com health checks |

**Não troque** framework, banco, gerenciador de pacotes ou biblioteca de estado sem decisão explícita do time.

---

## 5. Estrutura do repositório

Monorepo. Backend e frontend têm builds e deploys independentes.

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
│   │   ├── config/           # env, timeouts, validação de configuração
│   │   ├── httpapi/          # router, handlers, middleware, respostas
│   │   ├── domain/           # entidades, enums, erros de negócio
│   │   ├── service/          # regras de negócio e casos de uso
│   │   ├── repository/       # MongoDB
│   │   └── platform/         # clock, IDs, log, senha, WhatsApp (fase seguinte)
│   ├── scripts/seed/         # seed de desenvolvimento e teste
│   ├── vendor/               # dependências Go para build Docker offline
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router/
│   │   ├── stores/
│   │   ├── services/        # cliente HTTP e endpoints tipados
│   │   ├── components/      # componentes reutilizáveis
│   │   ├── views/           # páginas do cliente e do admin
│   │   ├── features/        # regras de UI por domínio
│   │   ├── composables/      # (fase seguinte)
│   │   ├── types/            # tipos compartilhados
│   │   ├── utils/            # normalização e utilitários puros
│   │   └── styles/
│   ├── e2e/                  # Playwright (smoke da home e navegação para o cardápio)
│   ├── Dockerfile
│   ├── Dockerfile.runtime
│   ├── nginx.conf
│   ├── index.html
│   ├── package.json
│   ├── playwright.config.ts
│   ├── pnpm-lock.yaml
│   ├── vite.config.ts
│   ├── vitest.config.ts
│   └── tsconfig.json
├── scripts/
│   ├── check_design_system.py
│   ├── check_docs.py
│   ├── check_realtime_proxy.py
│   ├── check_sentry.py
│   ├── test_check_design_system.py
│   ├── test_check_docs.py
│   ├── test_check_realtime_proxy.py
│   └── test_check_sentry.py
├── .github/
│   └── workflows/ci.yml      # CI: docs/design, proxies, Sentry, backend, frontend e E2E
├── .githooks/
│   └── pre-commit           # gate obrigatório antes de commit
├── design-system/           # referência visual; não editar o protótipo
│   ├── README.md
│   ├── brocados-lanches-design-system-v2.5.html
│   ├── tokens.css
│   ├── tokens.json
│   └── seed.json
├── docs/
├── briefing/
├── assets/
└── referencias/
```

**Nomes:** pastas em `kebab-case`; componentes Vue em `PascalCase.vue`; stores `useXStore.ts`; composables `useX.ts`; arquivos Go em `snake_case.go`; pacotes Go curtos e minúsculos.

**Dependências:** `backend/vendor/` é versionado para o build Docker não depender de rede. Depois de mudar dependências Go, rode `make vendor` e commite o resultado. `frontend/node_modules/`, `frontend/dist/`, cobertura e binários não são versionados.

**Testes:** ficam ao lado do código (`*_test.go`, `*.test.ts`) e são obrigatórios para todo comportamento novo.

---

## 6. Comandos

O `Makefile` é o contrato de entrada. Se um alvo não existir, crie-o junto com a funcionalidade em vez de exigir comandos soltos na documentação.

| Comando | Efeito |
|---|---|
| `make setup` | Instala dependências de backend e frontend |
| `make vendor` | Atualiza `backend/vendor` após mudar dependências Go |
| `make dev` | Sobe API e frontend em modo desenvolvimento |
| `make dev-api` | Só a API |
| `make dev-web` | Só o frontend |
| `make test` | Documentação, design system, proxies, Sentry, seed, backend e frontend |
| `make test-api` | `go test -race ./...` |
| `make test-web` | Vitest |
| `make test-e2e` | Playwright; exige MongoDB e sobe a API com `brocados_test` |
| `make bench` | Benchmarks Go do hub WebSocket e do rate limiter; sob demanda |
| `make test-load` | Teste de carga HTTP contra API em execução; aceita `LOADTEST_ARGS` |
| `make test-docs` | `unittest` dos validadores de documentação |
| `make test-design` | `unittest` do validador de design system |
| `make test-config` | `unittest` do validador de proxies de tempo real |
| `make seed` | Aplica `design-system/seed.json` em `brocados_dev` |
| `make seed-reset` | Esvazia as coleções semeadas e reaplica os dados de desenvolvimento |
| `make seed-check` | Valida o seed sem acessar o MongoDB |
| `make migrate-reference-check` | Valida o plano da migração versionada sem acessar o MongoDB |
| `make migrate-reference` | Aplica a migração versionada; exige `MIGRATION_CONFIRM` e alvo coerente |
| `make lint` | `golangci-lint` (ou `go vet`) + ESLint + `vue-tsc --noEmit` |
| `make fmt` | `gofmt` + Prettier |
| `make fmt-check` | Verifica formatação sem alterar arquivos |
| `make design-check` | Valida tokens e uso do design system no frontend |
| `make docs-check` | Valida links relativos da documentação |
| `make config-check` | Valida se Vite e nginx repassam o upgrade WebSocket |
| `make sentry-check` | Valida DSN por ambiente, sanitização, escopo de erros e CSP do Sentry |
| `make test-sentry` | `unittest` do validador do Sentry; entra em `make test` |
| `make build` | Binário Go em `backend/bin/api` e build do frontend |
| `make up` | Sobe MongoDB, API e web via Compose e espera os health checks |
| `make up-offline` | Compila o frontend no host e sobe via Compose sem baixar npm no Docker |
| `make down` | Derruba API, web e MongoDB |
| `make logs` | Acompanha os logs do Compose |
| `make ps` | Mostra o estado dos containers |

O build Docker do backend usa `backend/vendor/` e não acessa a rede. O build da imagem web baixa dependências do npm; o primeiro build precisa de acesso ao registry. Se o Docker não alcançar o npm, `make up-offline` compila o frontend no host e usa `frontend/Dockerfile.runtime` com `compose.offline.yaml`.

`make test-e2e` não faz parte de `make test`: exige MongoDB em `localhost:27017`, navegador Playwright instalado e sobe a API com `brocados_test`. `make seed` e `make seed-reset` operam apenas em banco com sufixo `_dev` ou `_test` e recusam produção. Não crie alvo vazio que finge executar uma verificação.

Nunca rode seed ou teste contra banco de produção. Use `brocados_dev` ou `brocados_test`.

---

## 7. Backend — Go + Chi

### 7.1 Camadas

```text
handler → service → repository → MongoDB
```

- `cmd/api`: apenas composição, configuração e shutdown. Sem regra de negócio.
- `httpapi`: lê e valida entrada, chama o service, mapeia erro para HTTP.
- `service`: regras de negócio e transações. Não importa `chi`, `net/http` nem driver do Mongo.
- `repository`: queries e índices. Traduz erro do Mongo para erro de domínio.
- `domain`: entidades, enums e erros sentinela. Sem dependência de infraestrutura.
- Injeção de dependência por construtor em `main`. Proibido estado global mutável e `init()` com efeito colateral.

### 7.2 HTTP

- Toda rota pública fica em `/api/v1`. Health checks ficam fora: `/healthz` (vivo) e `/readyz` (ping no Mongo).
- `http.Server` com `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` e shutdown gracioso.
- Limite de corpo JSON em 1 MB. Decodifique uma única vez e rejeite campos desconhecidos.
- Middleware mínimo: request ID, recover, log estruturado, limite de taxa, autenticação admin.
- Handlers não acessam MongoDB diretamente e não calculam preço.
- Paginação obrigatória em listagens; `pageSize` máximo de 100. Ordenação por allowlist — nunca aceite nome de campo arbitrário.
- Busca textual: escape de regex antes de usar `$regex`; nunca `$where`.

### 7.3 Erros

- Erros de domínio são sentinelas ou tipos em `domain`, comparáveis com `errors.Is`/`errors.As`.
- O mapeamento para HTTP é centralizado. Não espalhe `http.Error` e strings soltas.
- Nunca devolva erro bruto do MongoDB, stack trace ou detalhe interno ao cliente.
- Log de 5xx com contexto interno; resposta pública genérica.
- Códigos de erro estáveis: `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `STORE_CLOSED`, `ZONE_UNAVAILABLE`, `PRODUCT_UNAVAILABLE`, `RATE_LIMITED`, `INTERNAL_ERROR`.

### 7.4 MongoDB

- Um único `mongo.Client` por processo, criado no start e desconectado no shutdown.
- `context.Context` com timeout em toda operação (padrão 5 s; checkout até 10 s).
- `_id` é `ObjectID`. A API expõe `id` como string hexadecimal; nunca devolva `_id`.
- Coleções em plural `snake_case`; campos em `camelCase`.
- Datas em BSON `datetime` UTC. `createdAt` e `updatedAt` em todo documento mutável.
- Índices são criados de forma idempotente na inicialização, em pacote dedicado e com log.
- Transação apenas quando a operação grava em mais de um documento. O contador de pedido usa `findOneAndUpdate` com `$inc`.
- Não faça `find` sem limite em coleção que cresce. Não use arrays ilimitados dentro de um documento.
- Testes de repositório usam Mongo real e isolado (testcontainers), nunca o banco de desenvolvimento.

### 7.5 Dinheiro, datas e fuso

- Valores monetários são `int64` em centavos, com sufixo `Cents`: `priceCents`, `feeCents`, `totalCents`.
- Percentuais, se existirem, usam pontos-base inteiros. Nunca `float64`.
- `time.Time` sempre UTC na persistência e na API (RFC 3339). Converta para `America/Fortaleza` só na apresentação.
- Carregue o fuso com `time.LoadLocation("America/Fortaleza")`; não use offset fixo.
- O seed tem `price` em reais (ex.: `23`, `54.9`). Na importação, converta para centavos (`2300`, `5490`). Não importe `float` para o banco.

### 7.6 Segurança e LGPD

- Senha com **Argon2id** (parâmetros documentados no código). Mínimo de 12 caracteres. Nunca logue senha.
- Sessão: token aleatório de 32 bytes, guardado apenas como SHA-256 no MongoDB; coleção `sessions` com TTL.
- Cookie `HttpOnly`, `Secure` em produção, `SameSite=Lax`, `Path=/`. Em produção, prefixo `__Host-`. Expiração ociosa de 8 h e absoluta de 7 dias.
- CSRF: valide `Origin`/`Referer` em métodos mutáveis. Se o deploy deixar de ser same-origin, implemente token double-submit.
- Rate limit: login (5/min por IP+email), criação de pedido (10/min por IP e 3/min por telefone). Resposta `429` sem revelar existência de conta.
- Login com mensagem genérica; rotacione a sessão no login e apague no logout.
- CORS: same-origin por padrão. Em desenvolvimento, use proxy do Vite. Allowlist explícita se precisar.
- Headers de segurança no proxy reverso: CSP compatível com o build, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`, HSTS em produção.
- Nunca logue `name`, `phone`, `address`, `note`, `Cookie`, `Authorization` ou corpo de checkout/login.
- Coleta mínima: nome, telefone, endereço de entrega e observação do pedido. Não peça CPF, e-mail ou dados de cartão na V1.
- Retenção proposta (a validar com o jurídico): pedidos 24 meses, depois anonimizar; sessões 7 dias; auditoria 12 meses.

---

## 8. Frontend — Vue 3 + Vite + TypeScript

### 8.1 Organização

- Composition API com `<script setup lang="ts">`. Proibido Options API e mixins.
- TypeScript `strict`. Proibido `any`; use `unknown` e narrowing.
- Views em `src/views`, componentes reutilizáveis em `src/components`, regra de domínio de UI em `src/features/<dominio>`.
- Componente apresenta e emite eventos; regra de negócio não mora no componente.
- Rotas do cliente sugeridas: `/`, `/cardapio`, `/produto/:slug`, `/carrinho`, `/checkout`, `/pedido/:publicId`, `/creditos`.
- Rotas do admin: `/admin/login`, `/admin`, `/admin/produtos`, `/admin/categorias`, `/admin/adicionais`, `/admin/locais`, `/admin/pedidos`, `/admin/conta`.
- Views do admin usam lazy loading. Guard de rota consulta `GET /admin/auth/me`; o guard é conveniência, a autorização real é no servidor.

### 8.2 Estado e dados

- Pinia no formato setup. Stores por domínio: `useCartStore`, `useMenuStore`, `useAdminAuthStore`.
- Carrinho persiste em `localStorage` apenas IDs, quantidades e escolhas — nunca preço ou total confiável. O token opaco de histórico fica em storage versionado em `src/features/orders/orderAccessToken.ts`, e é apagado em `/privacidade`; não vai para URL, log ou mensagem de erro.
- Preço exibido vem do catálogo; no checkout, o total exibido é o retorno do servidor.
- Após `201` do pedido, limpe o carrinho. Se a chamada falhar, preserve o carrinho.
- Um único cliente HTTP em `src/services/http.ts`, com tipagem, timeout, tratamento do envelope de erro e `credentials: 'include'`.
- Endpoints ficam em `src/services/<dominio>.ts`; componentes não chamam `fetch` diretamente.
- Toda tela tem estados de **carregando, vazio, erro e sucesso**. Erro de API vira mensagem em português; nunca mostre JSON cru.
- Não acesse `localStorage` fora de stores/utilitários.

### 8.3 UI e design system

- Tailwind compilado no build. Proibido CDN em qualquer ambiente.
- Cores, fontes e escala vêm de `design-system/tokens.css`. Não use hex solto em componente; use token ou classe do tema.
- `make design-check` é obrigatório: bloqueia cor literal, CDN e divergência entre `tokens.css` e `tokens.json`.
- O frontend consome os tokens por alias/import do `design-system/` ou por cópia gerada. Nunca edite a cópia à mão.
- Fonte Bebas Neue e Inter self-hosted em produção, com `font-display: swap`. Sem Google Fonts em produção.
- Reaproveite componentes (`BaseButton`, `ProductCard`, `QuantityStepper`, `PriceText`) antes de criar variante nova.
- Mobile-first, referência de 375 px; valide também em 1280 px.
- Formate dinheiro com `Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' })` apenas na exibição.
- Imagens em WebP/AVIF, com `width`, `height`, `loading="lazy"` e `alt` real. Não use o ícone genérico do protótipo como foto de produto.

### 8.4 Acessibilidade e SEO

- Meta: WCAG 2.1 AA — contraste, foco visível, navegação por teclado, rótulos em inputs, erro associado ao campo e alvo de toque de 44 px.
- Não comunique status só por cor. Respeite `prefers-reduced-motion`.
- Título e metadados por rota com `@unhead/vue` (ou equivalente). Sitemap e JSON-LD `Restaurant`/`Menu` na V1.
- A SPA limita SEO orgânico. Se SEO virar prioridade, reavalie Nuxt com o time; não migre por conta própria.
- Analytics só com consentimento e sem PII. Não instale script de terceiro sem revisão.

### 8.5 Gate visual

- Nenhum componente entra sem validação contra `design-system/README.md`.
- `make design-check` roda no hook e no CI; falha bloqueia o commit.
- O validador cobre tokens, cores literais e CDNs. Revisão visual humana cobre espaçamento, hierarquia, estados e responsividade em 375 px e 1280 px.
- Mudança de padrão visual exige aprovação de design e atualização de `tokens.css` + `tokens.json` + documentação.

---

## 9. Contrato de API

### 9.1 Convenções

- Base: `/api/v1`. JSON em `camelCase`.
- Sucesso: `{ "data": ... }`. Lista paginada: `{ "data": [...], "meta": { "page": 1, "pageSize": 20, "total": 0, "totalPages": 0 } }`.
- Erro: `{ "error": { "code": "VALIDATION_ERROR", "message": "...", "details": [{ "field": "phone", "code": "INVALID", "message": "..." }] } }`.
- `201` com `Location` ao criar pedido. `204` em logout/exclusão sem corpo.
- `POST /orders` exige `Idempotency-Key`. A mesma chave devolve o mesmo pedido; não cria duplicata.
- Datas em RFC 3339 UTC. IDs como string hexadecimal de 24 caracteres.
- Mudança incompatível exige `/api/v2`; não quebre contrato em silêncio. Atualize `docs/06-contrato-api.md` (ou OpenAPI) na mesma alteração.

### 9.2 Endpoints públicos

| Método | Rota | Função |
|---|---|---|
| GET | `/store` | Status efetivo, horários, contato, tempo de entrega |
| GET | `/menu` | Categorias, produtos, adicionais e opções ativos |
| GET | `/delivery-zones` | Zonas ativas com taxa |
| GET | `/uploads/{id}` | Imagem enviada pelo painel, por identificador opaco (D-045) |
| POST | `/customers/profile` | Consulta de conveniência por telefone, com opt-in e resposta mínima (D-039) |
| POST | `/orders/quote` | Cotação server-side sem persistir pedido |
| POST | `/orders` | Cria pedido validado no servidor e devolve token de histórico |
| GET | `/orders/history` | Histórico do navegador por token (`Authorization: Bearer`), paginado e sem PII |
| GET | `/orders/{publicId}` | Acompanhamento por identificador opaco |
| WS | `/orders/{publicId}/events` | Mudanças de status por `publicId`, sem PII |

O pedido enviado pelo cliente contém apenas:

```json
{
  "items": [
    { "productId": "...", "quantity": 2, "optionIds": ["..."], "note": "sem cebola" }
  ],
  "fulfillment": "delivery",
  "zoneId": "...",
  "customer": { "name": "...", "phone": "+5585999999999" },
  "address": {
    "street": "...", "number": "...", "complement": "",
    "district": "...", "city": "Aquiraz", "state": "CE", "zipCode": "..."
  },
  "payment": { "method": "cash", "changeForCents": 10000 },
  "saveCustomerProfile": true
}
```

`saveCustomerProfile` é opcional (padrão `true`); `false` desativa o perfil de conveniência no servidor sem alterar o pedido. Com a mesma opção habilitada, `POST /customers/profile` recebe apenas `{ "phone": "..." }` e devolve `{ "found": false }` ou os dados mínimos de conveniência; opt-out devolve `found: false`. Telefone vai no corpo, nunca na URL, e a resposta usa `Cache-Control: no-store`.

Preço, taxa, total e disponibilidade **nunca** vêm do cliente. A resposta inclui `orderNumber`, `publicId`, `status`, `totalCents`, `whatsappUrl` e `accessToken` do pedido já persistido. O token é opaco, vai no cabeçalho `Authorization` e nunca em URL, log ou listagem; só seu hash é gravado no pedido.

`POST /orders/quote` recebe apenas itens, `fulfillment` e `zoneId`, sem dados pessoais. Ele valida catálogo/zona e devolve subtotal, taxa e total sem persistir pedido; o `POST /orders` recalcula tudo antes de gravar.

### 9.3 Endpoints administrativos

| Método | Rota | Função |
|---|---|---|
| POST | `/admin/auth/login` | Login com sessão |
| POST | `/admin/auth/logout` | Encerra sessão |
| GET | `/admin/auth/me` | Usuário e papel |
| GET/POST/PATCH | `/admin/products` | CRUD de produtos |
| GET/POST/PATCH | `/admin/categories` | CRUD de categorias |
| GET/POST/PATCH | `/admin/addons` | CRUD de adicionais |
| GET/POST/PATCH | `/admin/delivery-zones` | CRUD de zonas e taxas |
| POST | `/admin/uploads` | Envia imagem de produto em `multipart/form-data` (D-045) |
| GET/PATCH | `/admin/store` | Status, horários e contato |
| GET | `/admin/orders` | Fila com filtro por status e data |
| GET | `/admin/orders/{id}` | Detalhe, com `customerProfile` quando houver perfil reconhecido |
| PATCH | `/admin/orders/{id}/status` | Transição de status |
| WS | `/admin/orders/events` | Criações e mudanças de status; exige sessão |
| GET/POST/PATCH | `/admin/users` | Usuários admin; somente `owner` |
| GET | `/admin/audit-logs` | Auditoria de mutações; somente `owner` |

Catálogo não é apagado fisicamente: use `active: false`.

**Transições de status:** `pending → confirmed | cancelled`; `confirmed → preparing | cancelled`; `preparing → ready | cancelled`; `ready → out_for_delivery` (entrega) ou `picked_up` (retirada); `out_for_delivery → delivered`; `delivered`, `picked_up` e `cancelled` são finais. Toda transição inválida retorna `409`.

### 9.4 Papéis

- `owner`: tudo, incluindo usuários admin, configurações da loja e auditoria.
- `operator`: catálogo, zonas, pedidos e status da loja.
- Autorização é verificada no servidor em toda rota admin. O frontend só esconde o que o papel não pode fazer.

---

## 10. Modelo de dados (MongoDB)

**Banco:** `brocados_dev`, `brocados_test`, `brocados` (produção). Nunca compartilhe banco entre ambientes.

| Coleção | Papel | Índices principais |
|---|---|---|
| `store_settings` | Documento único `_id: "main"` | — |
| `categories` | Categorias do cardápio | `slug` único; `active + sortOrder` |
| `products` | Produtos e opções obrigatórias | `slug` único; `categoryId + active + sortOrder` |
| `addons` | Adicionais reutilizáveis | `slug` único; `active` |
| `delivery_zones` | Bairros e taxas | `slug` único; `active + name` |
| `uploads` | Imagens enviadas pelo painel (D-045) | `createdAt` |
| `orders` | Pedidos com snapshot | `number` único; `publicId` único; `idempotencyKey` único sparse; `status + createdAt`; `createdAt`; `customer.phone + createdAt` |
| `customer_profiles` | Perfil de conveniência por telefone (D-038) | `phoneKey` único; `expiresAt` TTL; `lastOrderAt` |
| `admin_users` | Usuários do painel | `email` único |
| `sessions` | Sessões do admin | `tokenHash` único; `expiresAt` TTL; `userId` |
| `audit_logs` | Auditoria de mutações | `createdAt`; `actorId + createdAt` |
| `counters` | Sequência do número do pedido | `_id` do contador |

Campos canônicos:

- **Produto:** `name`, `slug`, `description`, `priceCents`, `categoryId`, `imageUrl`, `imageAlt`, `allergens[]`, `available`, `featured`, `active`, `sortOrder`, `optionGroups[]`, `addonIds[]`, `createdAt`, `updatedAt`.
- **Zona:** `name`, `slug`, `feeCents`, `active`, `sortOrder`, `createdAt`, `updatedAt`.
- **Imagem enviada:** `contentType`, `sizeBytes`, `data`, `createdAt`. Tipos permitidos: `image/jpeg`, `image/png`, `image/webp`, `image/avif`; teto de 2 MiB.
- **Pedido:** `number`, `publicId`, `status`, `fulfillment`, `customer{name,phone}`, `address`, `zone{id,name,feeCents}`, `items[]`, `subtotalCents`, `deliveryFeeCents`, `totalCents`, `payment{method,changeForCents}`, `idempotencyKey`, `requestHash`, `accessTokenHash` (opcional), `statusHistory[]`, `createdAt`, `updatedAt`, `confirmedAt`, `cancelledAt`, `cancelReason`. `accessToken` existe só na resposta e no navegador.
- **Item do pedido:** `productId`, `name`, `unitPriceCents`, `quantity`, `options[]` (nome e preço congelados), `note`, `lineTotalCents`.
- **Perfil de cliente:** `phoneKey` (normalizado), `phone`, `name`, `address`, `fulfillment`, `zone`, `paymentMethod`, `orderCount`, `lastOrderPublicId`, `lastOrderAt`, `optedOut`, `createdAt`, `updatedAt`, `expiresAt`. Sem `note` ou `changeForCents`.
- **Usuário admin:** `name`, `email` (normalizado), `passwordHash`, `role`, `active`, `lastLoginAt`, `createdAt`, `updatedAt`.
- **Auditoria:** `actorId`, `action`, `entity`, `entityId`, `before`, `after`, `createdAt`. Sem senha, token ou dado pessoal desnecessário.

Regras:

- Todo documento mutável tem `createdAt`/`updatedAt` UTC.
- `publicId` é aleatório e não sequencial (128 bits, base64url).
- `number` é sequencial e exibido como `#1042`; nunca use para acesso público.
- Pedido guarda snapshot; não faça `lookup` de preço atual para renderizar histórico.
- `store_settings` guarda `name`, `isOpen`, `openingHours`, `deliveryTime`, `whatsapp`, `instagram`, `address`, `createdAt`, `updatedAt`.
- Horário é avaliado em `America/Fortaleza`; status efetivo = horário permite **e** `isOpen` ligado.
- `products.imageUrl` aceita URL absoluta ou caminho relativo (`/products/...` do seed ou `/api/v1/uploads/{id}` dos uploads); o tipo servido vem da assinatura do arquivo, nunca do `Content-Type` do cliente (D-045).
- Pagamento V1: `pix`, `card`, `cash` — todos pagos na entrega/retirada, sem gateway online. Nenhum dado de cartão é armazenado.
- Perfil de cliente é apenas conveniência: escrita best-effort após o pedido gravado; a leitura pública restrita a `POST /customers/profile` devolve só o necessário para preencher o checkout e nunca histórico, observação, troco ou taxa; `optedOut` mantém tombstone mínimo, devolve `found: false` e impede recriação automática (D-038, D-039).
- Migrações de dados são scripts versionados em `backend/scripts`; não altere documento em produção por comando manual sem registro.

---

## 11. Testes e qualidade

- Backend: testes de tabela (table-driven) para service, `httptest` para handlers e testcontainers para repositório. Rode `go test -race ./...`.
- Scripts Python: `unittest` para os validadores de documentação, design system, proxies de tempo real e Sentry, executado por `make test-docs`, `make test-design`, `make test-config`, `make test-sentry` e pelo hook de commit.
- Testes de integração do repositório: pulam por padrão; use `cd backend && MONGO_TEST_URI=mongodb://localhost:27017 go test -race ./internal/repository -v` com um Mongo real e descartável. O job Backend do CI sobe MongoDB 8 e define a variável.
- Frontend: Vitest + Vue Test Utils para componentes e stores; Playwright para o fluxo crítico.
- Fluxo E2E mínimo: cardápio → personalizar → carrinho → checkout → pedido persistido → pedido aparece no admin → mudança de status.
- Não use snapshot como única prova de UI crítica.
- Não mocke MongoDB em teste de repositório. Use instância isolada e descartável.
- Teste os casos de borda: loja fechada, produto esgotado, zona inativa, retirada, troco, pedido duplicado por idempotência, transição inválida de status.
- Cobertura não é meta; fluxos críticos cobertos são.
- Lint: `golangci-lint` no backend; ESLint + Prettier + `vue-tsc --noEmit` no frontend.
- CI deve rodar fmt, lint, typecheck, testes e build em todo PR. Provedor de CI ainda em aberto.

---

## 12. Git, commits e PRs

- O projeto é versionado com Git. A branch principal é `main`; todo trabalho entra por branch curta e PR.
- Ative o gate de commit uma vez por clone: `git config core.hooksPath .githooks`.
- Branches curtas: `feat/cardapio`, `fix/total-pedido`, `chore/ci`.
- **Conventional Commits em português**, com escopo quando útil:
  - `feat(api): cria endpoint de pedido com idempotência`
  - `fix(web): corrige total após troca de zona`
  - `docs: registra decisão de autenticação`
- Tipos: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `build`, `ci`.
- Um PR = uma mudança lógica. Evite diff gigante e mistura de assuntos.
- Descrição do PR: contexto, o que mudou, como testar, riscos e evidência (comando/print) quando for UI.
- Nunca commite `.env`, segredo, `node_modules/`, `dist/`, cobertura, binário ou dump do Mongo.
- Não force push em `main`; não commite direto em `main` após a Fase 1.

### 12.1 Gate de commit (obrigatório)

Nenhum commit é criado sem passar por esta sequência:

1. **Documentação:** atualize o documento afetado — `AGENTS.md` para decisão de arquitetura/processo, `README.md` para estrutura ou uso, `docs/` para comportamento e contrato, `design-system/README.md` para padrão visual. Links relativos precisam resolver.
2. **Testes:** todo arquivo novo de código (endpoint, componente, store, script, migração) nasce com teste automatizado correspondente. Todo comportamento novo ou corrigido exige teste. Correção de bug sem teste de regressão não é aceita, salvo impossibilidade técnica justificada no PR. **Todos os testes, novos e existentes, precisam passar antes do commit.**
3. **Validação:** rode `make fmt` (ou `make fmt-check`), `make lint`, `make test`, `make docs-check`, `make design-check`, `make config-check` e `make sentry-check`. Tudo precisa passar antes do commit.
4. **Inspeção:** revise `git diff --cached` e confirme que não há segredo, PII, arquivo gerado, cor/fonte fora do design system ou mudança fora do escopo.
5. **Commit:** use Conventional Commits em português. Não use `--no-verify` para contornar o hook.

O hook versionado em `.githooks/pre-commit` executa as verificações mínimas (segredos, whitespace, design system, links de documentação, validação do Sentry e, quando o `Makefile` existir, `make lint` e `make test`). Se o commit for apenas de documentação, os links e a estrutura ainda são validados.

---

## 13. Definição de pronto

Uma entrega só termina quando:

- [ ] O comportamento pedido funciona em mobile (375 px) e desktop (1280 px).
- [ ] Nenhum arquivo de código novo ficou sem teste correspondente.
- [ ] Teste automatizado novo ou alterado cobre o comportamento e passa localmente.
- [ ] Preço, taxa e total são calculados e validados no servidor.
- [ ] O pedido é persistido antes do handoff e sobrevive a recarregar/fechar a aba.
- [ ] Loja fechada, item esgotado, zona inativa e transição inválida têm tratamento explícito.
- [ ] Rotas admin exigem sessão válida e autorização por papel.
- [ ] Nenhum dado pessoal aparece em log, analytics ou mensagem de erro.
- [ ] `make fmt`, `make lint`, `make test`, `make docs-check`, `make design-check`, `make config-check` e `make sentry-check` passam (ou a exceção está justificada).
- [ ] Contrato de API e documentação foram atualizados, com links válidos.
- [ ] Auditoria registra mutações administrativas relevantes.
- [ ] Acessibilidade básica verificada: teclado, foco, rótulos, contraste.
- [ ] UI validada contra `design-system/README.md`, sem cor, fonte ou espaçamento fora do padrão.

---

## 14. O que não fazer

- Não publique o HTML do protótipo como produto.
- Não use Tailwind via CDN, Google Fonts via CDN ou dados em memória como persistência.
- Não calcule total, taxa ou disponibilidade só no frontend.
- Não confie em preço, total ou status enviados pelo cliente.
- Não use `float` para dinheiro nem `parseFloat` para preço.
- Não exponha `_id`, ID sequencial de pedido ou listagem pública de pedidos.
- Não crie rota admin sem autenticação e autorização.
- Não logue telefone, endereço, nome, observação, cookie ou token.
- Não armazene dados de cartão.
- Não invente cardápio, preços, bairros, horários ou contatos.
- Não crie cor, fonte, espaçamento, raio, sombra ou componente fora do padrão documentado no design system.
- Não edite `design-system/brocados-lanches-design-system-v2.5.html` nem os tokens sem aprovação de design; mudança de token exige atualizar `tokens.css`, `tokens.json` e documentação.
- Não desative, ignore ou contorne `make design-check`.
- Não troque stack, banco ou gerenciador de pacotes por preferência pessoal.
- Não adicione dependência sem necessidade comprovada.
- Não use `any` no TypeScript.
- Não faça `find` sem limite nem query com campo vindo cru do cliente.
- Não rode seed ou teste contra banco de produção.
- Não commite sem documentação e teste correspondentes.
- Não crie código sem teste nem commite com teste falhando.
- Não commite com `--no-verify` nem desative o hook para “resolver” uma falha.

---

## 15. Decisões ainda abertas

| Decisão | Premissa atual da V1 |
|---|---|
| Handoff do WhatsApp | Link `wa.me` gerado após persistir o pedido; API oficial só se contratada |
| Pagamento online | Não; pagamento na entrega/retirada (Pix, cartão, dinheiro) |
| Hospedagem e domínio | Em aberto; API e web same-origin atrás de proxy reverso |
| Catálogo real, fotos, logo e alergênicos | Aguardando cliente; seed é fictício |
| Horários, feriados e tempo de entrega | Aguardando cliente; avaliar em `America/Fortaleza` |
| MFA e gestão de usuários | Sessão + papel na V1; MFA desejável antes do lançamento |
| Impressão, KDS e som de pedido | Fora da V1 até confirmação |
| Retenção LGPD e política de privacidade | Proposta de 24 meses para pedidos; validar com jurídico |
| Analytics e metas de conversão | Sem ferramenta definida; nada com PII |
| SEO orgânico | SPA na V1; reavaliar Nuxt se virar prioridade |

Toda mudança nessas premissas deve atualizar este arquivo e o briefing.

---

## 16. Referências

- `README.md` — visão geral, como rodar localmente e via Docker, limites do protótipo e do esqueleto.
- `compose.yaml` — MongoDB, API e web no ambiente integrado.
- `.env.example` — variáveis de ambiente do desenvolvimento local.
- `briefing/chamado.md` — escopo, pendências e dados a confirmar.
- `docs/01-inventario-prototipo.md` — o que o protótipo faz e não faz.
- `docs/02-plano-de-producao.md` — plano histórico da Fase 1.
- `docs/README.md` — índice da documentação completa.
- `docs/03-prd.md` — produto, personas, escopo e requisitos.
- `docs/04-arquitetura.md` — arquitetura, fluxos e segurança.
- `docs/05-modelo-de-dados.md` — coleções, campos, índices e retenção.
- `docs/06-contrato-api.md` — contrato HTTP detalhado.
- `docs/07-design-system-e-experiencia.md` — telas, componentes e acessibilidade.
- `docs/08-backlog.md` — épicos, histórias, tarefas e aceite.
- `docs/09-qualidade-e-testes.md` — estratégia e matriz de testes.
- `docs/10-operacao-e-lancamento.md` — deploy, backup, monitoramento e LGPD.
- `docs/11-roadmap.md` — fases, marcos e caminho crítico.
- `docs/12-referencia-ola-click.md` — snapshot funcional do site de referência.
- `docs/13-decisoes-e-pendencias.md` — decisões e perguntas abertas.
- `docs/14-auditoria-design-system.md` — auditoria de conformidade visual de 25/11/2026 e tarefas do Épico 12.
- `docs/adr/README.md` — decisões arquiteturais, incluindo a ADR 0007 de monitoramento de erros.
- `design-system/tokens.css` — fonte dos tokens de marca.
- `design-system/tokens.json` — tokens para ferramentas.
- `design-system/seed.json` — dados fictícios para desenvolvimento.
- `design-system/README.md` — regra visual e validação obrigatória do design system.
- `design-system/brocados-lanches-design-system-v2.5.html` — protótipo visual; somente leitura.
