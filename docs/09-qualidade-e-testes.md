# Qualidade e Testes — Brocados Lanches

**Versão:** 1.1  
**Atualizado em:** 23/11/2026  
**Status:** estratégia da V1 definida; seed e E2E implementados em 20/11/2026; testes de integração do catálogo no CI e testes de service/handler da API pública em 21/11/2026; testes de home, cardápio, personalização, autenticação, shell/guard, fila e transições do admin em 22/11/2026; testes de CRUD de catálogo e configuração da loja no painel em 23/11/2026.  
**Relacionados:** `03-prd.md`, `04-arquitetura.md`, `06-contrato-api.md`, `08-backlog.md`, `10-operacao-e-lancamento.md`.

---

## 1. Princípios

1. **Teste não é opcional.** Todo comportamento novo ou corrigido tem teste automatizado correspondente.
2. **Fluxos críticos primeiro.** Cobertura não é meta; pedido, preço, disponibilidade, autenticação e transição de status são.
3. **Servidor é a verdade.** O teste precisa provar que o cliente não consegue fraudar preço, taxa, total ou status.
4. **Mongo real em repositório.** Sem mock de MongoDB em teste de repositório.
5. **Sem PII em evidência.** Logs, prints e fixtures de teste não podem conter dados reais.
6. **Teste determinístico.** Sem dependência de horário real, rede externa ou ordem de execução.
7. **Regressão tem nome.** Todo bug corrigido ganha teste que falharia antes da correção.
8. **Evidência no PR.** Comando, resultado e, quando UI, print ou vídeo curto.

---

## 2. Níveis de teste

| Nível | Ferramenta | Escopo | Quando |
|---|---|---|---|
| Domínio/service | `testing` table-driven | Regras puras, cálculo, status, validação | Toda regra nova |
| Handler | `httptest` | JSON, status HTTP, envelope de erro, auth | Todo endpoint |
| Repositório | `testing` + testcontainers | Queries, índices, TTL, concorrência | Toda query/índice |
| Frontend unidade | Vitest | Stores, composables, services, utilitários | Toda lógica de UI |
| Frontend componente | Vitest + Vue Test Utils | Estados, eventos, acessibilidade básica | Todo componente relevante |
| E2E | Playwright | Fluxo completo em navegador | Todo fluxo crítico |
| Segurança | testes de API + revisão | Auth, CSRF, rate limit, enumeração, PII | Antes do go-live |
| Acessibilidade | axe + manual | Teclado, foco, contraste, rótulos | Telas críticas |
| Performance | Lighthouse/k6 ou equivalente | Orçamentos de carregamento e API | Antes do go-live |

---

## 3. Ambientes e dados

| Ambiente | Banco | Uso | Regra |
|---|---|---|---|
| Local | `brocados_dev` | Desenvolvimento | Dados fictícios; nunca produção |
| Teste | `brocados_test` | Testes automatizados | Limpo antes/depois; descartável |
| Preview | banco isolado | Validação de PR | Sem dados reais |
| Produção | `brocados` | Operação | Nunca usar em teste |

Regras:

- Seed de desenvolvimento é fictício e marcado como tal.
- Teste de integração usa Mongo real e descartável; o teste atual pula se `MONGO_TEST_URI` não estiver definido.
- Não usar telefone, endereço ou nome reais em fixture.
- Relógio é injetável (`platform/clock`) para testar horário da loja.
- IDs e `publicId` são gerados por interface injetável para testes determinísticos quando necessário.
- Nenhum teste depende do site de referência ou da internet.

### 3.1 Comandos

```bash
make seed              # aplica design-system/seed.json em brocados_dev
make seed-reset        # esvazia as coleções semeadas e reaplica
make seed-check        # valida o seed sem acessar o MongoDB
make test              # docs, design, seed, Go e Vitest
make test-e2e          # Playwright com API isolada em brocados_test
```

- `make seed` e `make seed-reset` recusam `APP_ENV=production`, o banco `brocados` e nomes sem sufixo `_dev` ou `_test`.
- `make test-e2e` exige MongoDB acessível em `localhost:27017` (ou `MONGO_URI` definido) e sobe a API na porta 8080 com `brocados_test`.
- O E2E não faz parte de `make test`; o gate de commit continua rápido e sem navegador.
- O smoke atual cobre home, health, readiness, alvo de toque em 375 px e navegação da home para o cardápio.
- O job Backend do CI sobe MongoDB 8 e define `MONGO_TEST_URI`; localmente, sem essa variável, os testes de integração são pulados.

---

## 4. Matriz de testes por área

| Área | Testes obrigatórios | Histórias |
|---|---|---|
| Loja e horário | Aberto/fechado por horário, kill switch, virada de meia-noite, fuso | US-002, US-021, US-065 |
| Catálogo | Ativo/inativo, disponível/esgotado, categoria vazia, adicional não vinculado | US-020–US-024 |
| Opções | Obrigatória, mínimo, máximo, opção indisponível, preço adicional | US-024 |
| Carrinho | Persistência, versão, quantidade zero, remoção, clear | US-030, US-031 |
| Cotação | Delivery, retirada, zona inativa, item esgotado, troco | US-032 |
| Pedido | Persistência, snapshot, número, `publicId`, `201`/`Location` | US-040 |
| Idempotência | Mesmo corpo, corpo diferente, concorrência, chave ausente | US-041 |
| WhatsApp | URL, falha de abertura, reenvio, pedido preservado | US-042, US-044 |
| Acompanhamento | ID válido, inválido, status final, sem PII | US-050, US-051 |
| Auth | Login, logout, me, sessão expirada, senha errada, rate limit | US-060 |
| Papéis | Operator vs owner, 401 vs 403, último owner | US-061, US-066 |
| Catálogo admin (API) | Criar/editar/desativar, centavos, slug único, paginação, allowlist de ordenação, 401/403/CSRF, auditoria | US-022, US-062–US-064 |
| Catálogo admin (painel) | Lista/filtro/paginação, formulários, slug, reais → centavos, opções, adicionais, estados de carregamento/vazio/erro | US-062–US-064 |
| Loja admin (API e painel) | Horários válidos/não sobrepostos, E.164, kill switch, PATCH parcial, auditoria sem PII; formulário com validação por campo | US-065 |
| Pedidos admin | Filtros, paginação, detalhe, PII mascarada; fila e detalhe no painel com estados de carregamento/vazio/erro | US-070, US-071 |
| Status | Todas as transições válidas e inválidas, motivo, concorrência; ações contextuais e confirmação no painel | US-072 |
| Auditoria | Mutação gera log, owner lê, sem PII | US-073 |
| SEO | Título, descrição, JSON-LD, sitemap | US-081 |
| A11y | Teclado, foco, rótulos, contraste, 44 px, reduced motion | US-082 |
| LGPD | PII fora de log, retenção, acompanhamento mínimo | US-080, US-093 |

---

## 5. Cenários críticos

Estes cenários precisam existir em teste automatizado e, quando possível, em E2E.

### 5.1 Loja fechada por horário

- **Dado** que o relógio está fora do horário de funcionamento.
- **Quando** o cliente tenta cotar ou criar pedido.
- **Então** recebe `409 STORE_CLOSED`; nenhum pedido é criado; a UI explica e preserva o carrinho.

### 5.2 Loja fechada pelo kill switch

- **Dado** que `store_settings.isOpen = false` dentro do horário.
- **Quando** o cliente tenta pedir.
- **Então** recebe `409 STORE_CLOSED`; o admin vê o status fechado; a mudança é auditada.

### 5.3 Produto esgotado

- **Dado** produto ativo com `available: false`.
- **Quando** o cardápio carrega e o cliente tenta adicionar.
- **Então** aparece “Esgotado”, a ação é bloqueada e o servidor rejeita `POST /orders` com `409 PRODUCT_UNAVAILABLE`.

### 5.4 Produto inativo

- **Dado** produto com `active: false`.
- **Quando** o cardápio carrega.
- **Então** ele não aparece e não pode ser usado nem por ID direto.

### 5.5 Adicional inválido

- **Dado** adicional inativo ou não vinculado ao produto.
- **Quando** o cliente envia `optionIds`.
- **Então** `422 VALIDATION_ERROR`; nenhum pedido é criado.

### 5.6 Opção obrigatória ausente

- **Dado** grupo obrigatório com `minSelections: 1`.
- **Quando** o cliente envia sem opção.
- **Então** `422 VALIDATION_ERROR` com detalhe do grupo.

### 5.7 Retirada com taxa zero

- **Dado** carrinho válido.
- **Quando** o cliente escolhe retirada.
- **Então** `deliveryFeeCents = 0`, `zone = null`, `address` não é exigido e o total é subtotal.

### 5.8 Delivery com zona inativa

- **Dado** zona com `active: false`.
- **Quando** o cliente tenta cotar/pedir.
- **Então** `409 ZONE_UNAVAILABLE`; a zona não aparece na lista.

### 5.9 Troco

- **Dado** pagamento em dinheiro.
- **Quando** `changeForCents < totalCents`.
- **Então** `422 VALIDATION_ERROR`.
- **Quando** `changeForCents >= totalCents`.
- **Então** pedido válido e valor de troco exibido corretamente.

### 5.10 Preço adulterado no cliente

- **Dado** que o cliente altera o preço no carrinho/localStorage.
- **Quando** envia o pedido.
- **Então** o servidor ignora o preço enviado e calcula o valor do catálogo; o total persistido é o do servidor.

### 5.11 Pedido duplicado por idempotência

- **Dado** a mesma `Idempotency-Key` e o mesmo corpo.
- **Quando** a requisição é repetida.
- **Então** o mesmo pedido é retornado, sem duplicata, com `200` e `Idempotent-Replayed: true`.

### 5.12 Mesma chave, corpo diferente

- **Dado** a mesma chave com outro conteúdo.
- **Quando** a requisição é repetida.
- **Então** `409 CONFLICT`; o pedido original não muda.

### 5.13 Falha de handoff

- **Dado** pedido persistido e link do WhatsApp bloqueado.
- **Quando** o cliente volta à confirmação.
- **Então** o pedido continua acessível, o link pode ser copiado/reenviado na confirmação imediata, e o carrinho já foi limpo. Após recarregar, o cliente usa o contato da loja e informa o número do pedido.

### 5.14 Transição inválida

- **Dado** pedido `pending`.
- **Quando** o operador tenta `delivered`.
- **Então** `409 CONFLICT`; status e auditoria não mudam.

### 5.15 Acesso indevido ao admin

- **Dado** visitante sem sessão.
- **Quando** acessa rota admin ou API admin.
- **Então** `401`; nenhuma escrita ocorre e o frontend redireciona para `/admin/login` com `redirect`.
- **Dado** operador tentando rota de owner.
- **Então** `403` na API; no frontend o guard leva a `/admin/sem-acesso`.
- Coberto por testes de guard/store e pelo E2E anônimo (US-060 e US-061).

### 5.16 PII em log

- **Dado** checkout com nome, telefone, endereço e observação.
- **Quando** a requisição é processada, inclusive em erro.
- **Então** os logs não contêm esses valores, cookie ou token.

### 5.17 Fuso da operação

- **Dado** um horário limite em `America/Fortaleza`.
- **Quando** o servidor calcula status.
- **Então** usa o fuso correto, independentemente do fuso do host.

### 5.18 Enumeração de pedido

- **Dado** `publicId` inválido ou de outro pedido.
- **Quando** o cliente consulta.
- **Então** `404` genérico, sem revelar existência ou dados.

---

## 6. Backend

### 6.1 Domínio e service

- Table-driven para cálculo de item, subtotal, taxa, total e troco.
- Testar valores limite: 0, 1 centavo, quantidade máxima, opção grátis, adicional caro.
- Testar loja aberta/fechada com relógio fake.
- Testar matriz de status completa.
- Testar validação de opções e adicionais.
- Não usar `float` em nenhum teste.

### 6.2 Handlers

- `httptest` para cada endpoint.
- Verificar status, envelope, `Content-Type`, `Location`, `Set-Cookie`.
- Verificar rejeição de campo desconhecido e corpo > 1 MB.
- Verificar erros de validação por campo.
- Verificar que 5xx não vaza detalhe.
- Verificar rate limit e `Retry-After`.

### 6.3 Repositório

- Testcontainers com Mongo real.
- Criar índices no setup e testar unicidade.
- Testar TTL de sessão (quando suportado no ambiente).
- Testar concorrência de `findOneAndUpdate` do contador.
- Testar `publicId` e `idempotencyKey` duplicados.
- Testar queries de listagem e paginação.
- Nunca mockar MongoDB.

### 6.4 Execução

```bash
cd backend
go test -race ./...
```

Integração:

```bash
cd backend
MONGO_TEST_URI=mongodb://localhost:27017 go test -race ./internal/repository -v
```

---

## 7. Frontend

### 7.1 Stores e fluxos

- Catálogo: carregar uma vez, forçar recarga, erro, retry, produto/categoria por slug e somente destaques disponíveis.
- Home: estados de carregando, vazio, erro e sucesso; navegação para o cardápio com categoria.
- Cardápio: busca sem acento/caixa, filtro por categoria, contagem, busca sem resultado e produto esgotado.
- Personalização: obrigatoriedade, mínimo/máximo, adicionais, observação, preço estimado, esgotado e produto não encontrado.
- Carrinho: adicionar, somar quantidade, remover, limpar, persistir, migrar versão.
- Checkout: delivery/retirada, zona, troco, cotação, erro.
- Admin auth: login, logout, 401, papel, bootstrap de sessão único, recarga forçada e mensagens de erro por código.
- Guard do admin: visitante redirecionado com `redirect`, operador bloqueado em rota de owner, owner liberado, login redirecionado quando já autenticado.
- Shell do admin: menu por papel, atalhos mobile e logout.
- Filtros e paginação.
- Catálogo administrativo: carregar, filtrar, paginar, criar/editar e desativar categorias, adicionais, zonas e produtos; erro de sessão, conflito de slug, validação por campo e rejeição de valor monetário inválido (não convertido em zero).
- Configuração da loja: carregar, validar horários/contato, salvar, sessão expirada e erro de permissão.

### 7.2 Componentes

- `BaseButton`: loading, disabled, teclado.
- `QuantityStepper`: limites e acessibilidade.
- `OptionGroup`: obrigatório, mínimo, máximo.
- `ProductCard`: esgotado, preço, imagem.
- `CartLine`: opções, remoção, quantidade.
- `StatusBadge`: texto além de cor.
- `DataTable`: vazio, erro, paginação (ainda não abstraído).
- `AdminCatalogFilters`: busca, situação e limpeza.
- `AdminCatalogList`: dados, ações, paginação, erro e vazio.
- `AdminCatalogForm`: slug gerado, validação, edição e conversão de reais para centavos.
- `AdminProductForm`: grupos/opções, adicionais, alergênicos e validação.
- `AdminStoreView`: horários, dias da semana, validação de WhatsApp e salvamento.
- `Modal`/`Drawer`: foco, ESC, retorno de foco.

### 7.3 Services

- Cliente HTTP trata envelope de erro e preserva `details` por campo.
- Timeout e erro de rede viram mensagem em português.
- `credentials: 'include'`.
- Nenhum componente chama `fetch`.

### 7.4 Execução

```bash
cd frontend
pnpm test
pnpm typecheck
pnpm lint
```

---

## 8. E2E

### 8.1 Fluxo mínimo

1. Abrir cardápio.
2. Abrir produto.
3. Escolher opção obrigatória e adicional.
4. Adicionar ao carrinho.
5. Abrir checkout.
6. Escolher delivery e zona.
7. Preencher nome, telefone e endereço.
8. Escolher dinheiro e troco.
9. Ver total cotado.
10. Enviar pedido.
11. Ver confirmação e link de WhatsApp.
12. Abrir admin, autenticar.
13. Ver pedido na fila.
14. Mudar status para `confirmed` e `preparing`.
15. Abrir acompanhamento público e ver status.

### 8.2 Variações

- Retirada.
- Loja fechada.
- Produto esgotado.
- Zona inativa.
- Troco inválido.
- Duplo clique no envio.
- Sessão expirada no admin.
- Transição inválida.
- Mobile 375 px e desktop 1280 px.

### 8.3 Regras

- Playwright sobe ambiente isolado com banco de teste.
- Sem dependência de produção.
- Evidência: vídeo/trace no CI em caso de falha.
- Teste não depende de ordem; cada um prepara seu estado.

### 8.4 Cobertura implementada

- Smoke da home, navegação para o cardápio e health/readiness pelo proxy.
- Acesso anônimo a rota protegida redirecionado para `/admin/login` com `redirect`, e credenciais inválidas exibindo mensagem genérica (US-061).
- Pedido ponta a ponta: cardápio → personalização → carrinho → checkout → confirmação → fila → detalhe → `Confirmar pedido`, em desktop e mobile (US-040 a US-043, US-070 a US-072).
- CRUD de catálogo e configuração da loja cobertos por Vitest; E2E dessas telas ainda pendente.

---

## 9. Segurança e privacidade

| Teste | Critério |
|---|---|
| Auth | Rota admin sem sessão retorna 401 |
| Papel | Operator não acessa owner; 403 |
| CSRF | POST com Origin inválido é rejeitado |
| Rate limit | Login e pedido retornam 429 após limite |
| Enumeração | Login não revela e-mail; pedido não revela ID |
| PII em log | Nome/telefone/endereço/nota ausentes |
| Cookie | `HttpOnly`, `Secure` em produção, `SameSite=Lax` |
| Senha | Argon2id; nunca em resposta/log |
| Sessão | Expira, rotaciona e logout invalida |
| Card | Nenhum campo de cartão é recebido |
| Headers | CSP, nosniff, referrer, permissions |
| Dependências | Scanner sem vulnerabilidade crítica conhecida |

---

## 10. Acessibilidade

- axe-core nas telas: home, cardápio, produto, carrinho, checkout, confirmação, admin login, admin pedidos, admin catálogo e admin loja.
- Navegação apenas por teclado no fluxo de checkout.
- Foco visível e ordem lógica.
- Modal/drawer com foco preso.
- Rótulos e erros associados.
- Alvos de 44 px.
- Contraste conforme `07-design-system-e-experiencia.md`.
- `prefers-reduced-motion`.
- Zoom 200% sem perda de conteúdo.
- Teste manual com leitor de tela em pelo menos um fluxo.

Falha grave de acessibilidade bloqueia o lançamento.

---

## 11. Performance

| Alvo | Meta | Medição |
|---|---|---|
| LCP do cardápio | ≤ 2,5 s em 4G | Lighthouse/WebPageTest |
| API de leitura p95 | ≤ 300 ms | métricas |
| Criação de pedido p95 | ≤ 800 ms | métricas |
| Erros 5xx | < 0,5% | monitoramento |
| Peso inicial | orçamento definido após fotos | bundle analyzer |
| Imagens | WebP/AVIF, dimensões e lazy | auditoria |

Teste de carga básico:

- 50 requisições concorrentes em `GET /menu`;
- 10 pedidos concorrentes com chaves diferentes;
- 2 pedidos concorrentes com a mesma chave;
- medir latência, erros e conexões.

Não é teste de stress; é verificação de comportamento sob carga inicial.

---

## 12. Gates de qualidade

Antes de merge:

```bash
make fmt
make lint
make test
make docs-check
make design-check
make build
```

Antes de lançamento:

- [ ] Todos os cenários críticos cobertos.
- [ ] E2E mobile e desktop passando.
- [ ] axe sem erro grave.
- [ ] Performance dentro do orçamento.
- [ ] Segurança revisada.
- [ ] Backup/restore testado.
- [ ] Nenhuma PII em log.
- [ ] Go-live checklist assinado.

Falha em gate bloqueia merge/lançamento; não usar `--no-verify` nem desativar verificação.

---

## 13. Bug, regressão e evidência

- Bug reproduzível vira teste que falha antes da correção.
- Se não for possível testar, justificar no PR e registrar risco.
- PR inclui:
  - contexto;
  - o que mudou;
  - como testar;
  - riscos;
  - evidência (comando, print, vídeo).
- Teste instável é corrigido ou removido com justificativa; não ignorado em silêncio.
- Snapshot não é prova única de UI crítica.

---

## 14. Rastreabilidade

| História | Teste principal |
|---|---|
| US-023 | Home/cardápio no Vitest + smoke E2E de navegação |
| US-024 | Opções e adicionais no service + componente `OptionGroup` |
| US-032 | Cotação no service + handler + store |
| US-040 | Criação de pedido no service + handler + repository |
| US-041 | Concorrência de idempotência no repository |
| US-042 | Geração de URL + falha de handoff no frontend |
| US-050 | Handler público sem PII |
| US-060 | Auth handler + cookie + rate limit |
| US-062 | `AdminProductForm` + `AdminProductsView` no Vitest; service/store do catálogo |
| US-063 | `AdminCatalogForm` + views de categorias/adicionais no Vitest |
| US-064 | `AdminZonesView` no Vitest; conversão de taxa em centavos |
| US-065 | `AdminStoreView` no Vitest; validação de horários/E.164 + API |
| US-070 | Filtros/paginação no service + repository + handler; PII mascarada |
| US-071 | Detalhe no repository + handler com snapshot completo |
| US-072 | Matriz de status no domínio + CAS/auditoria no repository + handler |
| US-082 | axe + teste de teclado |
| US-093 | Revisão de log + testes de segurança |

---

## 15. Pendências

- Definir ferramenta de CI.
- Definir testcontainers ou Mongo efêmero no CI.
- Definir se Playwright roda em container.
- Definir ferramenta de load test.
- Definir se haverá teste de contrato OpenAPI.
- Definir responsável por teste manual de acessibilidade.
