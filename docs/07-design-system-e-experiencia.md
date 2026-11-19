# Design System e Experiência — Brocados Lanches

**Versão:** 1.0  
**Atualizado em:** 18/11/2026  
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
- Em produção, self-host das fontes com `font-display: swap`; sem Google Fonts.

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
| `AdminShell` | Sidebar desktop, bottom nav mobile, header |
| `AdminPageHeader` | Título, ação primária, breadcrumb |
| `DataTable` | Produtos, pedidos, usuários, auditoria; paginação |
| `TableEmpty` | Estado vazio com CTA |
| `ProductForm` | Nome, slug, categoria, preço, descrição, imagem, opções, adicionais |
| `CategoryForm` | Nome, slug, ordem, ativo |
| `AddonForm` | Nome, slug, preço, ativo |
| `ZoneForm` | Nome, slug, taxa, ativo |
| `StoreSettingsForm` | Kill switch, horários, contato, prazo |
| `OrderFilters` | Status, data, fulfillment, busca |
| `OrderDetailPanel` | Snapshot, cliente, endereço, pagamento, histórico |
| `StatusTransitionActions` | Ações válidas para o status atual; confirmação |
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
- Barra fixa de carrinho no mobile; sidebar no desktop.

**Estados:**

- carregando: skeleton do hero e cards;
- erro: mensagem + tentar novamente;
- loja fechada: status claro, CTA de cardápio ainda visível, aviso de que o pedido não pode ser concluído;
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
- Subtotal, taxa estimada e total.
- Aviso de que o total final é confirmado no checkout.
- CTA “Continuar”.
- Estado vazio com CTA para o cardápio.

### 5.5 Checkout — `/checkout`

**Objetivo:** coletar dados e criar pedido.

Etapas visuais:

1. Recebimento (delivery/retirada).
2. Endereço/zona (somente delivery).
3. Dados do cliente.
4. Pagamento e troco.
5. Revisão e envio.

Regras:

- Retirada oculta endereço e zera taxa.
- Zona inativa não aparece.
- Telefone com máscara e normalização.
- Erros por campo, com foco no primeiro erro.
- Total vem de `/orders/quote`; se a cotação falhar, não enviar.
- Botão de envio desabilitado durante a requisição.
- Se a loja fechar, mensagem clara e ação de voltar ao cardápio.
- Se o pedido for criado, limpar carrinho e ir para confirmação.
- Se falhar, preservar carrinho e dados.

### 5.6 Confirmação — `/pedido/:publicId`

**Objetivo:** confirmar e orientar o próximo passo.

- Número do pedido, status, total e prazo.
- Botão “Abrir WhatsApp” e link copiável.
- Aviso de que o pedido já está registrado.
- Link para acompanhar.
- Se o WhatsApp falhar, manter pedido e oferecer reenvio.

### 5.7 Acompanhamento — `/pedido/:publicId`

**Objetivo:** consultar status sem expor PII.

- Timeline: pending → confirmed → preparing → ready → out_for_delivery/picked_up → delivered/picked_up.
- Itens e total.
- Contato da loja.
- Atualizar manualmente.
- Estado inválido/não encontrado sem revelar dados.

### 5.8 Contato e privacidade

- Endereço, horários, WhatsApp e Instagram.
- Link para mapa, se aprovado.
- Política de privacidade e termos com data.
- Canal do titular.

---

## 6. Telas administrativas

### 6.1 Login — `/admin/login`

- E-mail e senha.
- Erro genérico.
- Loading e bloqueio de envio.
- Sem “lembrar-me” na V1.
- Após login, redirecionar para `/admin`.

### 6.2 Dashboard — `/admin`

- Status da loja e kill switch.
- Pedidos por status.
- Produtos disponíveis/total.
- Zonas ativas.
- Pedidos recentes.
- Ações rápidas.

### 6.3 Produtos — `/admin/produtos`

- Lista com busca, categoria, status.
- Criar/editar.
- Preço em reais na UI, convertido para centavos na API.
- Ativar/desativar, disponível/esgotado, destaque.
- Grupos de opções e adicionais.
- Imagem com alt.
- Auditoria ao salvar.

### 6.4 Categorias — `/admin/categorias`

- Nome, slug, ordem, ativa.
- Não excluir; desativar.

### 6.5 Adicionais — `/admin/adicionais`

- Nome, slug, preço, ativo, ordem.
- Vínculo com produtos editado no produto.

### 6.6 Locais e taxas — `/admin/locais`

- Nome, slug, taxa, ativo, ordem.
- Taxa em reais na UI, centavos na API.
- Aviso de que a taxa vale para novos pedidos.

### 6.7 Loja — `/admin/loja`

- Kill switch.
- Horários por dia, múltiplos intervalos.
- Prazo, WhatsApp, Instagram, endereço.
- Preview do status efetivo.
- Auditoria.

### 6.8 Pedidos — `/admin/pedidos`

- Filtros por status, data e fulfillment.
- Paginação.
- Linha com número, cliente mascarado, total, status e horário.
- Detalhe com snapshot, endereço, pagamento e histórico.
- Ações de transição válidas.
- Cancelamento com motivo.

### 6.9 Usuários — `/admin/usuarios` (owner)

- Lista, criar, editar papel, ativar/desativar.
- Não desativar o último owner.
- Sem senha em resposta.

### 6.10 Auditoria — `/admin/auditoria` (owner)

- Filtros por ator, entidade e data.
- Sem PII.
- Somente leitura.

### 6.11 Conta — `/admin/conta`

- Nome, e-mail, papel.
- Troca de senha (P1; se não houver, owner redefine).
- Sair.

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

- [ ] Todo input tem rótulo visível ou `aria-label`.
- [ ] Erro de campo associado com `aria-describedby`.
- [ ] Foco visível em todos os controles.
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

- WebP/AVIF em produção.
- `width` e `height` para evitar layout shift.
- `loading="lazy"` fora da primeira dobra.
- `alt` descritivo; não usar “imagem de produto”.
- Fallback para imagem ausente.
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
| Créditos | Manter como página institucional, se aprovado |

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

1. Aprovar tokens de espaçamento, raio, sombra, movimento e z-index.
2. Definir warning/info ou manter apenas aberto/erro.
3. Definir uso de emoji na interface; proposta: não usar como ícone.
4. Definir se a página de créditos permanece na V1.
5. Definir fotos, logo e direitos de uso.
6. Definir se o admin mobile usa bottom nav ou drawer.
7. Definir se o carrinho desktop é sidebar fixa ou drawer.
