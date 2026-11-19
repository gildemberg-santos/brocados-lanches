# ADR 0005 — Design system como fonte visual

- **Status:** aceita
- **Data:** 18/11/2026
- **Decisores:** design e engenharia
- **Relacionados:** `../07-design-system-e-experiencia.md`, `../../design-system/README.md`

## Contexto

Existem duas referências visuais possíveis: o design system v2.5 entregue pelo cliente e o site público da OlaClick. Elas têm identidade, layout e tecnologia diferentes. O protótipo também usa Tailwind CDN, Google Fonts e dados fictícios, o que não pode ir para produção.

## Decisão

- O **design system v2.5** é a fonte de verdade visual.
- O site da OlaClick é referência **funcional e de conteúdo**, não visual.
- Tokens vêm de `design-system/tokens.css` e `tokens.json`; não usar cor/fonte literal.
- Tailwind é compilado; CDN é proibido.
- Bebas Neue e Inter são self-hosted; Google Fonts é proibido em produção.
- Componentes reutilizáveis antes de criar variante.
- `make design-check` é gate obrigatório.
- Mudança de token/componente exige aprovação de design e atualização conjunta de `tokens.css`, `tokens.json` e documentação.
- O protótipo HTML não é editado nem publicado como produto.

## Alternativas consideradas

| Alternativa | Problema |
|---|---|
| Copiar o visual da OlaClick | Marca de terceiro; perde identidade |
| Criar visual novo | Sem aprovação; retrabalho |
| Usar CDN do protótipo | Viola AGENTS; performance/segurança |
| Editar tokens por conta própria | Quebra contrato visual |

## Consequências

Positivas:

- consistência;
- aprovação clara;
- validação automática;
- identidade própria.

Negativas:

- faltam tokens de espaçamento, raio, sombra, movimento e z-index;
- o protótipo tem problemas de acessibilidade que precisam ser corrigidos;
- qualquer evolução visual exige design.

Mitigações:

- lacunas documentadas em `../07-design-system-e-experiencia.md`;
- valores derivados do protótipo até aprovação;
- revisão visual humana;
- checklist de acessibilidade.

## Referências

- `../07-design-system-e-experiencia.md`
- `../12-referencia-ola-click.md`
- `../../design-system/README.md`
- `../../design-system/tokens.css`
- `../../design-system/tokens.json`
