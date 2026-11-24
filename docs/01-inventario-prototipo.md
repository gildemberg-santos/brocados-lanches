# Inventário do Protótipo — Brocados Lanches

**Arquivo:** `../design-system/brocados-lanches-design-system-v2.5.html`  
**Versão do design system:** 2.5  
**Tamanho:** ~130 KB, 2.023 linhas  
**Data da análise:** 18/11/2026

---

## 1. Ficha técnica

| Item | Detalhe |
|---|---|
| Tipo | Protótipo navegável em página única |
| HTML | Estrutura semântica básica, `lang="pt-BR"` |
| CSS | Tailwind CSS via CDN + estilos inline no `<head>` |
| JavaScript | Vanilla JS, sem framework e sem build |
| Fontes | Bebas Neue (display) e Inter (interface), via Google Fonts |
| Estado | Objeto `appState` em memória |
| Persistência | Nenhuma (`localStorage` não é usado) |
| Backend/API | Nenhum |
| Imagens | Ícones SVG desenhados inline; sem fotos |
| Exportação | CSS/Tailwind e JSON gerados no navegador |

**Consequência direta:** o protótipo serve para apresentar e validar o visual. Recarregar a página restaura todos os dados originais.

---

## 2. Design System

Seis frames documentados no próprio protótipo:

| Frame | Conteúdo |
|---|---|
| 01 — Identidade e paleta | Cores de fundo, superfície, input, borda, primária, hover, neon, textos e status |
| 02 — Escala tipográfica | Display 36 px, H1 22 px, H2 16 px, preço 16 px, corpo 14 px, legenda 12 px |
| 03 — Botões e controles | Estados default, hover e disabled; chips; voltar; stepper; radio e checkbox |
| 04 — Cards e inputs | Card de produto 1:1, linha de adicional, grade responsiva e sidebar de carrinho |
| 05 — Responsivo e rodapé | Grids, navegação mobile, navbar desktop, painel e rodapé persistente |
| 06 — Admin e live sync | Controle master da loja, status de produto e sincronização com a visão do cliente |

### Personalização ao vivo

O topo do protótipo permite alterar:

- cor primária;
- cor neon;
- cor de superfície;
- resetar as cores.

As mudanças atualizam variáveis CSS e afetam os componentes na hora. Não são salvas ao recarregar.

### Exportações

- **CSS & Tailwind:** variáveis `:root` e extensão de tema.
- **JSON:** `store`, `categories`, `products` e `delivery_zones`.

Os tokens também foram extraídos para `../design-system/tokens.css`, `tokens.json` e `seed.json`.

---

## 3. Telas do cliente (6)

| Tela | O que faz |
|---|---|
| Home | Hero de destaque, status da loja, tempo médio, categorias, mais pedidos, endereço e carrinho |
| Cardápio | Busca por nome/descrição, filtro por categoria, grade de produtos, selo “Esgotado” e adição rápida |
| Personalizar | Ponto da carne, adicionais pagos, observações e preço final antes de adicionar |
| Carrinho | Itens, quantidade, subtotal, taxa da zona selecionada, total e avanço para checkout |
| Checkout | Delivery ou retirada, bairro/taxa, endereço, pagamento (Pix, Cartão, Dinheiro), troco e total |
| Créditos | Apresentação do projeto, stack visual e atalhos para Home e Cardápio |

### Regras de negócio embutidas

- A taxa de entrega vem da zona selecionada; retirada zera a taxa.
- Produto indisponível não abre personalização e aparece como “Esgotado”.
- Zonas e categorias inativas deixam de aparecer no fluxo do cliente.
- Ponto da carne é obrigatório (`Ao ponto` ou `Bem passado`).
- Adicionais: Bacon Crocante (R$ 5), Cheddar Cremoso Extra (R$ 4) e Blend 160g Extra (R$ 9).
- O carrinho inicia com um item de exemplo.

### Checkout — comportamento real

O botão **“Finalizar no WhatsApp”**:

1. verifica se a loja está aberta;
2. verifica se o carrinho tem itens;
3. cria um pedido fictício em `recentOrders`;
4. mostra um toast de sucesso.

Ele **não** abre o WhatsApp, **não** envia mensagem, **não** grava pedido e **não** chama nenhuma API. É uma simulação de fluxo.

---

## 4. Painel administrativo (5 telas)

| Tela | Recursos |
|---|---|
| Dashboard e Status | Controle master aberto/fechado, contadores de produtos e bairros, tempo médio e pedidos recentes |
| Gestão de Produtos | Cadastro de nome, preço, categoria e descrição; alternância disponível/esgotado; destaque |
| Categorias | Cadastro e ativação/desativação de categorias |
| Locais e Taxas | Cadastro de bairro e taxa; ativação/desativação de zonas |
| Conta e Contato | Nome da loja, WhatsApp, tempo de entrega, Instagram e endereço |

Todas as alterações são refletidas imediatamente na visão do cliente **dentro da mesma página**. Nada é persistido.

### Observações do admin

- Não existe login, perfil ou permissão.
- Não existe histórico de pedidos além dos dois exemplos.
- Não existe impressão, KDS, som ou notificação.
- O cadastro de produto não inclui foto, alergênicos, estoque ou disponibilidade por horário.
- O formulário de conta não valida telefone, endereço ou redes sociais.

---

## 5. Dados de demonstração

### Loja

```json
{
  "name": "Brocados Lanches",
  "isOpen": true,
  "deliveryTime": "35-50 min",
  "whatsapp": "(85) 99999-9999",
  "instagram": "@brocadoslanches",
  "address": "Av. Principal, 1200 - Centro, Aquiraz - CE"
}
```

### Catálogo

| Categoria | Produto | Preço |
|---|---|---|
| Hambúrgueres | Brocado's Burguer Cheddar | R$ 23,00 |
| Hambúrgueres | Brocadão Duplo Cheddar | R$ 30,00 |
| Combos | Combo Casal Brocado | R$ 54,90 |
| Petiscos | Batata Cheddar & Bacon | R$ 22,00 |
| Bebidas | Coca-Cola Lata 350ml | R$ 6,00 |

### Zonas de entrega

| Zona | Taxa |
|---|---|
| Centro | R$ 5,00 |
| Prainha | R$ 8,00 |
| Porto das Dunas | R$ 12,00 |

Os dados completos estão em `../design-system/seed.json`.

> No protótipo, telefone, endereço, preços, taxas e produtos são fictícios. O seed de desenvolvimento atual usa categorias, produtos e imagens copiadas do catálogo público como referência visual (ver `docs/13-decisoes-e-pendencias.md`), mas loja, zonas e pedidos seguem fictícios. Não usar em produção sem validação do cliente.

---

## 6. Lacunas e riscos

| Lacuna | Impacto | Prioridade sugerida |
|---|---|---|
| Pedido não é registrado no servidor | Pedido pode ser perdido; não há comprovante nem histórico | P0 |
| Preços e taxas calculados só no navegador | Cliente pode adulterar valores | P0 |
| Painel sem autenticação | Qualquer pessoa altera catálogo e status | P0 |
| Dados em memória | Recarregar perde tudo | P0 |
| Tailwind via CDN | Dependência externa, peso e risco em produção | P1 |
| Sem backend, banco ou API | Não há operação real | P1 |
| Sem integração de pagamento | Checkout não cobra | P1 |
| Sem política de privacidade/LGPD | Risco jurídico se houver dados pessoais | P1 |
| Sem fotos reais | Baixa conversão e aparência de protótipo | P2 |
| Sem SEO, acessibilidade auditada ou testes | Qualidade não mensurada | P2 |
| Textos e preços fictícios | Divergência com o negócio | P1 |

---

## 7. Recomendação

Usar o protótipo como **referência visual e funcional**, não como código de produção.

O caminho mais seguro é:

1. validar marca, cardápio, taxas e fluxo de pedido;
2. implementar frontend e backend com os tokens extraídos;
3. registrar o pedido no servidor antes de qualquer envio;
4. proteger o admin com autenticação e permissões;
5. substituir o CDN por build;
6. auditar acessibilidade, SEO, segurança e LGPD antes do lançamento.

O plano detalhado está em `11-roadmap.md` e `08-backlog.md`. A análise do site público usado como referência funcional está em `12-referencia-ola-click.md`; ela não substitui a validação do cliente.
