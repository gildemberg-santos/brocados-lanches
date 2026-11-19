# ADR 0006 — Cotação server-side antes do pedido

- **Status:** aceita no planejamento
- **Data:** 18/11/2026
- **Decisores:** produto e engenharia
- **Relacionados:** `../03-prd.md`, `../04-arquitetura.md`, `../06-contrato-api.md`

## Contexto

O checkout precisa mostrar subtotal, taxa e total antes de o cliente enviar o pedido. Se o frontend calcular, ele pode divergir do servidor, usar preço velho ou ser adulterado. Se o total só aparecer depois do `POST /orders`, o cliente confirma sem saber o valor final.

O contrato original tinha apenas `POST /orders` e `GET /orders/{publicId}`.

## Decisão

Adicionar `POST /orders/quote`:

- recebe itens, `fulfillment` e `zoneId`, sem dados pessoais;
- valida loja, produtos, opções, adicionais e zona;
- devolve subtotal, taxa, total e itens normalizados;
- não persiste pedido, não reserva estoque e não gera número;
- usa a mesma função de cálculo do `POST /orders`;
- tem rate limit próprio.

O `POST /orders` recalcula tudo e pode rejeitar se algo mudou entre a cotação e o envio. A cotação é uma estimativa autoritativa do momento, não uma garantia de preço futuro.

## Alternativas consideradas

| Alternativa | Problema |
|---|---|
| Calcular no frontend | Servidor deixa de ser a única fonte; fraude/divergência |
| Só mostrar total após criar pedido | Cliente confirma sem saber o valor |
| Criar pedido em rascunho | Persistência desnecessária, limpeza, estados extras |
| `dryRun` no `POST /orders` | Contamina semântica e idempotência |
| Não ter cotação | Pior UX e risco de erro |

## Consequências

Positivas:

- total confiável antes do envio;
- menos erro de preço;
- uma função de cálculo compartilhada;
- sem PII no endpoint.

Negativas:

- mais um endpoint para manter;
- possibilidade de mudança entre quote e order;
- exige rate limit e testes de consistência.

Mitigações:

- cálculo compartilhado no domínio/service;
- testes de equivalência quote/order;
- mensagem clara se o preço mudar;
- `POST /orders` sempre revalida.

## Contrato

```http
POST /api/v1/orders/quote
```

Ver `../06-contrato-api.md` seção 6.

## Referências

- `../03-prd.md`
- `../04-arquitetura.md`
- `../06-contrato-api.md`
- `../13-decisoes-e-pendencias.md` (D-008)
