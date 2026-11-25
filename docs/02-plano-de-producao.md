# Plano de Produção — Brocados Lanches

**Status:** documento histórico da Fase 1; planejamento executável atualizado em `11-roadmap.md` e `08-backlog.md`.  
**Premissa:** o HTML recebido é referência de design e fluxo, não base de produção.  
**Stack:** decidida em 18/11/2026 — Go + Chi, Vue 3 + Vite + TypeScript, MongoDB e Docker Compose.

> Este é o plano inicial da descoberta. O escopo funcional detalhado está em `03-prd.md`; a arquitetura em `04-arquitetura.md`; o contrato em `06-contrato-api.md`; e a execução em `08-backlog.md` + `11-roadmap.md`. Em caso de conflito, prevalecem os documentos atualizados.

---

## 1. Decisões que precisam vir antes do código

| Decisão | Por que importa |
|---|---|
| Pedido fecha no WhatsApp ou dentro do site? | Define se haverá checkout próprio, pagamento e painel de pedidos |
| Haverá pagamento online? | Muda escopo, provedor, conciliação e responsabilidade |
| Quem opera o painel? | Define login, permissões, treinamento e suporte |
| Onde será hospedado? | Afeta stack, custo, domínio e backup |
| Qual o catálogo real? | Sem isso, não há como validar preços, adicionais e alergênicos |
| Qual o prazo e o orçamento? | Define se a V1 será enxuta ou completa |

Sem essas respostas, o plano abaixo é uma referência — não um compromisso de entrega.

---

## 2. Fases propostas

### Fase 0 — Descoberta e fechamento de escopo

- Preencher `../briefing/chamado.md`.
- Confirmar marca, contato, endereço, horários e redes.
- Levantar cardápio, adicionais, zonas e taxas reais.
- Definir o fluxo do pedido e o papel do WhatsApp.
- Acordar critérios de aceite, prazo e responsáveis.

**Saída:** briefing aprovado e escopo congelado para a V1.

### Fase 1 — Fundação técnica

- Criar repositório, ambientes e pipeline.
- Escolher stack e hospedagem.
- Importar `../design-system/tokens.css`, configurar o tema e ativar `make design-check`.
- Configurar lint, formatação, testes e deploy de preview.
- Definir modelo de dados inicial.

**Entregue em 18/11/2026:**

- monorepo com backend Go + Chi, frontend Vue 3 + Vite + TypeScript e MongoDB via Compose;
- health checks `/healthz` e `/readyz`, logs JSON sem PII e shutdown gracioso;
- tela inicial Vue consumindo o readiness da API;
- tokens do design system importados no build, Tailwind compilado e validação `make design-check`;
- testes de backend, frontend, documentação e design system;
- Dockerfiles de API e web, Nginx same-origin e `compose.yaml`;
- caminho offline (`frontend/Dockerfile.runtime`) para ambientes sem acesso npm dentro do Docker;
- dependências Go vendorizadas para build Docker offline.

**Na Fase 1:** provedor de CI, modelo de dados inicial, seed e testes entregues. Continuam pendentes o deploy de preview e a definição de hospedagem; o plano executável atual está em `11-roadmap.md` e `08-backlog.md`.

**Saída:** aplicação vazia no ar em ambiente de preview, com tema aplicado.

### Fase 2 — Catálogo e conteúdo

- Modelar categorias, produtos, adicionais, zonas e horários.
- Popular com dados reais validados.
- Implementar listagem, busca, filtros e estados de esgotado.
- Adicionar fotos otimizadas e textos com alergênicos.
- Criar páginas institucionais e de contato.

**Saída:** cardápio navegável e conteúdo revisado.

### Fase 3 — Carrinho e checkout

- Carrinho persistente entre recarregamentos.
- Cálculo de subtotal, taxa e total no servidor.
- Delivery ou retirada, endereço e zona.
- Formas de pagamento conforme decisão da Fase 0.
- Validação de preço, taxa e disponibilidade no backend.
- Registro do pedido antes de qualquer envio.
- Página de confirmação e acompanhamento.

**Saída:** pedido completo criado e recuperável, mesmo se o cliente fechar a aba.

### Fase 4 — Painel administrativo

- Autenticação com senha forte e, se possível, MFA.
- Perfis e permissões.
- CRUD de produtos, categorias, adicionais e zonas.
- Abrir/fechar loja e pausar itens.
- Fila de pedidos com status, som e impressão/KDS, se contratado.
- Auditoria das alterações.

**Saída:** operação consegue receber e acompanhar pedidos sem depender do desenvolvedor.

### Fase 5 — Qualidade, segurança e lançamento

- Testes unitários, de integração e ponta a ponta nos fluxos críticos.
- Auditoria visual contra o design system: tokens, componentes, estados e 375/1280 px.
- Acessibilidade (axe, navegação por teclado e foco).
- SEO local, metadados, sitemap e dados estruturados.
- Política de privacidade, termos e retenção de dados.
- Backup, restauração e monitoramento.
- Revisão de segurança e rate limit.
- Deploy no domínio definitivo e checklist de go-live.

**Saída:** V1 publicada com critérios de aceite comprovados.

### Fase 6 — Pós-lançamento

- Medir funil: visita, cardápio, carrinho, checkout, pedido.
- Acompanhar erros, tempo de resposta e pedidos perdidos.
- Corrigir atritos das primeiras duas semanas.
- Planejar melhorias com base em dados reais.

---

## 3. Definição de pronto da V1

- [ ] Cardápio, preços, taxas e adicionais validados pelo cliente.
- [ ] Pedido gravado no servidor com identificador único.
- [ ] Total recalculado e validado no backend.
- [ ] Painel protegido por autenticação.
- [ ] Fluxo completo testado em mobile e desktop.
- [ ] Loja fechada, item esgotado e zona indisponível tratados.
- [ ] Nenhum dado pessoal em logs ou analytics.
- [ ] Política de privacidade e termos publicados.
- [ ] Backup e restauração testados.
- [ ] Deploy, domínio e suporte documentados.

---

## 4. Riscos principais

| Risco | Mitigação |
|---|---|
| Escopo crescer durante o desenvolvimento | Congelar a V1 e registrar mudanças fora do escopo |
| Catálogo e taxas desatualizados | Definir responsável e rotina de atualização no admin |
| Dependência do WhatsApp para fechar pedido | Registrar o pedido no site antes do handoff |
| Vazamento de dados | Minimização, autenticação, LGPD e revisão de logs |
| Protótipo ser confundido com produto pronto | Comunicar limites e não publicar o HTML como está |
| Falta de suporte após a entrega | Definir contrato, canal e prazo de manutenção |

---

## 5. Stack definida

A decisão foi registrada em 18/11/2026. As regras de implementação, contrato de API, modelo de dados e definição de pronto estão em `../AGENTS.md`.

- monorepo com `backend/` e `frontend/`;
- backend em Go + Chi, com MongoDB 8+ e driver oficial;
- frontend em Vue 3 + Vite + TypeScript + Pinia, com Tailwind compilado e os tokens atuais;
- autenticação admin com sessão em cookie `HttpOnly` e Argon2id;
- validação explícita no backend e Zod no frontend;
- testes com `testing`/`httptest`/testcontainers, Vitest, Testing Library, Playwright e axe;
- gate de documentação, testes e design system antes de cada commit;
- deploy com preview por pull request e API/web same-origin;
- observabilidade com logs estruturados sem PII e alerta de erro;
- integração de pedido/WhatsApp por adapter, para permitir troca sem reescrever o fluxo.

Não troque a stack sem atualizar `../AGENTS.md` e registrar a decisão.

---

## 6. Próximo passo imediato

Fechar o briefing e iniciar a Fase 2 (catálogo), mantendo o esqueleto da Fase 1 como base. Sem número do chamado, solicitante, prazo, catálogo real e definição do fluxo de pedido, qualquer estimativa é chute.
