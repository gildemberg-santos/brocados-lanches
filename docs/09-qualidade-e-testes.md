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

- Seed de desenvolvimento é marcado como tal; categorias, produtos e imagens vêm do catálogo público de referência, contato/horários seguem o status publicado em `/contato`, zonas e taxas foram fornecidas em 25/11/2026, e pedidos são fictícios.
- `seed.Build` valida a configuração da loja com as mesmas regras da API (inclusive `whatsapp` em E.164 e `openingHours`), para que um `PATCH /admin/store` parcial nunca seja bloqueado por dado inválido do próprio seed.
- Com `APP_ENV=test`, `seed.ForEnvironment` força `isOpen: true` e remove `openingHours` para o E2E não depender da hora da execução; o seed de desenvolvimento mantém os horários publicados.
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
- Os testes E2E cobrem smoke (home, health, readiness, alvo de toque em 375 px e navegação para o cardápio), fluxo completo de compra em desktop/mobile, acesso por papel no painel e cenários críticos de falha (loja fechada pelo kill switch, produto esgotado, zona inativa, retirada, troco, idempotência, transição inválida e sessão encerrada).
- Evidência de 25/11/2026: `make test-e2e` executado com MongoDB local e `brocados_test` — 28 testes, 26 aprovados e 2 pulados no mobile. Os testes que criam pedido rodam apenas no desktop porque os dois projetos compartilham o limite de 10 pedidos/min por IP da API. Ainda falta injetar falha de handoff/retry no E2E.
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
| Exclusão manual | 204 após remoção, 404 para ID inválido/removido, 403 para operator, qualquer status, auditoria sem PII, sem webhook, evento só no painel, confirmação e erro amigável na fila | US-077, D-051 |
| Auditoria | Mutação gera log, owner lê, sem PII | US-073 |
| Webhook | URL absoluta com qualquer esquema e host, exigência de URL quando ativo, payload `{number, storeNumber, message}` com os dois destinatários e número da loja ausente sem envio, cabeçalho de evento, replay sem duplicata, falha de transporte sem bloqueio, 403 para operator, auditoria sem URL e formulário owner-only | US-075, D-046 |
| Seleção de eventos do webhook | Chaves válidas/repetidas/desconhecidas, lista vazia, configuração legada com todos os eventos, filtro do notificador por criação/status, cabeçalho inalterado, auditoria sem URL e checkboxes no painel | US-076, D-048 |
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

### 5.19 Falha de webhook

- **Dado** webhook ativo apontando para destino indisponível ou que responde `500`.
- **Quando** o pedido é criado ou o status é transicionado.
- **Então** a operação persiste normalmente e o log registra apenas categoria/status, sem URL, corpo ou PII.

### 5.20 Replay de webhook

- **Dado** um `POST /orders` repetido com a mesma `Idempotency-Key`.
- **Quando** o servidor devolve o pedido gravado (`200` com `Idempotent-Replayed`).
- **Então** nenhum evento de webhook é publicado para esse replay.

### 5.21 Telefone impossível

- **Dado** telefone com DDI duplicado, DDD inexistente ou celular de 9 dígitos começando com `8`.
- **Quando** o cliente tenta finalizar o checkout ou consultar `POST /customers/profile`.
- **Então** o frontend bloqueia com erro no campo e o servidor responde `422 VALIDATION_ERROR`; nenhum pedido é criado e nenhum número inválido é usado como destinatário do webhook.

### 5.22 Webhook com dois destinatários

- **Dado** webhook ativo e loja com `whatsapp` preenchido.
- **Quando** o pedido é criado ou transicionado.
- **Então** o POST traz `number` com o telefone do cliente e `storeNumber` com o da loja, em um único evento, sem PII em log (D-046).
- **Dado** loja sem `whatsapp` preenchido: nenhum POST é enviado, o aviso não contém telefone e o pedido permanece gravado.

### 5.23 Seleção de eventos do webhook (D-048)

- **Dado** webhook ativo com `events: ["order.created", "order.confirmed"]`.
- **Quando** o pedido é criado, transiciona para `confirmed` e depois para `preparing`.
- **Então** apenas os dois primeiros eventos geram POST, com `X-Brocados-Event` igual a `order.created` e `order.updated`, respectivamente.
- **Dado** configuração antiga sem o campo `events`.
- **Quando** qualquer criação ou transição persistida ocorre.
- **Então** o envio continua como antes, e a API devolve a lista completa de eventos para o painel.
- **Dado** `events: []` ou uma chave desconhecida/repetida no `PATCH`.
- **Quando** a configuração é salva e um pedido é criado.
- **Então** a lista vazia não envia nada sem alterar `active`, e a chave inválida é rejeitada com `422 VALIDATION_ERROR` sem gravar nem disparar.

### 5.24 Monitoramento de erros sem PII (D-047)

- **Dado** DSN vazio na API e no build.
- **Quando** a aplicação inicia e uma requisição falha.
- **Então** nenhum SDK é inicializado e nenhum evento tenta sair do processo.
- **Dado** erro de domínio `422` no cliente HTTP.
- **Quando** a requisição termina.
- **Então** o navegador registra apenas breadcrumb com método, rota sanitizada, status e duração arredondada; não há `captureException`.
- **Dado** evento do Sentry com `user`, `request`, tags, breadcrumbs, stacktrace e contexto contendo telefone, e-mail, CEP, query e token.
- **Quando** `BeforeSend` roda no Go ou no Vue.
- **Então** `user`/`request` são removidos, chaves sensíveis são descartadas e os padrões restantes são mascarados; os testes não dependem de rede.
- **Dado** painel com CSP sem `connect-src` do Sentry ou código com tracing/Replay/DSN literal.
- **Quando** `make sentry-check` roda.
- **Então** a validação falha com o motivo e bloqueia o CI.
- **Dado** snippet do GTM sem container, com nonce fixo, com `unsafe-inline` no `script-src` ou política de privacidade que nega analytics.
- **Quando** `make analytics-check` roda.
- **Então** a validação falha com o motivo e bloqueia o CI.
- **Dado** duas respostas HTML do Vite com o placeholder `__CSP_NONCE__`.
- **Quando** o plugin de CSP processa cada uma.
- **Então** cada resposta recebe um nonce diferente no header e no snippet, e nenhuma PII é enviada ao `dataLayer` pelo código da loja.

### 5.26 Exclusão manual de pedido (D-051)

- **Dado** pedido em qualquer status e sessão de `owner`.
- **Quando** `DELETE /admin/orders/{id}` é chamado.
- **Então** o documento é removido, a resposta é `204`, a auditoria registra número/status/forma de entrega/total sem PII e nenhum POST de webhook é feito.
- **Dado** `operator`, ID inválido ou pedido já removido.
- **Quando** a mesma rota é chamada.
- **Então** a API responde respectivamente `403`, `404` e `404`, sem apagar documento algum.
- **Dado** fila aberta no painel.
- **Quando** o `owner` clica em excluir.
- **Então** aparece confirmação explícita; só depois dela o pedido sai da fila, e falhas viram mensagem em português sem dados internos.
- **Dado** assinatura no canal administrativo.
- **Quando** a exclusão é persistida.
- **Então** chega `order.deleted` sem PII; o canal público não recebe o sinal e o acompanhamento passa a devolver `404`.

### 5.27 Notificação visual e sonora de status (D-052)

- **Dado** acompanhamento público, fila ou detalhe do painel com pedido já carregado.
- **Quando** chega `order.updated` com status diferente do conhecido.
- **Então** aparece `StatusToast` com número e novo status, o som é tocado uma vez e o refetch HTTP atualiza a tela.
- **Dado** mensagem `connected`, reconexão, carga inicial, pedido de outra tela ou evento repetido.
- **Quando** o cliente processa o sinal.
- **Então** nenhum toast nem som é disparado.
- **Dado** operador que confirma uma transição no detalhe.
- **Quando** a API responde `200` com o novo status.
- **Então** o toast e o som aparecem imediatamente, e o `order.updated` seguinte não repete o aviso.
- **Dado** navegador sem Web Audio ou com autoplay bloqueado.
- **Quando** o som é solicitado.
- **Então** a chamada falha silenciosamente e a notificação visual permanece.

### 5.28 Pix cobrado após gravar o pedido (D-053)

- **Dado** Pix ativo, cotação válida e forma de pagamento `pix`.
- **Quando** o cliente escolhe Pix no checkout.
- **Então** o checkout não gera BR Code; ele informa que o pedido será gravado primeiro e o QR aparecerá na confirmação.
- **Quando** `POST /orders` responde `201`.
- **Então** a confirmação `/pedido/:publicId` monta localmente BR Code com `54` igual ao `totalCents` persistido, `62.05=***` e CRC16-CCITT/FFFF, exibindo QR em PNG e o Pix Copia e Cola, sem gateway ou confirmação automática.
- **Dado** pedido gravado com Pix e configuração ativa.
- **Quando** a confirmação é recarregada.
- **Então** `GET /orders/{publicId}` devolve `paymentMethod: "pix"` e o BR Code é refeito com o mesmo total gravado.
- **Dado** chave, nome ou cidade ausentes/inválidos, configuração desativada ou falha ao carregar `GET /store`.
- **Quando** a confirmação tenta exibir a cobrança de um pedido com Pix.
- **Então** nenhum BR Code aparece e a mensagem orienta a falar com a loja; o pedido gravado não é alterado.
- **Dado** `operator` ou chamada pública à configuração do Pix.
- **Quando** tenta ler ou alterar `GET/PATCH /admin/integrations/pix`.
- **Então** recebe `403`/`401`; a auditoria de alterações registra apenas `active` e se os campos estão preenchidos, nunca a chave, o nome ou a cidade.
- **Dado** integração Pix desativada no servidor e pedido enviado com `payment.method = "pix"`.
- **Quando** uma aba antiga ou chamada direta à API tenta criar o pedido.
- **Então** a API responde `422 VALIDATION_ERROR` com `details[0].field = "payment.method"` sem persistir; o checkout orienta a escolher outra forma de pagamento.
- **Dado** total cotado exibido e alteração de preço ou taxa no servidor antes do envio.
- **Quando** o cliente envia o pedido com `expectedTotalCents` divergente do total recalculado.
- **Então** a API responde `409 CONFLICT` sem persistir, o checkout recota e mantém o carrinho para nova tentativa; o BR Code ainda não existe nessa etapa.

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
- `StatusToast`: `aria-live`, mensagem, fechamento manual e expiração em 4 s.
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

- Retirada — coberta em `failure-scenarios.spec.ts`.
- Loja fechada — coberta pelo kill switch, com cotação, UI do checkout e restauração do estado.
- Produto esgotado — coberto.
- Zona inativa — coberto.
- Troco — coberto no pedido; troco inválido ainda pendente.
- Idempotência — replay com a mesma chave e conflito com corpo diferente cobertos.
- Duplo clique no envio — pendente.
- Sessão encerrada no admin — coberta por logout; expiração ociosa/absoluta permanece em teste de unidade.
- Transição inválida — coberta.
- Mobile 375 px e desktop 1280 px.

### 8.3 Regras

- Playwright sobe ambiente isolado com banco de teste.
- Sem dependência de produção.
- Evidência: vídeo/trace no CI em caso de falha.
- Testes independentes preparam seu estado; os que alteram configuração da loja usam `describe.serial`, operador dedicado e restauração no `afterAll`.
- Cenários que criam pedido rodam apenas no projeto desktop para não esgotar o limite de 10 pedidos/min por IP compartilhado entre desktop e mobile.

### 8.4 Cobertura implementada

- Smoke da home, navegação para o cardápio e health/readiness pelo proxy.
- Acesso anônimo a rota protegida redirecionado para `/admin/login` com `redirect`, e credenciais inválidas exibindo mensagem genérica (US-061).
- Pedido ponta a ponta: cardápio → personalização → carrinho → checkout → confirmação → fila → detalhe → `Confirmar pedido`, em desktop e mobile (US-040 a US-043, US-070 a US-072); o link de handoff continua disponível após recarregar a confirmação.
- Cenários de falha em `e2e/failure-scenarios.spec.ts`: `STORE_CLOSED`, `PRODUCT_UNAVAILABLE`, `ZONE_UNAVAILABLE`, retirada com taxa zero, troco congelado, idempotência, `409` em transição inválida e `UNAUTHORIZED` após logout.
- CRUD de catálogo e configuração da loja cobertos por Vitest; E2E dessas telas ainda pendente.
- Monitoramento de erros coberto por testes Go/Vitest do sanitizador, da captura só de 5xx e do recoverer, além do validador estático executado por `make test-sentry` e `make sentry-check`.

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
| Sentry | DSN vazio não inicializa SDK; nenhum teste envia rede |
| Sentry PII | `BeforeSend` remove user/request e mascara telefone, e-mail, CEP, query e token nos dois lados |
| Sentry escopo | 4xx e panics de regra de negócio não viram evento; 5xx/panic e timeout são capturados |
| Sentry build | Validador rejeita DSN literal, tracing, Replay, source maps e CSP sem `connect-src` do Sentry |
| GTM | Snippet padrão, nonce por resposta, hosts do Google no CSP e política de privacidade declarada; nenhum dado do pedido no `dataLayer` |
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

### 11.1 Teste de carga HTTP automatizado

O comando `backend/cmd/loadtest` executa as três frentes acima contra uma API
já em execução e imprime contagens, códigos de status e p50/p95/p99:

```bash
make up                       # ou make dev-api com MongoDB disponível
make test-load                # usa http://localhost:8080
make test-load LOADTEST_ARGS="-url http://127.0.0.1:8080 -orders-concurrency 5"
```

Por padrão o pedido é descoberto em `GET /menu` e `GET /delivery-zones`: o
primeiro produto disponível com opções obrigatórias atendidas e a primeira zona
ativa. Para usar um carrinho real, passe `-payload /caminho/pedido.json` com o
mesmo corpo aceito por `POST /orders`. O relatório nunca imprime nome, telefone
ou endereço; o pedido gerado é fictício e desativa o perfil de conveniência.

O limite de 10 pedidos/min por IP descrito no contrato pode produzir `429` na
carga padrão (10 chaves + 2 duplicatas). O comando reporta o `429` como aviso;
para medir sem esse ruído, reduza `-orders-concurrency`, aumente o limite no
ambiente de teste ou distribua as chamadas entre IPs. Erros de transporte e
`5xx` falham o comando. Os testes unitários da ferramenta rodam em `make test`
com `httptest` e não exigem MongoDB.

### 11.2 Benchmarks de tempo real

Os benchmarks cobrem o custo de abrir/fechar assinaturas, o `fan-out` do hub
para 1, 10 e 100 assinantes (consumidor rápido e lento), a contenção em tópicos
públicos distintos e o rate limiter com 1 e 1024 chaves:

```bash
make bench
```

Eles não entram em `make test` e não fixam limite de latência: servem para
acompanhar regressões de alocação e contenção entre versões. Os limites de
capacidade do hub (10 assinantes por pedido, 200 no tópico administrativo e
2000 tópicos por instância) continuam descritos em
[`04-arquitetura.md`](04-arquitetura.md).

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

O backend usa `golangci-lint` v2 quando instalado; sem ele, `make lint` faz apenas `go vet` e pode deixar problemas passar. O frontend exige Node 22.12 ou superior (fixado em `.tool-versions` para asdf).

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
| US-040 | Criação de pedido no service + handler + repository + E2E de retirada/troco |
| US-041 | Concorrência de idempotência no repository + E2E de replay/conflito |
| US-042 | Geração de URL + falha de handoff no frontend + link disponível após recarregar no E2E |
| US-061 | Login no E2E: redirecionamento, mensagem genérica e logout |
| US-050 | Handler público sem PII |
| US-060 | Auth handler + cookie + rate limit |
| US-062 | `AdminProductForm` + `AdminProductsView` no Vitest; service/store do catálogo |
| US-063 | `AdminCatalogForm` + views de categorias/adicionais no Vitest |
| US-064 | `AdminZonesView` no Vitest; conversão de taxa em centavos |
| US-065 | `AdminStoreView` no Vitest; validação de horários/E.164 + API |
| US-070 | Filtros/paginação no service + repository + handler; PII mascarada |
| US-071 | Detalhe no repository + handler com snapshot completo |
| US-072 | Matriz de status no domínio + CAS/auditoria no repository + handler + E2E de transição inválida |
| US-082 | axe + teste de teclado |
| US-093 | Revisão de log + testes de segurança |

---

## 15. Pendências

Resolvido: CI no GitHub Actions (D-016), job de backend com MongoDB 8 como service e `MONGO_TEST_URI`, job de E2E com Chromium instalado pelo Playwright, ambas em `.github/workflows/ci.yml`.

Pendências:

- Adotar testcontainers para executar testes de repositório localmente sem depender de Mongo externo; hoje `MONGO_TEST_URI` ausente faz os testes de integração pularem.
- Definir ferramenta de load test.
- Definir se haverá teste de contrato OpenAPI.
- Definir responsável por teste manual de acessibilidade e executar axe nas telas críticas.
- Injetar falha de handoff e retry no E2E; hoje a disponibilidade do link após recarregar é verificada, mas a falha de envio não é simulada.
- Completar os cenários de troco inválido e duplo clique no envio no E2E.
