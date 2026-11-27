# ADR 0008 — Analytics com Google Tag Manager

- **Status:** aceita e implementada
- **Data:** 27/11/2026
- **Decisores:** solicitante e engenharia
- **Relacionados:** `../03-prd.md`, `../05-modelo-de-dados.md`, `../08-backlog.md` (US-084), `../10-operacao-e-lancamento.md`, `../13-decisoes-e-pendencias.md` (D-050), `../../AGENTS.md` §8.4

## Contexto

O solicitante pediu a instalação do container `GTM-NLTLDV4N` no frontend. Até então `D-014` e `AGENTS.md` §8.4 exigiam consentimento prévio e revisão antes de qualquer script de terceiro; P-019 e US-084 estavam pendentes, e `/privacidade` declarava que a loja não usava analytics.

O snippet entregue misturava o carregador do GTM (`gtm.js?id=GTM-...`) com a chamada `gtag('js', new Date())` do GA4. O CSP dos proxies só permitia `script-src 'self'`, bloqueando tanto o domínio do Google quanto o `<script>` inline do container.

## Decisão

Carregar o **snippet padrão do container GTM** em `frontend/index.html`, em todas as visitas, **sem banner de opt-in**, com autorização explícita do solicitante para a exceção à regra de consentimento (waiver registrado em D-050):

- `window.dataLayer` recebe apenas o evento mínimo `gtm.start` do próprio snippet; o código da loja não envia nome, telefone, endereço, observação, troco, token ou identificador do pedido;
- o nonce do `<script>` inline é gerado por resposta no Vite (`AsyncLocalStorage` + `randomBytes`) e no Nginx (`$request_id` com `sub_filter`), sem `'unsafe-inline'` em `script-src`;
- o CSP libera `https://www.googletagmanager.com` em `script-src` e `connect-src`, e `https://www.google-analytics.com` em `connect-src` e `img-src`, mantendo `https://*.sentry.io` e `default-src 'self'`;
- `/privacidade` passa a declarar o GTM, o carregamento em todas as visitas, o compartilhamento com o Google e o prazo definido pelo provedor;
- `scripts/check_analytics.py` valida o snippet, o nonce, o CSP e a política no CI, no hook e em `make test`.

## Alternativas consideradas

| Alternativa | Motivo da rejeição |
|---|---|
| Banner de consentimento antes do GTM | O solicitante escolheu explicitamente o carregamento em todas as visitas |
| Snippet colado com `gtag('js', ...)` | Mistura GTM e GA4; sem um ID `G-...`, a chamada não configura medição alguma |
| `'unsafe-inline'` no `script-src` | Enfraquece o CSP para todos os scripts inline da aplicação |
| Nonce fixo no build | Reutilizável por quem ler o HTML; o nonce deve variar por resposta |
| Injetar dados do checkout no `dataLayer` | Contraria a minimização de dados pessoais de `AGENTS.md` |

## Consequências

Positivas:

- o container solicitado é carregado em dev e produção com o snippet oficial;
- mesmo com script de terceiro, nenhum dado do pedido é passado pelo código da loja;
- o CSP continua negando scripts inline sem nonce e hosts não listados.

Negativas e mitigações:

- o Google pode processar dados técnicos da visita, como endereço IP e identificadores de cookie, conforme a política do provedor; isso está declarado em `/privacidade` e foi aceito no waiver;
- tags configuradas dentro do container que usem outros domínios serão bloqueadas pelo CSP até serem explicitamente liberadas — isso é intencional;
- `$request_id` do Nginx é usado como nonce: o valor é imprevisível antes da resposta e muda a cada requisição;
- o Vite apenas aplica o nonce no modo dev; o build gera o placeholder `__CSP_NONCE__`, substituído pelo Nginx em produção.

## Verificação

- `make analytics-check` valida estática e deterministicamente a integração.
- `make test-analytics` cobre o validador.
- `pnpm test` cobre o nonce por resposta em `frontend/csp.test.ts`.
- Em produção, o Nginx deve servir o HTML com `sub_filter` ativo; a ausência do nonce deixa o snippet bloqueado pelo CSP, sem quebrar a aplicação.
