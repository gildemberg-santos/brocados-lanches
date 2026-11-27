# ADR 0007 — Monitoramento de erros com Sentry

- **Status:** aceita e implementada
- **Data:** 27/11/2026
- **Decisores:** produto e engenharia
- **Relacionados:** `../04-arquitetura.md`, `../06-contrato-api.md`, `../09-qualidade-e-testes.md`, `../10-operacao-e-lancamento.md`, `../13-decisoes-e-pendencias.md` (D-047)

## Contexto

A API e o navegador precisam registrar falhas inesperadas para diagnóstico fora do horário da loja. Os logs JSON estruturados já respondem por correlação por `requestId`, mas não agrupam ocorrências nem mostram a distribuição por rota e release.

Ao mesmo tempo, o fluxo de pedido trata nome, telefone, endereço, observação, troco, token de histórico e sessão administrativa. Qualquer telemetria precisa respeitar a regra de dados pessoais mínimos de `../../AGENTS.md` e a política publicada em `/privacidade`.

## Decisão

Adotar Sentry **somente para erros**, com DSN por ambiente e sem credenciais no repositório:

- backend Go `github.com/getsentry/sentry-go`, inicializado em `cmd/api` apenas quando `SENTRY_DSN` não está vazio;
- frontend `@sentry/vue`, inicializado em `main.ts` apenas quando `VITE_SENTRY_DSN` não está vazio;
- projetos separados para API e web, com `SENTRY_ENVIRONMENT` e `SENTRY_RELEASE` vindos do ambiente/build;
- captura automática de `5xx` e panics na API, e de `5xx`, timeout e falha de carregamento de rota no navegador;
- `4xx` de regra de negócio, health checks e reconexões WebSocket viram no máximo breadcrumb no cliente, nunca evento;
- `BeforeSend` sanitiza ambos os lados: remove `user` e `request`, descarta chaves sensíveis e máscara telefone, e-mail, CEP, link `wa.me`, query string e tokens;
- CSP do Nginx libera `connect-src` para `https://*.sentry.io`, sem tunnel próprio;
- `sentry.Flush` no shutdown da API; falha de inicialização do SDK nunca impede a subida do serviço e o envio é best-effort;
- um validador estático (`scripts/check_sentry.py`) bloqueia DSN literal, tracing, Replay, captura de PII e CSP incompatível.

Não entram nesta decisão: source maps, upload com `SENTRY_AUTH_TOKEN`, tracing, Session Replay, métricas, logs do Sentry e identificação de usuário.

## Alternativas consideradas

| Alternativa | Motivo da rejeição |
|---|---|
| Apenas logs JSON | Sem agrupamento, release ou visibilidade de falhas do navegador |
| DSN fixo no código | Credencial de escrita versionada e impossibilidade de separar ambientes |
| `sentryhttp` no backend | Anexa headers/query e pode criar transações; o middleware próprio mantém o evento mínimo |
| Tracing/APM | Fora do escopo pedido; aumenta volume e cardinalidade |
| Source maps no build | Exige token e upload; não autorizado nesta rodada |
| Tunnel no mesmo domínio | Complexidade extra sem necessidade; o CSP aberto ao domínio do Sentry atende |

## Consequências

Positivas:

- falha inesperada agrupada por rota, release e ambiente;
- nenhum dado do pedido sai pela telemetria;
- sem DSN, o comportamento é idêntico ao projeto sem Sentry;
- build e deploy não dependem de token do Sentry.

Negativas e mitigações:

- o DSN do navegador fica no bundle por ser `VITE_*`: é credencial de ingestão, não de leitura, e por isso não deve receber permissão administrativa;
- o CSP precisa ser ajustado se o projeto contratado usar domínio próprio em vez de `sentry.io`;
- stacks sem source maps mostram nomes ofuscados: aceito nesta entrega;
- o hub do Sentry no backend usa contexto por requisição, sem estado global mutável.

## Verificação

- `make sentry-check` valida a integração estaticamente.
- `make test-sentry` testa o validador.
- Testes Go e Vitest cobrem sanitização, captura só de erros inesperados e ausência de PII.

## Referências

- `../../AGENTS.md`
- `../06-contrato-api.md` seção 14
- `../10-operacao-e-lancamento.md` seção 8
- `../../frontend/src/features/monitoring/sentry.ts`
- `../../backend/internal/httpapi/sentry.go`
