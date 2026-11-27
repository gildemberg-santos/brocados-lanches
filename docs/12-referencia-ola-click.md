# Referência Pública — OlaClick / Brocados Lanches

**URL:** https://cardpiobrocadoslanchess.ola.click/  
**Snapshot:** 18/11/2026  
**Tipo:** aplicação Nuxt/white-label OlaClick  
**Status:** referência funcional e de conteúdo; **não** é fonte automática de produção.  
**Relacionados:** `03-prd.md`, `07-design-system-e-experiencia.md`, `13-decisoes-e-pendencias.md`.

---

## 1. Por que este documento existe

O site público mostra como a Brocados Lanches já vende hoje: cardápio, categorias, taxas por bairro, formas de pagamento e checkout por WhatsApp. Ele ajuda a responder:

- quais telas o cliente espera;
- quais dados de negócio existem;
- o que o protótipo cobre e o que não cobre;
- o que a V1 deve reproduzir, adaptar ou deliberadamente não copiar.

O site é uma plataforma de terceiros. **Não copie** marca, código, textos, imagens ou layout sem autorização. O visual do produto continua sendo o design system v2.5.

---

## 2. Método

- Leitura das páginas públicas `/`, `/products`, `/info` e de páginas de produto.
- Extração do JSON-LD (`Restaurant`, `Menu`, `Product`, `BreadcrumbList`).
- Leitura do payload público do Nuxt para configurações expostas.
- Nenhum login, painel ou dado privado foi acessado.
- Nenhum dado foi importado para o banco.
- Valores foram arredondados apenas para leitura; a API usa centavos.

Limitações:

- O site pode mudar após o snapshot.
- Taxas por bairro não aparecem no HTML público; precisam ser validadas.
- O painel administrativo da OlaClick não é público.
- O fluxo de checkout depende de JavaScript e não foi reproduzido inteiro neste levantamento.

---

## 3. Dados públicos observados

| Campo | Valor no snapshot |
|---|---|
| Nome | Brocados Lanches |
| Endereço | R. Lourenço Alves, 364 - Tapera, Aquiraz - CE, 61700-000 |
| WhatsApp | `+55 85 99235-8708` |
| Fuso | `America/Fortaleza` |
| Entrega | 30–90 min |
| Retirada | Sim |
| Checkout | WhatsApp |
| Pagamento | Dinheiro, cartão e Pix, pagos na entrega/retirada |
| Site | OlaClick, versão pública `4.00.68f345e-v2` |

> Contato e endereço são dados públicos da loja, mas ainda precisam de confirmação antes do lançamento.

### 3.1 Horários

| Dia | Horário | Situação |
|---|---|---|
| Domingo | 17:45–23:30 | Aberto |
| Segunda | 18:00–23:30 | Aberto |
| Terça | — | Fechado |
| Quarta | 18:00–23:30 | Aberto |
| Quinta | 18:00–23:30 | Aberto |
| Sexta | 17:30–23:30 | Aberto |
| Sábado | 18:00–23:30 | Aberto |

O site mostra terça como “Fechado”; a configuração interna ainda tem uma faixa 18:00–23:00, mas `is_active: false`. A V1 deve tratar o dia como fechado.

### 3.2 Bairros de entrega

O snapshot lista 30 regiões:

Araçá da Mônica, Araçá do biel, Araçá do manduca, Área verde, Barro preto, Canoa, Canto verde, Catu, Centro de Aquiraz, Divineia, Engenho velho, Estrada Nova, Fagundes, Iguape, Lagoa de cima até o açaí, Machuca, Novo Iguape, Patacas, Pau pombo, Portelinha, Praia bella, Presídio (até a Praia), Presídio (até a rua da bica), Riviera, Sítio extremo, Sítios Novos, Smart city, Tapera, Vila Isabel, Vila pagã.

As taxas não são exibidas no HTML público. **Não invente valores.** A lista e as taxas precisam ser validadas no painel/planilha do cliente.

### 3.3 Pagamentos

| Método | Entrega | Retirada | Online |
|---|---|---|---|
| Dinheiro | Sim | Sim | Não |
| Cartão | Sim | Sim | Não |
| Pix | Sim | Sim | Não |

O site exibe uma chave Pix e orienta enviar o comprovante pelo WhatsApp. A V1 deve mostrar instruções de pagamento **depois** de criar o pedido, sem armazenar dados de cartão. A chave Pix não deve ser copiada para documentação pública sem necessidade.

---

## 4. Catálogo observado

### 4.1 Resumo

| Seção | Itens | Faixa de preço | Esgotados no snapshot |
|---|---:|---:|---:|
| Promoção 🔥 | 3 | R$ 49,99–199,99 | 0 |
| Hambúrguer artesanal 🍔 | 18 | R$ 15,99–34,99 | 4 |
| Hambúrguer tradicionais 🍔 | 4 | R$ 15,00–40,00 | 0 |
| Porções | 8 | R$ 10,00–29,99 | 1 |
| Bebidas 🥤 | 5 | R$ 6,00–13,00 | 3 |
| Pastel | 2 | R$ 8,00–8,00 | 0 |
| **Total** | **40** | — | **8** |

### 4.2 Itens observados

Tabela de conferência; não é catálogo oficial nem seed de produção.

| Seção | Item | Preço |
|---|---|---:|
| Promoção | Combo 3 hambúrguer | 49,99 |
| Promoção | Combo 5 hambúrguer | 89,99 |
| Promoção | Combo 10 hambúrguers | 199,99 |
| Artesanal | Burger Frescurinha | 18,00 |
| Artesanal | Burger Salad | 19,00 |
| Artesanal | Bacon Cheddar | 22,00 |
| Artesanal | Calabreso | 22,00 |
| Artesanal | Nordestino (esgotado) | 24,90 |
| Artesanal | X-Tudo (o fim da dieta) | 28,00 |
| Artesanal | Duplo Cheddar | 26,00 |
| Artesanal | Crocante (esgotado) | 20,00 |
| Artesanal | Onion Bacon | 26,90 |
| Artesanal | Big Salad | 26,00 |
| Artesanal | Tropical Bacon (esgotado) | 28,90 |
| Artesanal | Burger Mexicano | 25,00 |
| Artesanal | Duplo Chedda Bacon | 28,00 |
| Artesanal | Triplo Bacon | 34,99 |
| Artesanal | Broca Doce (esgotado) | 15,99 |
| Artesanal | Big Calabreso | 32,00 |
| Artesanal | Crispy Burger | 24,99 |
| Artesanal | Hambúrguer de Frango | 23,00 |
| Tradicional | Hambúrguer Tradicional | 15,00 |
| Tradicional | Big Broca | 40,00 |
| Tradicional | Brocado | 20,00 |
| Tradicional | Brocadão | 25,00 |
| Porção | Batata P (esgotada) | 10,00 |
| Porção | Batata G | 15,00 |
| Porção | Batata com Cheddar | 18,00 |
| Porção | Batata com Cheddar e Calabresa | 20,00 |
| Porção | Batata com Cheddar e Bacon | 20,00 |
| Porção | Batata Tudão | 29,99 |
| Porção | Anel de Cebola com Geleia de Pimenta | 18,00 |
| Porção | Anel de Cebola | 16,00 |
| Bebida | Coca Lata (esgotada) | 6,00 |
| Bebida | Coca-Cola 1L | 11,00 |
| Bebida | Guaraná 1L | 9,00 |
| Bebida | Fanta Lata (esgotada) | 6,00 |
| Bebida | Coca Zero 1,5L (esgotada) | 13,00 |
| Pastel | Pastel de Chocolate | 8,00 |
| Pastel | Pastel de Doce de Leite | 8,00 |

### 4.3 Nomes e apresentação

- O site usa nomes em caixa alta e emoji em algumas seções.
- A V1 pode manter o nome comercial, mas deve padronizar a exibição e remover emoji de identificadores.
- Descrições, ingredientes e alergênicos não foram copiados; precisam de aprovação.
- Fotos do site pertencem à plataforma/loja; uso exige autorização.

---

## 5. Fluxo observado

### 5.1 Cliente

1. Abre a home e vê status, prazo e link para o cardápio.
2. Navega por seções e busca produtos.
3. Abre o produto, vê descrição, preço, imagem e disponibilidade.
4. Adiciona ao carrinho.
5. Escolhe entrega ou retirada.
6. Informa bairro/endereço e dados.
7. Escolhe pagamento.
8. Finaliza; o checkout é do tipo WhatsApp.

### 5.2 O que não foi possível confirmar

- se o site cria pedido persistido antes do WhatsApp;
- se há idempotência;
- se o link de acompanhamento existe;
- se o painel admin tem papéis e auditoria;
- se há taxa por bairro e qual é;
- se há grupos de opções/adicionais nos produtos (as páginas públicas observadas não exibem);
- se retirada tem fluxo completo no checkout.

Essas lacunas são exatamente o que a V1 precisa resolver melhor que a referência.

---

## 6. Matriz de paridade

| Capacidade | Referência pública | Protótipo v2.5 | Decisão V1 |
|---|---|---|---|
| Cardápio por categorias | Sim | Sim | Sim |
| Busca/filtro | Sim | Sim | Sim |
| Detalhe de produto | Sim | Sim | Sim |
| Grupos de opções | Não visível | Sim | Sim, após validar |
| Adicionais | Não visível | Sim | Sim, após validar |
| Carrinho | Sim | Sim | Sim |
| Delivery por bairro | Sim (30 bairros) | Sim | Sim, com taxas validadas |
| Retirada | Sim | Sim | Sim, taxa zero |
| Dinheiro/cartão/Pix | Sim | Sim | Sim, offline |
| Checkout WhatsApp | Sim | Sim | Sim, após persistir |
| Pedido persistido | Não observável | Não | Sim |
| Idempotência | Não observável | Não | Sim |
| Link público de acompanhamento | Não observável | Não | Sim |
| Admin autenticado | Não observável | Simulado | Sim |
| Papéis owner/operator | Não observável | Não | Sim |
| Auditoria | Não observável | Não | Sim |
| SEO/JSON-LD | Sim | Não | Sim |
| PWA | Não validado | Não | Fora da V1 |
| Analytics | Configuração nula no snapshot | Não | Só com consentimento |
| Pagamento online | Não | Não | Fora da V1 |
| Fidelidade/cupom | Não observável | Não | Fora da V1 |
| Avaliações | Não observável | Não | Fora da V1 |
| KDS/impressão | Não observável | Não | Fora da V1 |
| Multilíngue | Plataforma suporta; público em pt | Não | Só pt-BR |
| Consumo no local | Não observável | Não | Fora da V1 |

---

## 7. O que adotar

1. **Cardápio direto ao ponto:** categorias, busca, preço e disponibilidade.
2. **Delivery e retirada:** deixar a escolha explícita e refletir a taxa.
3. **Bairros como zonas:** usar o conceito, mas validar nomes e taxas.
4. **Pagamento offline:** dinheiro, cartão e Pix; sem gateway.
5. **WhatsApp como fechamento:** manter o hábito do cliente, mas persistir antes.
6. **SEO:** metadados e JSON-LD.
7. **Mobile-first:** o cliente já usa o site no celular.

## 8. O que não copiar

- marca, logo, fotos, textos e layout da OlaClick;
- dependência de plataforma white-label;
- emojis como ícones de interface;
- checkout sem persistência;
- analytics/pixels sem consentimento;
- chave Pix exposta em documentação;
- qualquer taxa, bairro ou item sem validação.

---

## 9. Perguntas para o cliente

1. O catálogo do site é o oficial? Há itens fora do site?
2. Quais grupos de opções e adicionais existem por produto?
3. Quais bairros e taxas valem hoje?
4. O prazo de 30–90 min é real? Varia por bairro?
5. O WhatsApp `+55 85 99235-8708` é o correto?
6. A retirada tem endereço/horário específicos?
7. O Pix do site continua o mesmo? Como o comprovante é conferido?
8. Há feriados, férias ou fechamentos temporários?
9. Quem opera o painel e qual papel?
10. Quais fotos e textos podem ser usados?

---

## 10. Evidências e atualização

- Snapshot feito em 18/11/2026.
- Páginas consultadas: `/`, `/products`, `/info` e páginas de produto.
- Dados copiados manualmente para conferência; nenhuma importação automática.
- Se o site mudar, este documento continua sendo histórico; a fonte operacional é o cliente.
- Atualizar este documento apenas quando houver novo snapshot ou validação.

---

## 11. Links

- Site: https://cardpiobrocadoslanchess.ola.click/
- Cardápio: https://cardpiobrocadoslanchess.ola.click/products
- Informações: https://cardpiobrocadoslanchess.ola.click/info

Links externos não são validados por `make docs-check`.
