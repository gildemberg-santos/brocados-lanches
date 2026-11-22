# Roadmap — Brocados Lanches V1

**Versão:** 1.1  
**Atualizado em:** 22/11/2026  
**Status:** Fase 1 com CI, seed e E2E entregues; preview por PR pendente. Fase 2 com US-020, US-021, US-023 e US-024 concluídas; Fase 3 com US-030 a US-034 concluídas; Fase 4 com US-060, US-061, US-022, US-065 (API) e US-070 a US-072 (API e painel) concluídas. Informações públicas da loja (US-025) e telas do painel para catálogo e loja pendentes.  
**Relacionados:** `03-prd.md`, `08-backlog.md`, `10-operacao-e-lancamento.md`, `13-decisoes-e-pendencias.md`.

---

## 1. Premissas

- A V1 é uma loja, um turno e um painel administrativo.
- O fechamento continua no WhatsApp; não há gateway de pagamento.
- Catálogo, taxas, horários e textos ainda precisam de validação.
- Hospedagem, domínio, CI e política de retenção ainda estão em aberto.
- O time e a velocidade não foram informados; por isso, não há cronograma comprometido.
- O caminho crítico assume um desenvolvedor full-stack; design, conteúdo e operação entram conforme disponibilidade.

> As faixas de esforço abaixo são **relativas e preliminares**. Elas servem para priorização, não para promessa de prazo. Qualquer compromisso exige briefing fechado e equipe definida.

---

## 2. Marcos

| Marco | Nome | Critério de saída |
|---|---|---|
| M0 | Descoberta fechada | Dados de negócio, conteúdo, escopo e decisões bloqueantes validados |
| M1 | Fundação pronta | CI, fixtures, observabilidade base e E2E smoke funcionando |
| M2 | Catálogo navegável | Cardápio real em staging, com estados e personalização |
| M3 | Pedido ponta a ponta | Cotação, pedido persistido, idempotência e WhatsApp em staging |
| M4 | Operação pronta | Login, catálogo admin, fila, status e auditoria |
| M5 | Lançamento | Qualidade, segurança, backup, monitoramento e go-live aprovados |
| M6 | Estabilização | Duas semanas de operação com métricas e correções |

---

## 3. Fases

### Fase 0 — Descoberta e conteúdo

**Objetivo:** eliminar invenção de dados e fechar o que bloqueia o código.

**Entregáveis:**

- catálogo validado (itens, preços, opções, adicionais);
- zonas, taxas, prazo e horários;
- WhatsApp, endereço, Instagram e pagamentos;
- fotos, logo, alergênicos e textos legais;
- hospedagem, domínio, CI e retenção decididos;
- `13-decisoes-e-pendencias.md` atualizado.

**Saída:** M0 assinado. Sem isso, não começar catálogo real.

**Esforço relativo:** G, dependente do cliente.

---

### Fase 1 — Fundação

**Objetivo:** garantir que o esqueleto atual vire base repetível.

**Entregáveis:**

- CI com fmt, lint, testes, docs, design e build;
- seed/reset de desenvolvimento e teste;
- logs, request ID e health;
- Playwright configurado;
- preview por PR;
- documentação de ambiente.

**Saída:** M1. CI, seed/reset e harness Playwright entregues em 20/11/2026; preview por PR pendente.

**Esforço relativo:** M.

---

### Fase 2 — Catálogo

**Objetivo:** cardápio navegável e personalizável.

**Entregáveis:**

- modelo e repositórios;
- `GET /menu` e `GET /delivery-zones`;
- home, cardápio, busca, filtros, detalhe;
- grupos de opções e adicionais;
- estados de carregando/vazio/erro/sucesso;
- imagens e informações da loja.

**Status em 22/11/2026:** US-020 concluída com domínio, validação, repositórios e índices idempotentes; US-021 concluída com `GET /menu` e `GET /delivery-zones`; US-023 e US-024 concluídas com home, cardápio, busca/filtros, detalhe e personalização. Falta US-025 e a validação do catálogo real com o cliente. O total do item é estimado a partir do catálogo; cotação e pedido continuam server-side.

**Saída:** M2. Demonstração em staging sem criação de pedido.

**Esforço relativo:** GG.

**Pode paralelizar com:** E6 (auth/admin) depois que o modelo estiver estável.

---

### Fase 3 — Carrinho, checkout e pedido

**Objetivo:** fechar o fluxo de compra com preço server-side.

**Entregáveis:**

- carrinho persistente — entregue (US-030/US-031);
- cotação;
- formulário de checkout;
- criação de pedido com snapshot;
- idempotência;
- confirmação;
- handoff para WhatsApp;
- resiliência de rede.

**Saída:** M3. Pedido real em staging, com banco e link.

**Esforço relativo:** GG.

**Depende de:** Fase 2.

---

### Fase 4 — Operação

**Objetivo:** permitir que a loja opere sem desenvolvedor.

**Entregáveis:**

- ~~auth admin~~ (US-060 concluída);
- ~~shell e guard~~ (US-061 concluída);
- ~~API CRUD de catálogo~~ (US-022 concluída);
- ~~API de configuração da loja~~ (US-065 concluída);
- ~~API da fila de pedidos~~ (US-070 concluída);
- ~~API de detalhe do pedido~~ (US-071 concluída);
- ~~API de transições de status~~ (US-072 concluída);
- telas de produtos, categorias, adicionais, zonas e loja (US-062 a US-065 no front);
- ~~telas de fila, detalhe e ações contextuais no painel~~ (US-070 a US-072 no front concluídas);
- consulta de auditoria;
- usuários.

**Saída:** M4. Operador treinado em staging.

**Esforço relativo:** GG.

**Pode paralelizar com:** Fase 3 depois de auth e modelo.

---

### Fase 5 — Qualidade, conteúdo e lançamento

**Objetivo:** deixar a V1 apresentável, segura e operável.

**Entregáveis:**

- páginas legais e institucionais;
- SEO/JSON-LD;
- acessibilidade;
- imagens/fontes;
- testes E2E, segurança e performance;
- staging/produção;
- backup/restore;
- monitoramento/alertas;
- go-live.

**Saída:** M5. Checklist de `10-operacao-e-lancamento.md` assinado.

**Esforço relativo:** G.

---

### Fase 6 — Estabilização

**Objetivo:** acompanhar operação real e corrigir atritos.

**Entregáveis:**

- revisão de funil e erros;
- correções de bugs;
- ajustes de conteúdo;
- priorização do próximo ciclo;
- postmortem do lançamento.

**Saída:** M6. Relatório de duas semanas e roadmap seguinte.

**Esforço relativo:** M.

---

## 4. Cenário preliminar de esforço

Cenário apenas para planejamento, **não compromisso**:

| Cenário | Premissa | Faixa indicativa |
|---|---|---|
| Conservador | 1 dev full-stack, design/QA parciais, cliente responde em até 3 dias | 10–14 semanas |
| Acelerado | 2 devs, design/QA dedicados, conteúdo pronto | 7–9 semanas |
| Bloqueado | Conteúdo/hospedagem atrasam | Sem previsão |

Não inclui:

- espera por conteúdo do cliente;
- contratação de hospedagem;
- revisão jurídica;
- desenvolvimento de pagamento online;
- integrações fora do escopo.

A estimativa deve ser refeita após M0.

---

## 5. Caminho crítico

```text
M0 Descoberta
  → Catálogo (E2)
  → Carrinho/Checkout (E3)
  → Pedido/WhatsApp (E4)
  → Operação (E7)
  → Qualidade/Segurança (E9)
  → Deploy/Go-live (E10)
  → M5
```

### 5.1 Paralelização possível

| Trilha | Pode começar | Depende de |
|---|---|---|
| Admin auth/catálogo | Após modelo de catálogo | E2 |
| Conteúdo/SEO | Após cardápio | E2 |
| Acessibilidade | Após primeiras telas | E2 |
| Infra/CI | Imediato | — |
| Testes E2E | Após pedido | E4 |

### 5.2 Gargalos

- validação de catálogo e taxas;
- definição de hospedagem/domínio;
- revisão legal;
- fotos e logo;
- disponibilidade da operação para testar o admin.

---

## 6. Dependências entre fases

| Fase | Depende de | Bloqueia |
|---|---|---|
| 0 | Cliente | 2, 5 |
| 1 | — | 2–5 |
| 2 | 0, 1 | 3, 4 |
| 3 | 2 | 4, 5 |
| 4 | 2, 3 | 5 |
| 5 | 3, 4 | 6 |
| 6 | 5 | Próximo ciclo |

---

## 7. Riscos e mitigação

| Risco | Impacto | Mitigação |
|---|---|---|
| Conteúdo atrasa | Alto | Congelar catálogo mínimo; usar placeholder apenas em dev |
| Escopo cresce | Alto | V1 congelada; mudanças em 13 |
| Operação não testa | Alto | Sessões guiadas; checklist; prazo de feedback |
| Hospedagem indefinida | Alto | Decidir antes de M3; usar staging simples |
| Taxa errada | Alto | Teste por zona; revisão diária no início |
| WhatsApp bloqueado | Médio | Link copiável e pedido persistido |
| Design system incompleto | Médio | Aprovar tokens faltantes antes de escalar UI |
| Time pequeno | Médio | Fatias verticais; cortar P2 |
| Dados pessoais em log | Alto | Teste e revisão de log antes do go-live |

---

## 8. Critérios de corte

Se o prazo apertar, cortar nesta ordem:

1. P2 (analytics, atualização em tempo real, sitemap).
2. Usuários/auditoria? **Não cortar auditoria de mutações**; pode adiar a tela de consulta, mas o registro fica.
3. SEO avançado; manter metadados básicos.
4. Página de créditos.
5. Filtros avançados de pedidos.
6. MFA (se risco aceito e registrado).

Não cortar:

- persistência do pedido;
- preço server-side;
- idempotência;
- auth admin;
- transições válidas;
- logs sem PII;
- backup;
- acessibilidade básica;
- testes críticos.

---

## 9. Governança

- Reunião semanal de progresso e bloqueios.
- Revisão de escopo a cada marco.
- Decisão registrada em `13-decisoes-e-pendencias.md`.
- Mudança de arquitetura vira ADR.
- Critério de pronto por história e por marco.
- Nenhum marco avança com teste crítico falhando.

---

## 10. Próximo passo

1. Fechar M0 com o cliente.
2. Configurar CI e fixtures (M1).
3. Iniciar catálogo somente leitura (M2).
4. Não iniciar checkout antes de catálogo e taxas validados.
