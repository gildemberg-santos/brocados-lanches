# Decisões e Pendências — Brocados Lanches

**Versão:** 1.1  
**Atualizado em:** 21/11/2026  
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
| D-014 | 18/11/2026 | Não usar CDN, Google Fonts ou analytics sem consentimento em produção | Decidido | Segurança, privacidade e performance | `../AGENTS.md`, `10-operacao-e-lancamento.md` |
| D-015 | 18/11/2026 | Hospedagem, domínio, MongoDB gerenciado e retenção final ainda não decididos | Pendente | Bloqueia deploy/go-live | Este documento |
| D-016 | 20/11/2026 | CI no GitHub Actions com jobs de documentação/design, backend, frontend e E2E | Decidido | Pipeline versionado em `.github/workflows/ci.yml`; PR bloqueado por falha | `.github/workflows/ci.yml` |
| D-017 | 22/11/2026 | Total do item no detalhe é estimado a partir dos preços do catálogo | Decidido | A cotação e o pedido recalculam no servidor; nenhum valor do cliente é confiável | `08-backlog.md`, `09-qualidade-e-testes.md`, `AGENTS.md` §8.2 |

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

### 4.2 Acompanhamento público mínimo

- **Problema:** link público não pode expor PII.
- **Decisão:** retornar número, status, itens (nome/quantidade), total, prazo e link do WhatsApp; nunca telefone, endereço, observação ou `_id`.
- **Consequência:** o cliente vê o essencial; a loja continua sendo contatada pelo WhatsApp.

### 4.3 Usuários e auditoria

- **Problema:** `owner` precisa gerenciar usuários e consultar auditoria, mas o contrato original não listava endpoints.
- **Decisão:** adicionar `/admin/users` e `/admin/audit-logs`, somente owner.
- **Consequência:** mais superfície de auth; testes de autorização obrigatórios.

### 4.4 Seed de desenvolvimento

- **Problema:** o protótipo tem dados fictícios e não pode alimentar produção nem ser confundido com catálogo validado.
- **Decisão:** `make seed` e `make seed-reset` importam `design-system/seed.json`, convertem reais para centavos sem ponto flutuante e gravam `seed.fictitious: true` em cada documento. O comando recusa `APP_ENV=production`, o banco `brocados` e qualquer nome sem sufixo `_dev` ou `_test`.
- **Consequência:** o reset apaga apenas `categories`, `products`, `delivery_zones` e `store_settings`; usuários, sessões, pedidos e auditoria não são tocados.
- **Risco:** o seed sobrescreve campos do catálogo a cada execução. Use apenas em desenvolvimento e teste.

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
