# Contrato de API — Brocados Lanches

**Versão:** 1.4  
**Base:** `/api/v1`  
**Atualizado em:** 26/11/2026  
**Relacionados:** `04-arquitetura.md`, `05-modelo-de-dados.md`, `09-qualidade-e-testes.md`.

---

## 1. Convenções

- JSON em `camelCase`.
- Sucesso: `{ "data": ... }`.
- Lista paginada: `{ "data": [...], "meta": { "page": 1, "pageSize": 20, "total": 0, "totalPages": 0 } }`.
- Erro: `{ "error": { "code": "VALIDATION_ERROR", "message": "...", "details": [...] } }`.
- Datas em RFC 3339 UTC.
- IDs públicos em string hexadecimal de 24 caracteres, exceto `publicId` (base64url opaco).
- Dinheiro em centavos (`int64`).
- `Content-Type: application/json` em corpos JSON.
- Limite de corpo: 1 MB.
- Campos desconhecidos são rejeitados.
- `POST /orders` exige `Idempotency-Key` e aceita `Authorization: Bearer <accessToken>` opcional.
- `GET /orders/history` exige `Authorization: Bearer <accessToken>`; o token nunca vai em query string, URL ou log.
- `201` inclui `Location`; `204` não tem corpo.
- Respostas com dados pessoais ou estado de pedido usam `Cache-Control: no-store`.
- Mudança incompatível exige `/api/v2`.

### 1.1 Health checks

Fora de `/api/v1`:

| Método | Rota | Resposta |
|---|---|---|
| GET | `/healthz` | `200` se o processo está vivo |
| GET | `/readyz` | `200` se o Mongo responde; `503` caso contrário |

---

## 2. Endpoints públicos

| Método | Rota | Função |
|---|---|---|
| GET | `/store` | Status efetivo, horários, contato, prazo |
| GET | `/menu` | Categorias, produtos, opções e adicionais ativos |
| GET | `/delivery-zones` | Zonas ativas com taxa |
| POST | `/customers/profile` | Consulta de conveniência por telefone (D-039) |
| POST | `/orders/quote` | Cotação server-side sem persistir |
| POST | `/orders` | Cria pedido persistido |
| GET | `/orders/history` | Histórico do navegador por token opaco |
| GET | `/orders/{publicId}` | Acompanhamento público mínimo |
| WS | `/orders/{publicId}/events` | Sinais de mudança de status do pedido |

---

## 3. `GET /store`

Retorna o status efetivo calculado no servidor.

```json
{
  "data": {
    "name": "Brocados Lanches",
    "status": "open",
    "isOpenManual": true,
    "timezone": "America/Fortaleza",
    "deliveryTime": "30-90 min",
    "whatsapp": "+5585992358708",
    "instagram": "@brocadoslanches",
    "address": "R. Lourenço Alves, 364 - Tapera, Aquiraz - CE, 61700-000",
    "openingHours": [
      { "day": "sunday", "closed": false, "intervals": [{ "open": "17:45", "close": "23:30" }] },
      { "day": "monday", "closed": false, "intervals": [{ "open": "18:00", "close": "23:30" }] },
      { "day": "tuesday", "closed": true, "intervals": [] }
    ],
    "nextOpeningAt": "2026-11-19T21:00:00Z",
    "updatedAt": "2026-11-18T20:00:00Z"
  }
}
```

Regras:

- `status` é `open` ou `closed`.
- `status = open` somente se o horário permitir e `isOpenManual` for `true`.
- `nextOpeningAt` pode ser `null`.
- O frontend não recalcula o status.

---

## 4. `GET /menu`

Retorna somente categorias ativas, produtos ativos, opções disponíveis e adicionais ativos.

```json
{
  "data": {
    "categories": [
      {
        "id": "665f1c2a9b1e4a0012a3b4c1",
        "name": "Hambúrguer artesanal",
        "slug": "hamburguer-artesanal",
        "sortOrder": 1,
        "products": [
          {
            "id": "665f1c2a9b1e4a0012a3b4c6",
            "name": "BACON CHEDDAR",
            "slug": "bacon-cheddar",
            "description": "Pão brioche, carne artesanal 120g, cheddar e bacon.",
            "priceCents": 2200,
            "imageUrl": "https://cdn.example.com/products/bacon-cheddar.webp",
            "imageAlt": "Hambúrguer com bacon e cheddar",
            "allergens": ["glúten", "leite"],
            "available": true,
            "featured": true,
            "sortOrder": 1,
            "optionGroups": [
              {
                "id": "doneness",
                "name": "Ponto da carne",
                "required": true,
                "minSelections": 1,
                "maxSelections": 1,
                "options": [
                  { "id": "medium", "name": "Ao ponto", "priceCents": 0, "available": true }
                ]
              }
            ],
            "addons": [
              { "id": "665f1c2a9b1e4a0012a3b4d1", "name": "Bacon Crocante", "priceCents": 500, "available": true }
            ]
          }
        ]
      }
    ],
    "updatedAt": "2026-11-18T20:00:00Z"
  }
}
```

Regras:

- Produto ativo e `available: false` aparece para permitir “Esgotado”.
- Produto inativo não aparece.
- Categoria sem produto ativo é omitida.
- Adicional só aparece se ativo e vinculado ao produto; como só adicionais ativos são retornados, `available` é `true` na V1.
- Opção indisponível não aparece. Um grupo sem nenhuma opção disponível permanece com `options: []` para que o cliente bloqueie a compra.
- Categorias, produtos, opções e adicionais são ordenados por `sortOrder` e, em empate, por `name`; grupos de opções preservam a ordem definida no produto.
- `updatedAt` é o maior `updatedAt` entre categorias, produtos e adicionais retornados; é omitido quando o cardápio está vazio.
- O cliente não envia preço; o servidor é a fonte.
- Cache HTTP curto pode ser usado; nenhuma resposta cacheada pode sobreviver a uma mudança crítica sem invalidação.

---

## 5. `GET /delivery-zones`

```json
{
  "data": [
    { "id": "665f1c2a9b1e4a0012a3b4e1", "name": "Tapera", "feeCents": 800, "sortOrder": 1 },
    { "id": "665f1c2a9b1e4a0012a3b4e2", "name": "Centro de Aquiraz", "feeCents": 1000, "sortOrder": 2 }
  ]
}
```

Somente zonas ativas, ordenadas por `sortOrder` e, em empate, por `name`. A taxa é sempre do servidor. Lista vazia retorna `data: []`.

---

## 5.1. `POST /customers/profile`

Consulta de conveniência usada pelo checkout depois que o cliente autoriza o tratamento e informa um telefone brasileiro completo. Não cria, altera nem apaga perfil, não consulta pedidos e não serve para marketing. O telefone vai no corpo — nunca na URL — e toda resposta usa `Cache-Control: no-store`.

### Request

```http
POST /api/v1/customers/profile
Content-Type: application/json
```

```json
{ "phone": "(85) 99999-9999" }
```

### Response

```json
{
  "data": {
    "found": true,
    "profile": {
      "name": "Maria Silva",
      "phone": "(85) 99999-9999",
      "fulfillment": "delivery",
      "zone": { "id": "665f1c2a9b1e4a0012a3b4e1", "name": "Tapera" },
      "address": {
        "street": "Rua das Flores",
        "number": "450",
        "complement": "Apto 2",
        "district": "Tapera",
        "city": "Aquiraz",
        "state": "CE",
        "zipCode": "61700000"
      },
      "paymentMethod": "pix"
    }
  }
}
```

Perfil ausente ou desativado por opt-out retorna `{"data":{"found":false}}`, sem distinção entre os dois casos. A resposta nunca inclui `feeCents`, `orderCount`, `lastOrderPublicId`, `note`, `changeForCents`, histórico de pedidos ou `phoneKey`. Zona retornada é apenas sugestão: o cliente descarta zona inativa e a taxa continua sendo calculada por `POST /orders/quote` e revalidada em `POST /orders`.

`422 VALIDATION_ERROR` para telefone incompleto ou fora do formato; `400 VALIDATION_ERROR` para JSON inválido ou campo desconhecido. Limites em `§15`.

---

## 6. `POST /orders/quote`

Calcula totais sem persistir pedido e sem receber PII.

### Request

```http
POST /api/v1/orders/quote
Content-Type: application/json
```

```json
{
  "items": [
    {
      "productId": "665f1c2a9b1e4a0012a3b4c6",
      "quantity": 2,
      "optionIds": ["medium"],
      "addonIds": ["665f1c2a9b1e4a0012a3b4d1"],
      "note": "sem cebola"
    }
  ],
  "fulfillment": "delivery",
  "zoneId": "665f1c2a9b1e4a0012a3b4e1"
}
```

O cliente envia somente identificadores, quantidades e escolhas. `optionIds` contém escolhas dos grupos de opções do produto; `addonIds` contém adicionais vinculados ao produto. Campos desconhecidos são rejeitados. Limites: 50 itens, quantidade de 1 a 99 por item e observação de até 140 caracteres. Corpo JSON de até 1 MB.

### Response `200`

```json
{
  "data": {
    "subtotalCents": 5400,
    "deliveryFeeCents": 800,
    "totalCents": 6200,
    "currency": "BRL",
    "items": [
      {
        "productId": "665f1c2a9b1e4a0012a3b4c6",
        "name": "BACON CHEDDAR",
        "quantity": 2,
        "unitPriceCents": 2200,
        "optionsPriceCents": 500,
        "lineTotalCents": 5400,
        "options": [
          { "name": "Ao ponto", "priceCents": 0 },
          { "name": "Bacon Crocante", "priceCents": 500 }
        ]
      }
    ],
    "warnings": []
  }
}
```

Regras:

- Não cria pedido, não reserva estoque e não gera número. Não há necessidade de `Idempotency-Key` porque nada é gravado.
- Valida loja, produto, opções, adicionais e zona. Status efetivo da loja = kill switch ligado **e** horário vigente em `America/Fortaleza`; quando não há `openingHours` cadastrados, o kill switch decide.
- `pickup` ignora `zoneId` e retorna taxa zero.
- `unitPriceCents` é o preço base do produto; `optionsPriceCents` soma opções e adicionais; `lineTotalCents = (unitPriceCents + optionsPriceCents) * quantity`.
- Pode retornar `STORE_CLOSED`, `PRODUCT_UNAVAILABLE`, `ZONE_UNAVAILABLE` ou `VALIDATION_ERROR`.
- Rate limit ativo: 30/min por IP; detalhes em `§15 Rate limits`.
- O total da cotação é uma estimativa autoritativa do momento; o `POST /orders` recalcula e pode rejeitar se algo mudou. O cálculo está centralizado em `domain.BuildQuote` e é reutilizado pelo `POST /orders`.

---

## 7. `POST /orders`

### Headers

```http
POST /api/v1/orders
Content-Type: application/json
Idempotency-Key: 4d1f6b1e-7c2a-4e2f-9f2a-1c3d5e7f9a0b
Authorization: Bearer 7h3K9pQ2vX8mR5tY1wZ0aB6cD4eF7gH9jK2lM5nP8sA (opcional)
```

O `Authorization` é opcional na criação. Quando ausente, o servidor emite um token a partir da chave de idempotência e de `ORDER_ACCESS_SECRET`; quando presente, o pedido é associado ao token já guardado no navegador. Um cabeçalho `Authorization` malformado ou com token inválido retorna `400 VALIDATION_ERROR`.

### Request

```json
{
  "items": [
    {
      "productId": "665f1c2a9b1e4a0012a3b4c6",
      "quantity": 2,
      "optionIds": ["medium"],
      "addonIds": ["665f1c2a9b1e4a0012a3b4d1"],
      "note": "sem cebola"
    }
  ],
  "fulfillment": "delivery",
  "zoneId": "665f1c2a9b1e4a0012a3b4e1",
  "customer": {
    "name": "Maria Silva",
    "phone": "+5585999999999"
  },
  "address": {
    "street": "Rua das Flores",
    "number": "450",
    "complement": "Apto 2",
    "district": "Tapera",
    "city": "Aquiraz",
    "state": "CE",
    "zipCode": "61700-000"
  },
  "payment": {
    "method": "cash",
    "changeForCents": 10000
  },
  "saveCustomerProfile": true
}
```

`saveCustomerProfile` é opcional; ausência significa `true`. Quando `false`, o servidor desativa o perfil de conveniência daquele telefone e remove os dados dele, sem afetar o pedido gravado. Preço, taxa, total e disponibilidade nunca vêm do cliente; a preferência também não entra no hash de idempotência.

Para retirada:

```json
{
  "items": [{ "productId": "665f1c2a9b1e4a0012a3b4c6", "quantity": 1, "optionIds": ["medium"] }],
  "fulfillment": "pickup",
  "customer": { "name": "Maria Silva", "phone": "+5585999999999" },
  "payment": { "method": "pix" }
}
```

### Response `201`

```json
{
  "data": {
    "orderNumber": 1042,
    "publicId": "3Jk9Qm2pX7vR5sT1aB8cD4",
    "status": "pending",
    "fulfillment": "delivery",
    "subtotalCents": 5400,
    "deliveryFeeCents": 800,
    "totalCents": 6200,
    "createdAt": "2026-11-18T20:15:00Z",
    "whatsappUrl": "https://wa.me/5585992358708?text=...",
    "accessToken": "7h3K9pQ2vX8mR5tY1wZ0aB6cD4eF7gH9jK2lM5nP8sA"
  }
}
```

`accessToken` é opaco, tem 43 caracteres em base64url e é devolvido tanto na criação quanto no replay idempotente. O servidor grava apenas `SHA-256` dele; o token em claro nunca entra no banco, em log nem na URL. O navegador guarda esse valor e o reenvia em `Authorization` nos próximos pedidos.

Headers:

```http
Location: /api/v1/orders/3Jk9Qm2pX7vR5sT1aB8cD4
Cache-Control: no-store
```

No replay idempotente, o header `Idempotent-Replayed: true` também é devolvido.

### Idempotência

- Primeira requisição: `201`.
- Repetição com a mesma chave e o mesmo corpo: `200`, mesmo corpo, header `Idempotent-Replayed: true`.
- Mesma chave e corpo diferente: `409 CONFLICT`.
- Chave ausente: `400 VALIDATION_ERROR`.
- Chave com 16 a 128 caracteres; recomendado UUID v4.
- O corpo canônico é usado para `requestHash`; campos desconhecidos são rejeitados.
- O cliente reutiliza a chave em retries do mesmo conteúdo e gera outra quando o pedido muda; isso evita duplicata se a resposta da primeira requisição se perder.

### Validações

| Regra | Erro |
|---|---|
| Loja fechada | `409 STORE_CLOSED` |
| Produto inativo/indisponível | `409 PRODUCT_UNAVAILABLE` |
| Zona inativa/inexistente | `409 ZONE_UNAVAILABLE` |
| Opção obrigatória ausente | `422 VALIDATION_ERROR` |
| Adicional não vinculado | `422 VALIDATION_ERROR` |
| Quantidade inválida | `422 VALIDATION_ERROR` |
| Troco insuficiente | `422 VALIDATION_ERROR` |
| Telefone inválido | `422 VALIDATION_ERROR` |
| Endereço ausente em delivery | `422 VALIDATION_ERROR` |
| Chave idempotente conflitante | `409 CONFLICT` |
| Limite de taxa | `429 RATE_LIMITED` |

> O código HTTP de validação é `422` para erros semânticos; `400` para JSON malformado ou campo desconhecido. O mapeamento é centralizado e testado.

### Contrato do `whatsappUrl`

- Gerado após a persistência.
- Contém número da loja e mensagem com número do pedido, itens, total, pagamento e dados de entrega.
- O servidor não faz chamada à API do WhatsApp na V1.
- O cliente abre o link; se o navegador bloquear, mostra link copiável.
- A URL não é logada.

---

## 8. `GET /orders/{publicId}`

Retorna o acompanhamento público mínimo.

```json
{
  "data": {
    "orderNumber": 1042,
    "publicId": "3Jk9Qm2pX7vR5sT1aB8cD4",
    "status": "preparing",
    "fulfillment": "delivery",
    "items": [
      { "name": "BACON CHEDDAR", "quantity": 2 }
    ],
    "subtotalCents": 5400,
    "deliveryFeeCents": 800,
    "totalCents": 6200,
    "createdAt": "2026-11-18T20:15:00Z",
    "updatedAt": "2026-11-18T20:30:00Z",
    "estimatedTime": "30-90 min"
  }
}
```

Não retorna:

- telefone;
- endereço;
- observações;
- e-mail;
- dados de outros pedidos;
- `_id`;
- histórico interno de auditoria;
- `whatsappUrl` (a URL contém a mensagem com PII e só é devolvida na criação do pedido).

Após recarregar a confirmação, o cliente vê número, status, itens e total pelo `publicId`. A URL completa do WhatsApp não é devolvida pelo acompanhamento: ela fica apenas na resposta de criação e, para permitir o reenvio na mesma sessão, no `sessionStorage` do navegador sob a chave `brocados:handoff:{publicId}`. Ela não é gravada no servidor nem enviada em logs.

`404 NOT_FOUND` para ID inválido ou inexistente. Rate limit ativo: 60/min por IP; detalhes em `§15 Rate limits`.

### 8.1. `GET /orders/history`

Lista os pedidos associados ao token opaco guardado no navegador. Não há senha, e-mail ou consulta por telefone.

```http
GET /api/v1/orders/history?page=1&pageSize=20
Authorization: Bearer 7h3K9pQ2vX8mR5tY1wZ0aB6cD4eF7gH9jK2lM5nP8sA
```

`page` começa em `1`; `pageSize` vai de `1` a `100`, com padrão `20`. A ordenação é fixa, do mais recente para o mais antigo; o cliente não escolhe campo de ordenação.

```json
{
  "data": [
    {
      "orderNumber": 1042,
      "publicId": "3Jk9Qm2pX7vR5sT1aB8cD4",
      "status": "preparing",
      "fulfillment": "delivery",
      "items": [{ "name": "BACON CHEDDAR", "quantity": 2 }],
      "subtotalCents": 5400,
      "deliveryFeeCents": 800,
      "totalCents": 6200,
      "createdAt": "2026-11-18T20:15:00Z",
      "updatedAt": "2026-11-18T20:30:00Z",
      "estimatedTime": "30-90 min"
    }
  ],
  "meta": { "page": 1, "pageSize": 20, "total": 1, "totalPages": 1 }
}
```

Cada item usa exatamente o formato mínimo de `GET /orders/{publicId}`: sem telefone, endereço, observação, troco, `whatsappUrl`, `_id` ou `accessToken`. Resposta com `Cache-Control: no-store`. Token ausente retorna `400 VALIDATION_ERROR`; token com forma válida mas sem pedidos associados retorna `200` com `data: []`. A página serve para descobrir `publicId`s; detalhe e atualizações continuam em `GET /orders/{publicId}` e `WS /orders/{publicId}/events`.

### 8.2. `WS /orders/{publicId}/events`

Canal WebSocket público de acompanhamento. O `publicId` vai no caminho; nenhum dado pessoal é enviado na conexão. O servidor envia primeiro `connected` e, depois, um sinal por mudança persistida:

```json
{ "type": "connected" }
```

```json
{
  "type": "order.updated",
  "publicId": "3Jk9Qm2pX7vR5sT1aB8cD4",
  "orderNumber": 1042,
  "status": "ready",
  "updatedAt": "2026-11-18T20:45:00Z"
}
```

- Tipos: `connected`, `order.created` e `order.updated`; o sinal de criação usa o mesmo formato.
- O evento público **nunca** contém `orderId`, cliente, endereço, observação, troco, valores, taxa, total ou `whatsappUrl`. Para renderizar os dados completos, o cliente refaz `GET /orders/{publicId}`.
- Handshake com `Origin` diferente do host é rejeitado. O limite de 60/min por IP do acompanhamento também vale para o upgrade.
- O cliente deve reconectar com backoff exponencial após queda; o servidor mantém a assinatura apenas enquanto a conexão estiver aberta e não guarda eventos para clientes ausentes.

---

## 9. Admin — autenticação

Todas as rotas administrativas aplicam CSRF por `Origin`/`Referer` em métodos mutáveis. Requisições sem esses cabeçalhos dependem do `SameSite=Lax` do cookie. A autorização real é sempre verificada no servidor.

O primeiro usuário `owner` é criado no bootstrap da API com `ADMIN_BOOTSTRAP_NAME`, `ADMIN_BOOTSTRAP_EMAIL` e `ADMIN_BOOTSTRAP_PASSWORD`. O bootstrap é idempotente e só atua quando `admin_users` está vazio; depois disso, a gestão de usuários passa pelo painel (US-066). Nenhuma rota pública cria usuário admin.

### 9.1 `POST /admin/auth/login`

```json
{
  "email": "operador@example.com",
  "password": "senha-forte-com-12-caracteres"
}
```

Resposta `200`:

```json
{
  "data": {
    "user": {
      "id": "665f1c2a9b1e4a0012a3b4f1",
      "name": "Operador",
      "email": "operador@example.com",
      "role": "operator"
    }
  }
}
```

- Cookie `HttpOnly`, `SameSite=Lax`, `Path=/`; `Secure` e prefixo `__Host-` em produção (nome padrão: `brocados_admin`).
- Sessão ociosa de 8 h; absoluta de 7 dias. O token opaco de 32 bytes é gravado apenas como SHA-256 em `sessions`.
- Login rotaciona a sessão e atualiza `lastLoginAt`.
- Erro genérico `401 UNAUTHORIZED`; não revelar se o e-mail existe nem se a conta está inativa.
- Corpo com campos desconhecidos retorna `400 VALIDATION_ERROR`.
- Rate limit ativo: 5/min por IP + e-mail (o e-mail entra no limite como hash); resposta `429 RATE_LIMITED` com `Retry-After: 60`.

### 9.2 `POST /admin/auth/logout`

Resposta `204`. Apaga a sessão e o cookie. É idempotente: sessão inexistente também retorna `204` e limpa o cookie.

### 9.3 `GET /admin/auth/me`

Retorna `{ "data": { "user": { "id", "name", "email", "role" } } }`; `401 UNAUTHORIZED` se não houver sessão válida ou se a conta estiver inativa. Nunca devolve `passwordHash` nem dados de sessão.

---

## 10. Admin — catálogo

Todas as rotas exigem sessão válida, papel `operator` ou `owner` e origem same-origin (CSRF por `Origin`/`Referer` em `POST`/`PATCH`). Mutações geram auditoria em `audit_logs`.

| Método | Rota | Papel | Função |
|---|---|---|---|
| GET | `/admin/products` | operator/owner | Lista com paginação/filtros |
| POST | `/admin/products` | operator/owner | Cria produto (`201`) |
| PATCH | `/admin/products/{id}` | operator/owner | Atualiza produto |
| GET | `/admin/categories` | operator/owner | Lista categorias |
| POST | `/admin/categories` | operator/owner | Cria categoria (`201`) |
| PATCH | `/admin/categories/{id}` | operator/owner | Atualiza categoria |
| GET | `/admin/addons` | operator/owner | Lista adicionais |
| POST | `/admin/addons` | operator/owner | Cria adicional (`201`) |
| PATCH | `/admin/addons/{id}` | operator/owner | Atualiza adicional |
| GET | `/admin/delivery-zones` | operator/owner | Lista zonas |
| POST | `/admin/delivery-zones` | operator/owner | Cria zona (`201`) |
| PATCH | `/admin/delivery-zones/{id}` | operator/owner | Atualiza zona |

Listagens usam o envelope paginado `{ "data": [...], "meta": { "page": 1, "pageSize": 20, "total": 0, "totalPages": 0 } }` e aceitam:

```http
GET /api/v1/admin/products?page=1&pageSize=20&active=true&categoryId=...&q=burguer&sort=priceCents&order=desc
```

- `page` ≥ 1, `pageSize` entre 1 e 100 (padrão 20); valores inválidos retornam `422`.
- `q` busca em nome/slug (regex escapada no servidor); `active` filtra por visibilidade.
- Ordenação permitida: `name`, `slug`, `sortOrder`, `createdAt`, `updatedAt` para todas as entidades; `priceCents` para produtos e adicionais; `feeCents` para zonas. `order` aceita `asc`/`desc`. Qualquer outro campo retorna `422`.
- `categoryId` só filtra produtos e precisa ser um ID hexadecimal de 24 caracteres.
- `q`, `active`, `sort` e `order` são opcionais.

Corpo de criação (`POST`) usa os campos canônicos de `docs/05-modelo-de-dados.md`; `id`, `createdAt` e `updatedAt` são atribuídos pelo servidor. `active` e `available` são `true` por padrão quando omitidos; `featured` é `false`. `PATCH` aceita apenas os campos enviados; corpo vazio, campo desconhecido ou JSON inválido retornam erro (`422` e `400`, respectivamente). O resultado validado é devolvido em `{ "data": ... }`.

Regras:

- Não há `DELETE`; desative com `active: false`.
- `priceCents`/`feeCents` são inteiros em centavos; valores negativos retornam `422`.
- `slug` é único por coleção; duplicata retorna `409 CONFLICT`.
- ID desconhecido ou malformado retorna `404 NOT_FOUND`.
- Produto pode ter `optionGroups` e `addonIds` (IDs existentes); os grupos são validados quanto a mínimo/máximo e opções duplicadas.
- Toda criação/atualização registra `action` (`create`/`update`), `entity`, `entityId`, `actorId` e snapshots mínimos sem PII.

Exemplo de `PATCH /admin/products/{id}`:

```json
{
  "available": false,
  "priceCents": 2300,
  "featured": true
}
```

---

## 11. Admin — loja

| Método | Rota | Papel | Função |
|---|---|---|---|
| GET | `/admin/store` | operator/owner | Configuração completa em `{ "data": ... }` |
| PATCH | `/admin/store` | operator/owner | Atualiza status, horários e contato |

Exemplo de resposta/corpo parcial:

```json
{
  "name": "Brocados Lanches",
  "isOpen": false,
  "openingHours": [
    { "day": "sunday", "closed": false, "intervals": [{ "open": "18:00", "close": "23:30" }] }
  ],
  "deliveryTime": "30-90 min",
  "whatsapp": "+5585992358708",
  "instagram": "@brocados",
  "address": "Rua das Palmeiras, 123, Aquiraz-CE"
}
```

- `PATCH` é parcial e rejeita campos desconhecidos (`400`) e corpo vazio (`422`).
- A configuração resultante é validada por completo: `name`, `deliveryTime`, `whatsapp` e `address` são obrigatórios; `whatsapp` usa E.164 (`+` e 8 a 15 dígitos).
- `openingHours` aceita no máximo 7 dias (`sunday`…`saturday`), sem duplicata; dia `closed: true` ignora intervalos; demais dias exigem ao menos um intervalo `HH:MM`, com `close` diferente de `open` e sem sobreposição. `close` menor que `open` indica virada de meia-noite.
- Alterar `isOpen` (ou qualquer outro campo) gera auditoria com snapshots mínimos de `name`, `isOpen` e `deliveryTime`; endereço e contato não entram na auditoria.
- Horários são avaliados em `America/Fortaleza`; o status efetivo exige kill switch ligado **e** horário vigente.
- A escrita usa `_id: "main"` com upsert; `createdAt` é preservado.

### 11.1 Integrações — webhook de notificações (owner)

| Método | Rota | Papel | Função |
|---|---|---|---|
| GET | `/admin/integrations/webhook` | owner | Configuração atual em `{ "data": ... }` |
| PATCH | `/admin/integrations/webhook` | owner | Atualiza URL e estado da integração |

Resposta/corpo parcial:

```json
{
  "url": "https://integracao.exemplo.com/pedidos",
  "active": true,
  "createdAt": "2026-11-26T12:00:00Z",
  "updatedAt": "2026-11-26T12:00:00Z"
}
```

- `PATCH` é parcial e rejeita corpo vazio; `url` aceita qualquer URL absoluta, com qualquer esquema e qualquer host, inclusive IP e `localhost`. Não há allowlist de host, exigência de `http`/`https` ou bloqueio de rede interna: o operador assume o risco de SSRF e de envio de PII a terceiros (D-042). Esquemas não HTTP são aceitos na configuração e falham apenas no envio.
- Ativar sem URL retorna `422 VALIDATION_ERROR`. A URL é validada novamente no servidor, mesmo já validada no painel.
- A configuração fica no documento `store_settings` (`webhook`), sem nova coleção. A auditoria registra `entity: webhook_settings` e apenas `active` + `urlConfigured` antes/depois; a URL nunca entra em auditoria ou log.
- A leitura é restrita a `owner` para não expor o destino ao papel `operator`.
- O disparo é best-effort e assíncrono: um único POST por evento, timeout de 3 s, sem retry, sem fila persistida e sem seguir redirecionamentos. Falha de conexão, timeout, resposta 3xx/4xx/5xx ou configuração inválida nunca altera o resultado da criação/transição; 3xx é registrado como falha no log.
- O PATCH da loja (`/admin/store`) não altera `webhook`; o PATCH da integração não altera campos operacionais da loja. Não há sobrescrita entre configurações lidas em momentos diferentes.
- Corpo exato (sem campos adicionais):

```json
{
  "number": "+5511999991234",
  "message": "Mensagem formatada para envio ao cliente"
}
```

- `number` é o telefone do cliente congelado no pedido; `message` reaproveita o texto do handoff do WhatsApp e inclui o status legível após transições. O cabeçalho `X-Brocados-Event` traz `order.created` para criação e `order.updated` para qualquer transição persistida. Não existe cadastro de statuses: um status futuro entra automaticamente pelo evento genérico.
- O evento é ignorado em replay de `Idempotency-Key` (resposta `200` com `Idempotent-Replayed: true`), para não duplicar notificação.
- Os logs registram apenas categoria do erro, status HTTP e tipo do evento: sem URL completa, corpo, telefone, nome ou endereço.

---

## 12. Admin — pedidos

| Método | Rota | Papel | Função |
|---|---|---|---|
| GET | `/admin/orders` | operator/owner | Fila paginada |
| GET | `/admin/orders/{id}` | operator/owner | Detalhe completo |
| PATCH | `/admin/orders/{id}/status` | operator/owner | Transição de status |
| WS | `/admin/orders/events` | operator/owner | Sinais de criação e mudança de status da fila |

Listagem:

```http
GET /api/v1/admin/orders?status=pending&from=2026-11-18T00:00:00Z&to=2026-11-19T00:00:00Z&page=1&pageSize=20&sort=createdAt:desc
```

- Parâmetros: `page` (padrão `1`), `pageSize` (padrão `20`, máximo `100`), `sort` (padrão `createdAt`) e `order` (`asc` ou `desc`, padrão `desc`).
- Filtros: `status`, `from`, `to` (RFC 3339 UTC), `fulfillment` (`delivery`/`pickup`) e `q` (número ou dígitos do telefone; aceita o telefone mascarado da fila).
- Ordenação allowlist: `createdAt`, `number`, `status`, `totalCents`. Qualquer outro campo retorna `422 VALIDATION_ERROR`.
- A fila devolve `customerName` com iniciais e `phone` com os últimos quatro dígitos (`****8708`); o detalhe completo exige sessão e inclui snapshot, cliente, endereço, zona, pagamento, itens e `statusHistory`.
- Quando existe perfil de conveniência para o telefone do pedido, o detalhe inclui `customerProfile` com `name`, `phone`, `orderCount`, `lastOrderAt`, `address`, `fulfillment`, `zoneName`, `paymentMethod`, `optedOut`, `updatedAt` e `retainedUntil`. A chave normalizada `phoneKey` nunca é devolvida. Falha na leitura do perfil não falha o detalhe do pedido.
- A única consulta pública por telefone é `POST /customers/profile` (§5.1), restrita à conveniência do checkout; histórico e contagem de pedidos de terceiros continuam apenas no detalhe administrativo autenticado. O histórico do próprio cliente é consultado por token opaco em `GET /orders/history` (§8.1), sem telefone.
- Dados pessoais só para admin autenticado; não vão para log nem para auditoria.

Transição:

```json
{
  "status": "confirmed"
}
```

Cancelamento:

```json
{
  "status": "cancelled",
  "reason": "Cliente solicitou cancelamento"
}
```

- A resposta de `PATCH` é o detalhe atualizado em `{ "data": ... }`.
- Transição inválida (inclusive estado final, mesmo status e desfecho incompatível com `fulfillment`): `409 CONFLICT`.
- `reason` é obrigatório em `cancelled`, com no máximo 280 caracteres; o texto é operacional e não deve conter dados do cliente.
- Cada transição atualiza `status`, `updatedAt`, `statusHistory` e, quando aplicável, `confirmedAt`/`cancelledAt` + `cancelReason`. A gravação usa compare-and-set no status lido: uma transição concorrente retorna `409` em vez de sobrescrever a primeira.
- Auditoria registra somente `status`, `fulfillment` e `totalCents` antes/depois, com `actorId`, `action: update`, `entity: orders` e `entityId` (ID hexadecimal do pedido).

### 12.1. `WS /admin/orders/events`

Canal WebSocket autenticado pela mesma sessão em cookie `HttpOnly` das demais rotas administrativas. Não aceita token em query string. O servidor envia `connected` na conexão e publica `order.created` / `order.updated` para toda criação ou transição persistida:

```json
{
  "type": "order.updated",
  "orderId": "665f1c2a9b1e4a0012a3b4f1",
  "publicId": "3Jk9Qm2pX7vR5sT1aB8cD4",
  "orderNumber": 1042,
  "status": "ready",
  "updatedAt": "2026-11-18T20:45:00Z"
}
```

- Mesmo com sessão válida, o evento contém apenas identificadores, número, status e data — sem PII e sem valores. A fila e o detalhe continuam sendo lidos por `GET /admin/orders` e `GET /admin/orders/{id}`.
- O upgrade entra no limite de 120/min por sessão e na proteção CSRF por `Origin`/`Referer` do grupo administrativo.
- A publicação é em memória e por instância: com mais de uma réplica da API, é necessário um barramento compartilhado antes de considerar a fila globalmente consistente (ver `04-arquitetura.md`).

---

## 13. Admin — usuários e auditoria

### 13.1 Usuários (owner)

| Método | Rota | Função |
|---|---|---|
| GET | `/admin/users` | Lista usuários |
| POST | `/admin/users` | Cria usuário |
| PATCH | `/admin/users/{id}` | Atualiza nome, papel ou `active` |

- Nunca retornar `passwordHash`.
- Não permitir desativar o último owner ativo.
- Não permitir autoexclusão do último owner.
- Toda mutação é auditada.
- Criação exige senha temporária forte; troca no primeiro acesso é P2.

### 13.2 Auditoria (owner)

```http
GET /api/v1/admin/audit-logs?actorId=...&entity=products&from=...&to=...&page=1&pageSize=20
```

- Somente leitura.
- Sem PII.
- `pageSize` máximo 100.
- Ordenação por `createdAt` desc.

---

## 14. Erros

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Não foi possível validar o pedido.",
    "details": [
      { "field": "customer.phone", "code": "INVALID", "message": "Telefone inválido." }
    ]
  }
}
```

| Código | HTTP | Uso |
|---|---|---|
| `VALIDATION_ERROR` | 400/422 | JSON inválido ou regra semântica |
| `UNAUTHORIZED` | 401 | Sem sessão/credencial inválida |
| `FORBIDDEN` | 403 | Papel sem permissão |
| `NOT_FOUND` | 404 | Recurso inexistente |
| `CONFLICT` | 409 | Idempotência, transição ou estado conflitante |
| `STORE_CLOSED` | 409 | Loja fechada |
| `ZONE_UNAVAILABLE` | 409 | Zona inativa/inexistente |
| `PRODUCT_UNAVAILABLE` | 409 | Produto inativo/esgotado |
| `RATE_LIMITED` | 429 | Limite excedido |
| `INTERNAL_ERROR` | 500 | Erro inesperado |

Regras:

- Não devolver erro bruto do Mongo, stack trace ou detalhe interno.
- Logar 5xx com request ID e contexto técnico.
- Resposta pública genérica em 5xx.
- `429` sem revelar existência de conta.
- `404` de acompanhamento não revela se o ID já existiu.

---

## 15. Rate limits

| Rota | Limite | Chave |
|---|---|---|
| `POST /admin/auth/login` | 5/min | IP + e-mail |
| `POST /orders` | 10/min | IP |
| `POST /orders` | 3/min | telefone |
| `POST /customers/profile` | 10/min | IP |
| `POST /customers/profile` | 3/min | telefone |
| `POST /orders/quote` | 30/min | IP |
| `GET /orders/history` | 60/min | IP |
| `GET /orders/history` | 30/min | hash do token |
| `GET /orders/{publicId}` | 60/min | IP |
| `WS /orders/{publicId}/events` | 60/min | IP (handshake) |
| `GET /menu` | 120/min | IP |
| Admin autenticado | 120/min | sessão |

Respostas `429` incluem `Retry-After: 60` e o código `RATE_LIMITED`. Os limites são por instância, em janela fixa de 1 minuto; chaves de telefone, token e sessão entram como hash SHA-256, nunca em texto claro. Em deployment com múltiplas réplicas, o contador precisa passar para um store compartilhado.

A resposta de sucesso inclui o cabeçalho `X-Request-Id`, que correlaciona a requisição aos logs estruturados sem dados pessoais.

---

## 16. Segurança HTTP

- Same-origin por padrão.
- `Origin`/`Referer` validados em métodos mutáveis.
- Cookie de sessão `HttpOnly`; nunca acessível por JS.
- CORS allowlist explícita apenas em desenvolvimento.
- Headers no proxy: CSP, `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`, HSTS em produção.
- Nenhum dado de cartão é recebido ou armazenado.
- Nenhum PII em log, analytics ou URL de tracking.

---

## 17. Versionamento e evolução

- `/api/v1` é a versão da V1.
- Adicionar campo opcional é compatível.
- Remover/renomear campo, mudar tipo ou semântica exige `/api/v2`.
- `docs/06-contrato-api.md` e OpenAPI (quando existir) são atualizados no mesmo PR.
- Testes de contrato cobrem o envelope e os campos críticos.

---

## 18. OpenAPI

A V1 pode começar com este contrato em Markdown e evoluir para OpenAPI 3.1. Quando o OpenAPI existir:

- ele é a fonte executável do contrato;
- exemplos são validados no CI;
- o frontend gera tipos ou mantém tipos espelhados testados;
- mudanças incompatíveis são detectadas.

Até lá, este documento é a referência e deve ser atualizado junto com o código.
