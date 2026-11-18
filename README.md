# Brocados Lanches — Projeto

**Status:** protótipo de design system recebido; escopo de produção ainda não fechado.  
**Data de entrada:** 18/11/2026  
**Origem:** `~/Downloads/brocados_lanches_design_system_responsive_prototype_studio.html`  
**Classificação:** material de cliente — uso interno.

---

## 1. Objetivo da pasta

Reunir o material do novo chamado **Brocados Lanches** em um único lugar:

- o protótipo responsivo e o design system recebidos;
- os tokens de marca e os dados de demonstração separados do HTML;
- o inventário do que já está pronto e do que ainda é simulação;
- o briefing do chamado e as pendências para virar produção.

Esta pasta **não é um projeto pronto para deploy**. O arquivo principal é um protótipo navegável feito em HTML, CSS e JavaScript, com dados em memória.

**Stack de produção definida:** Go + Chi (backend), Vue 3 + Vite + TypeScript (frontend) e MongoDB. As regras de arquitetura, testes e commit estão em `AGENTS.md`.

---

## 2. O que já existe

O arquivo `design-system/brocados-lanches-design-system-v2.5.html` contém:

| Bloco | Conteúdo |
|---|---|
| Design System | 6 frames: identidade e paleta, escala tipográfica, botões e controles, cards e inputs, responsividade e rodapé, componentes admin e live sync |
| Protótipo do cliente | 6 telas: Home, Cardápio, Personalizar, Carrinho, Checkout e Créditos |
| Painel administrativo | 5 telas: Dashboard e Status, Gestão de Produtos, Categorias, Locais e Taxas, Conta e Contato |
| Responsividade | Visualização em mobile (375 px) e desktop (1280 px), com ajuste de escala no desktop |
| Live sync | Alterações do admin refletem na visão do cliente dentro da mesma sessão |
| Exportação | Exporta variáveis CSS, extensão de tema Tailwind e JSON com loja, categorias, produtos e zonas de entrega |

### Identidade visual documentada

| Token | Valor |
|---|---|
| Fundo principal | `#0A0A0A` |
| Superfície | `#161616` |
| Input | `#1F1F1F` |
| Borda | `#2E2E2E` |
| Primária | `#FF8C00` |
| Hover | `#E07B00` |
| Neon | `#FFB703` |
| Texto principal | `#FFFFFF` |
| Texto secundário | `#A3A3A3` |
| Loja aberta | `#22C55E` |
| Erro / fechado | `#EF4444` |

**Tipografia:** Bebas Neue para display e Inter para interface.

### Dados de demonstração

O protótipo vem com 4 categorias, 5 produtos, 3 zonas de entrega e 2 pedidos recentes fictícios. Os dados foram extraídos para `design-system/seed.json`.

> **Atenção:** telefone `(85) 99999-9999`, endereço `Av. Principal, 1200 - Centro, Aquiraz - CE` e todos os preços são **placeholders**. Precisam ser confirmados com o cliente antes de qualquer uso público.

---

## 3. Estrutura

```text
brocados-lanches/
├── README.md
├── AGENTS.md
├── Makefile
├── .gitignore
├── .githooks/
│   └── pre-commit
├── scripts/
│   ├── check_docs.py
│   └── test_check_docs.py
├── briefing/
│   └── chamado.md
├── design-system/
│   ├── brocados-lanches-design-system-v2.5.html
│   ├── tokens.css
│   ├── tokens.json
│   └── seed.json
├── docs/
│   ├── 01-inventario-prototipo.md
│   └── 02-plano-de-producao.md
├── assets/
└── referencias/
```

- `assets/`: logos, fotos e arquivos originais do cliente.
- `referencias/`: links, concorrentes, materiais de apoio e versões anteriores.

---

## 4. Como abrir o protótipo

Abra o arquivo diretamente no navegador:

```bash
xdg-open design-system/brocados-lanches-design-system-v2.5.html
```

O protótipo usa Tailwind via CDN e Google Fonts. Portanto, **precisa de internet** para carregar o visual completo. Não há build, servidor ou instalação.

---

## 5. Limites do protótipo

O protótipo é uma peça de apresentação e validação. Ele **não** entrega:

- persistência: recarregar a página zera carrinho, produtos e configurações;
- pedido real: o botão “Finalizar no WhatsApp” apenas simula o envio e adiciona um pedido recente na tela do admin — não abre o WhatsApp, não envia mensagem e não grava nada;
- backend, banco de dados, API ou autenticação;
- integração de pagamento;
- controle de acesso ao painel administrativo;
- imagens reais dos produtos (os ícones são SVGs genéricos);
- SEO, acessibilidade auditada, testes automatizados ou política de privacidade;
- Tailwind compilado: em produção, o CDN deve ser substituído por build próprio.

Esses pontos estão detalhados em `docs/01-inventario-prototipo.md` e priorizados em `docs/02-plano-de-producao.md`.

---

## 6. Próximos passos

1. Preencher `briefing/chamado.md` com número do chamado, solicitante, prazo e escopo.
2. Confirmar dados reais da loja: nome, endereço, WhatsApp, Instagram, horários e tempo médio de entrega.
3. Validar cardápio, categorias, preços, adicionais e zonas de entrega.
4. Definir se o pedido será fechado no WhatsApp ou dentro do site.
5. Definir hospedagem, domínio e responsável pela operação (stack já decidida: Go + Chi, Vue 3 + Vite + TypeScript e MongoDB).
6. Só então iniciar a implementação de produção.

---

## 7. Documentos relacionados

- `AGENTS.md` — regras de arquitetura, testes, commit e operação para agentes e time.
- `briefing/chamado.md` — o que se sabe sobre o chamado e o que falta.
- `docs/01-inventario-prototipo.md` — inventário funcional e técnico do HTML.
- `docs/02-plano-de-producao.md` — fases propostas para tirar o protótipo do papel.
- `Makefile` — comandos de entrada do projeto.
- `scripts/check_docs.py` — validador de links da documentação.
- `design-system/tokens.css` — variáveis de marca prontas para uso.
- `design-system/tokens.json` — os mesmos tokens em formato consumível por ferramentas.
- `design-system/seed.json` — dados de demonstração para popular um banco de desenvolvimento.

---

## 8. Versionamento e validação

O projeto é versionado com Git. Antes de qualquer commit, documentação e testes precisam estar atualizados e validados.

```bash
git config core.hooksPath .githooks   # ativa o gate uma vez por clone
make docs-check                       # valida links da documentação
make test                             # roda os testes disponíveis
```

O hook `.githooks/pre-commit` bloqueia arquivos de ambiente e segredos, valida a documentação e executa os testes. Nenhum commit deve ser criado com `--no-verify`.
