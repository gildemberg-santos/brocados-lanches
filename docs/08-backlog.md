# Backlog — Brocados Lanches V1

**Versão:** 1.1  
**Atualizado em:** 22/11/2026  
**Status:** Épico 1 em andamento (CI, seed e harness E2E entregues; preview por PR pendente); Épico 2 com US-020, US-021, US-023 e US-024 concluídas; Épico 3 com US-030 a US-034 concluídas; Épico 4 com US-040 a US-043 concluídas.  
**Relacionados:** `03-prd.md`, `04-arquitetura.md`, `05-modelo-de-dados.md`, `06-contrato-api.md`, `07-design-system-e-experiencia.md`, `09-qualidade-e-testes.md`, `11-roadmap.md`.

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

**Status em 22/11/2026:** US-010 e US-012 concluídos. US-013 com harness pronto e smoke da home + navegação para o cardápio em desktop/mobile. US-011 já está entregue no esqueleto (request ID, logs JSON e health checks); falta um teste dedicado de correlação e mascaramento de PII.

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
| US-025 | Informações da loja | P1 | M | US-021 | Status, horários, endereço, contato e prazo; estados |

**Status em 22/11/2026:** US-020 concluída (domínio, validação, repositórios, índices idempotentes e testes de integração com MongoDB real). US-021 concluída (`GET /menu` e `GET /delivery-zones`, filtros e testes de service/handler). US-023 concluída (home, cardápio, busca sem acento/caixa, filtro por categoria, grid, esgotado, estados e testes). US-024 concluída (detalhe, grupos obrigatórios com mínimo/máximo, adicionais, observação, total estimado e testes). O CTA de adicionar valida a personalização e alimenta o carrinho desde a US-030. US-025 planejada.

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

**Status em 22/11/2026:** US-030 concluída (`useCartStore` com schema `v1`, persistência de IDs/escolhas/quantidades, limpeza de versões antigas e validação contra o catálogo), US-031 concluída (`/carrinho` com `CartLine`, `CartSummary` e `QuantityStepper`, quantidade, remoção, limpeza, estado vazio e totais estimados), US-032 concluída (`POST /orders/quote` com validação de itens, opções, adicionais, zona e status da loja; totais calculados no servidor por `domain.BuildQuote`; sem persistência) e US-033/US-034 concluídas (`/checkout` com entrega/retirada, zona, endereço, pagamento, troco, validação por campo, foco no primeiro erro, cotação no servidor, bloqueio de envio e estados de carregamento/erro/sucesso). O CTA “Continuar” do carrinho leva ao checkout. A taxa estimada usa a menor zona ativa enquanto nenhuma zona foi escolhida; o total definitivo vem da cotação da US-032. O limite de taxa previsto no contrato ainda não está ativo no esqueleto.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-030 | Store do carrinho | P0 | M | US-024 | Persiste IDs/quantidades/escolhas; versão; testes |
| US-031 | UI do carrinho | P0 | M | US-030 | Quantidade, remover, limpar, vazio, subtotal e taxa; testes |
| US-032 | API de cotação | P0 | M | US-021 | `POST /orders/quote` valida e devolve totais; sem persistir; testes |
| US-033 | Formulário de checkout | P0 | G | US-031, US-032 | Delivery/retirada, zona, endereço, pagamento, troco e validação; testes |
| US-034 | Revisão e envio | P0 | M | US-033 | Total cotado, bloqueio de envio, estados de erro e sucesso |

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

---

## 8. Épico 4 — Pedido e WhatsApp

**Objetivo:** criar o pedido persistido, idempotente e com handoff recuperável.

**Status em 22/11/2026:** US-040 concluída (`POST /orders` recalcula o total com `domain.BuildQuote`, persiste snapshot com número via contador atômico, `publicId` opaco, índices únicos e `201` com `Location`), US-041 concluída (chave de 16 a 128 caracteres, `requestHash` canônico, replay `200` com `Idempotent-Replayed` e conflito `409`), US-042 concluída (link `wa.me` gerado após a persistência, com reenvio na mesma sessão), US-043 concluída (`/pedido/:publicId` mostra número/status/totais e limpa o carrinho somente após sucesso). US-044 ainda depende de teste E2E de rede.

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
- [ ] Testar falha de Mongo, falha de link e duplo clique.
- [ ] Testar concorrência com a mesma chave.

---

## 9. Épico 5 — Acompanhamento

**Objetivo:** permitir consulta pública segura do pedido.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-050 | API pública de acompanhamento | P1 | M | US-040 | Retorna status/itens/total sem PII sensível; rate limit; testes |
| US-051 | Tela de acompanhamento | P1 | M | US-050 | Timeline, atualização, não encontrado e erro; testes |
| US-052 | Reenvio de WhatsApp | P1 | P | US-042 | Link completo disponível na confirmação imediata; após recarregar, contato da loja com número do pedido |

**Tarefas técnicas:**

- [ ] Definir resposta pública mínima.
- [ ] Ocultar telefone, endereço, observação e `_id`.
- [ ] Implementar rate limit.
- [ ] Criar `OrderStatusTimeline`.
- [ ] Testar ID inválido, ID inexistente e status final.
- [ ] Atualizar contrato.

---

## 10. Épico 6 — Admin: autenticação e catálogo

**Objetivo:** dar autonomia segura à operação.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-060 | Autenticação admin | P0 | G | E1 | Login/logout/me, Argon2id, cookie, rate limit, CSRF; testes |
| US-061 | Shell e guard do admin | P0 | M | US-060 | Rotas lazy, guard, 401/403, menu por papel; testes |
| US-022 | API admin de catálogo | P0 | G | US-020, US-060 | CRUD de categorias/produtos/adicionais/zonas; auditoria; testes |
| US-062 | Gestão de produtos | P0 | G | US-022, US-061 | Lista/form/editar/ativar/disponível/destaque; testes |
| US-063 | Categorias e adicionais | P0 | G | US-022, US-061 | CRUD, slug, ordem, ativo; testes |
| US-064 | Zonas e taxas | P0 | M | US-022, US-061 | CRUD, taxa em centavos, ativo; testes |
| US-065 | Configuração da loja | P0 | M | US-061 | Kill switch, horários, contato, prazo; auditoria; testes |
| US-066 | Usuários e papéis | P1 | M | US-060 | Owner cria/edita/desativa; último owner protegido; testes |

**Tarefas técnicas:**

- [ ] Implementar Argon2id e política de senha.
- [ ] Implementar sessão com TTL e rotação.
- [ ] Implementar CSRF por Origin/Referer.
- [ ] Criar middleware de auth e autorização.
- [ ] Implementar CRUD admin de categorias, produtos, adicionais e zonas.
- [ ] Criar `AdminShell`, `DataTable`, `AdminPageHeader`.
- [ ] Criar formulários com validação e conversão de centavos.
- [ ] Implementar auditoria em toda mutação.
- [ ] Testar acesso anônimo, papel errado e sessão expirada.

---

## 11. Épico 7 — Operação de pedidos

**Objetivo:** permitir que a operação receba, acompanhe e finalize pedidos sem erro.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-070 | Fila de pedidos | P0 | G | US-040, US-060 | Lista/filtros/paginação; status e totais; testes |
| US-071 | Detalhe do pedido | P0 | M | US-070 | Snapshot, cliente, endereço, pagamento e histórico; testes |
| US-072 | Transições de status | P0 | G | US-071 | Matriz válida; `409` inválido; motivo de cancelamento; auditoria |
| US-073 | Auditoria | P1 | M | US-072 | Owner consulta mutações sem PII; filtros; testes |
| US-074 | Atualização em tempo real | P2 | M | US-070 | Fila atualiza sem recarregar ou polling curto; opcional |

**Tarefas técnicas:**

- [ ] Implementar listagem com filtros allowlist.
- [ ] Mascarar PII na lista; detalhe autorizado.
- [ ] Implementar máquina de estados no domínio.
- [ ] Gravar `statusHistory`, timestamps e auditoria.
- [ ] Criar ações contextuais por status.
- [ ] Testar todas as transições e concorrência.
- [ ] Criar tela de auditoria somente leitura.

---

## 12. Épico 8 — Conteúdo, SEO, acessibilidade e LGPD

**Objetivo:** deixar a V1 apresentável, encontrável, acessível e legalmente defensável.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-080 | Páginas institucionais e legais | P1 | M | E0 | Contato, privacidade e termos publicados com links |
| US-081 | SEO e dados estruturados | P1 | M | US-023 | Metadados, JSON-LD, sitemap e robots; testes |
| US-082 | Acessibilidade | P1 | G | US-023, US-033, US-061 | Teclado, foco, rótulos, contraste, 44 px e reduced motion |
| US-083 | Imagens e performance | P1 | M | US-023 | WebP/AVIF, dimensões, lazy, fontes self-hosted |
| US-084 | Consentimento e analytics | P2 | M | US-080 | Sem PII; opt-in; eventos documentados |

**Tarefas técnicas:**

- [ ] Criar páginas e conteúdo aprovado.
- [ ] Implementar `@unhead/vue` ou equivalente.
- [ ] Gerar JSON-LD `Restaurant`, `Menu` e `Product`.
- [ ] Gerar sitemap/robots.
- [ ] Rodar axe nas telas críticas.
- [ ] Corrigir foco, rótulos e alvos de toque.
- [ ] Self-host de fontes.
- [ ] Otimizar imagens e cache.
- [ ] Revisar retenção e canal do titular.

---

## 13. Épico 9 — Qualidade e segurança

**Objetivo:** provar que os fluxos críticos funcionam e que o sistema resiste a erro e abuso.

| ID | História | Prioridade | Tam. | Depende de | Aceite |
|---|---|---|---|---|---|
| US-090 | Testes de backend | P0 | G | E2–E7 | Service, handler e repository com Mongo real; `-race` |
| US-091 | Testes de frontend | P0 | G | E2–E7 | Stores, componentes e services; Vitest |
| US-092 | E2E crítico | P0 | G | US-040, US-072 | Cardápio → pedido → admin → status em mobile/desktop |
| US-093 | Segurança e privacidade | P0 | M | US-060, US-040 | Rate limit, CSRF, auth, enumeração e PII em logs |
| US-094 | Performance e carga | P1 | M | E2–E7 | Orçamentos de `03-prd.md` medidos; gargalos corrigidos |

**Tarefas técnicas:**

- [ ] Criar matriz de testes de `09-qualidade-e-testes.md`.
- [ ] Configurar testcontainers.
- [ ] Configurar Playwright.
- [ ] Testar todos os cenários críticos de `09-qualidade-e-testes.md`.
- [ ] Revisar logs e mascaramento.
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

## 16. Rastreabilidade requisito → história

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

---

## 17. Bloqueios atuais

| Bloqueio | Histórias afetadas | Ação |
|---|---|---|
| Catálogo não validado | US-001, US-020–US-024 | Fechar com cliente |
| Taxas/horários não validados | US-002, US-032, US-033 | Fechar com cliente |
| Hospedagem/domínio indefinidos | US-100–US-104 | Decidir |
| Texto legal indefinido | US-080 | Jurídico/cliente |
| Fotos/logo indefinidos | US-004, US-083 | Cliente |
| CI provider indefinido | US-010 | Decidir |
| WhatsApp/API indefinido | US-042 | Decidir |

---

## 18. Regras de gestão

- Uma história por PR; se crescer, divida.
- Não iniciar P1 antes de fechar P0 do mesmo fluxo.
- Não marcar como concluída sem evidência de teste.
- Mudança de escopo entra em `13-decisoes-e-pendencias.md` e ajusta este backlog.
- O backlog não é uma promessa de data; ver `11-roadmap.md`.
