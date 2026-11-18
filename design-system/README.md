# Design System — Brocados Lanches

**Versão:** 2.5  
**Fonte de verdade:** `tokens.css` (variáveis CSS) e `tokens.json` (consumo por ferramentas).  
**Referência visual:** `brocados-lanches-design-system-v2.5.html`.

---

## 1. Regra do projeto

Nenhuma tela, componente ou estilo pode criar cor, fonte, escala, espaçamento ou padrão visual fora deste design system.

- Use os tokens de `tokens.css`; não escreva hex, `rgb()` ou `hsl()` no frontend.
- Use Tailwind compilado e o tema derivado dos tokens; CDN é proibido.
- Bebas Neue e Inter são self-hosted em produção; Google Fonts via CDN é proibido.
- O protótipo HTML é referência visual e funcional, não código de produção.
- Alterar token, componente ou padrão visual exige aprovação de design e atualização de `tokens.css` e `tokens.json` juntos.

---

## 2. Validação obrigatória

```bash
make design-check
```

O validador `scripts/check_design_system.py` confere:

- consistência entre `tokens.css` e `tokens.json`;
- cores literais (`#...`, `rgb`, `rgba`, `hsl`, `hsla`) em `frontend/`;
- Tailwind CDN e Google Fonts CDN em `frontend/`.

O mesmo check roda no hook `.githooks/pre-commit`. Nenhum commit passa sem ele.

---

## 3. Mudanças no design system

1. Obtenha aprovação de design antes de alterar tokens, componentes ou identidade.
2. Atualize `tokens.css` e `tokens.json` na mesma alteração.
3. Atualize a documentação afetada (`AGENTS.md` e este arquivo).
4. Rode `make design-check`, `make docs-check` e `make test`.
5. Registre a mudança no commit e no PR, com evidência visual quando houver UI.
