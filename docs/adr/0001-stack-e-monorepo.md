# ADR 0001 — Stack e monorepo

- **Status:** aceita
- **Data:** 18/11/2026
- **Decisores:** engenharia do projeto
- **Relacionados:** `../../AGENTS.md`, `../04-arquitetura.md`

## Contexto

O material recebido é um protótipo HTML com Tailwind via CDN, dados em memória e nenhum backend. A V1 precisa de persistência, autenticação, cálculo server-side, testes e deploy. O plano inicial em `docs/02-plano-de-producao.md` mencionava Next.js + Postgres, mas essa decisão foi substituída antes da implementação.

Restrições:

- prazo e equipe pequenos;
- uma loja, um turno;
- necessidade de build offline no Docker;
- familiaridade com Go e Vue;
- MongoDB disponível;
- simplicidade operacional.

## Decisão

Adotar um monorepo com:

- **Backend:** Go + Chi, MongoDB 8+ com driver oficial v2;
- **Frontend:** Vue 3 + Vite + TypeScript, Composition API, Pinia, Vue Router;
- **Estilo:** Tailwind compilado, tokens do design system;
- **Pacotes:** pnpm;
- **Ambiente:** Docker Compose;
- **Testes:** `testing`/`httptest`/testcontainers no backend; Vitest/Vue Test Utils/Playwright no frontend.

Backend e frontend têm builds e deploys independentes, mas são servidos same-origin atrás de um proxy.

## Alternativas consideradas

| Alternativa | Prós | Contras | Por que não |
|---|---|---|---|
| Next.js + Postgres | Ecossistema, SSR | Mudaria stack e exigiria novo plano; SPA atende | Já havia decisão por Go/Vue |
| SPA + Firebase | Rápido | Vendor lock-in, regras no cliente, custo | Servidor precisa ser autoritativo |
| PHP/Laravel | Maturidade | Fora da stack e da equipe | Sem justificativa |
| Microserviços | Escala independente | Complexidade desnecessária | Uma loja, um time |

## Consequências

Positivas:

- uma linguagem de backend simples e testável;
- frontend moderno e tipado;
- build offline com `vendor/`;
- deploy same-origin;
- testes em camadas.

Negativas:

- duas toolchains;
- mais arquivos de build;
- Mongo exige modelagem e índices cuidadosos;
- transações exigem replica set se forem necessárias.

Mitigações:

- Makefile como contrato;
- CI único para os dois lados;
- `backend/vendor/` versionado;
- documentação de arquitetura e dados.

## Referências

- `../../AGENTS.md`
- `../04-arquitetura.md`
- `../05-modelo-de-dados.md`
- `../09-qualidade-e-testes.md`
