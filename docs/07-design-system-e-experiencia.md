# Design System e Experiência — Brocados Lanches

**Versão:** 1.1  
**Atualizado em:** 26/11/2026  
**Fonte visual:** `../design-system/brocados-lanches-design-system-v2.5.html`  
**Tokens:** `../design-system/tokens.css` e `../design-system/tokens.json`  
**Referência funcional:** `12-referencia-ola-click.md`  
**Relacionados:** `03-prd.md`, `06-contrato-api.md`, `08-backlog.md`, `09-qualidade-e-testes.md`.

---

## 1. Papel deste documento

Este documento traduz o design system v2.5 em regras de produto e implementação. Ele responde:

- quais tokens existem e quais faltam;
- quais componentes precisam existir;
- como cada tela se comporta em 375 px e 1280 px;
- quais estados (carregando, vazio, erro, sucesso) são obrigatórios;
- como garantir acessibilidade e consistência;
- o que aproveitar do protótipo e o que substituir.

O site público da OlaClick é referência de fluxo e conteúdo, **não** de identidade visual. Não copie cores, fontes, textos, imagens ou layout sem aprovação.

---

## 2. Identidade visual

### 2.1 Cores

| Token | Valor | Uso |
|---|---|---|
| `--bg-primary` | `#0A0A0A` | Fundo principal |
| `--bg-surface` | `#161616` | Cards, superfícies |
| `--bg-input` | `#1F1F1F` | Inputs e áreas internas |
| `--bg-border` | `#2E2E2E` | Bordas |
| `--brand-primary` | `#FF8C00` | Ação principal, marca |
| `--brand-hover` | `#E07B00` | Hover da ação |
| `--brand-neon` | `#FFB703` | Preços, destaque, foco |
| `--text-main` | `#FFFFFF` | Texto principal |
| `--text-sub` | `#A3A3A3` | Texto secundário |
| `--status-open` | `#22C55E` | Loja aberta/sucesso |
| `--status-error` | `#EF4444` | Erro; “fechado” deve usar neutro + texto até existir token específico |

Contraste calculado:

| Par | Razão | Situação |
|---|---:|---|
| Texto principal sobre fundo | 19,8:1 | Excelente |
| Texto secundário sobre fundo | 7,85:1 | AA/AAA |
| Texto secundário sobre superfície | 7,17:1 | AA/AAA |
| Primária sobre fundo | 8,49:1 | Excelente |
| Neon sobre fundo | 11,34:1 | Excelente |
| Preto sobre primária | 8,49:1 | Excelente |
| Erro sobre superfície | 4,81:1 | Passa AA para texto normal; validar em textos pequenos |

Regras:

- Nunca usar cor literal em componente; usar token ou classe do tema.
- Status nunca é comunicado só por cor; sempre acompanhar de texto e/ou ícone.
- Não usar vermelho de erro para “loja fechada” sem decisão de design; fechado é estado operacional, não falha.
- Não usar a primária como fundo de texto pequeno sem verificar contraste.
- Não criar nova cor sem aprovação e atualização de `tokens.css` + `tokens.json`.

### 2.2 Tipografia

| Token | Valor | Uso |
|---|---|---|
| `--font-display` | Bebas Neue | Títulos de marca, hero, números grandes |
| `--font-sans` | Inter | Interface, corpo, formulários |
| `--font-size-display` | 36 px / 1.0 | Display |
| `--font-size-h1` | 22 px | Título de seção |
| `--font-size-h2` | 16 px | Nome de produto/card |
| `--font-size-price` | 16 px | Preço |
| `--font-size-body` | 14 px | Corpo |
| `--font-size-caption` | 12 px | Legenda |

Regras:

- Bebas Neue só em display; não usar em corpo.
- Inter para todos os controles e textos longos.
- Preço sempre com peso e cor de destaque.
- Não usar tamanhos abaixo de 12 px em conteúdo essencial.
- Fontes self-hosted em `frontend/public/fonts/` com `font-display: swap` e licenças SIL OFL; sem Google Fonts.

### 2.3 Breakpoints

| Nome | Largura | Uso |
|---|---:|---|
| Mobile | 375 px | Referência principal; mínimo 320 px |
| Tablet | 768 px | Ajustes intermediários |
| Desktop | 1280 px | Layout completo |
| Wide | 1440 px | Container máximo |

---

## 3. Lacunas de token

O design system atual documenta cores, tipografia e breakpoints. Faltam tokens para:

- espaçamento;
- raio;
- sombra;
- movimento;
- z-index;
- estados de foco, disabled, overlay e warning/info.

Enquanto não houver aprovação, use os valores derivados do protótipo e registre a decisão. A proposta abaixo **não** está aprovada; não altere os tokens sem design.

| Grupo | Proposta | Origem no protótipo |
|---|---|---|
| Espaçamento | base 4 px; `4, 8, 12, 16, 20, 24, 32, 40, 48, 64` | classes Tailwind |
| Raio | `8, 12, 16, 24, 999` px | `rounded-lg/xl/2xl/3xl/full` |
| Sombra | botão `0 4px 16px rgba(255,140,0,.32)`; card `0 12px 32px rgba(0,0,0,.55)`; modal `0 24px 48px rgba(0,0,0,.85)` | CSS do protótipo |
| Movimento | `150ms` e `250ms`, `ease-out`; respeitar `prefers-reduced-motion` | transições existentes |
| z-index | header 40, overlay 50, toast 60 | protótipo |
| Foco | anel de 2 px `--brand-neon` com offset 2 px | CSS atual do app |
| Overlay | `rgba(0,0,0,.75)` | modal |
| Warning | a definir com design | não existe |
| Info | a definir com design | não existe |

**Decisão necessária:** aprovar ou ajustar esses valores e atualizar `tokens.css`, `tokens.json`, `design-system/README.md` e este documento juntos.

---

## 4. Componentes

### 4.1 Base

| Componente | Variantes/estados | Regras |
|---|---|---|
| `BrandLogo` | completo, símbolo | SVG; não usar imagem raster sem necessidade |
| `BaseButton` | primary, secondary, ghost, danger; loading, disabled | Altura mínima 44 px; foco visível; loading bloqueia duplo clique |
| `IconButton` | neutral, primary, danger | Área de toque 44 px; `aria-label` obrigatório |
| `Chip` / `FilterChip` | ativo, inativo | Não usar só cor; estado selecionado com `aria-pressed` |
| `SearchInput` | default, focus, error | Rótulo visível ou `aria-label`; limpar busca |
| `QuantityStepper` | default, min, max | Botões ≥ 44 px; valor acessível; não permitir zero silencioso |
| `PriceText` | normal, destaque, riscado | Formatar BRL; nunca calcular no componente |
| `StatusBadge` | aberto, fechado, pending, confirmed, preparing, ready, out_for_delivery, delivered, picked_up, cancelled | Texto + cor/ícone; não só cor |
| `Toast` | sucesso, erro, info | `aria-live`; não substitui erro de formulário |
| `Modal` / `ConfirmDialog` | default, danger | Foco preso; ESC fecha; retorna foco |
| `Drawer` / `BottomSheet` | carrinho, filtros, admin | Foco preso; scroll lock; fechar por gesto e botão |
| `EmptyState` | catálogo, carrinho, pedidos | Explica o vazio e oferece ação |
| `ErrorState` | rede, servidor, não encontrado | Mensagem em português; botão “Tentar novamente” |
| `LoadingSkeleton` | card, lista, detalhe | Respeita reduced motion |
| `ImageWithFallback` | produto, logo | `width`, `height`, `loading`, `alt`; fallback visual |

### 4.2 Catálogo e carrinho

| Componente | Uso |
|---|---|
| `ProductCard` | Grade 2 colunas mobile, 3 desktop; imagem, nome, descrição curta, preço, ação |
| `ProductCardCompact` | Home/destaques; linha horizontal |
| `CategoryPill` | Filtro por categoria |
| `OptionGroup` | Radio (escolha única) ou checkbox (múltipla); obrigatório/mín/máx |
| `AddonRow` | Nome, preço, checkbox, disponibilidade |
| `CartLine` | Nome, opções, observação, preço, stepper, remover |
| `CartSummary` | Subtotal, taxa, total; nota de retirada |
| `CheckoutStep` | Indicador de etapa; não usar só cor |
| `FulfillmentSelector` | Delivery/retirada com cartões |
| `ZoneSelect` | Zona + taxa; estado sem zonas |
| `PaymentSelector` | Pix/cartão/dinheiro; troco condicional |
| `OrderSummary` | Itens e total para revisão |
| `OrderStatusTimeline` | Status atual e histórico público |

### 4.3 Admin

| Componente | Uso |
|---|---|
| `AdminShell` | Sidebar desktop, drawer mobile com todos os itens permitidos pelo papel e header — **implementado em US-061, navegação mobile revisada na US-126** |
| `AdminPageHeader` | Título, ação primária, breadcrumb |
| `DataTable` | Produtos, pedidos, usuários, auditoria; paginação (ainda não abstraído) |
| `TableEmpty` | Estado vazio com CTA |
| `AdminCatalogFilters` | Busca e situação do catálogo — **implementado em US-062 a US-064** |
| `AdminCatalogList` | Cartões do catálogo com ações, estado vazio/erro e paginação — **implementado em US-062 a US-064** |
| `AdminCatalogForm` | Nome, slug, ordem, ativo, preço/taxa — **implementado em US-063/US-064** (substitui `CategoryForm`, `AddonForm` e `ZoneForm`) |
| `AdminProductForm` | Nome, slug, categoria, preço, descrição, imagem, alergênicos, opções, adicionais — **implementado em US-062** |
| `StoreSettingsForm` | Kill switch, horários, contato, prazo — **implementado inline em `AdminStoreView` (US-065)** |
| `AdminOrderFilters` | Status, data, fulfillment, busca — **implementado em US-070** |
| `OrderStatusBadge` | Status na fila e no detalhe — **implementado em US-070** |
| `AdminOrderDetailView` | Snapshot, cliente, endereço, pagamento, histórico — **implementado em US-071** |
| `OrderStatusActions` | Ações válidas para o status atual; confirmação de cancelamento — **implementado em US-072** |
| `AuditLogTable` | Ator, ação, entidade, data; sem PII |
| `UserForm` | Nome, e-mail, papel, ativo |

---

## 5. Telas do cliente

### 5.1 Home — `/`

**Objetivo:** apresentar a loja e levar ao cardápio.

- Header com logo, status e prazo.
- Hero com destaque do dia e CTA “Abrir cardápio”.
- Categorias ativas em chips horizontais.
- “Mais pedidos” com produtos disponíveis.
- Cartão de endereço, WhatsApp e horário.
- Barra fixa de carrinho no mobile; sidebar no desktop (a sidebar fica pendente de decisão em §14).

Implementado em `frontend/src/views/HomeView.vue` (US-125): a barra `CartSummaryBar` aparece somente com itens, exibe contagem e subtotal calculados a partir do catálogo e leva a `/carrinho`; some a partir de 768 px para não concorrer com a navegação desktop.

**Estados:**

- carregando: skeleton do hero e cards;
- erro: mensagem + tentar novamente;
- loja fechada: status claro, CTA de cardápio ainda visível, aviso de que o pedido não pode ser concluído (implementado na US-125, sem usar cor de erro);
- catálogo vazio: estado próprio, sem hero quebrado.

### 5.2 Cardápio — `/cardapio`

**Objetivo:** encontrar e adicionar produtos.

- Busca por nome/descrição.
- Filtro por categoria.
- Grade 2 colunas no mobile, 3 no desktop.
- Produto esgotado com selo e ação desabilitada.
- Contagem de resultados.
- Carrinho persistente.

**Estados:** carregando, vazio por busca, erro, sucesso ao adicionar.

### 5.3 Produto — `/produto/:slug`

**Objetivo:** personalizar e adicionar.

- Imagem, nome, descrição, alergênicos e preço base.
- Grupos de opções com obrigatoriedade, mínimo/máximo.
- Adicionais com preço.
- Observação com limite.
- Preço final atualizado.
- CTA fixo no mobile.

**Estados:**

- carregando;
- produto não encontrado;
- produto esgotado/inativo;
- opção obrigatória não escolhida;
- limite de adicionais;
- erro ao carregar.

### 5.4 Carrinho — `/carrinho`

**Objetivo:** revisar antes do checkout.

- Itens com opções, observação, quantidade e remover.
- Subtotal apenas: nenhuma taxa ou total com entrega é exibido nesta etapa.
- Aviso de que a taxa aparece no checkout, após a escolha entre entrega e retirada, e de que o total é calculado no servidor.
- CTA “Continuar” habilitado quando há itens disponíveis; leva a `/checkout`.
- Estado vazio com CTA para o cardápio.

### 5.5 Checkout — `/checkout`

**Objetivo:** coletar dados e criar pedido.

Etapas visuais:

1. Dados do cliente (nome e telefone).
2. Recebimento (delivery/retirada).
3. Endereço/zona (somente delivery).
4. Pagamento e troco.
5. Revisão e envio.

Regras:

- O cliente é identificado primeiro. Se o telefone corresponder às preferências locais, o checkout as aplica sem consultar a API. Caso contrário, com a opção de conveniência habilitada e um telefone brasileiro completo, o checkout consulta `POST /customers/profile` e preenche entrega, endereço, zona e pagamento; nome digitado, observação e troco não são sobrescritos (D-039).
- A consulta ao perfil não devolve taxa, histórico de pedidos, observação, troco nem `phoneKey`. Zona inativa retornada pelo perfil é descartada; preço e taxa continuam vindo da cotação autoritativa.
- Retirada oculta endereço e zera taxa.
- Zona inativa não aparece.
- Telefone com máscara e normalização. Colar o número com prefixo `+55` preserva o DDD: a máscara remove o código do país antes de formatar, e a busca envia a forma mascarada; servidor e banco comparam apenas os dígitos normalizados com prefixo `55`. A validação exige DDD existente, celular com 9 dígitos começando em `9` ou fixo com 8 dígitos começando de `2` a `5`; número com DDI duplicado é bloqueado antes do envio (D-043).
- Erros por campo, com foco no primeiro erro.
- Total vem de `/orders/quote`; se a cotação falhar, não enviar.
- Na entrega, a cotação só é solicitada após escolher a zona. Depois de recarregar a página, o resumo pede a zona em vez de enviar uma cotação incompleta (`422`).
- Botão de envio desabilitado durante a requisição e enquanto a cotação não chega.
- Se a loja fechar, mensagem clara e ação de voltar ao cardápio.
- Se o pedido for criado (`201`), limpar carrinho e ir para confirmação; no replay idempotente (`200`), o mesmo destino é usado.
- Se falhar, preservar carrinho e dados; a chave idempotente é mantida em retries do mesmo pedido e renovada somente quando o conteúdo da requisição muda.
- Após o pedido gravado (`201` ou replay idempotente), salvar neste navegador entrega/retirada, zona, nome, telefone, endereço e forma de pagamento para o próximo pedido. Observação e valor de troco não são salvos; zona que deixou de existir é descartada ao carregar as preferências (D-037).
- Checkbox “Buscar e salvar meus dados para agilizar meus pedidos”, marcado por padrão. Desmarcado, não há consulta ao perfil, as preferências locais são apagadas e o pedido envia `saveCustomerProfile: false`; o servidor desativa o perfil por telefone e remove os dados de conveniência, mantendo o pedido (D-038).

Implementado em `frontend/src/views/CheckoutView.vue` (US-122). A última seção é “Revisão e envio”, com o total cotado pelo servidor e o botão de envio dentro dela; na retirada a etapa de endereço é oculta e a numeração é ajustada.

### 5.6 Confirmação — `/pedido/:publicId`

**Objetivo:** confirmar e orientar o próximo passo.

- Número do pedido, status, itens, total e prazo.
- Botão “Abrir WhatsApp” e link copiável.
- Aviso de que o pedido já está registrado.
- Link para voltar ao cardápio; o próprio caminho `/pedido/:publicId` serve de acompanhamento.
- Se o WhatsApp falhar, manter pedido e oferecer reenvio pelo link guardado no `sessionStorage` na mesma sessão.

### 5.7 Acompanhamento — `/pedido/:publicId`

**Objetivo:** consultar status sem expor PII.

- Timeline: pending → confirmed → preparing → ready → out_for_delivery/picked_up → delivered/picked_up (histórico completo depende da fila administrativa).
- Itens e total nesta entrega.
- Contato da loja e atualização manual previstos para a US-051.
- Enquanto a página está aberta, o status é atualizado automaticamente por WebSocket e o texto discreto abaixo do título informa se a conexão está ativa ou em reconexão (D-040); estados finais encerram a assinatura.
- Estado inválido/não encontrado sem revelar dados.

### 5.8 Contato e privacidade

- Endereço, horários, WhatsApp e Instagram.
- Link para mapa, se aprovado.
- Política de privacidade e termos com data de publicação. Desde a US-126 as páginas exibiam a data do rascunho (25/11/2026); em 25/11/2026 o cliente confirmou os textos e eles passaram a ser apresentados como versão vigente, publicada nessa data (D-033). A privacidade e os termos foram detalhados com base em modelos externos, adaptados apenas ao fluxo real (dados do pedido, sessão administrativa, IP para limite de requisições, link `wa.me`, retenção proposta, registro/idempotência, regras de preço/entrega/pagamento e acesso ao painel), sem copiar prestadores, assinaturas ou recursos que a loja não usa.
- Canal do titular: os canais já publicados na `/contato`, sem inventar CNPJ, razão social ou e-mail (D-033).
- A política descreve as preferências do checkout guardadas no navegador e oferece o botão “Apagar preferências salvas neste navegador” na seção de retenção; a versão publicada em 25/11/2026 foi atualizada em 26/11/2026 para cobrir esse tratamento (D-037), o perfil de conveniência no servidor, com opt-out no checkout, retenção de 24 meses e exclusão pelos canais da loja (D-038), e a consulta por telefone com finalidade exclusiva de preencher o próximo pedido (D-039). Em 26/11/2026 passou a descrever também o token opaco de acesso ao histórico e o botão “Apagar acesso aos meus pedidos”, com o aviso de que isso remove o histórico naquele navegador (D-041).

### 5.9 Meus pedidos — `/meus-pedidos`

**Objetivo:** rever os pedidos feitos neste navegador a qualquer momento (D-041).

- Rota pública, acessível pelo link “Pedidos” no cabeçalho, sem login e sem consulta por telefone.
- Lista paginada com número do pedido, status, data, itens, forma de entrega e total, no mesmo formato mínimo do acompanhamento; cada cartão leva ao `/pedido/:publicId`.
- O token opaco emitido na criação é guardado no `localStorage` versionado e enviado em `Authorization: Bearer`; ele não aparece na URL, em log nem na listagem.
- Texto informa que o acesso fica apenas neste dispositivo e que limpar os dados do navegador remove o histórico. A seção de retenção em `/privacidade` oferece a limpeza explícita do token.
- Estados: carregamento com skeleton, vazio sem token, vazio sem pedidos, erro com “Tentar novamente” e paginação com “Anterior”/“Próxima”.
- Cartões usam superfície, borda, raio e cores de status dos tokens existentes; sem novo padrão visual. Alvos de toque de 44 px, `aria-live` na paginação e foco visível preservados.
- O histórico não substitui o acompanhamento em tempo real: detalhe e status continuam em `/pedido/:publicId` com WebSocket.

---

## 6. Telas administrativas

### 6.1 Login — `/admin/login`

- E-mail e senha.
- Erro genérico.
- Loading e bloqueio de envio.
- Sem “lembrar-me” na V1.
- Após login, redirecionar para `/admin` ou para o destino em `?redirect=`, desde que comece com `/admin`.
- Implementado em `frontend/src/views/admin/AdminLoginView.vue` (US-061).

### 6.2 Dashboard — `/admin`

- Status da loja e kill switch.
- Pedidos por status.
- Produtos disponíveis/total.
- Zonas ativas.
- Pedidos recentes.
- Ações rápidas.

Implementado em `frontend/src/views/admin/AdminDashboardView.vue` (US-121). O interruptor manual usa `PATCH /admin/store` com confirmação e auditoria; os indicadores vêm das APIs administrativas e o status efetivo continua calculado no servidor. Contagens indisponíveis aparecem como `—`, e a lista recente mostra apenas número, status e data.

### 6.3 Produtos — `/admin/produtos`

- Lista com busca, categoria, status.
- Criar/editar.
- Preço em reais na UI, convertido para centavos na API.
- Ativar/desativar, disponível/esgotado, destaque.
- Grupos de opções e adicionais.
- Imagem com alt, por upload ou por URL/caminho.
- Auditoria ao salvar.

Implementado em `frontend/src/views/admin/AdminProductsView.vue` e `frontend/src/components/admin/AdminProductForm.vue` (US-062). A lista carrega categorias e adicionais com `pageSize=100` para alimentar o formulário; preços e opções são enviados em centavos.

O campo de imagem aceita tanto uma URL absoluta quanto um caminho relativo (ex.: `/products/burger.png`, usado pelo seed) e mostra prévia com `alt` descritivo. O envio de arquivo usa botão nativo com `accept` de JPEG/PNG/WebP/AVIF, teto de 2 MB, estados de enviando, sucesso e erro em texto, e preenche o campo com a URL devolvida pelo servidor (D-045). O arquivo vai em `multipart/form-data` para `POST /api/v1/admin/uploads`; nenhuma imagem é convertida em `data:` no navegador. Se o upload falhar, o valor anterior do campo é mantido e o produto não é salvo automaticamente.

### 6.4 Categorias — `/admin/categorias`

- Nome, slug, ordem, ativa.
- Não excluir; desativar.

Implementado em `frontend/src/views/admin/AdminCategoriesView.vue` (US-063), com `AdminCatalogFilters`, `AdminCatalogForm` e `AdminCatalogList`.

### 6.5 Adicionais — `/admin/adicionais`

- Nome, slug, preço, ativo, ordem.
- Vínculo com produtos editado no produto.

Implementado em `frontend/src/views/admin/AdminAddonsView.vue` (US-063). O preço é digitado em reais e convertido para centavos antes do `POST`/`PATCH`.

### 6.6 Locais e taxas — `/admin/locais`

- Nome, slug, taxa, ativo, ordem.
- Taxa em reais na UI, centavos na API.
- Aviso de que a taxa vale para novos pedidos.

Implementado em `frontend/src/views/admin/AdminZonesView.vue` (US-064), com aviso de que pedidos antigos mantêm a taxa congelada no snapshot.

### 6.7 Loja — `/admin/loja`

- Kill switch.
- Horários por dia, múltiplos intervalos.
- Prazo, WhatsApp, Instagram, endereço.
- Preview do status efetivo a partir de `GET /api/v1/store`.
- Auditoria.

Implementado em `frontend/src/views/admin/AdminStoreView.vue` (US-065/US-025). O formulário valida E.164, horários `HH:MM` e sobreposição de intervalos antes do PATCH; o envio com `PATCH` vazio é bloqueado no cliente e o badge mostra o status calculado no servidor.

### 6.7.1 Integrações — `/admin/integracoes/webhook` (owner)

- Seção “Integrações → Webhook de Notificações” dentro das configurações da loja.
- Campo de URL e opção “Webhook ativo” com os estados de carregamento, vazio, erro e sucesso do padrão do painel.
- Grupo “Eventos enviados” com um checkbox por momento mapeado: pedido criado, confirmado, em preparo, pronto, saído para entrega, retirado, entregue e cancelado. Configurações antigas aparecem com todos marcados; nenhum marcado exibe aviso explícito e salva lista vazia (D-048).
- Aviso explícito de que o POST contém telefone do cliente e que falhas não bloqueiam o pedido, além da nota de que o cabeçalho do evento permanece genérico.
- Validação de UX aceita qualquer URL absoluta, sem exigir HTTPS ou permitir apenas `http`/`https`; o servidor permanece autoritativo (D-042).
- Implementado em `frontend/src/views/admin/AdminWebhookView.vue` (US-075 e US-076), com rota e item de menu restritos a `owner`.

### 6.8 Pedidos — `/admin/pedidos`

- Filtros por status, data e fulfillment.
- Paginação.
- Linha com número, cliente mascarado, total, status e horário.
- Detalhe com snapshot, endereço, pagamento e histórico; quando há perfil reconhecido pelo telefone, seção “Cliente recorrente” com contagem de pedidos, último endereço/zona, preferências e prazo de retenção, ou aviso de perfil desativado.
- Ações de transição válidas.
- Cancelamento com motivo.
- Aviso discreto de conexão em tempo real; criações e transições feitas por outro operador recarregam a fila e o detalhe sem recarga da página (D-040).

Implementado em `frontend/src/views/admin/AdminOrdersView.vue` e `frontend/src/views/admin/AdminOrderDetailView.vue` (US-070 a US-072, US-074). As datas dos filtros usam UTC; a exibição usa `America/Fortaleza`.

### 6.9 Usuários — `/admin/usuarios` (owner)

- Lista paginada com busca, criar, editar papel e ativar/desativar.
- Senha temporária forte no cadastro; senha nunca volta na resposta.
- Não desativar o último owner ativo.
- Implementado em `AdminUsersView.vue`, `AdminUserList.vue` e `AdminUserForm.vue`.

### 6.10 Auditoria — `/admin/auditoria` (owner)

- Filtros por entidade e ação; paginação e ordenação por data desc.
- Sem PII no registro exibido.
- Somente leitura, implementado em `AdminAuditView.vue`.

### 6.11 Conta — `/admin/conta`

- Nome, e-mail e papel da sessão atual.
- Logout disponível; troca de senha/MFA ficam fora da V1.
- Implementado em `AdminAccountView.vue`.

---

## 7. Estados obrigatórios

Toda tela que busca dados precisa tratar:

| Estado | Regra |
|---|---|
| Carregando | Skeleton ou indicador; não mostrar conteúdo vazio como se fosse resultado |
| Vazio | Explicar por que está vazio e oferecer próxima ação |
| Erro | Mensagem em português, sem JSON; botão de retry quando fizer sentido |
| Sucesso | Feedback claro; não usar toast como única confirmação de pedido |
| Sem permissão | `403` com mensagem e link para área permitida |
| Não encontrado | `404` sem revelar existência de recurso de terceiros |
| Offline | Avisar e preservar dados; não perder carrinho |

---

## 8. Acessibilidade

Meta: WCAG 2.1 AA.

Checklist:

- [x] Todo input tem rótulo visível ou `aria-label`.
- [x] Erro de campo associado com `aria-describedby` e `aria-invalid="true"` no controle.
- [x] Foco visível em todos os controles; o link “Pular para o conteúdo” é a primeira parada do Tab e leva ao único destino `#conteudo`, com `tabindex="-1"`.
- [x] Um único `<main>` por página. No shell administrativo, o destino do skip-link é o contêiner do conteúdo, e as views não repetem o `id="conteudo"`.
- [ ] Navegação por teclado em modais, drawers e menus.
- [ ] Foco preso em modal/drawer; ESC fecha.
- [ ] Alvo de toque ≥ 44 px.
- [ ] Contraste ≥ 4,5:1 para texto normal e ≥ 3:1 para texto grande.
- [ ] Status não depende só de cor.
- [ ] `prefers-reduced-motion` respeitado.
- [ ] Imagens com `alt` real; decorativas com `alt=""`.
- [ ] Ícones com `aria-hidden` quando acompanhados de texto.
- [ ] Preço e total anunciados de forma compreensível.
- [ ] Toast com `aria-live="polite"`; erro crítico com `role="alert"`.
- [ ] Não usar `autofocus` em mobile sem necessidade.
- [ ] Testar zoom 200% e 375 px.

Problemas do protótipo a corrigir:

- botões de stepper com 24–32 px;
- ausência de foco visível em vários controles;
- emoji como ícone;
- status baseado em cor;
- texto pequeno em 10–11 px;
- imagens ilustrativas genéricas;
- modais sem gestão de foco.

---

## 9. Conteúdo e microcopy

| Situação | Texto recomendado |
|---|---|
| Loja fechada | “A loja está fechada agora. Você pode ver o cardápio e voltar quando abrirmos.” |
| Item esgotado | “Este item está esgotado. Escolha outro para continuar.” |
| Zona indisponível | “Ainda não entregamos nesse bairro. Escolha outra região ou retire na loja.” |
| Carrinho vazio | “Seu carrinho está vazio. Veja o cardápio para começar.” |
| Troco inválido | “O valor para troco precisa ser maior ou igual ao total.” |
| Erro de rede | “Não conseguimos concluir agora. Seus dados foram preservados. Tente novamente.” |
| Pedido criado | “Pedido #1042 registrado. Abra o WhatsApp para enviar à loja.” |
| Handoff falhou | “O pedido está salvo. Abra o WhatsApp manualmente ou copie a mensagem.” |
| Pedido não encontrado | “Não encontramos este pedido. Confira o link ou fale com a loja.” |

Regras:

- Não usar “Erro 500”, “undefined”, “null” ou JSON.
- Não culpar o cliente.
- Não prometer prazo não confirmado.
- Usar “retirada”, não “pickup”, na interface.
- Usar “taxa de entrega”, não “frete”, se for o termo da loja.

---

## 10. Movimento e feedback

- Transições curtas (150–250 ms).
- Toast 2–4 s; erro crítico permanece até ação.
- Skeleton em vez de spinner em listas.
- Respeitar `prefers-reduced-motion`.
- Não usar `animate-ping` contínuo em produção.
- Confirmação de pedido não pode depender de animação.

---

## 11. Imagens e mídia

- WebP/AVIF em produção: o seed usa `.webp` e `ProductImage` oferece AVIF/WebP via `<picture>`, com o arquivo original como fallback.
- `width` e `height` para evitar layout shift.
- `loading="lazy"` fora da primeira dobra.
- `alt` descritivo; não usar “imagem de produto”.
- Fallback para imagem ausente.
- Uploads do painel são servidos em `/api/v1/uploads/{id}`, sem extensão no caminho; `ProductImage` usa o `<img>` direto nesse caso.
- Não usar o ícone genérico do protótipo como foto real.
- Fotos e logo precisam de autorização do cliente.

---

## 12. Mapeamento protótipo → produção

| Item do protótipo | Decisão |
|---|---|
| Dark mode + neon | Manter como identidade |
| Bebas Neue + Inter | Manter, self-hosted |
| Tailwind CDN | Substituir por build |
| Dados em memória | Substituir por API/Mongo |
| Checkout simulado | Substituir por pedido persistido |
| Admin sem login | Substituir por sessão e papéis |
| Ícones emoji | Substituir por SVG consistente |
| Cards sem foto | Usar fotos reais aprovadas |
| Status por cor | Adicionar texto/ícone |
| Stepper 24–32 px | Aumentar para 44 px |
| Sem estados de erro | Implementar todos |
| Sem acompanhamento | Criar tela pública |
| Sem auditoria | Criar no admin |
| Créditos | Mantida como página institucional, preenchida com créditos de desenvolvimento a partir do currículo do autor (D-032) e publicada em 25/11/2026 com `noindex` removido (D-033); mídia da loja segue pendente |

---

## 13. Checklist de QA visual

- [ ] 375 px sem scroll horizontal.
- [ ] 1280 px com sidebar e grid correto.
- [ ] 320 px sem quebra crítica.
- [ ] Tokens usados; nenhuma cor literal.
- [ ] Nenhuma fonte via CDN.
- [ ] Botões com 44 px.
- [ ] Foco visível em todos os controles.
- [ ] Estados de carregando/vazio/erro/sucesso.
- [ ] Textos em português e sem jargão.
- [ ] Preços formatados em BRL.
- [ ] Imagens com alt e dimensões.
- [ ] Modal/drawer com foco preso.
- [ ] `prefers-reduced-motion` testado.
- [ ] `make design-check` passando.
- [ ] Revisão visual humana aprovada.

---

## 14. Decisões abertas

1. Aprovar tokens de espaçamento, raio, sombra, movimento e z-index. Em 25/11/2026 foi decidido manter os valores atuais; padronizar exige aprovação e atualização conjunta de `tokens.css` e `tokens.json`.
2. Definir warning/info ou manter apenas aberto/erro. Mantido apenas aberto/erro; “Loja fechada” usa cor neutra desde a US-125.
3. Uso de emoji na interface decidido em 25/11/2026: não usar como ícone. Os emojis dos nomes de categoria do seed foram removidos na US-126.
4. Página de créditos mantida na V1, preenchida com os dados profissionais do autor (D-032) e publicada em 25/11/2026 (D-033); logotipo e fotos seguem pendentes de validação.
5. Definir fotos, logo e direitos de uso.
6. Navegação mobile do painel decidida em 25/11/2026: drawer completo, com todos os itens permitidos pelo papel, aberto pelo botão “Menu” no header e fechável por Escape, pelo botão “Fechar” ou pelo fundo. Implementado na US-126.
7. Definir se o carrinho desktop é sidebar fixa ou drawer. A barra fixa no mobile foi entregue na US-125; a versão desktop segue em aberto.

A auditoria de 25/11/2026 comparou a implementação com estas decisões e registrou os desvios e as tarefas de correção em `14-auditoria-design-system.md`.
