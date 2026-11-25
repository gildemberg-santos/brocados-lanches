# Decisões e Pendências — Brocados Lanches

**Versão:** 1.1  
**Atualizado em:** 25/11/2026  
**Status:** registro vivo.  
**Relacionados:** `03-prd.md`, `04-arquitetura.md`, `08-backlog.md`, `adr/`.

---

## 1. Como usar

- **Decisão** é algo já escolhido e que deve ser seguido.
- **Proposta** é uma recomendação que ainda precisa de aprovação.
- **Pendência** é uma informação ausente que pode bloquear execução.
- Toda decisão relevante tem data, impacto e link para o documento/ADR.
- Mudou a decisão? Atualize a linha, o documento afetado e, se for arquitetura, crie/atualize um ADR.
- Não apague histórico; marque como substituída.

---

## 2. Registro de decisões

| ID | Data | Decisão | Status | Impacto | Referência |
|---|---|---|---|---|---|
| D-001 | 18/11/2026 | Stack: Go + Chi, MongoDB 8+, Vue 3 + Vite + TypeScript, Pinia, Vue Router, Tailwind compilado, pnpm, Docker Compose | Decidido | Base de todo o produto | `../AGENTS.md`, `adr/0001-stack-e-monorepo.md` |
| D-002 | 18/11/2026 | O design system v2.5 é a fonte visual; o site OlaClick é referência funcional, não visual | Decidido | Evita copiar marca/layout de terceiros | `adr/0005-design-system-e-ui.md` |
| D-003 | 18/11/2026 | O servidor é a única fonte de verdade para preço, taxa, total, disponibilidade e status | Decidido | Nenhum cálculo confiável no cliente | `../AGENTS.md`, `04-arquitetura.md` |
| D-004 | 18/11/2026 | O pedido é persistido antes do handoff para WhatsApp | Decidido | Pedido sobrevive a falha do link | `adr/0002-persistencia-do-pedido-e-handoff-whatsapp.md` |
| D-005 | 18/11/2026 | Pagamento da V1 é offline: Pix, cartão ou dinheiro na entrega/retirada | Decidido | Sem gateway, sem dado de cartão | `03-prd.md` |
| D-006 | 18/11/2026 | Dinheiro em centavos (`int64`); persistência UTC e exibição em `America/Fortaleza` | Decidido | Evita erro de arredondamento e fuso | `adr/0003-dinheiro-e-fuso.md` |
| D-007 | 18/11/2026 | Admin com sessão em cookie `HttpOnly`, Argon2id, papéis `owner`/`operator` | Decidido | Segurança do painel | `adr/0004-autenticacao-admin.md` |
| D-008 | 18/11/2026 | Adicionar `POST /orders/quote` para cotação server-side sem persistir | Decidido no planejamento | Checkout mostra total autoritativo antes do envio | `adr/0006-cotacao-server-side.md` |
| D-009 | 18/11/2026 | Acompanhamento público usa `publicId` opaco e retorna o mínimo, sem telefone/endereço/observação | Decidido | Reduz exposição de PII | `06-contrato-api.md` |
| D-010 | 18/11/2026 | Adicionar endpoints administrativos de usuários e auditoria para o papel `owner` | Decidido no planejamento | Fecha lacuna do contrato original | `06-contrato-api.md` |
| D-011 | 18/11/2026 | Retirada zera taxa e não exige endereço/zona | Decidido | Regra de negócio | `03-prd.md` |
| D-012 | 18/11/2026 | Carrinho persiste apenas IDs, quantidades e escolhas; nunca preço confiável | Decidido | Evita fraude e dado obsoleto | `03-prd.md`, `07-design-system-e-experiencia.md` |
| D-013 | 18/11/2026 | V1 em português; código/API em inglês | Decidido | Consistência | `../AGENTS.md` |
| D-014 | 18/11/2026 | Não usar CDN, Google Fonts ou analytics sem consentimento em produção | Decidido e implementado (fontes) | Segurança, privacidade e performance | `../AGENTS.md`, `10-operacao-e-lancamento.md`, `07-design-system-e-experiencia.md` §2.2 |
| D-015 | 18/11/2026 | Hospedagem, domínio, MongoDB gerenciado e retenção final ainda não decididos; Render + Atlas estão mapeados apenas como proposta em `../README.md` e `04-arquitetura.md` | Pendente | Bloqueia deploy/go-live | Este documento |
| D-016 | 20/11/2026 | CI no GitHub Actions com jobs de documentação/design, backend, frontend e E2E | Decidido | Pipeline versionado em `.github/workflows/ci.yml`; PR bloqueado por falha | `.github/workflows/ci.yml` |
| D-017 | 22/11/2026 | Total do item no detalhe é estimado a partir dos preços do catálogo | Decidido | A cotação e o pedido recalculam no servidor; nenhum valor do cliente é confiável | `08-backlog.md`, `09-qualidade-e-testes.md`, `AGENTS.md` §8.2 |
| D-018 | 22/11/2026 | No carrinho, a taxa estimada é a menor taxa ativa até a escolha da zona no checkout | Decidido | Evita taxa inventada e deixa claro que o total definitivo é server-side | `07-design-system-e-experiencia.md` §5.4, `08-backlog.md` §7 |
| D-019 | 22/11/2026 | Pedido gravado com snapshot, número sequencial, `publicId` opaco e chave idempotente; replay devolve `200` | Decidido e implementado | Nenhum duplo pedido; histórico não depende do catálogo atual | `06-contrato-api.md` §7, `08-backlog.md` §8 |
| D-020 | 22/11/2026 | O link `wa.me` sai na criação e fica no `sessionStorage` por pedido para reenvio na mesma sessão; não é devolvido no acompanhamento | Decidido e implementado | Handoff recuperável sem expor PII na rota pública | `06-contrato-api.md` §7 e §8, `07-design-system-e-experiencia.md` §5.6 |
| D-021 | 22/11/2026 | O guard do painel é apenas UX (`401` → login com `redirect`, `403` → `/admin/sem-acesso`); a autorização real continua no servidor. No desenvolvimento o proxy do Vite alinha `Origin` ao `Host` da API para não falsificar o CSRF | Decidido e implementado | Painel navegável em dev/produção sem enfraquecer a proteção | `04-arquitetura.md` §5.5, `frontend/vite.config.ts`, `08-backlog.md` §10 |
| D-022 | 22/11/2026 | A fila admin mascara nome/telefone; o detalhe completo exige sessão; transição usa CAS no status e auditoria sem PII | Decidido e implementado | Operação segura com rastro, sem vazar dados na listagem | `06-contrato-api.md` §12, `05-modelo-de-dados.md` §3.6, `08-backlog.md` §11 |
| D-023 | 23/11/2026 | No painel, preço/taxa são digitados em reais e convertidos para centavos no cliente antes do envio; slug, horários e WhatsApp são validados no cliente, mas o servidor continua autoritativo | Decidido e implementado | Evita `float` e PATCH vazio; mantém a API como fonte da verdade | `../AGENTS.md` §7.5 e §8.2, `08-backlog.md` §10 |
| D-024 | 23/11/2026 | Rate limits em janela fixa de 1 minuto, em memória por instância; telefone e sessão entram como hash; `X-Request-Id` correlaciona logs com `Retry-After: 60` em `429` | Decidido e implementado | Atende ao contrato sem depender de nova infraestrutura; contadores precisam ser compartilhados se houver múltiplas réplicas | `06-contrato-api.md` §15, `backend/internal/httpapi/ratelimit.go` |
| D-025 | 23/11/2026 | Metadados por rota implementados diretamente no router, sem adicionar `@unhead/vue`; JSON-LD e `robots.txt` entregues, sitemap e canonical seguem pendentes até domínio definido (D-015) e conteúdo aprovado | Decidido e implementado parcialmente | Evita dependência estrutural e dado de negócio inventado sem abandonar o requisito de SEO do `../AGENTS.md` | `08-backlog.md` §12, `frontend/src/router/meta.ts`, `frontend/src/features/seo/structuredData.ts` |
| D-026 | 25/11/2026 | Dados estruturados são gerados somente com `GET /store` e `GET /menu`, sem URL absoluta inventada; o `robots.txt` usa caminhos relativos e bloqueia rotas privadas | Decidido e implementado | SEO útil sem depender de domínio, telefone ou conteúdo não aprovado | `08-backlog.md` §12, `frontend/public/robots.txt` |
| D-027 | 25/11/2026 | Cliente HTTP usa `AbortController` com timeout configurável (`VITE_HTTP_TIMEOUT_MS`, padrão 10 s) e o cabeçalho sinaliza conexão perdida sem PII; textos legais/institucionais ficam provisórios até aprovação | Decidido e implementado | Resiliência em rede instável sem inventar dado de negócio nem texto legal | `frontend/src/services/http.ts`, `frontend/src/components/SiteHeader.vue`, `08-backlog.md` §12 |
| D-028 | 25/11/2026 | O seed valida a configuração da loja com `domain.StoreSettings.Validate` e usa WhatsApp fictício em E.164; os cenários E2E de falha usam operador dedicado e só criam pedidos no projeto desktop | Decidido e implementado | Um PATCH parcial de `/admin/store` nunca é rejeitado por dado inválido do próprio seed, e os testes não esgotam os rate limits por IP/e-mail | `backend/internal/seed/seed.go`, `frontend/e2e/failure-scenarios.spec.ts`, `09-qualidade-e-testes.md` §8 |

### 2.1 Contrato refletido no AGENTS

As decisões D-008 (`POST /orders/quote`) e D-010 (`/admin/users` e `/admin/audit-logs`) já estão refletidas em `../AGENTS.md` §9.2 e §9.3 e em `06-contrato-api.md`. A verificação de 20/11/2026 não encontrou divergência.

---

## 3. Pendências de negócio

| ID | Pergunta | Por que importa | Bloqueia? | Responsável | Status |
|---|---|---|---|---|---|
| P-001 | Número do chamado, solicitante, prazo e prioridade | Governança e estimativa | Sim | Cliente/gestão | Pendente |
| P-002 | O catálogo do site é o oficial? | Evita inventar cardápio | Sim | Cliente | Pendente |
| P-003 | Quais grupos de opções/adicionais por produto? | Personalização e preço | Sim | Cliente | Pendente; UI pronta e testada com fixtures, mas o seed ainda não contém grupos/adicionais reais |
| P-004 | Quais bairros e taxas valem hoje? | Checkout e entrega | Sim | Cliente | Pendente |
| P-005 | Horários, feriados e prazo reais | Status da loja | Sim | Cliente | Pendente |
| P-006 | WhatsApp, endereço, Instagram e Pix | Contato e handoff | Sim | Cliente | Pendente |
| P-007 | Retirada: endereço, horário e regras | Fluxo de retirada | Sim | Cliente | Pendente |
| P-008 | Pagamento e regra de troco | Checkout | Sim | Cliente | Pendente |
| P-009 | Quem opera o painel e com qual papel? | Auth e treinamento | Sim | Cliente | Pendente |
| P-010 | Fotos, logo e direitos de uso | UI e SEO | Sim | Cliente | Pendente |
| P-011 | Alergênicos e descrições | Obrigação e confiança | Sim | Cliente | Pendente |
| P-012 | Textos legais e canal do titular | LGPD | Sim | Cliente/jurídico | Pendente |
| P-013 | Retenção e anonimização | LGPD | Sim | Cliente/jurídico | Pendente |
| P-014 | Domínio e hospedagem | Deploy/TLS/cookies | Sim | Cliente/engenharia | Pendente |
| P-015 | Provedor de CI | Pipeline | Não para código; sim para lançamento | Engenharia | Resolvido em 20/11/2026 (D-016) |
| P-016 | MongoDB gerenciado ou container | Backup/transações | Sim para produção | Engenharia | Pendente |
| P-017 | WhatsApp link ou API oficial | Handoff | Não para V1; decidir antes do lançamento | Cliente/engenharia | Pendente |
| P-018 | MFA antes ou depois do lançamento | Segurança | Não; risco a aceitar | Cliente/engenharia | Pendente |
| P-019 | Analytics e consentimento | Medição | Não | Cliente/engenharia | Pendente |
| P-020 | Página de créditos permanece? | Conteúdo | Não | Cliente | Pendente |

---

## 4. Decisões técnicas já fechadas

### 4.1 Cotação antes do pedido

- **Problema:** o checkout precisa mostrar total confiável antes de criar o pedido.
- **Decisão:** `POST /orders/quote` valida itens, opções, adicionais e zona, e devolve subtotal, taxa e total sem persistir.
- **Consequência:** o frontend não calcula total; o `POST /orders` recalcula e pode rejeitar se algo mudou.
- **Risco:** duplicação de lógica entre quote e order. Mitigação: uma única função de cálculo no domínio/service.
- **Estado em 22/11/2026:** US-032 implementada e reutilizada pelo `POST /orders` (US-040). `domain.BuildQuote` é a função única de cálculo; o service carrega catálogo, zonas e `store_settings`, e os handlers expõem `POST /api/v1/orders/quote` sem gravar e `POST /api/v1/orders` com persistência. O status efetivo da loja usa o kill switch e, quando existem horários cadastrados, a faixa vigente em `America/Fortaleza`.

### 4.2 Acompanhamento público mínimo

- **Problema:** link público não pode expor PII.
- **Decisão:** retornar número, status, itens (nome/quantidade), total e prazo; nunca link do WhatsApp, telefone, endereço, observação ou `_id`.
- **Consequência:** o cliente vê o essencial sem PII; o link completo com PII só aparece na criação do pedido e pode ser reenviado pelo `sessionStorage` na mesma sessão (D-020).

### 4.3 Usuários e auditoria

- **Problema:** `owner` precisa gerenciar usuários e consultar auditoria, mas o contrato original não listava endpoints.
- **Decisão:** adicionar `/admin/users` e `/admin/audit-logs`, somente owner.
- **Consequência:** mais superfície de auth; testes de autorização obrigatórios.

### 4.4 Seed de desenvolvimento

- **Problema:** o protótipo tem dados fictícios e não pode alimentar produção nem ser confundido com catálogo validado.
- **Decisão:** `make seed` e `make seed-reset` importam `design-system/seed.json`, convertem reais para centavos sem ponto flutuante e gravam `seed.fictitious: true` em cada documento. O comando recusa `APP_ENV=production`, o banco `brocados` e qualquer nome sem sufixo `_dev` ou `_test`.
- **Consequência:** o reset apaga apenas `categories`, `products`, `delivery_zones` e `store_settings`; usuários, sessões, pedidos e auditoria não são tocados.
- **Risco:** o seed sobrescreve campos do catálogo a cada execução. Use apenas em desenvolvimento e teste.

### 4.5 Catálogo de referência no seed de desenvolvimento

- **Problema:** os cinco produtos fictícios do protótipo não permitiam validar apresentação, busca, filtros e personalização com o catálogo real.
- **Decisão:** em 23/11/2026, importar do catálogo público as 5 categorias, 37 produtos, preços, descrições, disponibilidade e as 37 imagens originais de 800 px. Nomes de arquivo seguem o slug do produto em `frontend/public/products/`; o seed passa a preencher `imageUrl` e `imageAlt`. A origem fica registrada em `_meta.catalogo_original` no JSON.
- **Consequência:** o marcador `seed.fictitious: true` continua identificando documentos gerados pelo seed, mas agora convive com campos copiados do catálogo público. Dados de loja, zonas e pedidos seguem fictícios.
- **Risco:** preços, disponibilidade e imagens são um retrato do catálogo na data da extração e exigem validação do cliente antes de produção. Imagens são mantidas no formato original para fidelidade visual; a conversão para WebP/AVIF deve acontecer no preparo para produção.
- Para popular a home sem inventar produtos, `prod_1` (Combo 3 hambúrguer), `prod_5` (BURGER FRESCURINHA) e `prod_24` (Batata G) recebem `featured: true` apenas no seed. O catálogo de origem não informa popularidade; a seleção não deve ser interpretada como “mais pedidos” real.

---

## 5. Pendências que não bloqueiam o início

- Página de créditos.
- Sitemap/robots.
- Atualização em tempo real da fila.
- MFA.
- Analytics.
- Relatórios avançados.
- Impressão/KDS.

Elas entram em P2 ou no roadmap seguinte.

### 5.1 Decisões abertas de design — auditoria de 25/11/2026

Origem: `14-auditoria-design-system.md` e §14 de `07-design-system-e-experiencia.md`. Nenhuma foi decidida pela auditoria; todas precisam de aprovação de design antes de virar padrão.

| Decisão | Situação atual | História afetada |
|---|---|---|
| Navegação mobile do painel: bottom nav de 4 atalhos ou drawer completo | Bottom nav implementado; telas secundárias dependem dos atalhos do dashboard | US-126 |
| Carrinho no desktop: sidebar fixa ou drawer | Não implementado; mobile também pendente | US-125 |
| Emoji na interface e em nomes do seed | Emojis presentes nos dados de desenvolvimento | US-126 |
| Tokens de espaçamento, raio, sombra, movimento e z-index | Valores aplicados no CSS, sem token canônico | Mudança de padrão; exige `tokens.css`, `tokens.json` e documentação |
| Token de aviso/info ou manter apenas aberto/erro | Somente `--status-open` e `--status-error` | US-125 |

---

## 6. Processo de mudança

1. Registrar a pergunta/decisão aqui.
2. Classificar impacto: produto, arquitetura, dados, API, design, operação ou legal.
3. Atualizar o documento dono.
4. Se for arquitetura, criar/atualizar ADR.
5. Se afetar escopo, atualizar `08-backlog.md` e `11-roadmap.md`.
6. Se afetar contrato, atualizar `06-contrato-api.md` e testes.
7. Comunicar no PR.

---

## 7. Histórico

| Data | Mudança |
|---|---|
| 18/11/2026 | Criação do registro; decisões da V1 e pendências iniciais |
| 20/11/2026 | D-016 (GitHub Actions); P-015 resolvido; decisão de seed registrada; §2.1 verificada sem divergência |
| 21/11/2026 | US-020 concluída: domínio, validação, repositórios e índices idempotentes de catálogo; job Backend do CI passa a rodar testes de integração com MongoDB 8 |
| 21/11/2026 | US-021 concluída: API pública de catálogo com `GET /menu` e `GET /delivery-zones`, filtros de ativos/disponibilidade, envelope de erro e testes de service/handler |
| 22/11/2026 | US-023 e US-024 concluídas: home, cardápio, busca/filtros, detalhe e personalização com opções, adicionais e total estimado; testes unitários e smoke E2E de navegação |
| 22/11/2026 | US-032 concluída: cotação server-side sem persistir, com `domain.BuildQuote` como única fonte de cálculo |
| 22/11/2026 | US-033/US-034 e US-040 a US-043 concluídas: checkout, pedido persistido e idempotente, handoff WhatsApp, confirmação com limpeza do carrinho e acompanhamento público; D-019 e D-020 registrados |
| 22/11/2026 | US-060, US-061, US-022 e US-065 concluídas: autenticação, shell/guard do painel, API CRUD de catálogo e API da loja |
| 22/11/2026 | US-070 a US-072 (API) concluídas: fila com PII mascarada, detalhe e transições com CAS/auditoria; D-021 e D-022 registrados |
| 22/11/2026 | US-070 a US-072 (painel) concluídas: fila com filtros/paginação, detalhe e ações de status; E2E cobre pedido → fila → transição |
| 23/11/2026 | US-062 a US-065 (painel) concluídas: CRUD de produtos, categorias, adicionais e zonas, e configuração da loja com kill switch e horários; D-023 registrado |
| 23/11/2026 | US-025, US-066 e US-073 concluídas: `GET /store` com status efetivo/`nextOpeningAt`, gestão de usuários e auditoria somente leitura no painel, com testes e proteção do último owner |
| 23/11/2026 | D-024 e D-025 registrados: rate limits do contrato, scrub de PII em logs, testes de concorrência/WhatsApp, metadados por rota e E2E completo do fluxo de compra; job de E2E versionado no CI; evidência de execução local, cenários críticos e itens de conteúdo/infra continuam pendentes |
| 25/11/2026 | Épico 8 parcial: rotas institucionais/legais e 404 com rodapé, JSON-LD/`robots.txt`, fontes self-hosted, WebP/AVIF, status da loja na home, dashboard com indicadores reais, timeout do cliente HTTP e aviso de conexão; D-014, D-025 atualizados e D-026/D-027 registrados. Textos aprovados, axe, sitemap/canonical e deploy seguem pendentes |
| 25/11/2026 | Gate completo executado: `make test` (407 testes frontend, backend com `-race`, validadores), `make lint`, `make fmt-check`, `make build` e `make test-e2e` aprovados; o E2E passou de 16 para 28 testes com cenários críticos de falha (26 aprovados e 2 pulados no mobile por rate limit compartilhado) |
| 25/11/2026 | Ferramentas alinhadas: `.tool-versions` fixa Node 22.23.5 e `golangci-lint` v2.15.0 passou a rodar no `make lint`; 16 achados do lint completo corrigidos (fechamento de cursor/arquivo e código morto em testes) |
