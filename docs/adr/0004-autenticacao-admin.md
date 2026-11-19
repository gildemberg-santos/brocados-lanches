# ADR 0004 — Autenticação e autorização do admin

- **Status:** aceita
- **Data:** 18/11/2026
- **Decisores:** engenharia e segurança
- **Relacionados:** `../../AGENTS.md`, `../04-arquitetura.md`, `../06-contrato-api.md`

## Contexto

O painel administrativo altera catálogo, taxas, loja e pedidos. O protótipo não tem login e qualquer visitante poderia, em tese, mudar tudo. A V1 precisa de autenticação simples, sem MFA obrigatório, mas com base segura.

## Decisão

- Senha com **Argon2id**; mínimo de 12 caracteres.
- Sessão com token aleatório de 32 bytes; somente o **SHA-256** do token é guardado no MongoDB.
- Cookie `HttpOnly`, `SameSite=Lax`, `Path=/`; `Secure` e prefixo `__Host-` em produção.
- Expiração ociosa de 8 h e absoluta de 7 dias; TTL no banco.
- Login rotaciona a sessão; logout apaga.
- Papéis: `owner` (tudo, incluindo usuários e auditoria) e `operator` (catálogo, zonas, pedidos e status da loja).
- Autorização é verificada no servidor em toda rota admin.
- CSRF: validar `Origin`/`Referer` em métodos mutáveis; se deixar de ser same-origin, usar token double-submit.
- Rate limit de login: 5/min por IP + e-mail; erro genérico.
- Não logar senha, token, cookie ou corpo de login.

## Alternativas consideradas

| Alternativa | Problema |
|---|---|
| JWT em `localStorage` | XSS rouba token; difícil revogar |
| Basic Auth | Sem sessão, sem papéis, sem rotação |
| OAuth externo | Complexidade e dependência para uma loja |
| Sem auth no admin | Inaceitável |
| MFA obrigatório na V1 | Aumenta escopo; pode ser P2 |

## Consequências

Positivas:

- token inacessível ao JavaScript;
- revogação simples;
- papéis claros;
- TTL e rotação.

Negativas:

- precisa de proteção CSRF;
- cookie entre domínios exige same-origin;
- recuperação de senha não está na V1;
- MFA fica como dívida de segurança.

Mitigações:

- same-origin;
- Argon2id;
- rate limit;
- auditoria;
- checklist de segurança no go-live;
- MFA reavaliado antes do lançamento.

## Referências

- `../../AGENTS.md`
- `../04-arquitetura.md`
- `../06-contrato-api.md`
- `../09-qualidade-e-testes.md`
