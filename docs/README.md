# Documentação — Brocados Lanches

**Versão:** 1.2  
**Atualizado em:** 25/11/2026  
**Status:** documentação da V1 concluída; fases 1 a 4 entregues no código, Épico 8 parcialmente entregue, épicos 9 e 10 e decisões de negócio em aberto; auditoria de conformidade visual registrada em `14-auditoria-design-system.md`, com correções no Épico 12 do backlog.  
**Entrada recomendada:** este arquivo.

---

## 1. Para que serve esta documentação

Este diretório reúne o planejamento completo da aplicação de pedidos da Brocados Lanches:

- o que será construído e por quê;
- como o sistema será organizado;
- quais dados e contratos existem;
- como o design system e o site de referência foram interpretados;
- quais tarefas precisam ser executadas, em que ordem e com quais critérios de aceite;
- como testar, operar, lançar e manter a V1.

O protótipo HTML continua sendo **referência visual e funcional**, não código de produção. O site público da OlaClick é **referência funcional e de conteúdo**, não a marca nem a base técnica do produto.

---

## 2. Hierarquia das fontes de verdade

Em caso de conflito, vale esta ordem:

1. `../AGENTS.md` — regras de engenharia, stack, segurança e qualidade.
2. `../briefing/chamado.md` — escopo contratado e dados confirmados pelo cliente.
3. `docs/` — decisões e planos detalhados.
4. `../design-system/brocados-lanches-design-system-v2.5.html` — referência visual.
5. Site público de referência — inspiração funcional, nunca fonte automática de dados.

Dentro de `docs/`, a precedência é:

| Assunto | Documento que prevalece |
|---|---|
| Produto, escopo e regras funcionais | `03-prd.md` |
| Arquitetura e decisões técnicas | `04-arquitetura.md` e `adr/` |
| Modelo de dados | `05-modelo-de-dados.md` |
| Contrato HTTP | `06-contrato-api.md` |
| Design, UX, componentes e acessibilidade | `07-design-system-e-experiencia.md` |
| Execução, tarefas e aceite | `08-backlog.md` |
| Testes e qualidade | `09-qualidade-e-testes.md` |
| Operação e lançamento | `10-operacao-e-lancamento.md` |
| Sequência e marcos | `11-roadmap.md` |
| Referência externa | `12-referencia-ola-click.md` |
| Decisões e pendências | `13-decisoes-e-pendencias.md` |
| Auditoria de conformidade visual | `14-auditoria-design-system.md` |

`01-inventario-prototipo.md` e `02-plano-de-producao.md` são documentos históricos da descoberta e da Fase 1. O plano executável atual está em `11-roadmap.md` e `08-backlog.md`.

---

## 3. Mapa dos documentos

| Documento | Para quem | Conteúdo | Status |
|---|---|---|---|
| `01-inventario-prototipo.md` | Produto, design, engenharia | Inventário do HTML, telas, dados fictícios e lacunas | Concluído |
| `02-plano-de-producao.md` | Gestão, engenharia | Plano inicial por fases e riscos | Histórico; complementado por 11 |
| `03-prd.md` | Todos | Problema, personas, escopo, requisitos e critérios de sucesso | Concluído |
| `04-arquitetura.md` | Engenharia | Contexto, camadas, fluxos, segurança, deploy e observabilidade | Concluído |
| `05-modelo-de-dados.md` | Backend, dados | Coleções, campos, índices, invariantes e retenção | Concluído |
| `06-contrato-api.md` | Backend, frontend, QA | Endpoints, envelopes, erros, idempotência e transições | Concluído |
| `07-design-system-e-experiencia.md` | Design, frontend, produto | Tokens, componentes, telas, estados e acessibilidade | Concluído |
| `08-backlog.md` | Produto, engenharia, QA | Épicos, histórias, tarefas, dependências e aceite | Vivo — épicos 8 a 10 em aberto |
| `09-qualidade-e-testes.md` | Engenharia, QA | Estratégia, matriz de testes, gates e cenários críticos | Vivo — cenários de erro, a11y e carga pendentes |
| `10-operacao-e-lancamento.md` | Engenharia, operação | Ambientes, deploy, backup, monitoramento, LGPD e go-live | Concluído |
| `11-roadmap.md` | Gestão, engenharia | Fases, marcos, caminho crítico e riscos | Concluído |
| `12-referencia-ola-click.md` | Produto, design, conteúdo | Análise do site público e matriz de paridade | Concluído |
| `13-decisoes-e-pendencias.md` | Todos | Registro de decisões e perguntas abertas | Vivo |
| `14-auditoria-design-system.md` | Design, frontend, produto | Aderência das telas ao design system, desvios e tarefas de correção | Vivo — Épico 12 no backlog |
| `adr/` | Engenharia | Decisões arquiteturais com contexto e consequências | Vivo |

---

## 4. Caminhos de leitura

### Pessoa nova no projeto

1. `03-prd.md`
2. `04-arquitetura.md`
3. `07-design-system-e-experiencia.md`
4. `11-roadmap.md`
5. `08-backlog.md`

### Backend

1. `04-arquitetura.md`
2. `05-modelo-de-dados.md`
3. `06-contrato-api.md`
4. `09-qualidade-e-testes.md`
5. `adr/`

### Frontend e design

1. `07-design-system-e-experiencia.md`
2. `14-auditoria-design-system.md`
3. `03-prd.md`
4. `06-contrato-api.md`
5. `09-qualidade-e-testes.md`
6. `12-referencia-ola-click.md`

### Produto, operação e gestão

1. `03-prd.md`
2. `08-backlog.md`
3. `10-operacao-e-lancamento.md`
4. `11-roadmap.md`
5. `13-decisoes-e-pendencias.md`

---

## 5. Legenda de status

| Rótulo | Significado |
|---|---|
| **Decidido** | Já existe decisão registrada e ela deve ser seguida. |
| **Proposto** | Há recomendação, mas falta aprovação do cliente, design ou responsável técnico. |
| **Pendente** | A informação não existe ou não foi confirmada. |
| **Bloqueado** | A execução não pode avançar sem resolver a pendência. |
| **Fora da V1** | Foi considerado e deliberadamente postergado. |

Prioridades:

- **P0** — bloqueia o lançamento.
- **P1** — importante para a V1, mas não impede um piloto controlado.
- **P2** — melhoria posterior.

---

## 6. Regras de manutenção

1. Toda mudança de comportamento, contrato, dado ou design atualiza o documento correspondente no mesmo commit.
2. Link relativo precisa resolver; rode `make docs-check`.
3. Não duplique regra: referencie `../AGENTS.md` quando a regra já estiver lá.
4. Marque claramente o que é fato, premissa, decisão ou pendência.
5. Não copie dados do site de referência para produção sem validação do cliente.
6. Não edite o protótipo HTML nem os tokens sem aprovação de design.
7. Decisão nova relevante vira ADR em `adr/` e linha em `13-decisoes-e-pendencias.md`.
8. Documento histórico não é apagado; ele recebe uma nota apontando para o substituto.

---

## 7. Glossário rápido

| Termo | Significado |
|---|---|
| **V1** | Primeira versão pública com pedido persistido, WhatsApp e painel operacional. |
| **Pedido** | Registro persistido no MongoDB, com preços e dados congelados. |
| **Quote** | Cotação server-side de subtotal, taxa e total, sem persistir pedido. |
| **Handoff** | Passagem do pedido persistido para o WhatsApp da loja. |
| **Kill switch** | Interruptor manual que fecha a loja independentemente do horário. |
| **Snapshot** | Cópia dos valores do catálogo no momento do pedido. |
| **P0/P1/P2** | Prioridade de execução. |
| **ADR** | Architecture Decision Record; registro de decisão arquitetural. |
| **LGPD** | Lei Geral de Proteção de Dados. |

---

## 8. Próximo passo

As próximas entregas de código estão em **Épico 12 — Conformidade com o design system**, com as correções priorizadas da auditoria em `14-auditoria-design-system.md`, além dos **Épicos 8 e 10** em `08-backlog.md`. O **Épico 0 — Descoberta e conteúdo** continua bloqueando catálogo real, taxas, horários, contatos, pagamentos e textos legais: o seed é fictício e não deve ir para produção sem validação do cliente.
