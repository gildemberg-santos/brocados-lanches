# Decisões e Pendências — Brocados Lanches

**Versão:** 1.3  
**Atualizado em:** 27/11/2026
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
| D-005 | 18/11/2026 | Pagamento da V1 é offline: Pix, cartão ou dinheiro na entrega/retirada | Substituída por D-053 em 27/11/2026 | Sem gateway, sem dado de cartão | `03-prd.md` |
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
| D-018 | 22/11/2026 | No carrinho, a taxa estimada é a menor taxa ativa até a escolha da zona no checkout | Substituída por D-035 em 26/11/2026 | A estimativa antiga antecipava uma cobrança que depende da escolha de entrega/retirada | `07-design-system-e-experiencia.md` §5.4, `08-backlog.md` §7 |
| D-019 | 22/11/2026 | Pedido gravado com snapshot, número sequencial, `publicId` opaco e chave idempotente; replay devolve `200` | Decidido e implementado | Nenhum duplo pedido; histórico não depende do catálogo atual | `06-contrato-api.md` §7, `08-backlog.md` §8 |
| D-020 | 22/11/2026 | O link `wa.me` sai na criação e fica no `sessionStorage` por pedido para reenvio na mesma sessão; não é devolvido no acompanhamento | Decidido e implementado | Handoff recuperável sem expor PII na rota pública | `06-contrato-api.md` §7 e §8, `07-design-system-e-experiencia.md` §5.6 |
| D-021 | 22/11/2026 | O guard do painel é apenas UX (`401` → login com `redirect`, `403` → `/admin/sem-acesso`); a autorização real continua no servidor. No desenvolvimento o proxy do Vite alinha `Origin` ao `Host` da API para não falsificar o CSRF | Decidido e implementado | Painel navegável em dev/produção sem enfraquecer a proteção | `04-arquitetura.md` §5.5, `frontend/vite.config.ts`, `08-backlog.md` §10 |
| D-022 | 22/11/2026 | A fila admin mascara nome/telefone; o detalhe completo exige sessão; transição usa CAS no status e auditoria sem PII | Decidido e implementado | Operação segura com rastro, sem vazar dados na listagem | `06-contrato-api.md` §12, `05-modelo-de-dados.md` §3.6, `08-backlog.md` §11 |
| D-023 | 23/11/2026 | No painel, preço/taxa são digitados em reais e convertidos para centavos no cliente antes do envio; slug, horários e WhatsApp são validados no cliente, mas o servidor continua autoritativo | Decidido e implementado | Evita `float` e PATCH vazio; mantém a API como fonte da verdade | `../AGENTS.md` §7.5 e §8.2, `08-backlog.md` §10 |
| D-024 | 23/11/2026 | Rate limits em janela fixa de 1 minuto, em memória por instância; telefone e sessão entram como hash; `X-Request-Id` correlaciona logs com `Retry-After: 60` em `429` | Decidido e implementado | Atende ao contrato sem depender de nova infraestrutura; contadores precisam ser compartilhados se houver múltiplas réplicas | `06-contrato-api.md` §15, `backend/internal/httpapi/ratelimit.go` |
| D-025 | 23/11/2026 | Metadados por rota implementados diretamente no router, sem adicionar `@unhead/vue`; JSON-LD e `robots.txt` entregues, sitemap e canonical seguem pendentes até domínio definido (D-015) e conteúdo aprovado | Decidido e implementado parcialmente | Evita dependência estrutural e dado de negócio inventado sem abandonar o requisito de SEO do `../AGENTS.md` | `08-backlog.md` §12, `frontend/src/router/meta.ts`, `frontend/src/features/seo/structuredData.ts` |
| D-026 | 25/11/2026 | Dados estruturados são gerados somente com `GET /store` e `GET /menu`, sem URL absoluta inventada; o `robots.txt` usa caminhos relativos e bloqueia rotas privadas | Decidido e implementado | SEO útil sem depender de domínio, telefone ou conteúdo não aprovado | `08-backlog.md` §12, `frontend/public/robots.txt` |
| D-027 | 25/11/2026 | Cliente HTTP usa `AbortController` com timeout configurável (`VITE_HTTP_TIMEOUT_MS`, padrão 10 s) e o cabeçalho sinaliza conexão perdida sem PII; textos legais/institucionais publicados em 25/11/2026 (D-033) | Decidido e implementado | Resiliência em rede instável sem inventar dado de negócio nem texto legal | `frontend/src/services/http.ts`, `frontend/src/components/SiteHeader.vue`, `08-backlog.md` §12 |
| D-028 | 25/11/2026 | O seed valida a configuração da loja com `domain.StoreSettings.Validate`, incluindo `openingHours`; em 25/11/2026 passou a usar o WhatsApp e os horários publicados em `/contato`, além das 30 zonas e taxas fornecidas. Com `APP_ENV=test`, o plano força a loja aberta e sem horários; os cenários E2E de falha usam operador dedicado e só criam pedidos no projeto desktop | Decidido e implementado | Um PATCH parcial de `/admin/store` nunca é rejeitado por dado inválido do próprio seed, os testes não esgotam os rate limits por IP/e-mail, o E2E não depende do relógio e o desenvolvimento reproduz as referências publicadas/fornecidas | `backend/internal/seed/seed.go`, `design-system/seed.json`, `frontend/e2e/failure-scenarios.spec.ts`, `09-qualidade-e-testes.md` §8 |
| D-029 | 25/11/2026 | A navegação mobile do painel usa drawer completo com todos os itens permitidos pelo papel, aberto pelo botão “Menu”, com foco movido para o botão “Fechar”, fechamento por Escape/fundo e trapped focus; a sidebar desktop permanece | Decidido e implementado | Aprovação de design substitui o bottom nav de quatro atalhos; nenhum item administrativo fica inacessível no mobile | `07-design-system-e-experiencia.md` §4.3 e §14, `frontend/src/components/admin/AdminShell.vue` |
| D-030 | 25/11/2026 | Emojis não são usados como ícone nem em nomes de categoria do seed de desenvolvimento; os presentes no protótipo foram removidos e um teste impede o retorno | Decidido e implementado | Dados fictícios não reforçam um padrão visual não aprovado; slugs das categorias permanecem inalterados | `design-system/seed.json`, `backend/internal/seed/seed_test.go`, `07-design-system-e-experiencia.md` §14 |
| D-031 | 25/11/2026 | Tokens de espaçamento, raio, sombra, movimento e z-index permanecem com os valores atuais; textos legais são publicados em 25/11/2026 com data vigente e canal do titular em `/contato` (D-033) | Decidido | Evita mudança de padrão visual sem aprovação e não inventa conteúdo jurídico; `tokens.css`/`tokens.json` seguem consistentes | `07-design-system-e-experiencia.md` §5.8 e §14, `frontend/src/views/PrivacyView.vue`, `frontend/src/views/TermsView.vue` |
| D-032 | 25/11/2026 | A página de créditos permanece na V1 e usa os dados profissionais do currículo fornecido pelo autor; telefone e e-mail pessoais não são publicados | Decidido e implementado | Atende ao pedido do autor sem expor contato pessoal | `07-design-system-e-experiencia.md` §12 e §14, `frontend/src/views/CreditsView.vue`, `frontend/src/router/index.ts` |
| D-033 | 25/11/2026 | Cliente confirmou a publicação de privacidade, termos e créditos em 25/11/2026: páginas passam a exibir versão vigente, o canal do titular é o já publicado em `/contato` (sem CNPJ/razão social inventados) e o `noindex` dos créditos é removido | Decidido e implementado | Textos aprovados não ficam mais como rascunho; nenhum dado empresarial é inventado e os créditos ficam elegíveis para indexação | `07-design-system-e-experiencia.md` §5.8, §12 e §14, `08-backlog.md` US-080/US-126, `frontend/src/views/PrivacyView.vue`, `frontend/src/views/TermsView.vue`, `frontend/src/views/CreditsView.vue`, `frontend/src/router/index.ts` |
| D-034 | 25/11/2026 | A migração `20261125-dados-referencia` reaproveita o plano do seed para atualizar loja, catálogo e zonas sem apagar documentos; em produção exige `-apply`, `APP_ENV=production`, `MONGO_DB=brocados` e confirmação explícita, preserva o kill switch atual e usa URI apenas por variável de ambiente | Preparada, não executada | Atende ao pedido de levar os dados de referência à produção sem transformar o seed em comando de produção nem registrar credenciais no repositório; documentos fora do plano são desativados e o rollback é backup/restauração ou migração corretiva | `backend/scripts/migrate/20261125-dados-referencia/`, `05-modelo-de-dados.md` §7, `10-operacao-e-lancamento.md` §6.2, `README.md` §4 |
| D-035 | 26/11/2026 | O carrinho mostra somente o subtotal; taxa de entrega e total aparecem no checkout, depois da escolha de entrega/retirada e da zona, calculados no servidor | Decidido e implementado | A cobrança antecipada sugeria taxa mesmo para retirada; a cotação autoritativa continua no `POST /orders/quote` | `07-design-system-e-experiencia.md` §5.4, `08-backlog.md` §7, `frontend/src/components/CartSummary.vue`, `frontend/src/views/CartView.vue` |
| D-036 | 26/11/2026 | A cotação no checkout só é enviada quando a entrega tem `zoneId` ou quando a forma é retirada; após recarregar, o resumo pede a zona antes de calcular | Decidido e implementado | O recarregamento reinicia o formulário sem zona e disparava `422` com `zoneId` vazio; o servidor continua validando a requisição | `07-design-system-e-experiencia.md` §5.5, `08-backlog.md` §7, `frontend/src/views/CheckoutView.vue` |
| D-037 | 26/11/2026 | Após o pedido gravado, o checkout salva no `localStorage` forma de entrega, zona, nome, telefone, endereço e pagamento para preencher o próximo pedido; observação e valor de troco ficam de fora, e as preferências podem ser apagadas em `/privacidade` | Decidido e implementado | O cliente pediu para reaproveitar as escolhas do “Finalizar pedido” automaticamente e depois excluiu observação e troco; zona inativa é descartada ao carregar, e nada é enviado ao servidor além do pedido | `07-design-system-e-experiencia.md` §5.5 e §5.8, `08-backlog.md` US-035, `10-operacao-e-lancamento.md` §10, `frontend/src/features/checkout/preferences.ts`, `frontend/src/views/CheckoutView.vue`, `frontend/src/views/PrivacyView.vue` |
| D-038 | 26/11/2026 | O servidor grava um perfil de conveniência por telefone normalizado depois que o pedido é gravado, sem observação nem troco, com opt-out no checkout, retenção de 24 meses e resumo no detalhe administrativo autenticado; a restrição de consulta pública foi revisada em D-039 | Decidido e implementado | Reconhece cliente recorrente no painel e agiliza pedidos futuros sem transformar marketing em finalidade; falha na escrita não impede o pedido e o replay idempotente não reescreve o perfil | `05-modelo-de-dados.md` §3.11, `06-contrato-api.md` §7 e §12, `08-backlog.md` US-036, `10-operacao-e-lancamento.md` §10, `backend/internal/domain/customer_profile.go`, `backend/internal/repository/customer.go`, `frontend/src/views/CheckoutView.vue` |
| D-039 | 26/11/2026 | O checkout pode consultar o perfil de conveniência por telefone em `POST /customers/profile`, somente com a opção habilitada e telefone brasileiro completo. A resposta é mínima (`found` e, quando houver, nome, telefone, endereço, entrega/retirada, zona sem taxa e pagamento); opt-out retorna `found: false`. Limites: 10/min por IP e 3/min por telefone em hash, com `Cache-Control: no-store` | Decidido e implementado | O cliente pediu para identificar o cliente primeiro e completar os demais dados. A consulta é explícita no checkout, não devolve histórico/observação/troco/taxa e usa resposta uniforme para não revelar existência de perfil; preço e taxa seguem autoritativos no servidor | `06-contrato-api.md` §5.1 e §15, `07-design-system-e-experiencia.md` §5.5 e §5.8, `frontend/src/views/PrivacyView.vue`, `backend/internal/httpapi/customer.go`, `backend/internal/service/customer.go`, `frontend/src/services/customers.ts` |
| D-040 | 26/11/2026 | A atualização de status usa WebSocket: `WS /orders/{publicId}/events` (público, pelo `publicId`) e `WS /admin/orders/events` (sessão administrativa). Os eventos trazem só identificador, número, status e timestamp; o cliente refaz a leitura HTTP para obter os dados completos | Decidido e implementado | Substitui recarga manual e polling curto na fila e no acompanhamento; a fila e o detalhe administrativos e a tela pública do cliente passam a reagir a criações e transições persistidas. O hub é em memória e por instância, sem PII nem valor monetário no canal | `06-contrato-api.md` §8.2 e §12.1, `04-arquitetura.md` §6.6, `08-backlog.md` US-074, `backend/internal/realtime/hub.go`, `frontend/src/services/realtime.ts` |
| D-041 | 26/11/2026 | O histórico do cliente é identificado por token opaco emitido pelo servidor, guardado no navegador e enviado em `Authorization: Bearer` a `GET /orders/history`; não há login nem consulta por telefone. Na criação, o token é derivado por HMAC-SHA256 da chave de idempotência com `ORDER_ACCESS_SECRET`, e só seu hash é gravado. Apagar os dados do navegador remove o acesso naquele dispositivo | Decidido e implementado | Atende ao pedido de rever pedidos a qualquer momento sem criar cadastro nem expor PII na listagem; a tela `/meus-pedidos` usa o mesmo formato mínimo do acompanhamento e leva ao `publicId`. Pedidos antigos são associados no replay com o token apresentado | `06-contrato-api.md` §7 e §8.1, `05-modelo-de-dados.md` §3.6, `08-backlog.md` Épico 13, `07-design-system-e-experiencia.md` §5.9, `frontend/src/features/orders/orderAccessToken.ts`, `frontend/src/views/MyOrdersView.vue` |
| D-042 | 26/11/2026 | Webhook de notificações com `GET/PATCH /admin/integrations/webhook` restrito a `owner`, qualquer URL absoluta, com qualquer esquema e host (sem allowlist, sem HTTPS obrigatório e sem bloqueio de SSRF) e envio best-effort assíncrono com um POST, timeout de 3 s e sem retry. Payload exato: `{number, message}`, ampliado para `{number, storeNumber, message}` em D-046; evento genérico `order.created`/`order.updated` cobre qualquer status persistido, inclusive futuro | Decidido e implementado | O cliente pediu integração com automações externas e aceitou explicitamente o risco de SSRF e de envio de PII ao destino; não bloquear criação/transição por falha externa. A URL fica em `store_settings.webhook`, a auditoria registra só `active` e `urlConfigured`, e o replay idempotente não dispara | `06-contrato-api.md` §11.1, `05-modelo-de-dados.md` §3.1 e §3.9, `08-backlog.md` US-075, `10-operacao-e-lancamento.md` §10, `backend/internal/platform/webhook/notifier.go`, `frontend/src/views/admin/AdminWebhookView.vue` |
| D-043 | 26/11/2026 | Telefone do cliente precisa ser brasileiro possível: DDD em uso pela Anatel, celular com 9 dígitos iniciados em `9` ou fixo com 8 dígitos iniciados de `2` a `5`, nacional ou em E.164 com DDI `55`. A mesma regra vale para `POST /orders` e `POST /customers/profile` | Decidido e implementado | Pedido de teste chegou ao gateway com DDI duplicado e celular de 9 dígitos iniciado em `8`; o WhatsApp respondeu que o destinatário não existe. Rejeitar antes da gravação evita perder a notificação e criar perfil de conveniência com número inválido. O `number` do webhook continua sendo o telefone do cliente | `06-contrato-api.md` §5.1 e §7, `07-design-system-e-experiencia.md` §5.5, `09-qualidade-e-testes.md` §5.21, `backend/internal/domain/phone.go`, `frontend/src/features/checkout/checkout.ts` |
| D-044 | 27/11/2026 | A mensagem do pedido usa negrito do WhatsApp, emojis de seção, valores em `pt-BR` com vírgula decimal, telefone e CEP formatados e blocos separados por linha em branco; transições acrescentam `✅ Status: <rótulo>` | Decidido e implementado | O texto anterior era um bloco corrido e difícil de ler no celular; o novo layout vale para o link `wa.me` e para o `message` do webhook, sem alterar os dados congelados do pedido nem o payload do webhook | `06-contrato-api.md` §11.1, `backend/internal/service/whatsapp.go`, `backend/internal/service/webhook_message.go` |
| D-045 | 27/11/2026 | Imagens de produto são enviadas pelo painel em MongoDB, na coleção `uploads`, e servidas por `GET /api/v1/uploads/{id}`; `POST /api/v1/admin/uploads` aceita multipart de até 2 MiB em JPEG/PNG/WebP/AVIF, detectados pela assinatura do arquivo. O campo de imagem do formulário aceita URL absoluta e caminho relativo | Decidido e implementado | O formulário exigia URL absoluta e rejeitava os caminhos `/products/...` já salvos pelo seed; a loja pediu upload pela própria tela. MongoDB evita volume novo e ajuste de proxy, e as imagens acompanham o backup do banco. Nenhum tipo executável é servido: o `Content-Type` do cliente é ignorado e a auditoria registra só tipo/tamanho | `06-contrato-api.md` §5.2 e §10.1, `05-modelo-de-dados.md` §3.13, `07-design-system-e-experiencia.md` §6.3, `08-backlog.md` US-067, `backend/internal/domain/upload.go`, `frontend/src/features/admin-catalog/imageUpload.ts` |
| D-046 | 27/11/2026 | O webhook passa a enviar `{number, storeNumber, message}`: `number` continua sendo o telefone congelado do cliente e `storeNumber` é o `whatsapp` da loja, lidos juntos do documento `store_settings` em um único POST por evento. Sem número da loja, o envio é ignorado com aviso sem PII | Decidido e implementado | A loja pediu que a automação notifique cliente e loja em toda criação/transição. Manter `number` e acrescentar campo é uma mudança aditiva; o formato de lista quebraria quem já lê `number`, e dois POST duplicariam entregas e o contrato de um evento. A configuração e o número vêm da mesma leitura para evitar duas consultas; o campo é validado na gravação da loja, e a falha externa continua best-effort | `06-contrato-api.md` §11.1, `05-modelo-de-dados.md` §3.9, `04-arquitetura.md` §6.7, `09-qualidade-e-testes.md` §5.22, `backend/internal/domain/webhook.go`, `backend/internal/platform/webhook/notifier.go`, `backend/internal/service/admin_webhook.go` |
| D-047 | 27/11/2026 | Monitoramento de erros com Sentry somente para falhas inesperadas: `5xx`/panics na API e `5xx`/timeout/erro de rota no navegador; DSN por ambiente (`SENTRY_DSN`/`VITE_SENTRY_DSN`), sem tracing, Session Replay, auth token ou upload de source maps. `BeforeSend` sanitiza user, request, PII e segredos nos dois lados; CSP libera apenas `https://*.sentry.io` | Decidido e implementado | Necessidade de diagnosticar falhas fora do horário da loja sem transformar telemetria em novo canal de dados pessoais. Sem DSN o projeto segue funcionando como antes; o DSN do navegador fica no bundle, por isso não recebe permissão administrativa e aponta para projeto separado da API. O cliente confirmou que os DSNs recebidos são apenas referência da documentação, não valores para versionar | `adr/0007-monitoramento-de-erros.md`, `06-contrato-api.md` §14, `09-qualidade-e-testes.md` §9, `10-operacao-e-lancamento.md` §8.5, `frontend/src/views/PrivacyView.vue`, `backend/internal/httpapi/sentry.go`, `frontend/src/features/monitoring/sentry.ts` |
| D-048 | 27/11/2026 | O `owner` escolhe no painel quais eventos disparam webhook: criação e cada um dos sete status mapeados. A seleção fica em `store_settings.webhook.events`, é validada no servidor e filtrada pelo notificador antes do POST; documento sem o campo equivale a todos habilitados, e lista vazia desativa todos sem mudar `active`. O cabeçalho `X-Brocados-Event` segue genérico (`order.created`/`order.updated`) e a chave específica do status é usada só no filtro e na auditoria | Decidido e implementado | A loja quer escolher o que enviar às automações sem receber notificações indesejadas. Manter o cabeçalho genérico preserva integrações já configuradas; a seleção explícita permite silenciar etapas específicas, e a configuração legada continua com o comportamento anterior até o primeiro `PATCH`. Nenhuma chave contém PII | `AGENTS.md`, `06-contrato-api.md` §11.1, `05-modelo-de-dados.md` §3.1, `07-design-system-e-experiencia.md` §6.7.1, `08-backlog.md` US-076, `09-qualidade-e-testes.md` §5.23, `backend/internal/domain/webhook.go`, `backend/internal/platform/webhook/notifier.go`, `frontend/src/views/admin/AdminWebhookView.vue` |
| D-049 | 27/11/2026 | A mensagem do webhook inclui `🔗 Ver pedido no painel: <ADMIN_BASE_URL>/admin/pedidos/<id>` quando `ADMIN_BASE_URL` está configurada. O padrão em desenvolvimento é `http://localhost:5173`, e em produção a variável deve ser HTTPS; sem valor, o link é omitido. O link usa o `id` hexadecimal interno, não o `publicId` opaco | Decidido e implementado | A automação precisa abrir diretamente o pedido no painel; o `publicId` não é aceito pela rota administrativa e o domínio de produção ainda não foi definido. A base fica no servidor, não no payload do cliente, e uma base ausente não quebra o envio | `06-contrato-api.md` §11.1, `backend/internal/config/config.go`, `backend/internal/service/webhook_message.go`, `backend/.air.toml` |
| D-050 | 27/11/2026 | O frontend carrega o container `GTM-NLTLDV4N` com o snippet padrão do Google Tag Manager em todas as visitas, sem banner de opt-in. O nonce do snippet vem de `AsyncLocalStorage` no Vite e de `$request_id` com `sub_filter` no Nginx; o CSP libera só os domínios necessários do Google. Nenhum dado do pedido, PII ou token é enviado ao `dataLayer` pelo código da loja | Decidido e implementado com waiver explícito | O solicitante pediu a instalação e autorizou expressamente a exceção às regras de consentimento e de script de terceiro, aceitando o tratamento de dados técnicos pelo Google. O nonce por resposta preserva o CSP sem `'unsafe-inline'`; a política de privacidade passa a declarar o GTM e seu carregamento em todas as visitas | `adr/0008-google-tag-manager.md`, `../AGENTS.md` §8.4, `03-prd.md`, `05-modelo-de-dados.md`, `08-backlog.md` US-084, `10-operacao-e-lancamento.md` §8.5, `frontend/index.html`, `frontend/csp.ts` |
| D-051 | 27/11/2026 | `DELETE /api/v1/admin/orders/{id}` permite ao `owner` excluir fisicamente um pedido em qualquer status, após confirmação no painel. A resposta é `204`; a auditoria guarda `number`, `status`, `fulfillment` e `totalCents` antes da exclusão, sem PII. A exclusão não dispara webhook nem apaga `customer_profiles`, e publica `order.deleted` somente no canal administrativo | Decidido e implementado | O painel precisava limpar pedidos de teste e lançamentos incorretos. Exclusão física é irreversível, por isso fica restrita a `owner` e exige confirmação explícita; qualquer status é elegível para não impedir essa limpeza. O evento público não é enviado: o acompanhamento passa a receber `404` na próxima consulta. A chave de idempotência desaparece com o documento, então um replay com a mesma chave pode criar um pedido novo | `../AGENTS.md` §9.3, `06-contrato-api.md` §12, `05-modelo-de-dados.md` §3.6 e §3.9, `08-backlog.md` US-077, `backend/internal/service/admin_orders.go`, `frontend/src/views/admin/AdminOrdersView.vue` |
| D-052 | 27/11/2026 | Mudanças reais de status exibem `StatusToast` (`aria-live`, 4 s, fechamento manual) e som curto gerado por Web Audio no acompanhamento público e na fila/detalhe administrativos. O disparo usa `order.updated`, compara com o status conhecido e deduplica por pedido/status/`updatedAt`; `connected`, reconexão, carga inicial e pedidos fora da página atual não avisam; a transição feita pelo próprio operador avisa na confirmação da chamada e o evento do servidor não duplica o som. Sem áudio disponível ou com autoplay bloqueado, a notificação visual segue | Decidido e implementado | O solicitante pediu notificação e som ao alterar status nos dois lados. Gerar o tom no navegador evita asset, rede e PII; reagir ao evento em vez do refetch evita duplicatas e ruído na abertura da página. O toast usa apenas tokens do design system e o áudio é best-effort para não afetar o pedido | `../AGENTS.md`, `07-design-system-e-experiencia.md` §4.1, §5.7, §6.8 e §10, `08-backlog.md` US-078, `frontend/src/components/StatusToast.vue`, `frontend/src/features/orders/statusNotifications.ts`, `frontend/src/features/orders/notificationSound.ts` |
| D-053 | 27/11/2026 | Pix com cobrança após a gravação: o checkout apenas escolhe `pix`; `POST /orders` persiste o pedido e a confirmação `/pedido/:publicId` monta localmente BR Code estático (EMV QRCPS-MPM) com tag `00=01`, merchant account `26`, `52=0000`, `53=986`, `54` com o `totalCents` gravado, `58=BR`, nome/cidade higienizados, `62.05=***` e CRC16-CCITT/FFFF, exibindo QR PNG em Data URI e “Pix Copia e Cola”. Sem gateway/PSP, sem confirmação automática e sem TXID por pedido nesta fase. A chave, o nome e a cidade são configurados pelo `owner` em `/admin/integracoes/pix` e publicados por `GET /store` somente quando ativos | Decidido e implementado; sequência ajustada em 27/11/2026 a pedido do solicitante, substitui D-005 | O solicitante pediu BR Code/QR e configurabilidade no painel e depois determinou que o pedido deve ser enviado antes de cobrar o cliente. Mover o QR para a confirmação elimina a dependência da cotação para pagamento e garante que o valor cobrado seja o snapshot gravado; `GET /orders/{publicId}` devolve `paymentMethod` para que a confirmação reconheça o Pix após recarregar, sem expor PII. `TXID` permanece `***`: não há promessa de conciliação nesta fase. `expectedTotalCents` continua apenas como guarda otimista da cotação exibida; em conflito o servidor não grava e o checkout recota antes de nova tentativa. A disponibilidade é validada no servidor: com `pix.active = false`, `payment.method = "pix"` é rejeitado com `422` mesmo em aba antiga. A dependência `qrcode` só desenha a imagem; TLV, sanitização e CRC ficam no código do projeto. Auditoria não expõe valores e `GET /store` não devolve chave/nome/cidade quando desativada, e nenhuma chave real é versionada | `../AGENTS.md`, `03-prd.md`, `05-modelo-de-dados.md` §3.1, `06-contrato-api.md` §3, §8, §11.2, `07-design-system-e-experiencia.md` §4.2, `08-backlog.md` US-038, `frontend/src/features/pix/brcode.ts`, `frontend/src/features/pix/qrcode.ts`, `frontend/src/components/PixPaymentPanel.vue`, `frontend/src/views/OrderView.vue`, `frontend/src/views/admin/AdminPixView.vue` |

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
| P-006 | WhatsApp, endereço, Instagram e chave Pix | Contato e handoff | Sim para os dados reais | Cliente | Parcial: WhatsApp/endereço/Instagram publicados em `/contato` e mecanismo de configuração do Pix entregue em 27/11/2026 (D-053); a chave definitiva ainda deve ser informada pelo proprietário no painel |
| P-007 | Retirada: endereço, horário e regras | Fluxo de retirada | Sim | Cliente | Pendente |
| P-008 | Pagamento e regra de troco | Checkout | Sim | Cliente | Pendente |
| P-009 | Quem opera o painel e com qual papel? | Auth e treinamento | Sim | Cliente | Pendente |
| P-010 | Fotos, logo e direitos de uso | UI e SEO | Sim | Cliente | Pendente |
| P-011 | Alergênicos e descrições | Obrigação e confiança | Sim | Cliente | Pendente |
| P-012 | Textos legais e canal do titular | LGPD | Sim | Cliente/jurídico | Resolvida em 25/11/2026 (D-033): textos publicados e canal do titular em `/contato` |
| P-013 | Retenção e anonimização | LGPD | Sim | Cliente/jurídico | Resolvida em 25/11/2026 (D-033): prazos publicados na política de privacidade |
| P-014 | Domínio e hospedagem | Deploy/TLS/cookies | Sim | Cliente/engenharia | Pendente |
| P-015 | Provedor de CI | Pipeline | Não para código; sim para lançamento | Engenharia | Resolvido em 20/11/2026 (D-016) |
| P-016 | MongoDB gerenciado ou container | Backup/transações | Sim para produção | Engenharia | Pendente |
| P-017 | WhatsApp link ou API oficial | Handoff | Não para V1; decidir antes do lançamento | Cliente/engenharia | Pendente |
| P-018 | MFA antes ou depois do lançamento | Segurança | Não; risco a aceitar | Cliente/engenharia | Pendente |
| P-019 | Analytics e consentimento | Medição | Não | Cliente/engenharia | Resolvido em 27/11/2026 (D-050): GTM em todas as visitas, com waiver explícito |
| P-020 | Página de créditos permanece? | Conteúdo | Não | Cliente | Resolvida em 25/11/2026 (D-032): mantida com créditos de desenvolvimento a partir do currículo do autor; mídia da loja segue pendente |

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
- **Consequência:** o marcador `seed.fictitious: true` continua identificando documentos gerados pelo seed, mas agora convive com campos copiados do catálogo público e fornecidos para o projeto. Contato e horários da loja passaram a seguir o status publicado em `/contato`, e as 30 zonas e taxas foram fornecidas em 25/11/2026; pedidos seguem fictícios.
- **Risco:** preços, disponibilidade e imagens são um retrato do catálogo na data da extração e exigem validação do cliente antes de produção. Imagens são mantidas no formato original para fidelidade visual; a conversão para WebP/AVIF deve acontecer no preparo para produção.
- Para popular a home sem inventar produtos, `prod_1` (Combo 3 hambúrguer), `prod_5` (BURGER FRESCURINHA) e `prod_24` (Batata G) recebem `featured: true` apenas no seed. O catálogo de origem não informa popularidade; a seleção não deve ser interpretada como “mais pedidos” real.

---

## 5. Pendências que não bloqueiam o início

- Sitemap/robots.
- Notificação ativa ao cliente fora do site (push do navegador ou WhatsApp no status `ready`).
- MFA.
- Analytics.
- Relatórios avançados.
- Impressão/KDS.

Elas entram em P2 ou no roadmap seguinte.

### 5.1 Decisões abertas de design — auditoria de 25/11/2026

Origem: `14-auditoria-design-system.md` e §14 de `07-design-system-e-experiencia.md`. As decisões registradas em 25/11/2026 estão consolidadas em D-029 a D-031; itens ainda sem aprovação seguem marcados como pendentes.

| Decisão | Situação atual | História afetada |
|---|---|---|
| Navegação mobile do painel: bottom nav de 4 atalhos, bottom nav completo ou drawer | Decidido em 25/11/2026 (D-029): drawer completo com todos os itens por papel; bottom nav removido | US-126 (AUD-13, concluída) |
| Carrinho no desktop: sidebar fixa ou drawer | Barra fixa entregue no mobile na US-125; versão desktop não implementada e ainda pendente | US-125 (parcial) |
| Emoji na interface e em nomes do seed | Decidido em 25/11/2026 (D-030): não usar como ícone; emojis removidos do seed e protegidos por teste | US-126 (AUD-12, concluída) |
| Tokens de espaçamento, raio, sombra, movimento e z-index | Decidido em 25/11/2026 (D-031): manter valores atuais; padronização segue pendente e exige `tokens.css`, `tokens.json` e documentação | Mudança de padrão |
| Token de aviso/info ou manter apenas aberto/erro | Mantido apenas aberto/erro; “Loja fechada” usa cor neutra desde a US-125 | US-125 (resolvido sem token novo) |
| Data de publicação de privacidade/termos e canal do titular | Resolvido em 25/11/2026 (D-033): versão vigente publicada em 25/11/2026 e canal do titular em `/contato` | US-126 (AUD-10, concluída) |

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
| 25/11/2026 | Épico 12 concluído com D-029/D-030/D-031: drawer completo no painel mobile, emojis removidos do seed, tokens mantidos e textos legais mantidos como rascunho até `E0` |
| 25/11/2026 | Política de privacidade detalhada com base em `menuaqui.com.br/privacidade`, ajustada ao tratamento real da Brocados; controlador completo, canal do titular e publicação definitiva seguem pendentes de `E0` |
| 25/11/2026 | Termos de uso detalhados com base em `menuaqui.com.br/termos`, ajustados a pedido gravado no servidor, valores calculados pela loja, pagamento na entrega/retirada e acesso restrito ao painel; versão vigente segue pendente de `E0` |
| 25/11/2026 | Página de créditos mantida e preenchida com D-032: nome, trajetória e links profissionais do autor a partir do currículo fornecido; contato pessoal omitido e mídia da loja ainda pendente |
| 25/11/2026 | D-033: cliente confirmou a publicação de privacidade, termos e créditos; avisos de rascunho removidos, canal do titular em `/contato` e `noindex` dos créditos liberado |
| 25/11/2026 | D-028 atualizado: o seed passou a carregar contato, tempo de entrega e `openingHours` do status publicado em `/contato`, 30 zonas e taxas fornecidas, e plano de teste sem dependência de horário |
| 25/11/2026 | D-034 registrada: migração versionada preparada para produção com confirmação explícita e sem credenciais no repositório; não executada contra produção |
| 27/11/2026 | D-047 registrada: Sentry de erros no backend e no frontend, com sanitização antes do envio, CSP compatível, validador estático no CI e nenhum DSN versionado |
| 27/11/2026 | D-052 registrada: notificação visual e som nas mudanças de status do acompanhamento e do painel, com deduplicação por evento e áudio best-effort |
| 27/11/2026 | D-053 registrada e ajustada: Pix cobrado na confirmação do pedido gravado por BR Code/QR local com total persistido, TXID `***` e configuração owner-only; D-005 substituída e chave definitiva ainda pendente em P-006 |
