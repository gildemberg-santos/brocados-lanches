# Backlog — Brocados Lanches V1

**Versão:** 1.2  
**Atualizado em:** 25/11/2026  
**Status:** Épico 1 em andamento (CI, seed e harness E2E entregues; preview por PR pendente); Épico 2 com US-020, US-021, US-023, US-024 e US-025 concluídas; Épico 3 com US-030 a US-034 concluídas; Épico 4 com US-040 a US-043 concluídas; Épico 6 com US-060 a US-066 concluídas (API e painel); Épico 7 com US-070 a US-073 concluídas (API e painel); Épico 8 parcial (rotas institucionais/legais com texto provisório, SEO básico, mídia otimizada e ajustes de acessibilidade entregues; textos aprovados, axe, sitemap e canonical pendentes); Épico 12 com US-123, US-120, US-121 e US-122 concluídas a partir da auditoria de conformidade visual.

**Relacionados:** `03-prd.md`, `04-arquitetura.md`, `05-modelo-de-dados.md`, `06-contrato-api.md`, `07-design-system-e-experiencia.md`, `09-qualidade-e-testes.md`, `11-roadmap.md`, `13-decisoes-e-pendencias.md`, `14-auditoria-design-system.md`.

---

## 1. Como usar

- Cada épico tem objetivo, histórias e tarefas técnicas.
- Histórias têm prioridade, tamanho, dependências e critérios de aceite.
- Tarefas técnicas são um checklist de implementação; não substituem os critérios de aceite.
- Um item só entra em “pronto para desenvolvimento” quando o Épico 0 estiver resolvido para o assunto.
- Toda história de código nasce com teste; toda mudança de contrato atualiza `06-contrato-api.md`.
- Não misture refatoração ampla com funcionalidade.

### 1.1 Prioridades

| Código | Significado |
|---|---|
| P0 | Bloqueia lançamento ou segurança |
| P1 | Necessário para V1 de qualidade, mas não impede piloto controlado |
| P2 | Pós-lançamento |

### 1.2 Tamanhos

| Código | Esforço relativo | Regra |
|---|---|---|
| P | Até 2 dias | Pequeno e isolado |
| M | 3–5 dias | Uma área, poucos arquivos |
| G | 6–10 dias | Múltiplas camadas ou tela completa |
| GG | > 10 dias | Deve ser dividido antes de iniciar |

Tamanhos são relativos e não comprometem prazo. Ver `11-roadmap.md`.

### 1.3 Status

`Planejado`, `Pronto`, `Em andamento`, `Bloqueado`, `Concluído`, `Cancelado`.

### 1.4 Definition of Ready

- Objetivo e critério de aceite claros.
- Dependências resolvidas ou explicitamente aceitas.
- Dados de negócio validados quando a história usa catálogo/taxa/horário.
- Contrato de API acordado quando há endpoint novo.
- Design de referência identificado.
- Estratégia de teste definida.

### 1.5 Definition of Done

- Código implementado na menor mudança coerente.
- Teste automatizado novo/alterado passando.
- `make fmt`, `make lint`, `make test`, `make docs-check` e `make design-check` passando.
- Documentação atualizada.
- Acessibilidade básica verificada.
- Nenhuma PII em log.
- Revisão de código aprovada.
- Evidência de teste ou print quando for UI.

---

## 2. Caminho crítico e fatias de entrega

### 2.1 Caminho crítico

```text
E0 (dados/conteúdo)
  → E2 (catálogo)
  → E3 (carrinho/checkout)
  → E4 (pedido/WhatsApp)
  → E7 (operação de pedidos)
  → E9 (qualidade/segurança)
  → E10 (deploy/go-live)
```

E6 (admin auth/catálogo) pode começar em paralelo a E2/E3, mas E7 depende de E6. E8 pode rodar em paralelo a partir de E2.

### 2.2 Fatias verticais

| Fatia | Conteúdo | Critério de demonstração |
|---|---|---|
| A — Catálogo somente leitura | E2 + parte de E8 | Cardápio real navegável em staging |
| B — Carrinho e cotação | E3 | Carrinho persistente e total server-side |
| C — Pedido e WhatsApp | E4 | Pedido persistido, idempotente e handoff |
| D — Admin de catálogo | E6 | Operador altera produto/zona/loja com login |
| E — Operação e acompanhamento | E5 + E7 | Fila, status e link público funcionando |
| F — Lançamento | E8 + E9 + E10 | Checklist de go-live completo |

---

## 3. Resumo dos épicos

| Épico | Objetivo | Prioridade | Tamanho | Depende de |
|---|---|---|---|---|
| E0 — Descoberta e conteúdo | Validar dados e fechar decisões | P0 | G | — |
| E1 — Fundação e entrega | CI, fixtures, observabilidade e E2E base | P0/P1 | M | E0 parcial |
| E2 — Catálogo | Modelo, API e storefront do cardápio | P0 | GG | E0, E1 |
| E3 — Carrinho e checkout | Carrinho, cotação e formulário | P0 | GG | E2 |
| E4 — Pedido e WhatsApp | Persistência, idempotência e handoff | P0 | GG | E3 |
| E5 — Acompanhamento | Link público e timeline | P1 | M | E4 |
| E6 — Admin auth e catálogo | Login, papéis e gestão | P0 | GG | E1, E2 |
| E7 — Operação de pedidos | Fila, detalhe, status e auditoria | P0 | GG | E4, E6 |
| E8 — Conteúdo, SEO, A11y e LGPD | Institucional, metadados, acessibilidade e privacidade | P1 | G | E2, E6 |
| E9 — Qualidade e segurança | Testes, E2E, segurança e performance | P0 | GG | E2–E7 |
| E10 — Deploy e go-live | Ambientes, backup, monitoramento e lançamento | P0 | G | E9 |
| E11 — Pós-lançamento | Medição, correções e evolução | P2 | M | E10 |
| E12 — Conformidade com o design system | Corrigir os desvios da auditoria de 25/11/2026 | P1/P2 | M | E4, E6, E8 |

---

## 4. Épico 0 — Descoberta e conteúdo

**Objetivo:** eliminar dúvidas de negócio que tornariam o catálogo e o checkout inventados.

**Fora do escopo:** implementar código de produto.

| ID | História | Prioridade | Tam. | Aceite |
|---|---|---|---|---|
| US-001 | Validar catálogo, preços, opções e adicionais | P0 | G | Cliente confirma item por item; grupos de opções e adicionais documentados; divergências do site de referência resolvidas |
| US-002 | Validar zonas, taxas, prazo e horários | P0 | M | Lista de bairros/taxas e horários aprovada; feriados definidos; retirada confirmada |
| US-003 | Validar contatos, endereço, pagamento e WhatsApp | P0 | M | WhatsApp, endereço, Instagram, Pix/cartão/dinheiro e regra de troco confirmados |
| US-004 | Aprovar marca, fotos, alergênicos e textos | P0 | M | Logo, fotos, descrições, alergênicos e textos legais aprovados com direito de uso |
| US-005 | Fechar hospedagem, domínio, CI e retenção | P0 | M | Responsáveis e decisões registrados em `13-decisoes-e-pendencias.md` |

**Tarefas:**

- [ ] Preencher identificação, prazo, prioridade e solicitante em `../briefing/chamado.md`.
- [ ] Percorrer o snapshot de `12-referencia-ola-click.md` com o cliente.
- [ ] Separar dado confirmado, dado a confirmar e dado fictício.
- [ ] Registrar aprovação de conteúdo e imagens.
- [ ] Atualizar `03-prd.md` e `05-modelo-de-dados.md` se o modelo mudar.

---

## 5. Épico 1 — Fundação e entrega

**Objetivo:** transformar o esqueleto da Fase 1 em base repetível de desenvolvimento e validação.

| ID | História | Prioridade | Tam. | Aceite |
|---|---|---|---|---|
| US-010 | Pipeline de CI | P0 | M | PR roda fmt, lint, testes, docs, design-check e build; falha bloqueia merge |
| US-011 | Observabilidade base | P1 | M | Request ID, logs JSON, health checks e correlação; sem PII |
| US-012 | Fixtures e seed de desenvolvimento | P0 | M | `brocados_dev` e `brocados_test` separados; reset idempotente; dados fictícios marcados |
| US-013 | Harness E2E | P1 | M | Playwright sobe ambiente isolado e executa um smoke de cardápio |

**Status em 25/11/2026:** US-010 e US-012 concluídos. US-013 com harness pronto e smoke da home + navegação para o cardápio em desktop/mobile, além do fluxo de compra e do acesso por papel. US-011 entregue no esqueleto (request ID, logs JSON e health checks), com teste de correlação e mascaramento de PII em `backend/internal/httpapi/router_test.go`. Falta apenas preview por PR.

**Tarefas:**

- [x] Monorepo, Docker Compose, health checks e build offline.
- [x] Testes iniciais de backend, frontend, docs e design system.
- [x] Hook de pre-commit.
- [x] Configurar CI no GitHub Actions (D-016).
- [x] Criar script de seed/reset com teste.
- [x] Configurar Playwright no CI.
- [ ] Definir preview por PR.
- [x] Documentar variáveis e secrets de CI.

---

## 6. Épico 2 — Catálogo

**Objetivo:** entregar categorias, produtos, opções, adicionais e zonas com API e storefront.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-020 | Modelo e repositórios de catálogo | P0 | G | E0 | Coleções, índices e testes de repositório com Mongo real |
| US-021 | API pública de catálogo | P0 | M | US-020 | `GET /menu` e `GET /delivery-zones` filtram ativos; testes de handler |
| US-023 | Home e cardápio | P0 | G | US-021 | Home, busca, filtros, grid e esgotado; estados completos; testes |
| US-024 | Detalhe e personalização | P0 | G | US-021 | Opções obrigatórias, min/máx, adicionais e preço; testes |
| US-025 | Informações da loja | P1 | M | US-021 | **Concluída** — `GET /store` com status efetivo/`nextOpeningAt`, home com status e painel com preview; testes |

**Status em 23/11/2026:** US-020 a US-025 concluídas. O catálogo público tem `GET /menu` e `GET /delivery-zones`; a home e o cardápio consomem catálogo e status; o detalhe tem personalização; `GET /store` calcula status efetivo e `nextOpeningAt` em `America/Fortaleza`, e o painel exibe o preview na configuração da loja.

**Tarefas técnicas:**

- [x] Definir tipos de domínio para categoria, produto, opção, adicional e zona.
- [x] Implementar validação de slug, preço, ordem e vínculos.
- [x] Criar índices idempotentes.
- [x] Implementar `GET /menu`, `GET /delivery-zones`.
- [x] Implementar filtros de ativos e disponibilidade.
- [x] Criar stores e services tipados no frontend.
- [x] Criar `ProductCard`, `CategoryPill`, `OptionGroup`, `AddonRow`, `PriceText`.
- [x] Implementar busca sem acento/caixa.
- [x] Testar produto esgotado, categoria vazia e erro de leitura no backend.
- [x] Testar busca sem resultado, erro de rede, retry e estados vazios no frontend.
- [x] Atualizar contrato da API.

---

## 7. Épico 3 — Carrinho e checkout

**Objetivo:** permitir montar o pedido, cotar no servidor e preencher o checkout sem perder dados.

**Status em 26/11/2026:** US-030 concluída (`useCartStore` com schema `v1`, persistência de IDs/escolhas/quantidades, limpeza de versões antigas e validação contra o catálogo), US-031 concluída (`/carrinho` com `CartLine`, `CartSummary` e `QuantityStepper`, quantidade, remoção, limpeza, estado vazio e subtotal), US-032 concluída (`POST /orders/quote` com validação de itens, opções, adicionais, zona e status da loja; totais calculados no servidor por `domain.BuildQuote`; sem persistência) e US-033/US-034 concluídas (`/checkout` com entrega/retirada, zona, endereço, pagamento, troco, validação por campo, foco no primeiro erro, cotação no servidor, bloqueio de envio e estados de carregamento/erro/sucesso). O CTA “Continuar” do carrinho leva ao checkout. Desde 26/11/2026 o carrinho não carrega zonas nem mostra taxa estimada; taxa e total vêm da cotação da US-032 depois da escolha de entrega/retirada e zona (D-035), sem requisição incompleta após recarregar a página (D-036). Desde 26/11/2026, após o pedido gravado o checkout salva no navegador entrega/retirada, zona, nome, telefone, endereço e pagamento para o próximo pedido; observação e troco ficam de fora, e a zona salva é descartada se deixar de existir (D-037). O limite de taxa previsto no contrato ainda não está ativo no esqueleto.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-030 | Store do carrinho | P0 | M | US-024 | Persiste IDs/quantidades/escolhas; versão; testes |
| US-031 | UI do carrinho | P0 | M | US-030 | Quantidade, remover, limpar, vazio e subtotal sem taxa; testes |
| US-032 | API de cotação | P0 | M | US-021 | `POST /orders/quote` valida e devolve totais; sem persistir; testes |
| US-033 | Formulário de checkout | P0 | G | US-031, US-032 | Delivery/retirada, zona, endereço, pagamento, troco e validação; testes |
| US-034 | Revisão e envio | P0 | M | US-033 | Total cotado, bloqueio de envio, estados de erro e sucesso |
| US-035 | Preferências do checkout | P1 | M | US-033, US-034 | Salvar automaticamente após pedido gravado; observação e troco de fora; apagar em `/privacidade`; testes |

**Tarefas técnicas:**

- [x] Definir schema versionado do carrinho.
- [x] Implementar `useCartStore` com setup e testes.
- [x] Validar IDs contra o catálogo carregado.
- [x] Criar `CartLine`, `CartSummary`, `QuantityStepper`.
- [x] Implementar cotação no backend com as mesmas regras do pedido.
- [x] Criar `CheckoutForm` com validação e erros por campo (helpers puros em `features/checkout`).
- [x] Normalizar telefone no cliente e no servidor.
- [x] Implementar troca de delivery/retirada e troca de zona.
- [x] Testar loja fechada, zona inativa, item esgotado e troco inválido.
- [x] Preservar carrinho em falha.
- [x] Salvar preferências do checkout apenas após pedido gravado, sem observação nem troco (US-035).
- [x] Validar preferências ao carregar e oferecer exclusão em `/privacidade` (US-035).

---

## 8. Épico 4 — Pedido e WhatsApp

**Objetivo:** criar o pedido persistido, idempotente e com handoff recuperável.

**Status em 25/11/2026:** US-040 concluída (`POST /orders` recalcula o total com `domain.BuildQuote`, persiste snapshot com número via contador atômico, `publicId` opaco, índices únicos e `201` com `Location`), US-041 concluída (chave de 16 a 128 caracteres, `requestHash` canônico, replay `200` com `Idempotent-Replayed` e conflito `409`), US-042 concluída (link `wa.me` gerado após a persistência, com reenvio na mesma sessão), US-043 concluída (`/pedido/:publicId` mostra número/status/totais e limpa o carrinho somente após sucesso). US-044 parcial: o E2E confirma que o link de handoff continua disponível após recarregar; falta injetar falha de envio/retry.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-040 | Criação de pedido | P0 | G | US-032 | Servidor recalcula e persiste snapshot; `201` + `Location`; testes |
| US-041 | Idempotência | P0 | M | US-040 | Mesma chave/corpo não duplica; conflito retorna `409`; testes |
| US-042 | Handoff WhatsApp | P0 | M | US-040 | URL com mensagem; falha preserva pedido; reenvio; testes |
| US-043 | Confirmação e limpeza do carrinho | P0 | M | US-040 | Carrinho limpo só após sucesso; confirmação com número e link |
| US-044 | Resiliência de rede | P0 | M | US-043 | Retry idempotente; dados preservados; mensagens claras |

**Tarefas técnicas:**

- [x] Implementar `POST /orders` com validação completa.
- [x] Gerar `number` via contador atômico.
- [x] Gerar `publicId` com 128 bits.
- [x] Calcular e congelar snapshot.
- [x] Criar índice único de `idempotencyKey` e gravar `requestHash` canônico.
- [x] Retornar `200` no replay idempotente.
- [x] Montar `whatsappUrl` sem logar.
- [x] Criar tela de confirmação e botão de reenvio.
- [ ] Testar falha de Mongo, falha de link e duplo clique (injeção de falha de handoff ainda pendente no E2E).
- [x] Testar concorrência com a mesma chave (repository + E2E de replay e conflito).

---

## 9. Épico 5 — Acompanhamento

**Objetivo:** permitir consulta pública segura do pedido.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-050 | API pública de acompanhamento | P1 | M | US-040 | Concluída: status/itens/total sem PII sensível; limite de 60/min por IP; testes |
| US-051 | Tela de acompanhamento | P1 | M | US-050 | Timeline, atualização, não encontrado e erro; testes |
| US-052 | Reenvio de WhatsApp | P1 | P | US-042 | Link completo disponível na confirmação imediata; após recarregar, contato da loja com número do pedido |

**Tarefas técnicas:**

- [x] Definir resposta pública mínima (US-050).
- [x] Ocultar telefone, endereço, observação e `_id` (US-050).
- [x] Implementar rate limit de 60/min por IP (US-050).
- [x] Atualizar contrato.
- [ ] Criar `OrderStatusTimeline` como componente dedicado (US-051).
- [ ] Testar ID inválido, ID inexistente e status final.

---

## 10. Épico 6 — Admin: autenticação e catálogo

**Objetivo:** dar autonomia segura à operação.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-060 | Autenticação admin | P0 | G | E1 | **Concluída** — login/logout/me, Argon2id, cookie, rate limit, CSRF; testes |
| US-061 | Shell e guard do admin | P0 | M | US-060 | **Concluída** — rotas lazy, guard, 401/403, menu por papel; testes |
| US-022 | API admin de catálogo | P0 | G | US-020, US-060 | **Concluída** — CRUD paginado de categorias/produtos/adicionais/zonas; validação, autorização, CSRF, auditoria e testes |
| US-062 | Gestão de produtos | P0 | G | US-022, US-061 | **Concluída** — lista com busca/filtro/paginação, formulário com preço em reais, opções, adicionais, disponibilidade/destaque/ativo e testes |
| US-063 | Categorias e adicionais | P0 | G | US-022, US-061 | **Concluída** — CRUD no painel com slug, ordem, ativo e testes; preço do adicional convertido para centavos |
| US-064 | Zonas e taxas | P0 | M | US-022, US-061 | **Concluída** — CRUD no painel, taxa digitada em reais e enviada em centavos, ativo e testes |
| US-065 | Configuração da loja (API) | P0 | M | US-061 | **Concluída** — `GET/PATCH /admin/store`, validação de horários em `America/Fortaleza`, E.164, auditoria, tela com kill switch e intervalos por dia, e testes |
| US-066 | Usuários e papéis | P1 | M | US-060 | **Concluída** — owner cria/edita/desativa; último owner protegido; auditoria sem PII; testes |

**Tarefas técnicas:**

- [x] Implementar Argon2id e política de senha (US-060).
- [x] Implementar sessão com TTL e rotação (US-060).
- [x] Implementar CSRF por Origin/Referer (US-060).
- [x] Criar middleware de auth e autorização (US-060).
- [x] Rate limit de login por IP+e-mail (US-060).
- [x] Implementar CRUD admin de categorias, produtos, adicionais e zonas (US-022).
- [x] Criar `AdminShell` e a tela de login (US-061).
- [x] Implementar `GET/PATCH /admin/store` com validação e auditoria (US-065).
- [x] Criar componentes de filtro, listagem e paginação do catálogo (US-062 a US-064).
- [x] Criar formulários com validação por campo e conversão de centavos (US-062 a US-065).
- [ ] Criar `DataTable` e `AdminPageHeader` se novas telas justificarem a abstração.
- [x] Implementar auditoria nas mutações de catálogo e configuração da loja (US-022/US-065).
- [x] Implementar `GET/POST/PATCH /admin/users` com autorização owner, auditoria e proteção do último owner (US-066).
- [x] Criar telas admin de usuários e conta (US-066).
- [x] Testar acesso anônimo e papel errado no guard e no E2E (US-061).
- [x] Testar sessão encerrada por logout no E2E; expiração ociosa/absoluta permanece coberta por teste de unidade.

---

## 11. Épico 7 — Operação de pedidos

**Objetivo:** permitir que a operação receba, acompanhe e finalize pedidos sem erro.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-070 | Fila de pedidos | P0 | G | US-040, US-060 | **Concluída** — API com lista/filtros/paginação e painel com busca, filtros, paginação e PII mascarado |
| US-071 | Detalhe do pedido | P0 | M | US-070 | **Concluída** — API e tela com snapshot, cliente, endereço, pagamento e histórico |
| US-072 | Transições de status | P0 | G | US-071 | **Concluída** — Matriz válida, `409` inválido, motivo de cancelamento, auditoria e ações contextuais no painel |
| US-073 | Auditoria | P1 | M | US-072 | **Concluída** — Owner consulta mutações sem PII; filtros por ator/entidade/ação/data; tela somente leitura e testes |
| US-074 | Atualização em tempo real | P2 | M | US-070 | Fila atualiza sem recarregar ou polling curto; opcional |

**Tarefas técnicas:**

- [x] Implementar listagem com filtros allowlist.
- [x] Mascarar PII na lista; detalhe autorizado.
- [x] Implementar máquina de estados no domínio.
- [x] Gravar `statusHistory`, timestamps e auditoria.
- [x] Criar ações contextuais por status e as telas admin de fila/detalhe.
- [x] Testar todas as transições e concorrência.
- [x] Implementar `GET /admin/audit-logs` e a tela somente leitura (US-073).

---

## 12. Épico 8 — Conteúdo, SEO, acessibilidade e LGPD

**Objetivo:** deixar a V1 apresentável, encontrável, acessível e legalmente defensável.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-080 | Páginas institucionais e legais | P1 | M | E0 | Parcial: `/contato`, `/privacidade`, `/termos`, `/creditos`, 404 e rodapé entregues; contato usa apenas `GET /store`; privacidade e termos detalhados em 25/11/2026 a partir de modelos externos adaptados ao fluxo real e publicados em 25/11/2026 (D-033), com canal do titular em `/contato`; créditos preenchidos com dados do currículo do autor (D-032) e publicados sem `noindex` (D-033) |
| US-081 | SEO e dados estruturados | P1 | M | US-023 | Parcial: título/descrição/robots por rota, JSON-LD `Restaurant`/`Menu`/`Product` e `robots.txt` entregues; faltam sitemap e canonical (domínio pendente) |
| US-082 | Acessibilidade | P1 | G | US-023, US-033, US-061 | Parcial: rótulos/`aria-describedby` no checkout, alvos de 44 px no rodapé, aviso de conexão e foco preservados; auditoria axe pendente |
| US-083 | Imagens e performance | P1 | M | US-023 | Concluída no seed: WebP/AVIF, dimensões, lazy, fontes self-hosted e validação no design-check; fotos definitivas dependem do cliente |
| US-084 | Consentimento e analytics | P2 | M | US-080 | Sem PII; opt-in; eventos documentados |

**Status em 25/11/2026:** o cliente HTTP central aplica timeout configurável (`VITE_HTTP_TIMEOUT_MS`, padrão 10 s) e o cabeçalho exibe aviso de conexão perdida sem PII; a home mostra status efetivo, prazo e contato a partir de `GET /store`; o dashboard do painel usa indicadores reais com estado neutro quando a fonte falha.

**Tarefas técnicas:**

- [x] Criar rotas e páginas institucionais/legais com conteúdo provisório (textos finais dependem de aprovação).
- [x] Implementar metadados por rota sem dependência adicional (`src/router/meta.ts`).
- [x] Gerar JSON-LD `Restaurant`, `Menu` e `Product` com dados de `GET /store` e `GET /menu`.
- [x] Publicar `robots.txt` protegendo admin, carrinho, checkout e acompanhamento.
- [ ] Gerar sitemap e URL canônica com domínio definido.
- [ ] Rodar axe nas telas críticas.
- [x] Corrigir rótulos e associações de erro no checkout e alvos de toque no rodapé.
- [x] Self-host de fontes com licenças e `font-display: swap`.
- [x] Gerar variantes WebP/AVIF das imagens do seed e validar no `make design-check`.
- [x] Retenção documentada e canal do titular publicado em `/contato` em 25/11/2026 (D-033).

---

## 13. Épico 9 — Qualidade e segurança

**Objetivo:** provar que os fluxos críticos funcionam e que o sistema resiste a erro e abuso.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-090 | Testes de backend | P0 | G | E2–E7 | Service, handler e repository com Mongo real; `-race` |
| US-091 | Testes de frontend | P0 | G | E2–E7 | Stores, componentes e services; Vitest |
| US-092 | E2E crítico | P0 | G | US-040, US-072 | Fluxo completo em `e2e/checkout.spec.ts` e cenários de falha em `e2e/failure-scenarios.spec.ts`; execução depende de MongoDB local |
| US-093 | Segurança e privacidade | P0 | M | US-060, US-040 | Concluída: rate limits do contrato, CSRF, auth, enumeração e scrub de PII em logs; testes cobrem correlação e mascaramento |
| US-094 | Performance e carga | P1 | M | E2–E7 | Orçamentos de `03-prd.md` medidos; gargalos corrigidos |

**Tarefas técnicas:**

- [x] Criar matriz de testes de `09-qualidade-e-testes.md`.
- [ ] Configurar testcontainers.
- [x] Configurar Playwright.
- [x] Cobrir concorrência de idempotência, falha de persistência, handoff indisponível, sessão expirada e PII em logs.
- [x] Revisar logs e mascaramento.
- [x] Cobrir no E2E loja fechada, esgotado, zona inativa, retirada, troco, idempotência, transição inválida e sessão encerrada.
- [ ] Injetar falha de handoff/retry e testar troco inválido e duplo clique no E2E.
- [ ] Rodar teste de carga básico.
- [ ] Registrar evidências no PR.

---

## 14. Épico 10 — Deploy, operação e go-live

**Objetivo:** publicar com domínio, TLS, backup, monitoramento e rollback.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-100 | Staging/preview | P0 | M | E9 | Deploy automático, banco isolado e smoke test |
| US-101 | Produção e domínio | P0 | G | US-100 | TLS, same-origin, headers, cookies `Secure`, índices |
| US-102 | Backup e restauração | P0 | M | US-101 | Backup diário e restore testado com evidência |
| US-103 | Monitoramento e alertas | P0 | M | US-101 | Health, 5xx, latência, pedidos e backup |
| US-104 | Go-live | P0 | M | US-102, US-103 | Checklist assinado; operação treinada; rollback definido |

**Tarefas técnicas:**

- [ ] Definir provedor e topologia.
- [ ] Configurar secrets.
- [ ] Configurar domínio/TLS.
- [ ] Criar índices em produção.
- [ ] Testar restore.
- [ ] Criar dashboards e alertas.
- [ ] Escrever runbooks.
- [ ] Treinar operação.
- [ ] Executar checklist de `10-operacao-e-lancamento.md`.

---

## 15. Épico 11 — Pós-lançamento

**Objetivo:** medir, corrigir e evoluir com dados reais.

| ID | História | Prioridade | Tam. | Aceite |
|---|---|---|---|---|
| US-110 | Revisão das duas primeiras semanas | P2 | M | Funil, erros, pedidos perdidos e atritos revisados |
| US-111 | Melhorias de conversão | P2 | M | Hipóteses priorizadas com dados, sem PII |
| US-112 | Roadmap seguinte | P2 | P | Itens fora da V1 reavaliados |

---

## 16. Épico 12 — Conformidade com o design system

**Origem:** `14-auditoria-design-system.md`, auditoria realizada em 25/11/2026.  
**Objetivo:** eliminar os desvios entre as telas implementadas e `07-design-system-e-experiencia.md` sem alterar o protótipo ou os tokens sem aprovação de design.  
**Fora do escopo:** mudanças de padrão visual. As decisões abertas do §14 do documento de design continuam em `13-decisoes-e-pendencias.md`.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-120 | Tela de pedido conforme o planejado | P1 | M | US-050, US-051 | Título varia por estado; timeline e contato da loja presentes; aviso de pedido registrado mesmo sem link do WhatsApp; 404 não revela dados; testes cobrem sucesso, erro e recarga |
| US-121 | Dashboard operacional completo | P1 | M | US-065, US-070 | Kill switch acionável com auditoria; pedidos por status; produtos disponíveis/total; prazo de entrega; pedidos recentes com link para o detalhe; sem PII; testes e evidência em 375/1280 px |
| US-122 | Checkout conforme o §5.5 | P1 | P | US-033, US-034 | Nenhum `var(--…)` inexistente; etapa “Revisão e envio” explícita com total cotado e botão de envio; retirada continua com etapa de endereço oculta; testes atualizados |
| US-123 | Gate de tokens inexistente | P1 | P | US-010 | `scripts/check_design_system.py` falha quando o frontend usa `var(--token)` não declarado; teste unitário cobre token válido e inválido; `make design-check` continua passando |
| US-124 | Acessibilidade dos formulários admin e do shell | P1 | M | US-061, US-062, US-065, US-066 | Concluída em 25/11/2026: erros associados por `aria-describedby`/`aria-invalid`, landmark `<main>` único, skip-link funcional, foco visível e alvos de 44 px preservados; testes de DOM |
| US-125 | Status fechado e carrinho persistente | P2 | M | US-023, US-065 | Concluída em 25/11/2026: “Loja fechada” em cor neutra no painel, barra fixa de carrinho no mobile com contagem e subtotal, aviso de que o pedido não pode ser concluído na home; prints em 375/1280 px |
| US-126 | Conteúdo e decisões pendentes de design | P2 | P | US-080, E0 | Concluída em 25/11/2026: páginas legais com data de rascunho e publicação pendente de E0, posteriormente publicadas em 25/11/2026 (D-033); emojis removidos do seed; drawer completo no painel mobile após decisão registrada; tokens mantidos; testes de conteúdo, seed e drawer |

**Tarefas técnicas:**

- [x] AUD-01: título dinâmico em `frontend/src/views/OrderView.vue` com microcopy do `07` §9.
- [x] AUD-02: componente/timeline pública e bloco de contato em `OrderView.vue`, reaproveitando `GET /store`.
- [x] AUD-03: ação de kill switch no dashboard via `PATCH /admin/store`, com confirmação e auditoria.
- [x] AUD-03: agregações de status, disponíveis/total, prazo e pedidos recentes no dashboard.
- [x] AUD-04: substituir `--bg-base` por `--bg-input` no checkout.
- [x] AUD-14: detectar token `var(--…)` inexistente no validador de design system e testar a falha.
- [x] AUD-07: numerar a revisão final como etapa no checkout e mover o envio para dentro dela.
- [x] AUD-05: associar erros aos inputs em `AdminCatalogForm`, `AdminProductForm`, `AdminStoreView` e `AdminUserForm`.
- [x] AUD-11: remover o `<main id="conteudo">` duplicado do shell ou das views, mantendo um único destino do skip-link.
- [x] AUD-06: trocar `--status-error` por cor neutra no badge “Loja fechada”.
- [x] AUD-08/AUD-09: barra fixa de carrinho no mobile e aviso de loja fechada na home.
- [x] AUD-10: privacidade e termos passaram a exibir data de rascunho (25/11/2026) e, após confirmação em 25/11/2026, a versão vigente com data de publicação e canal do titular em `/contato` (D-033).
- [x] AUD-12: emojis removidos dos nomes de categoria em `design-system/seed.json` após decisão de design (D-030); teste impede o retorno.
- [x] AUD-13: bottom nav de quatro atalhos substituído por drawer completo com todos os itens por papel, após decisão registrada em `13-decisoes-e-pendencias.md` (D-029).

**Ordem sugerida:** US-123 → US-120 → US-121 → US-122 → US-124 → US-125 → US-126. Nenhuma história entra em andamento sem teste automatizado previsto na Definition of Done da seção 1.5.

**Status em 25/11/2026:** US-123 concluída — validador de tokens inexistente com 3 testes novos (19 testes no total) e `--bg-base` substituído por `--bg-input`. US-120 concluída — título por estado, aviso de registro, `OrderStatusTimeline` com teste dedicado, contato da loja e mensagem genérica para `NOT_FOUND`; `OrderView.test.ts` cobre sucesso, recarga, finalização, carregamento, erro e 404. US-121 concluída — kill switch com confirmação, indicadores vindos da API, contagem por status e pedidos recentes sem PII; `AdminDashboardView.test.ts` cobre sucesso, alternância, status fechado, falha de fontes e contadores zerados, com prints em 375/1280 px. US-122 concluída — etapa final “Revisão e envio” com total cotado e botão dentro dela, numeração ajustada na retirada; `CheckoutView.test.ts` cobre sucesso, cotação falha e retirada, com prints em 375/1280 px. US-124 concluída — erros de campo associados nos quatro formulários administrativos, `<main>` duplicado removido do shell administrativo e alvos do skip-link tornados focáveis em todas as rotas; testes de DOM cobrem as associações e a estrutura. US-125 concluída — status fechado do painel em cor neutra, `CartSummaryBar` com contagem e subtotal na home mobile e aviso de indisponibilidade quando a loja está fechada; testes de DOM cobrem os três comportamentos. US-126 concluída — data de rascunho nas páginas legais, emojis removidos do seed, drawer completo no painel mobile com foco e fechamento acessíveis e decisões registradas em `13-decisoes-e-pendencias.md` (D-029 a D-031). Privacidade, termos e créditos publicados em 25/11/2026 (D-033), com canal do titular em `/contato` e `noindex` removido dos créditos. Épico 12 encerrado.

---

## 17. Rastreabilidade requisito → história

| Requisito | Histórias |
|---|---|
| FR-STORE-01/02 | US-021, US-023, US-025 |
| FR-MENU-01/02 | US-021, US-023 |
| FR-PRODUCT-01/02/03 | US-024 |
| FR-CART-01/02 | US-030, US-031 |
| FR-CHECKOUT-01..05 | US-032, US-033, US-034 |
| FR-ORDER-01..06 | US-040, US-041, US-042, US-043, US-050, US-072 |
| FR-ADMIN-AUTH-01/02 | US-060, US-061 |
| FR-ADMIN-CAT-01/02 | US-022, US-062, US-063 |
| FR-ADMIN-ZONE-01 | US-064 |
| FR-ADMIN-STORE-01 | US-065 |
| FR-ADMIN-ORDER-01..03 | US-070, US-071, US-072 |
| FR-ADMIN-USERS-01 | US-066 |
| FR-ADMIN-AUDIT-01 | US-073 |
| FR-CONTENT-01/02 | US-080 |
| FR-SEO-01..03 | US-081 |
| FR-A11Y-01/02 | US-082 |
| FR-LGPD-01 | US-080, US-093 |
| FR-OPS-01 | US-011, US-103 |
| Auditoria de conformidade AUD-01..14 | US-120 a US-126 |

---

## 18. Bloqueios atuais

| Bloqueio | Histórias afetadas | Ação |
|---|---|---|
| Catálogo não validado | US-001, US-020–US-024 | Fechar com cliente |
| Taxas/horários não validados | US-002, US-032, US-033 | Fechar com cliente |
| Hospedagem/domínio indefinidos | US-100–US-104 | Decidir |
| Fotos/logo indefinidos | US-004, US-083 | Cliente |
| CI provider indefinido | US-010 | Decidir |
| WhatsApp/API indefinido | US-042 | Decidir |
| Decisões abertas de design (§14 de `07-design-system-e-experiencia.md`) | US-125, US-126 | Aprovar nav mobile, carrinho desktop, emoji e tokens de espaço/sombra; registrar em `13-decisoes-e-pendencias.md` |

---

## 19. Regras de gestão

- Uma história por PR; se crescer, divida.
- Não iniciar P1 antes de fechar P0 do mesmo fluxo.
- Não marcar como concluída sem evidência de teste.
- Mudança de escopo entra em `13-decisoes-e-pendencias.md` e ajusta este backlog.
- O backlog não é uma promessa de data; ver `11-roadmap.md`.
