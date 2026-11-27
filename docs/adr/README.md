# ADR — Architecture Decision Records

**Atualizado em:** 27/11/2026

Este diretório registra decisões arquiteturais relevantes da Brocados Lanches.

## Índice

| ADR | Decisão | Status |
|---|---|---|
| [0001](0001-stack-e-monorepo.md) | Stack Go + Vue + MongoDB em monorepo | Aceita |
| [0002](0002-persistencia-do-pedido-e-handoff-whatsapp.md) | Persistir pedido antes do WhatsApp | Aceita |
| [0003](0003-dinheiro-e-fuso.md) | Dinheiro em centavos e fuso da operação | Aceita |
| [0004](0004-autenticacao-admin.md) | Sessão em cookie e papéis | Aceita |
| [0005](0005-design-system-e-ui.md) | Design system como fonte visual | Aceita |
| [0006](0006-cotacao-server-side.md) | Cotação server-side antes do pedido | Aceita e implementada |
| [0007](0007-monitoramento-de-erros.md) | Monitoramento de erros com Sentry, sem PII e sem tracing | Aceita e implementada |

## Formato

Cada ADR tem:

- **Status:** proposta, aceita, substituída ou rejeitada.
- **Contexto:** o problema e as restrições.
- **Decisão:** o que foi escolhido.
- **Alternativas:** o que foi considerado.
- **Consequências:** trade-offs e impactos.
- **Referências:** documentos e links.

## Regras

- Uma decisão por ADR.
- Não apagar ADR aceita; marcar como substituída.
- Mudança de arquitetura atualiza o ADR e o documento de arquitetura.
- ADR não substitui `../../AGENTS.md` nem o contrato de API.
