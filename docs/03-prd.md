# PRD — Produto de Pedidos Brocados Lanches

**Versão:** 1.0  
**Atualizado em:** 18/11/2026  
**Status:** V1 planejada; escopo sujeito à validação do cliente.  
**Responsável do produto:** a definir no briefing.  
**Documentos relacionados:** `04-arquitetura.md`, `06-contrato-api.md`, `07-design-system-e-experiencia.md`, `08-backlog.md`, `12-referencia-ola-click.md`.

---

## 1. Resumo executivo

A Brocados Lanches precisa receber pedidos digitais sem perder o controle do catálogo, dos preços e da operação. Hoje o protótipo demonstra a experiência visual, mas não persiste pedidos, não valida preços no servidor e não protege o painel administrativo.

A V1 será uma aplicação web **mobile-first**, com cardápio, personalização, carrinho, checkout, pedido persistido, acompanhamento por link opaco e painel administrativo autenticado. O fechamento do pedido continua no **WhatsApp**, mas só depois de o pedido existir no servidor. Preços, taxas, disponibilidade e status da loja serão sempre calculados e validados no backend.

O site público `https://cardpiobrocadoslanchess.ola.click/` é referência funcional e de conteúdo. O visual continua sendo o design system v2.5. Nenhum dado do site de referência entra em produção sem validação do cliente.

---

## 2. Problema e oportunidade

### 2.1 Problema

- O cliente não tem um cardápio digital confiável com preços, adicionais e taxa de entrega.
- O pedido depende de conversa manual no WhatsApp, o que aumenta erro, retrabalho e perda de mensagem.
- O protótipo não registra o pedido: fechar a aba apaga tudo.
- O painel atual não tem login, histórico, auditoria nem status de pedido.
- O operador não consegue abrir/fechar a loja, pausar itens e acompanhar a fila em um único lugar.
- Não há dados estruturados para medir conversão nem base para melhorar o atendimento.

### 2.2 Oportunidade

- Reduzir o tempo entre “quero pedir” e “pedido recebido”.
- Diminuir pedidos perdidos e erros de preço.
- Dar autonomia à operação para atualizar catálogo e disponibilidade.
- Criar uma base confiável para WhatsApp, atendimento e futuras integrações.
- Preservar a identidade visual já validada no design system.

---

## 3. Objetivos e métricas

### 3.1 Objetivos de negócio

1. Receber pedidos completos e válidos sem intervenção do desenvolvedor.
2. Garantir que todo pedido enviado ao WhatsApp exista no servidor.
3. Permitir que a operação gerencie catálogo, taxas, horários e status.
4. Reduzir erros de total, taxa e item indisponível.
5. Manter dados pessoais no mínimo necessário e fora de logs.

### 3.2 Objetivos do produto

- O cliente monta o pedido em poucos passos, no celular.
- O cliente entende taxa, total, forma de pagamento e prazo antes de enviar.
- O pedido pode ser recuperado e acompanhado por um link opaco.
- O operador vê a fila e muda o status sem ambiguidade.
- O owner consegue administrar usuários e consultar auditoria.

### 3.3 Métricas de sucesso

As metas serão medidas após o lançamento; ainda não há linha de base.

| Métrica | Como medir | Meta inicial |
|---|---|---|
| Conclusão do checkout | pedidos criados / sessões que abriram o checkout | A definir com dados reais |
| Handoff para WhatsApp | pedidos com WhatsApp aberto / pedidos criados | ≥ 95% |
| Pedido perdido | pedidos criados sem confirmação visual | 0 |
| Tempo para montar pedido | mediana entre abrir cardápio e criar pedido | A definir |
| Tempo de primeira resposta | criação do pedido até `confirmed` | A definir com a operação |
| Erros de preço/taxa | pedidos com divergência de total | 0 |
| Incidentes de PII | PII encontrada em log/analytics | 0 |
| Disponibilidade da API | monitoramento externo | ≥ 99,5% ao mês |

> Métricas de conversão só podem ser coletadas com consentimento e sem PII, salvo exceção explícita e registrada (D-050). O Google Tag Manager carrega em todas as visitas por decisão do solicitante, sem dados do pedido no `dataLayer`. Ver `10-operacao-e-lancamento.md` §8.6.

---

## 4. Personas e trabalhos a realizar

### 4.1 Cliente de delivery

- **Contexto:** está em casa, no celular, com fome e baixa tolerância a fricção.
- **Objetivo:** ver o cardápio, montar o lanche, saber o total e enviar o pedido.
- **Dores:** cardápio desatualizado, taxa surpresa, item esgotado, pedido que se perde.
- **Sucesso:** pedido criado, link de acompanhamento e confirmação clara.

### 4.2 Cliente de retirada

- **Contexto:** quer evitar taxa e buscar no balcão.
- **Objetivo:** escolher retirada, ver o endereço e o prazo, pagar na retirada.
- **Dores:** taxa cobrada indevidamente, endereço confuso, pedido não encontrado.
- **Sucesso:** taxa zerada, endereço e horário visíveis, pedido confirmado.

### 4.3 Operador

- **Contexto:** durante o serviço, em celular ou desktop, com vários pedidos ao mesmo tempo.
- **Objetivo:** ver pedidos novos, confirmar, preparar, despachar e pausar itens.
- **Dores:** pedido duplicado, status errado, item esgotado que continua vendendo.
- **Sucesso:** fila organizada, transições válidas e catálogo atualizado em segundos.

### 4.4 Owner

- **Contexto:** responsável pela loja e pelos usuários.
- **Objetivo:** configurar horários, taxas, contatos, usuários e consultar auditoria.
- **Dores:** falta de controle, mudanças sem rastro, acesso indevido.
- **Sucesso:** configuração centralizada, papéis claros e auditoria consultável.

### 4.5 Suporte e engenharia

- **Contexto:** precisa diagnosticar pedido, erro e indisponibilidade.
- **Objetivo:** consultar logs sem PII, health checks e auditoria.
- **Dores:** erro genérico, falta de correlação, dados sensíveis em log.
- **Sucesso:** request ID, logs estruturados e runbooks.

---

## 5. Escopo da V1

### 5.1 Dentro do escopo

| Área | Incluído |
|---|---|
| Storefront | Home, cardápio, busca/filtro, detalhe/personalização, carrinho, checkout, confirmação, acompanhamento, contato e privacidade |
| Catálogo | Categorias, produtos, grupos de opções, adicionais, disponibilidade, destaque e imagens |
| Entrega | Delivery por zona e retirada; taxa por zona; retirada com taxa zero |
| Pedido | Cotação server-side, criação persistida, snapshot, número sequencial, `publicId`, idempotência e handoff para WhatsApp |
| Pagamento | Pix cobrado **na confirmação do pedido gravado** por BR Code/QR local com o total persistido, sem gateway nem confirmação automática; cartão e dinheiro na entrega/retirada; troco; sem gateway online |
| Admin | Login, sessão, papéis, catálogo, categorias, adicionais, zonas, status da loja, fila de pedidos, detalhe, transições e auditoria |
| Conteúdo | Textos institucionais, contato, política de privacidade, termos e alergênicos |
| Qualidade | Acessibilidade AA, SEO básico, testes, observabilidade, backup e go-live |

### 5.2 Fora do escopo da V1

- Pagamento online, cartão armazenado, antifraude ou conciliação financeira.
- KDS, impressão térmica, som de novo pedido e integração com PDV.
- Programa de fidelidade, cupons, cashback e indicação.
- Avaliações, comentários públicos e ranking de produtos.
- Rastreamento de entregador em mapa.
- Aplicativo nativo, multi-loja, multimoeda e multilíngue.
- WhatsApp Business API oficial; a V1 usa link `wa.me`.
- MFA obrigatório; pode entrar como P2 antes do lançamento se o risco exigir.
- Tags de terceiros além do container GTM autorizado em D-050.
- Reserva de mesa e consumo no local.

### 5.3 Premissas

1. O cliente valida catálogo, preços, taxas, horários, contatos e textos antes do lançamento.
2. O WhatsApp da loja é o canal de fechamento; a API oficial não está contratada.
3. Pix é cobrado na confirmação do pedido gravado por BR Code/QR local com o total persistido, sem gateway nem confirmação automática; cartão e dinheiro são combinados na entrega/retirada.
4. A operação terá pelo menos um usuário `owner` e um `operator`.
5. A hospedagem e o domínio ainda serão definidos; a API e a web serão same-origin.
6. A loja opera em `America/Fortaleza`, com persistência em UTC.
7. O volume inicial é de uma loja, um turno e poucos pedidos simultâneos.
8. O design system v2.5 é a fonte visual; o site de referência não dita cores ou fontes.

---

## 6. Jornadas principais

### 6.1 Cliente — delivery

1. Abre a home e vê status da loja, prazo e destaques.
2. Entra no cardápio, busca ou filtra por categoria.
3. Abre um produto, escolhe opções obrigatórias e adicionais.
4. Adiciona ao carrinho e revisa quantidades.
5. No checkout, escolhe delivery, bairro e endereço.
6. Escolhe pagamento e, se dinheiro, informa troco.
7. Vê o total cotado pelo servidor e envia o pedido.
8. O servidor valida, persiste e devolve `orderNumber`, `publicId` e `whatsappUrl`.
9. O cliente é levado ao WhatsApp com a mensagem do pedido.
10. Recebe o link de acompanhamento e pode voltar depois.

**Falhas tratadas:** loja fechada, item esgotado, adicional inválido, zona inativa, troco insuficiente, falha de rede, handoff bloqueado.

### 6.2 Cliente — retirada

1. Monta o carrinho normalmente.
2. No checkout, escolhe retirada.
3. O endereço e a zona somem; a taxa é zero.
4. Vê endereço da loja, prazo e forma de pagamento.
5. Envia o pedido e recebe o link de acompanhamento.

### 6.3 Cliente — acompanhamento

1. Abre `/pedido/{publicId}`.
2. Vê número, status, itens, total e prazo.
3. Atualiza a página para ver a mudança de status.
4. Pode ver o contato da loja e informar o número do pedido no WhatsApp; o link completo com a mensagem não é exposto publicamente.
5. Não vê telefone, endereço ou observações de outros clientes.

### 6.4 Operador — fila

1. Faz login.
2. Vê pedidos por status e data.
3. Abre o detalhe com snapshot e dados de entrega.
4. Confirma, prepara, marca como pronto e despacha ou marca retirada.
5. Em caso de cancelamento, informa o motivo.
6. A transição inválida é bloqueada com mensagem clara.

### 6.5 Owner — configuração

1. Faz login e entra em configurações.
2. Ajusta horários, contato, endereço, prazo e kill switch.
3. Cadastra categorias, produtos, adicionais e zonas.
4. Consulta auditoria de mutações.
5. Gerencia usuários e papéis.

---

## 7. Regras de negócio

As regras abaixo complementam `../AGENTS.md`; em caso de conflito, vale o `AGENTS.md`.

### 7.1 Loja e disponibilidade

- A loja só aceita pedido se o horário permitir **e** o kill switch estiver ligado.
- O horário é interpretado em `America/Fortaleza`.
- Horários que cruzam a meia-noite devem ser suportados pelo modelo, mesmo que a operação atual feche às 23:30.
- O status efetivo é calculado no servidor; o frontend apenas exibe.
- Categoria inativa não aparece.
- Produto inativo não aparece; produto ativo e indisponível aparece como “Esgotado” e não pode ser adicionado.
- Adicional inativo ou não vinculado ao produto não pode ser usado.
- Zona inativa não aparece e não pode ser usada.

### 7.2 Preço e total

- Dinheiro é inteiro em centavos (`int64`); proibido `float`.
- O servidor recalcula subtotal, taxa e total a partir dos IDs e quantidades.
- O cliente nunca envia preço, taxa ou total confiável.
- O total exibido no checkout vem de uma cotação server-side.
- O total persistido é congelado no pedido.
- Mudanças posteriores no catálogo não alteram pedidos antigos.
- Retirada zera a taxa.
- Delivery usa a taxa da zona selecionada.
- O troco, quando informado, precisa ser maior ou igual ao total; o servidor calcula o valor a devolver na exibição.

### 7.3 Pedido

- Nenhum pedido existe só no navegador.
- O pedido é persistido antes do handoff para o WhatsApp.
- `Idempotency-Key` é obrigatório; a mesma chave e o mesmo corpo devolvem o mesmo pedido; mesma chave e corpo diferente retornam conflito.
- `publicId` é opaco, aleatório e não sequencial.
- `number` é sequencial e exibido como `#1042`; não serve para acesso público.
- O pedido guarda snapshot de nome, preço unitário, opções, adicionais, taxa e total.
- Se o handoff falhar, o pedido continua salvo e pode ser reenviado.
- O carrinho só é limpo após `201` ou replay idempotente bem-sucedido.

### 7.4 Status

- `pending → confirmed | cancelled`
- `confirmed → preparing | cancelled`
- `preparing → ready | cancelled`
- `ready → out_for_delivery` (delivery) ou `picked_up` (pickup)
- `out_for_delivery → delivered`
- `delivered`, `picked_up` e `cancelled` são finais.
- Toda transição inválida retorna `409`.
- Toda mutação administrativa gera auditoria.

### 7.5 Privacidade

- Coletar apenas nome, telefone, endereço de entrega e observação.
- Não pedir CPF, e-mail ou dados de cartão.
- Não logar nome, telefone, endereço, observação, cookie ou token.
- O acompanhamento público não expõe telefone, endereço, observação nem dados de terceiros.
- Analytics só com consentimento e sem PII, exceto pelo GTM autorizado em D-050.

---

## 8. Requisitos funcionais

A coluna “Aceite” resume o critério testável. O detalhamento de telas está em `07-design-system-e-experiencia.md`; o contrato está em `06-contrato-api.md`.

### 8.1 Storefront e catálogo

| ID | Requisito | Prioridade | Aceite |
|---|---|---|---|
| FR-STORE-01 | Exibir home com status efetivo, prazo, destaques e categorias ativas | P0 | Status e prazo vêm da API; categorias inativas não aparecem; estados de carregamento/erro existem |
| FR-STORE-02 | Exibir endereço, WhatsApp, Instagram e horários | P1 | Dados vêm da API; links abrem corretamente; texto em português |
| FR-MENU-01 | Listar categorias e produtos ativos | P0 | Produto ativo e esgotado aparece com selo; inativo não aparece |
| FR-MENU-02 | Buscar por nome/descrição e filtrar por categoria | P1 | Busca ignora caixa e acentos; filtro combina com busca; estado vazio claro |
| FR-PRODUCT-01 | Abrir detalhe do produto com imagem, descrição, preço e alergênicos | P0 | Preço formatado em BRL; imagem com alt; produto esgotado não permite adicionar |
| FR-PRODUCT-02 | Personalizar com grupos de opções e adicionais | P0 | Opções obrigatórias validadas; mínimo/máximo respeitados; adicional inválido bloqueado; preço recalculado |
| FR-PRODUCT-03 | Permitir observação do item | P1 | Limite de caracteres; conteúdo não vai para log; é congelado no pedido |
| FR-CART-01 | Persistir carrinho entre recarregamentos | P0 | Apenas IDs, quantidades e escolhas; nunca preço confiável; versão do schema tratada |
| FR-CART-02 | Alterar quantidade, remover e limpar | P0 | Quantidade zero remove; total é recalculado; carrinho vazio tem estado próprio |
| FR-CHECKOUT-01 | Escolher delivery ou retirada | P0 | Retirada zera taxa e oculta endereço; delivery exige zona e endereço |
| FR-CHECKOUT-02 | Selecionar zona ativa e calcular taxa | P0 | Taxa vem da API; zona inativa não aparece; troca de zona atualiza cotação |
| FR-CHECKOUT-03 | Coletar nome, telefone e endereço | P0 | Telefone normalizado; endereço obrigatório só em delivery; erros por campo |
| FR-CHECKOUT-04 | Escolher pagamento e troco | P0 | Pix escolhido no checkout e cobrado na confirmação por BR Code/QR local com o total gravado, sem gateway; cartão/dinheiro na entrega; troco validado no servidor; sem dados de cartão |
| FR-CHECKOUT-05 | Revisar total cotado pelo servidor | P0 | Subtotal, taxa e total vêm de `/orders/quote`; divergência bloqueia envio |
| FR-ORDER-01 | Criar pedido persistido | P0 | `201` + `Location`; pedido sobrevive a recarregar/fechar aba; número e `publicId` retornados |
| FR-ORDER-02 | Garantir idempotência | P0 | Mesma chave/corpo não duplica; mesma chave/corpo diferente retorna `409` |
| FR-ORDER-03 | Congelar snapshot | P0 | Mudança posterior de preço/nome não altera pedido antigo |
| FR-ORDER-04 | Fazer handoff para WhatsApp | P0 | URL contém número e mensagem do pedido; falha não apaga pedido; reenvio disponível |
| FR-ORDER-05 | Acompanhar pedido por link opaco | P1 | `/pedido/{publicId}` mostra status e itens sem PII sensível; ID inválido retorna estado claro |
| FR-ORDER-06 | Aplicar ciclo de status | P0 | Transições válidas funcionam; inválidas retornam `409`; timestamps registrados |
| FR-ORDER-07 | Exibir cobrança Pix após a gravação | P1 | A confirmação monta BR Code com o `totalCents` persistido e `paymentMethod` devolvido pelo acompanhamento; nenhum QR no checkout |

### 8.2 Administração

| ID | Requisito | Prioridade | Aceite |
|---|---|---|---|
| FR-ADMIN-AUTH-01 | Login/logout/me com sessão | P0 | Cookie `HttpOnly`; token opaco com hash; logout invalida; sessão expira |
| FR-ADMIN-AUTH-02 | Autorização por papel | P0 | `owner` e `operator`; rota proibida retorna `403`; guard é conveniência, servidor decide |
| FR-ADMIN-CAT-01 | CRUD de produtos | P0 | Criar/editar/ativar/desativar; preço em centavos; opções e adicionais; auditoria |
| FR-ADMIN-CAT-02 | CRUD de categorias e adicionais | P0 | Slug único; ativo/inativo; ordem; auditoria |
| FR-ADMIN-ZONE-01 | CRUD de zonas e taxas | P0 | Taxa em centavos; ativo/inativo; ordem; auditoria |
| FR-ADMIN-STORE-01 | Configurar loja, horários e contato | P0 | Kill switch; horários; prazo; WhatsApp; endereço; auditoria |
| FR-ADMIN-ORDER-01 | Listar e filtrar pedidos | P0 | Paginação; filtro por status/data; ordenação allowlist; sem PII desnecessária |
| FR-ADMIN-ORDER-02 | Ver detalhe do pedido | P0 | Snapshot completo; dados de entrega; histórico; link WhatsApp |
| FR-ADMIN-ORDER-03 | Mudar status | P0 | Matriz de transição; motivo no cancelamento; auditoria |
| FR-ADMIN-USERS-01 | Gerenciar usuários | P1 | Apenas owner; criar/desativar; papel; auditoria; sem senha em resposta |
| FR-ADMIN-AUDIT-01 | Consultar auditoria | P1 | Apenas owner; filtro por ator/entidade/data; sem PII desnecessária |

### 8.3 Conteúdo, SEO, acessibilidade e privacidade

| ID | Requisito | Prioridade | Aceite |
|---|---|---|---|
| FR-CONTENT-01 | Páginas de contato e informações | P1 | Endereço, horários, WhatsApp e mapa/link; links válidos |
| FR-CONTENT-02 | Política de privacidade e termos | P0 | Texto aprovado; link no rodapé e no checkout; data de vigência |
| FR-SEO-01 | Metadados por rota | P1 | Título/descrição/OG; sem duplicação; `lang=pt-BR` |
| FR-SEO-02 | Dados estruturados | P1 | `Restaurant`, `Menu` e `Product` válidos; sem preços divergentes |
| FR-SEO-03 | Sitemap e robots | P2 | Sitemap gerado; rotas públicas indexáveis; admin bloqueado |
| FR-A11Y-01 | Acessibilidade AA | P0 | Contraste, foco, teclado, rótulos, erro associado e alvo de 44 px |
| FR-A11Y-02 | Movimento reduzido | P1 | `prefers-reduced-motion` respeitado; sem animação essencial |
| FR-LGPD-01 | Minimização e retenção | P0 | Dados mínimos; retenção documentada; canal do titular; sem PII em log |
| FR-OPS-01 | Health checks e logs | P0 | `/healthz` e `/readyz`; logs JSON com request ID; sem PII |

---

## 9. Requisitos não funcionais

| Categoria | Requisito | Critério |
|---|---|---|
| Performance | Primeira renderização útil em 4G | ≤ 2,5 s no cardápio em produção |
| Performance | API de leitura | p95 ≤ 300 ms no ambiente de produção |
| Performance | Criação de pedido | p95 ≤ 800 ms, excluindo rede do cliente |
| Disponibilidade | API e web | ≥ 99,5% ao mês, medido externamente |
| Segurança | Autenticação admin | Sessão `HttpOnly`, `Secure`, `SameSite=Lax`; Argon2id; CSRF por Origin/Referer |
| Segurança | Rate limit | Login, pedido e acompanhamento limitados; `429` sem vazar existência |
| Privacidade | PII | Nunca em logs, analytics, mensagens de erro ou URLs de tracking |
| Acessibilidade | WCAG | 2.1 AA nas telas críticas; axe sem erros graves |
| Compatibilidade | Navegadores | Últimas 2 versões de Chrome, Safari, Firefox e Edge; mobile 375 px |
| Responsividade | Layout | 375 px e 1280 px sem quebra; desktop com sidebar |
| SEO | SPA | Metadados e JSON-LD; sitemap; limitações documentadas |
| Observabilidade | Diagnóstico | Request ID, logs estruturados, health checks e alertas |
| Manutenção | Testes | Fluxos críticos cobertos; gate de CI e commit |
| Fuso | Operação | Persistência UTC; exibição em `America/Fortaleza` |
| Dinheiro | Precisão | `int64` centavos; sem `float` |

---

## 10. Conteúdo e dados

### 10.1 Fontes

- **Visual:** design system v2.5.
- **Funcional:** protótipo e site público de referência.
- **Negócio:** cliente; nada entra em produção sem validação.

### 10.2 Dados que precisam de validação

- nome oficial, logo e variações;
- endereço, WhatsApp, Instagram e horários;
- cardápio, descrições, ingredientes, alergênicos e fotos;
- grupos de opções, adicionais e regras;
- zonas, taxas e prazo;
- formas de pagamento e política de troco;
- política de privacidade, termos e retenção;
- textos institucionais e créditos.

O snapshot público do site de referência está em `12-referencia-ola-click.md`. Ele serve para conferência, não como importação automática.

### 10.3 Diretrizes de conteúdo

- Português brasileiro, tom direto e acolhedor.
- Preço sempre formatado em BRL na exibição; API em centavos.
- Não usar “R$ 0,00” para taxa de retirada; escrever “Grátis” ou “Sem taxa”.
- Erros explicam o que fazer, sem culpar o cliente.
- Não expor detalhe técnico, stack trace ou JSON cru.
- Não prometer prazo que a operação não confirmou.

---

## 11. Dependências e restrições

| Dependência | Impacto | Situação |
|---|---|---|
| Catálogo e taxas validados | Bloqueia seed e lançamento | Pendente |
| Hospedagem e domínio | Bloqueia deploy e cookies `Secure` | Pendente |
| Texto legal e retenção | Bloqueia go-live | Pendente |
| WhatsApp definido | Define handoff e mensagem | Pendente |
| Fotos e logo | Afeta qualidade visual e SEO | Pendente |
| CI provider | Bloqueia pipeline automatizado | Pendente |
| MongoDB de produção | Define backup, transações e métricas | Pendente |

Restrições:

- monorepo Go + Vue + MongoDB, definido em `../AGENTS.md`;
- API e web same-origin;
- sem gateway de pagamento na V1; o BR Code do Pix é gerado localmente;
- sem dependência de CDN em produção;
- sem PII em logs;
- sem cálculo de preço no cliente.

---

## 12. Riscos de produto

| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| Catálogo desatualizado no lançamento | Alta | Alto | Validar e congelar conteúdo; responsável definido |
| Taxa errada | Média | Alto | Recalcular no servidor; teste por zona; conferência diária |
| WhatsApp não abrir no dispositivo | Média | Alto | Link alternativo, copiar mensagem e pedido persistido |
| Operação não usar o painel | Média | Alto | Treinamento, fluxo simples e suporte inicial |
| Escopo crescer | Alta | Médio | V1 congelada; mudanças registradas em 13 |
| Dados pessoais em log | Baixa | Alto | Revisão, testes e mascaramento |
| Design system incompleto | Média | Médio | Lacunas documentadas; aprovação antes de token novo |
| Dependência de terceiros | Média | Médio | Evitar CDN; hospedar fontes; contratos claros |
| Falta de prazo/owner | Alta | Alto | Fechar briefing antes de estimar compromisso |

---

## 13. Critérios de lançamento da V1

- [ ] Briefing e dados de negócio validados.
- [ ] Catálogo, preços, taxas e adicionais conferidos.
- [ ] Pedido persistido e recuperável.
- [ ] Total calculado no servidor e congelado.
- [ ] Idempotência comprovada.
- [ ] Handoff para WhatsApp funciona e falha de forma recuperável.
- [ ] Admin exige sessão e autoriza por papel.
- [ ] Transições de status e auditoria funcionam.
- [ ] Loja fechada, item esgotado, zona inativa e troco inválido tratados.
- [ ] Fluxo E2E mobile e desktop passa.
- [ ] Acessibilidade básica validada.
- [ ] Política de privacidade e termos publicados.
- [ ] Nenhuma PII em log ou analytics.
- [ ] Backup e restauração testados.
- [ ] Monitoramento e alertas ativos.
- [ ] `make fmt`, `make lint`, `make test`, `make docs-check` e `make design-check` passam.

---

## 14. Perguntas abertas

As perguntas e decisões estão centralizadas em `13-decisoes-e-pendencias.md`. As que bloqueiam o início do catálogo real são:

1. O cardápio do site de referência é o oficial?
2. Quais grupos de opções e adicionais existem por produto?
3. Quais bairros e taxas valem hoje?
4. Quais horários, feriados e prazos valem hoje?
5. Qual WhatsApp recebe os pedidos?
6. Haverá retirada? Em qual endereço e horário?
7. Quais formas de pagamento e regras de troco?
8. Quem opera o painel e com qual papel?
9. Qual hospedagem, domínio e política de retenção?
10. Quais fotos, logo e textos podem ser usados?

---

## 15. Glossário

| Termo | Definição |
|---|---|
| **Cotação** | Cálculo server-side de totais sem criar pedido. |
| **Snapshot** | Valores congelados do catálogo no pedido. |
| **Handoff** | Abertura do WhatsApp com a mensagem do pedido persistido. |
| **Kill switch** | Interruptor manual que fecha a loja. |
| **Zona** | Bairro/região de entrega com taxa própria. |
| **Opção** | Escolha obrigatória ou opcional dentro de um grupo (ex.: ponto da carne). |
| **Adicional** | Item pago vinculado ao produto (ex.: bacon extra). |
