# Chamado — Brocados Lanches

**Status:** aberto — briefing incompleto.  
**Aberto em:** 18/11/2026  
**Última atualização:** 18/11/2026

---

## 1. Identificação

| Campo | Valor |
|---|---|
| Cliente / projeto | Brocados Lanches |
| Número do chamado | não informado |
| Canal de origem | não informado |
| Solicitante | não informado |
| Responsável técnico | não informado |
| Prazo acordado | não informado |
| Prioridade | não informada |

> Preencher os campos acima antes de estimar esforço ou prometer data.

---

## 2. Pedido recebido

O material anexado é um **design system responsivo com protótipo navegável** da Brocados Lanches, versão 2.5, com telas de cliente e de administrador.

**Entendimento inicial (a confirmar):** criar a presença digital de pedidos da lanchonete, partindo deste protótipo como referência visual e funcional.

O chamado ainda não traz o texto do pedido do cliente, o escopo contratado nem o resultado esperado. Essas informações precisam ser confirmadas antes da execução.

---

## 3. O que já foi entregue

- Protótipo HTML standalone, responsivo (mobile e desktop).
- Design system com paleta, tipografia e componentes.
- Fluxo de cliente: Home → Cardápio → Personalizar → Carrinho → Checkout.
- Painel administrativo com produtos, categorias, zonas de entrega e dados da loja.
- Exportação de tokens CSS/Tailwind e JSON de dados.
- Esqueleto de produção em Go, Vue e MongoDB.
- Documentação completa de produto, arquitetura, dados, API, design, backlog, testes e operação em `../docs/README.md`.

Detalhes do protótipo em `../docs/01-inventario-prototipo.md`. Planejamento executável em `../docs/08-backlog.md` e `../docs/11-roadmap.md`.

---

## 4. Dados conhecidos

### 4.1 Protótipo (fictício)

| Dado | Valor no protótipo | Situação |
|---|---|---|
| Nome | Brocados Lanches | confirmar grafia oficial |
| WhatsApp | `(85) 99999-9999` | **placeholder** |
| Instagram | `@brocadoslanches` | **confirmar** |
| Endereço | `Av. Principal, 1200 - Centro, Aquiraz - CE` | **placeholder** |
| Tempo médio de entrega | `35-50 min` | confirmar |
| Categorias | Hambúrgueres, Combos, Petiscos, Bebidas | validar |
| Produtos | 5 itens de demonstração | substituir pelo cardápio real |
| Zonas de entrega | Centro (R$ 5), Prainha (R$ 8), Porto das Dunas (R$ 12) | validar bairros e taxas |
| Pagamentos | Pix, Cartão e Dinheiro | confirmar |
| Horário de funcionamento | não exibido | **falta** |

> Esses dados são de demonstração e **não** devem ir para produção.

### 4.2 Site público de referência (snapshot em 18/11/2026)

Fonte: `https://cardpiobrocadoslanchess.ola.click/`.

| Dado | Valor observado | Situação |
|---|---|---|
| Nome | Brocados Lanches | validar com cliente |
| WhatsApp | `+55 85 99235-8708` | **validar** |
| Endereço | `R. Lourenço Alves, 364 - Tapera, Aquiraz - CE, 61700-000` | **validar** |
| Fuso | `America/Fortaleza` | confirmado no site |
| Entrega | 30–90 min | **validar** |
| Retirada | sim | **validar regras** |
| Checkout | WhatsApp | confirmado no site |
| Pagamento | dinheiro, cartão e Pix, na entrega/retirada | **validar** |
| Catálogo | 6 seções, 40 itens; 8 esgotados no snapshot | **validar item a item** |
| Bairros | 30 regiões listadas; taxas não públicas | **validar taxas** |
| Horário | seg–dom, com terça fechada; ver detalhes | **validar feriados** |

Horários observados:

| Dia | Horário |
|---|---|
| Domingo | 17:45–23:30 |
| Segunda | 18:00–23:30 |
| Terça | fechado |
| Quarta | 18:00–23:30 |
| Quinta | 18:00–23:30 |
| Sexta | 17:30–23:30 |
| Sábado | 18:00–23:30 |

O catálogo, os bairros e as taxas do site **não** são importados automaticamente. A análise completa está em `../docs/12-referencia-ola-click.md`.

### 4.3 O que ainda falta validar

- cardápio oficial, descrições, ingredientes e alergênicos;
- grupos de opções e adicionais por produto;
- taxas por bairro e prazo real;
- feriados e fechamentos temporários;
- dados de contato e endereço;
- formas de pagamento e regra de troco;
- fotos, logo e direitos de uso.

---

## 5. Pendências para fechar o briefing

### Negócio

- [ ] Número e link do chamado.
- [ ] Nome completo e contato do solicitante.
- [ ] Escopo contratado: landing page, cardápio, pedidos, painel, integração?
- [ ] Prazo e prioridade.
- [ ] Domínio e hospedagem.
- [ ] Quem opera o painel no dia a dia.

### Marca e conteúdo

- [ ] Logo em vetor (SVG/AI/PDF) e variações.
- [ ] Fotos reais dos produtos e da loja.
- [ ] Cardápio oficial com nomes, descrições, ingredientes, alergênicos e preços.
- [ ] Adicionais e regras de personalização.
- [ ] Horários, feriados e disponibilidade.
- [ ] Endereço, WhatsApp, Instagram e demais canais.
- [ ] Textos institucionais, política de privacidade e termos.

### Operação

- [ ] Bairros atendidos e taxas por bairro.
- [ ] Tempo médio e raio de entrega.
- [ ] Formas de pagamento e necessidade de troco.
- [ ] Como o pedido chega hoje: WhatsApp, impressora, KDS, outro sistema?
- [ ] O que acontece se a loja estiver fechada ou um item esgotar.

### Técnico e jurídico

- [ ] Stack definida (Go + Vue + MongoDB); hospedagem, domínio e CI pendentes.
- [ ] Integração de pagamento (não prevista na V1; confirmar).
- [ ] Envio de pedidos para WhatsApp: link manual ou API oficial?
- [ ] Analytics e métricas de conversão (somente com consentimento).
- [ ] LGPD: base legal, retenção, consentimento e canal do titular.
- [ ] Responsável por suporte e manutenção após a entrega.

---

## 6. Critérios de aceite propostos

Só valem depois de o escopo ser confirmado:

- [ ] Cardápio e preços conferidos com o cliente.
- [ ] Pedido registrado no servidor antes de qualquer envio.
- [ ] Preços, taxas e adicionais validados no backend.
- [ ] Painel protegido por autenticação.
- [ ] Funciona em mobile e desktop sem quebra de layout.
- [ ] Loja fechada e item esgotado tratados com mensagem clara.
- [ ] Nenhum dado pessoal em logs ou analytics.
- [ ] Política de privacidade e termos publicados.
- [ ] Deploy, domínio e backup documentados.

---

## 7. Observações

- O protótipo usa dados fictícios e não deve ir para produção como está.
- O site público de referência é uma plataforma OlaClick; seus dados precisam de validação e seus textos/fotos não podem ser copiados sem autorização.
- O botão de finalizar pedido do protótipo **não envia nada hoje**; ele apenas simula o fluxo.
- O HTML do protótipo depende de CDN externo (Tailwind e Google Fonts).
- Antes de reutilizar qualquer conteúdo, confirmar autoria e licença das imagens e textos.
