# ADR 0002 — Persistência do pedido antes do handoff para WhatsApp

- **Status:** aceita
- **Data:** 18/11/2026
- **Decisores:** produto e engenharia
- **Relacionados:** `../03-prd.md`, `../04-arquitetura.md`, `../06-contrato-api.md`

## Contexto

O hábito da loja é fechar o pedido pelo WhatsApp. O protótipo monta o pedido em memória e não persiste nada. Se o cliente fechar a aba ou o WhatsApp falhar, o pedido desaparece. Além disso, sem registro no servidor não há fila, status, auditoria nem histórico.

## Decisão

1. O pedido é validado e persistido no MongoDB **antes** de gerar o link do WhatsApp.
2. A resposta `201` inclui `orderNumber`, `publicId`, `status`, `totalCents` e `whatsappUrl`.
3. O WhatsApp é apenas o canal de handoff; não é o banco.
4. Se o handoff falhar, o pedido continua salvo e pode ser reenviado.
5. `POST /orders` exige `Idempotency-Key` para evitar duplicata em retry.
6. O carrinho só é limpo após sucesso ou replay idempotente.
7. O acompanhamento público usa `publicId` opaco e não expõe PII sensível.

## Alternativas consideradas

| Alternativa | Problema |
|---|---|
| Abrir WhatsApp primeiro e salvar depois | Pedido pode ser perdido; não há garantia |
| Persistir só no navegador | Não sobrevive a fechar aba; não há fila |
| Salvar no WhatsApp e importar depois | Manual, sujeito a erro, sem contrato |
| Usar API oficial do WhatsApp na V1 | Custo/contratação e complexidade fora do escopo |
| Salvar sem idempotência | Duplicata em retry ou duplo clique |

## Consequências

Positivas:

- pedido recuperável;
- fila administrativa confiável;
- auditoria e status;
- cliente pode reabrir o link;
- falha do WhatsApp não perde venda.

Negativas:

- é preciso lidar com pedido criado e não enviado;
- exige idempotência e deduplicação;
- o link do WhatsApp pode ser bloqueado pelo navegador;
- a mensagem do WhatsApp pode divergir do pedido se não for gerada do snapshot.

Mitigações:

- confirmação com link copiável e reenvio;
- geração da mensagem a partir do snapshot persistido;
- botão de reenvio na confirmação imediata; após recarregar, contato da loja com o número do pedido;
- logs sem PII;
- testes de falha de handoff.

## Referências

- `../03-prd.md`
- `../04-arquitetura.md`
- `../06-contrato-api.md`
- `../08-backlog.md` (US-040 a US-044)
