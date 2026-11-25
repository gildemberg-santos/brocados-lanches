# Auditoria de conformidade com o design system — 25/11/2026

**Versão:** 1.0  
**Atualizado em:** 25/11/2026  
**Status:** auditoria concluída; achados convertidos no Épico 12 de `08-backlog.md`.  
**Relacionados:** `../AGENTS.md`, `07-design-system-e-experiencia.md`, `08-backlog.md`, `13-decisoes-e-pendencias.md`, `../design-system/README.md`.

---

## 1. Objetivo

Verificar se as telas implementadas no frontend estão dentro do planejado no design system e na documentação de experiência, sem alterar o protótipo ou os tokens. O resultado é um inventário de aderência, uma lista de desvios com evidência e as tarefas de correção correspondentes.

Esta auditoria **não altera a hierarquia de fontes de verdade** definida em `../AGENTS.md` e `README.md`: em conflito, valem `AGENTS.md`, `briefing/chamado.md`, `docs/`, protótipo HTML. Quando o protótipo e `07-design-system-e-experiencia.md` divergem, o documento de design prevalece — a divergência está registrada na seção 6.

---

## 2. Escopo e método

Foram consultados:

- `../design-system/README.md`, `../design-system/tokens.css` e `../design-system/tokens.json` na íntegra;
- `../design-system/brocados-lanches-design-system-v2.5.html` (2023 linhas): funções `buildClientScreenContent` (linhas 928–1348), `buildAdminScreenContent` (linhas 1446–1674), rodapé fixo, sidebar desktop e navegação;
- `07-design-system-e-experiencia.md`, seções 2 a 14;
- 11 views do cliente e 13 views do painel, além dos componentes compartilhados e do shell administrativo;
- `make design-check`: aprovado — tokens consistentes e 205 arquivos de frontend validados;
- navegação real em `http://localhost:8081` nos viewports de 375 px e 1280 px, com medição de `scrollWidth` (nenhum scroll horizontal nas rotas medidas) e login com usuário administrativo de desenvolvimento.

**Limitações da verificação:** a suíte completa (`make test`, `make lint`) não foi reexecutada nesta auditoria documental; as conclusões sobre comportamento vêm de leitura de código, do `design-check` e da navegação manual. As tarefas do Épico 12 exigem testes automatizados conforme a Definition of Done de `08-backlog.md`.

---

## 3. Veredito

A implementação está **majoritariamente dentro do planejado**. Identidade visual, tokens, tipografia, componentes base, estados de dados, regra de servidor autoritativo e acessibilidade básica atendem ao documento de design e ao protótipo. Os desvios relevantes concentram-se em quatro frentes:

1. tela pública de pedido (título, timeline e orientação de handoff);
2. dashboard administrativo (kill switch e indicadores do protótipo);
3. checkout (token CSS inexistente e etapa de revisão);
4. acessibilidade dos formulários administrativos e semântica do shell.

Nenhum desses achados bloqueia segurança ou persistência do pedido; por isso viraram histórias **P1/P2**, não P0.

---

## 4. Matriz de aderência

### 4.1 Telas do cliente

| Tela | Situação | Base | Observação |
|---|---|---|---|
| Home `/` | Parcial | `07` §5.1 e protótipo Tela 00 | Hero, categorias, destaques e contatos OK; falta barra fixa de carrinho no mobile e aviso de que o pedido não pode ser concluído com a loja fechada |
| Cardápio `/cardapio` | Conforme | `07` §5.2 e protótipo Tela 01 | Busca, filtro, grid 2/3 colunas, esgotado, contagem e estados presentes |
| Produto `/produto/:slug` | Conforme | `07` §5.3 e protótipo Tela 02 | Imagem, alergênicos, grupos, adicionais, observação, preço e CTA sticky presentes |
| Carrinho `/carrinho` | Conforme | `07` §5.4 e protótipo Tela 03 | Itens, quantidade, remoção, limpeza, totais, aviso de total final e estado vazio presentes |
| Checkout `/checkout` | Parcial | `07` §5.5 | Regras de negócio OK; 4 etapas (3 na retirada) em vez das 5 do documento e token CSS inválido |
| Pedido `/pedido/:publicId` | Fora do planejado | `07` §5.6/§5.7 | Título fixo “Pedido registrado” mesmo em erro; sem timeline, contato da loja e aviso de registro fora do handoff |
| Contato `/contato` | Conforme | `07` §5.8 | Dados vêm de `GET /store` |
| Privacidade / Termos | Parcial | `07` §5.8 e §9 | Sem data de publicação; canal do titular da privacidade ainda pendente |
| Créditos `/creditos` | Conforme | `07` §12 e protótipo Tela 05 | Página institucional mantida com `noindex`; créditos de desenvolvimento preenchidos em 25/11/2026 a partir do currículo do autor (D-032), com links profissionais públicos e sem contato pessoal |

### 4.2 Telas administrativas

| Tela | Situação | Base | Observação |
|---|---|---|---|
| Login `/admin/login` | Conforme | `07` §6.1 | Erro genérico, loading, sem “lembrar-me” e redirect seguro |
| Dashboard `/admin` | Fora do planejado | `07` §6.2 e protótipo Admin 01 | Faltam kill switch, pedidos por status, disponíveis/total, prazo de entrega e pedidos recentes |
| Produtos `/admin/produtos` | Conforme | `07` §6.3 e protótipo Admin 02 | CRUD, preço em reais, opções, adicionais e auditoria |
| Categorias e adicionais | Conforme | `07` §6.4/§6.5 e protótipo Admin 03 | CRUD com slug, ordem e ativo |
| Zonas `/admin/locais` | Conforme | `07` §6.6 e protótipo Admin 04 | Taxa em reais, centavos na API e aviso de snapshot |
| Loja `/admin/loja` | Parcial | `07` §6.7 | Kill switch, horários e preview OK; “Loja fechada” em cor de erro, mas o §2.1 pede neutro |
| Pedidos e detalhe | Conforme | `07` §6.8 | Filtros, paginação, PII mascarada, transições e histórico; datas UTC nos filtros conforme o documento |
| Usuários, auditoria e conta | Conforme | `07` §6.9–§6.11 | Papéis, último owner protegido, sem PII e logout |
| Shell do painel | Parcial | `07` §8 e §14 | Bottom nav mobile com 4 atalhos; `<main id="conteudo">` duplicado/aninhado |

---

## 5. Achados

Severidade: **Alta** = corrige comportamento fora do planejado ou informação enganosa; **Média** = fidelity visual/acessibilidade relevante; **Baixa** = ajuste de conteúdo ou decisão ainda aberta. A severidade documental não é prioridade de backlog; o mapeamento está na seção 8.

| ID | Sev. | Achado | Evidência | Origem | Tarefa |
|---|---|---|---|---|---|
| AUD-01 | Alta | Título “Pedido registrado” fixo, inclusive em erro/404 | `frontend/src/views/OrderView.vue:63` | `07` §5.7 e §9 | US-120 |
| AUD-02 | Alta | Falta timeline, contato da loja e aviso de registro quando não há link do WhatsApp | `frontend/src/views/OrderView.vue:61-128` | `07` §5.6/§5.7 | US-120 |
| AUD-03 | Alta | Dashboard sem kill switch, pedidos por status, disponíveis/total, prazo e recentes | `frontend/src/views/admin/AdminDashboardView.vue:73-112` | `07` §6.2 e protótipo Admin 01 | US-121 |
| AUD-04 | Alta | `var(--bg-base)` não existe; input do checkout fica com fundo transparente | `frontend/src/views/CheckoutView.vue:540` | `07` §2.1 e tokens | US-122 |
| AUD-05 | Média | Erros de formulário admin sem `aria-describedby`/`aria-invalid` no campo | `AdminCatalogForm.vue:117-131`, `AdminProductForm.vue:192+`, `AdminStoreView.vue:159`, `AdminUserForm.vue:83` | `07` §8 | US-124 |
| AUD-06 | Média | “Loja fechada” em `--status-error` no painel | `frontend/src/views/admin/AdminStoreView.vue:285-287` | `07` §2.1 | US-125 |
| AUD-07 | Média | Checkout sem etapa explícita “Revisão e envio” | `frontend/src/views/CheckoutView.vue:180-415` | `07` §5.5 | US-122 |
| AUD-08 | Média | Falta barra fixa de carrinho no mobile | `frontend/src/views/HomeView.vue:81-204` e `SiteHeader.vue` | `07` §5.1 e protótipo Tela 00 | US-125 |
| AUD-09 | Média | Loja fechada na home sem aviso de que o pedido não pode ser concluído | `frontend/src/views/HomeView.vue:123-142` | `07` §5.1 | US-125 |
| AUD-10 | Baixa | Páginas legais sem data de publicação e canal do titular pendente | `frontend/src/views/PrivacyView.vue`, `TermsView.vue` | `07` §5.8 e §9 | US-126 |
| AUD-11 | Média | `<main id="conteudo">` no shell e nas views, com IDs duplicados e landmark aninhado | `frontend/src/components/admin/AdminShell.vue:48`, views `admin/*.vue` | `07` §8 | US-124 |
| AUD-12 | Baixa | Emojis do seed em nomes de categoria | `design-system/seed.json` | `07` §12 e §14 | US-126 |
| AUD-13 | Baixa | Bottom nav do admin expõe só 4 atalhos | `frontend/src/components/admin/AdminShell.vue:53-62` | `07` §14 (decisão aberta) | US-126 |
| AUD-14 | Média | `make design-check` não detecta token CSS inexistente | `scripts/check_design_system.py` | Gate do `../AGENTS.md` §8.5/§12.1 | US-123 |

**Atualização em 25/11/2026:** AUD-04 e AUD-14 foram corrigidos junto com a US-123. O checkout passou a usar `--bg-input`, e `scripts/check_design_system.py` passou a reprovar qualquer `var(--…)` não declarado em `tokens.css`, inclusive com sintaxe de fallback. Três testes de regressão foram adicionados a `scripts/test_check_design_system.py`.

AUD-01 e AUD-02 foram corrigidos junto com a US-120. A tela de pedido passou a variar o título por estado (carregando, erro, registrado, acompanhamento, finalizado e cancelado), exibir timeline pública derivada do status sem inventar etapas, bloco de contato a partir de `GET /store` e aviso de registro mesmo sem link do WhatsApp. O erro `NOT_FOUND` ganhou mensagem genérica em `checkoutApiMessage`.

AUD-03 foi corrigido junto com a US-121. O dashboard passou a acionar o interruptor manual com confirmação (`PATCH /admin/store`, já auditado pelo servidor), mostrar o status efetivo calculado pelo backend, contar pedidos por status, exibir produtos disponíveis/total, zonas ativas, prazo de entrega e os cinco pedidos recentes com link para o detalhe. A lista não expõe PII.

AUD-07 foi corrigido junto com a US-122. O checkout passou a ter “Revisão e envio” como etapa final, com total cotado e botão de envio dentro dela, mantendo a etapa de endereço oculta na retirada.

AUD-05 e AUD-11 foram corrigidos junto com a US-124. Os formulários `AdminCatalogForm`, `AdminProductForm`, `AdminStoreView` e `AdminUserForm` passaram a marcar o controle inválido com `aria-invalid="true"` e a apontar, por `aria-describedby`, para a mensagem correspondente. O shell administrativo deixou de repetir `<main id="conteudo">`: as views admin mantêm um único `<main>`, e o shell oferece um contêiner focável com `tabindex="-1"` como destino do skip-link. Todos os alvos `#conteudo` das rotas públicas também receberam `tabindex="-1"`, e o link oculto foi reposicionado para o eixo horizontal a fim de permanecer na ordem de tabulação.

AUD-06, AUD-08 e AUD-09 foram corrigidos junto com a US-125. O painel passou a usar `--text-sub` para “Loja fechada”, sem cor de erro. A home exibe uma barra fixa no mobile quando há itens no carrinho, com contagem e subtotal derivados do catálogo, além de um aviso em superfície neutra de que o pedido não pode ser concluído com a loja fechada; o CTA do cardápio continua disponível. A US-126 foi concluída em seguida, com as decisões de design registradas.

### 5.1 Detalhamento dos achados de severidade Alta

**AUD-01 e AUD-02 — tela de pedido (corrigidos).** O `<h1>` era estático e aparecia mesmo quando `ErrorState` era renderizado. A rota serve simultaneamente como confirmação e acompanhamento, mas não apresentava a timeline prevista no §5.7 nem o contato da loja para o caso de o link do WhatsApp não estar mais na sessão; o aviso “Seu pedido já está salvo” só aparecia quando `handoffUrl` existia. A correção tornou o título dependente do estado, criou `OrderStatusTimeline` com a sequência de entrega/retirada a partir do status público, adicionou o bloco de contato com `GET /store` e moveu o aviso de registro para o resumo, visível em qualquer sucesso. A timeline não expõe `statusHistory` administrativo nem inventa horários para etapas intermediárias; para cancelamento e status fora da sequência, mostra apenas o estado atual.

**AUD-03 — dashboard (corrigido).** O protótipo Admin 01 traz “Controle Master da Loja” com botão ABERTO/FECHADO, “Produtos Disponíveis X / total”, “Bairros Ativos”, “Tempo Médio Delivery” e “Pedidos Recentes”. A versão anterior mantinha status somente leitura, substituía “pedidos recentes” pelo total da fila e mostrava apenas contagem de produtos ativos. A correção expôs a ação de kill switch com confirmação (reaproveitando `PATCH /admin/store`, já auditado pelo servidor), agregou contagem por status, disponíveis/total, prazo de entrega e lista de pedidos recentes com link para o detalhe. Campo sem resposta da API aparece como `—`.

**Evidência da US-121:** verificação em navegador headless após login administrativo, nos viewports de 375 px e 1280 px. Controle da loja, confirmação de abertura/fechamento, indicadores, contagens por status e pedidos recentes presentes; `scrollWidth === clientWidth` nas duas larguras. Prints em `/tmp/opencode/dashboard-us121-375.png` e `/tmp/opencode/dashboard-us121-1280.png` (fora do repositório). O `401` no console é a consulta preventiva do guard de rota antes da autenticação.

**AUD-04 — token inexistente (corrigido).** O valor resolvido de `--bg-base` era vazio no navegador; o input do checkout herdava `background-color: rgba(0, 0, 0, 0)`. O validador conferia cores literais e a consistência entre `tokens.css`/`tokens.json`, mas não conferia a existência de `var(--…)`. A correção trocou o token por `--bg-input` e reforçou o validador (AUD-14).

**AUD-07 — checkout (corrigido).** A implementação tinha quatro etapas (três na retirada) e deixava o botão de envio fora de uma etapa numerada. A correção tornou “Revisão e envio” a etapa final, com o total cotado pelo servidor e o botão dentro da seção; na retirada a etapa de endereço continua oculta e a numeração é ajustada.

**Evidência da US-122:** verificação em navegador headless com produto no carrinho e zona selecionada, nos viewports de 375 px e 1280 px. Título “5. Revisão e envio”, total cotado exibido, botão dentro da seção e `scrollWidth === clientWidth` nas duas larguras. Prints em `/tmp/opencode/checkout-us122-375.png` e `/tmp/opencode/checkout-us122-1280.png` (fora do repositório). O `422` no console é a primeira cotação, feita antes de escolher a zona.

**Evidência da US-124:** verificação em navegador headless nos viewports de 375 px e 1280 px. Há um único `#conteudo` com `tabindex="-1"`, o link “Pular para o conteúdo” recebe foco como primeira parada do Tab ao carregar a página e sua ativação move o foco para o contêiner. Nos formulários de loja, produto e usuário, os campos inválidos expõem `aria-invalid="true"` e `aria-describedby` apontando para mensagens visíveis; `scrollWidth === clientWidth` nas duas larguras. Print do formulário da loja em `/tmp/opencode/a11y-store-us124-375.png` e `/tmp/opencode/a11y-store-us124-1280.png` (fora do repositório). O `401` no console é a consulta preventiva do guard de autenticação antes do login.

**Evidência da US-125:** verificação em navegador headless com `GET /api/v1/store` interceptado como fechado e dois itens no carrinho, nos viewports de 375 px e 1280 px. A barra de carrinho fica visível em 375 px (“Ver carrinho (2)”, subtotal R$ 99,98 e altura mínima de 56 px) e oculta em 1280 px; o aviso de loja fechada aparece com texto nos dois tamanhos, e `scrollWidth === clientWidth` em ambos. Prints em `/tmp/opencode/home-us125-375.png` e `/tmp/opencode/home-us125-1280.png` (fora do repositório).

**Atualização da US-126 em 25/11/2026:** os três achados foram tratados após registro das decisões em `13-decisoes-e-pendencias.md` (D-029 a D-031). AUD-10: privacidade e termos exibem “Rascunho preparado em 25/11/2026. Publicação: pendente de E0”, sem apresentar texto provisório como vigente; a privacidade passou a descrever controlador, dados coletados, finalidades, compartilhamento, retenção, direitos e segurança, e os termos passaram a cobrir registro/idempotência, valores calculados no servidor, entrega/retirada, pagamento, status, acesso ao painel e propriedade intelectual, tudo conforme a implementação real e sem transferir para a loja prestadores, assinaturas ou recursos de modelos externos; a data de publicação efetiva e o canal do titular seguem dependentes de `E0`. AUD-12: os emojis foram removidos dos nomes de categoria do seed, mantendo os slugs, e um teste em `backend/internal/seed/seed_test.go` impede o retorno. AUD-13: o bottom nav de quatro atalhos foi substituído por um drawer completo com todos os itens permitidos pelo papel; o botão “Menu” abre o drawer e move o foco para “Fechar”, Escape e o fundo fecham, e o foco retorna ao botão de origem. Testes de DOM cobrem abertura, fechamento, retorno de foco e trapped focus.

**Evidência adicional da política de privacidade:** verificação em navegador headless em 375 px e 1280 px após a publicação do build no container web: 8 seções, link para `/contato` presente, fonte mínima de 14 px, `scrollWidth === clientWidth` nos dois tamanhos e nenhum erro de console. Prints em `/tmp/opencode/privacidade-375.png` e `/tmp/opencode/privacidade-1280.png` (fora do repositório).

**Evidência adicional dos termos de uso:** verificação em navegador headless em 375 px e 1280 px após a publicação do build no container web: 10 seções, links para `/contato` e `/privacidade` presentes, fonte mínima de 14 px, `scrollWidth === clientWidth` nos dois tamanhos e nenhum erro de console. Prints em `/tmp/opencode/termos-375.png` e `/tmp/opencode/termos-1280.png` (fora do repositório).

**Evidência adicional dos créditos:** verificação em navegador headless em 375 px e 1280 px após a publicação do build no container web: autor “Gildemberg Santos”, 5 seções, links externos com `rel="noopener noreferrer"`, altura mínima de 44 px nos links, `scrollWidth === clientWidth` nos dois tamanhos e nenhum erro de console. Prints em `/tmp/opencode/creditos-375.png` e `/tmp/opencode/creditos-1280.png` (fora do repositório).

**Evidência da US-126:** verificação em navegador headless nos viewports de 375 px e 1280 px após login administrativo. Em 375 px o drawer fica visível com 10 links, o botão “Menu” e o drawer visíveis, a sidebar oculta, o foco no botão “Fechar”, altura mínima de 44 px nos alvos e `scrollWidth === clientWidth` (375). Em 1280 px a sidebar permanece visível e o menu/drawer ficam ocultos, sem scroll horizontal (1280). Prints em `/tmp/opencode/admin-drawer-us126-375.png` e `/tmp/opencode/admin-drawer-us126-1280.png` (fora do repositório). O `401` no console é a consulta preventiva do guard de autenticação antes do login.

---

## 6. Diferenças entre protótipo e documento que **não** são desvios

- **Checkout:** o protótipo tem 3 etapas numeradas e botão “Finalizar no WhatsApp”; o `07` §5.5 define 5 etapas (4 na retirada), incluindo dados do cliente e revisão. A implementação segue o documento e mantém resumo lateral; a diferença em relação ao protótipo é de fidelidade, não desvio do padrão aprovado. AUD-07 era classificado como Média por ser uma lacuna em relação ao documento, não uma regressão do protótipo.
- **Conta:** o Admin 05 do protótipo reúne dados da loja; na produção isso foi separado em `/admin/loja`, enquanto `/admin/conta` mostra usuário/papel da sessão, conforme `07` §6.11.
- **Telas adicionais:** confirmação/acompanhamento, contato, privacidade, termos, adicionais, gestão de loja, fila de pedidos, usuários e auditoria são expansões previstas no `07`, ausentes do protótipo por ser ele apenas uma maquete sem backend.
- **Datas UTC nos filtros do admin:** o `07` §6.8 define UTC para filtros e `America/Fortaleza` para exibição; os rótulos “De (UTC)” são fiéis ao documento. A eventual troca por máscara local é decisão de microcopy, registrada como pendência, não como defeito.
- **Hero da home:** o protótipo usa gradiente laranja e a produção usa superfície neutra com card de destaque. Está dentro dos tokens e atende ao §5.1; trata-se de diferença de fidelidade visual, não de quebra de padrão.

---

## 7. Lacunas do próprio design system

`tokens.css` e `tokens.json` cobrem cores e escala tipográfica, mas não espaçamento, raio, sombra, movimento nem z-index. Valores como 12/16/18/24 px aparecem aplicados de forma consistente no frontend, porém sem token canônico. O `07` §3 e §14 já registram essas lacunas e as decisões abertas (warning/info, emoji, carrinho desktop, navegação mobile do admin). Elas não devem ser resolvidas unilateralmente em código: a decisão de design precisa atualizar `tokens.css`, `tokens.json` e a documentação juntos, conforme `../design-system/README.md`.

---

## 8. Tarefas geradas

As histórias abaixo foram adicionadas ao Épico 12 de `08-backlog.md`. Status inicial: `Planejado`.

| História | Prioridade | Tam. | Achados | Objetivo |
|---|---|---|---|---|
| US-120 | P1 | M | AUD-01, AUD-02 | Tela de pedido com estado correto, timeline e orientação de handoff |
| US-121 | P1 | M | AUD-03 | Dashboard operacional completo, com kill switch e indicadores |
| US-122 | P1 | P | AUD-04, AUD-07 | Checkout com token válido e etapa de revisão explícita |
| US-123 | P1 | P | AUD-14 | Validador de design system rejeita tokens inexistente |
| US-124 | P1 | M | AUD-05, AUD-11 | Acessibilidade dos formulários admin e semântica do shell |
| US-125 | P2 | M | AUD-06, AUD-08, AUD-09 | Status fechado neutro, carrinho fixo no mobile e aviso na home |
| US-126 | P2 | P | AUD-10, AUD-12, AUD-13 | Conteúdo legal datado, emojis do seed e decisões abertas de navegação |

Ordem sugerida: US-123 primeiro (fortalece o gate antes de mexer em UI), depois US-120/US-121/US-122, e por fim US-124/US-125/US-126. As decisões de design sobre emoji, navegação mobile e tokens foram registradas em `13-decisoes-e-pendencias.md` (D-029 a D-031) e implementadas na US-126; apenas a data de publicação e o canal do titular das páginas legais permanecem dependentes de `E0`.

---

## 9. Validação e evidência

Antes de marcar qualquer história como concluída:

1. `make fmt-check`, `make lint`, `make test`, `make docs-check` e `make design-check` passando.
2. Teste automatizado novo ou alterado cobrindo o comportamento (Vitest/Vue Test Utils no frontend; `unittest` em `scripts/`).
3. Para mudanças de UI: print ou evidência equivalente em 375 px e 1280 px, verificando foco visível, alvos de 44 px e ausência de scroll horizontal.
4. Contrato de API alterado? Atualizar `06-contrato-api.md` na mesma mudança.
5. Decisão de padrão visual alterada? Atualizar `tokens.css`, `tokens.json`, `../design-system/README.md` e este documento na mesma mudança, com aprovação de design.

**Evidência da US-120:** verificação em navegador headless com respostas interceptadas de `GET /api/v1/orders/{publicId}` e `GET /api/v1/store`, nos viewports de 375 px e 1280 px. Título por estado, aviso de registro, 6 etapas da timeline, contato da loja e ausência de link de handoff confirmados; `scrollWidth === clientWidth` nas duas larguras. Prints ficaram em `/tmp/opencode/order-us120-375.png` e `/tmp/opencode/order-us120-1280.png` (fora do repositório).

---

## 10. Estado desta auditoria

Nenhum arquivo de código foi alterado pela auditoria. O objetivo deste documento é servir de registro único dos achados e de ponte para as tarefas executáveis no backlog; a execução das histórias deve atualizar a coluna de status e anexar evidência conforme a seção 9.
