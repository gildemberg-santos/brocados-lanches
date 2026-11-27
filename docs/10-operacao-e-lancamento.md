# Operação e Lançamento — Brocados Lanches

**Versão:** 1.1  
**Atualizado em:** 26/11/2026  
**Status:** plano operacional da V1; provedores ainda em aberto.  
**Relacionados:** `04-arquitetura.md`, `09-qualidade-e-testes.md`, `11-roadmap.md`, `13-decisoes-e-pendencias.md`.

---

## 1. Objetivo

Garantir que a V1 saia do ambiente local e opere com:

- deploy repetível e reversível;
- banco protegido, com backup e restauração testada;
- logs e métricas sem PII;
- runbooks para falhas comuns;
- conformidade mínima com LGPD;
- critérios objetivos de go-live.

Este documento não define o provedor de nuvem. Ele define **requisitos** que qualquer provedor precisa atender.

---

## 2. Ambientes

| Ambiente | Banco | Domínio | Uso | Dados |
|---|---|---|---|---|
| Local | `brocados_dev` | `localhost` | Desenvolvimento | Fictícios |
| Teste | `brocados_test` | efêmero | Testes automatizados | Fictícios, descartáveis |
| Preview | banco isolado | subdomínio por PR | Validação | Fictícios |
| Produção | `brocados` | domínio oficial | Operação | Reais |

Regras:

- Nunca compartilhar banco entre ambientes.
- Nunca rodar seed/teste contra produção.
- Preview não acessa banco de produção.
- Produção não expõe MongoDB publicamente.
- Banco de produção tem usuário dedicado e permissão mínima.
- `APP_ENV=production` ativa cookies `Secure`, HSTS e CORS restrito.

---

## 3. Configuração e segredos

### 3.1 Variáveis previstas

Nomes e padrões abaixo são os lidos por `backend/internal/config/config.go` e por
`deploy/docker-entrypoint.sh`. Variáveis que não aparecem aqui não são lidas pela API.

| Variável | Obrigatória | Uso |
|---|---|---|
| `APP_ENV` | não (padrão `development`) | `development`, `test` ou `production`; em produção ativa cookie `Secure` com prefixo `__Host-`. Definir `production` no deploy |
| `PORT` | não (padrão `8080`) | Porta HTTP exposta; no Render é injetada pela plataforma |
| `API_PORT` | não (padrão `3000` no entrypoint) | Porta interna da API Go atrás do nginx no container único |
| `MONGO_URI` / `MONGODB_URI` | sim em produção | Conexão MongoDB; `MONGO_URI` tem prioridade. Padrão local: `mongodb://localhost:27017` |
| `MONGO_DB` | não (padrão `brocados_dev`) | Nome do banco. **Em produção definir `brocados`**; nunca usar `brocados_dev` fora de desenvolvimento |
| `ADMIN_SESSION_COOKIE` | não | Nome do cookie; padrão `brocados_admin`, com prefixo `__Host-` em produção |
| `ADMIN_COOKIE_SECURE` | não | Força `Secure` fora de produção; em produção é sempre true |
| `ADMIN_BOOTSTRAP_NAME` | não | Nome do primeiro `owner`; deve ser usado junto com e-mail e senha |
| `ADMIN_BOOTSTRAP_EMAIL` | não | E-mail do primeiro `owner`; só cria usuário se `admin_users` estiver vazio |
| `ADMIN_BOOTSTRAP_PASSWORD` | não | Senha do primeiro `owner` (12 a 128 caracteres); remover as três variáveis de bootstrap após criar o `owner` |
| `ORDER_ACCESS_SECRET` | sim em produção | Segredo HMAC-SHA256 para derivar tokens de histórico; mínimo de 32 caracteres. Em desenvolvimento há valor padrão fictício. Rodá-lo altera tokens gerados depois da troca; tokens já emitidos continuam consultando o histórico pelo hash gravado |

| `OTEL_*`/métricas | não | Observabilidade, se adotada |

`MONGODB_USERNAME` e `MONGODB_PASSWORD` não são lidas pela API: as credenciais
devem ir na própria URI de conexão. WhatsApp e demais dados de contato são
configurados em `store_settings` pelo painel, não por variável de ambiente.

O `.env.example` é a referência local. Nenhum segredo real entra no repositório.

### 3.2 Regras

- Segredos ficam no cofre do provedor/CI.
- Rotação documentada e testada.
- Nunca logar valor de segredo.
- Nunca commitar `.env`.
- Build não depende de segredo em tempo de compilação, exceto variáveis públicas do frontend (`VITE_*`).
- `VITE_*` nunca contém segredo.

---

## 4. Topologia de produção

```text
Internet
  → TLS (443)
  → Nginx / proxy reverso
      ├── /            → arquivos estáticos da SPA
      ├── /api/*       → Go API
      ├── /healthz     → Go API
      └── /readyz      → Go API
  → Go API (rede privada)
  → MongoDB (rede privada)
```

Requisitos:

- same-origin para web e API;
- TLS válido e renovação automática;
- HTTP redireciona para HTTPS;
- MongoDB sem IP público;
- API sem porta pública direta, se possível;
- headers de segurança no proxy;
- repasse do upgrade WebSocket em `/api` (`Upgrade`, `Connection` e `proxy_read_timeout` longo), validado por `make config-check`;
- compressão `gzip`/`brotli`;
- cache longo para assets com hash e curto para HTML;
- fallback da SPA para `index.html`.

---

## 5. Pipeline de entrega

### 5.1 Estágios

```text
commit → fmt/lint/test/docs/design/build → imagem → preview → aprovação → produção
```

### 5.2 Gates

- `make fmt`
- `make lint`
- `make test`
- `make docs-check`
- `make design-check`
- `make build`
- scanner de dependências
- smoke test no preview

### 5.3 Regras

- `main` protegida; PR obrigatório.
- Um PR = uma mudança lógica.
- Deploy de produção só a partir de commit aprovado.
- Rollback para imagem/commit anterior.
- Migração de banco é compatível para trás quando possível.
- Índices são criados no start, de forma idempotente.
- Nenhum deploy com teste falhando.

### 5.4 Estratégia de release

- Preferir deploy gradual ou janela de baixo movimento.
- Não fazer deploy durante pico sem necessidade.
- Após deploy: smoke test de `/healthz`, `/readyz`, `/`, `/api/v1/store`, `/api/v1/menu`.
- Se erro crítico: rollback imediato e comunicação.

### 5.5 Render (container único)

- Blueprint: `render.yaml` com serviço web `brocados-lanches`, `dockerfilePath: ./Dockerfile`, `healthCheckPath: /healthz`.
- Runtime: nginx serve SPA e faz proxy `/api` para Go local (`API_PORT=3000`); Render injeta `PORT`.
- Segredos: `MONGODB_URI`, `ADMIN_BOOTSTRAP_NAME`, `ADMIN_BOOTSTRAP_EMAIL`, `ADMIN_BOOTSTRAP_PASSWORD`.
- Banco: MongoDB Atlas acessível pelo Render; definir `MONGO_DB=brocados`.
- Após bootstrap do primeiro `owner`, remover `ADMIN_BOOTSTRAP_*`.
- Plano free tem cold start; considerar plano pago se houver necessidade operacional.

---

## 6. Banco de dados

### 6.1 Operação

- MongoDB 8+.
- Um `mongo.Client` por processo.
- Timeout de 5 s por operação; checkout até 10 s.
- Índices criados na inicialização com log.
- Conexões limitadas e monitoradas.
- Backup separado do volume da aplicação.

### 6.2 Migrações

- Scripts versionados em `backend/scripts/`.
- Idempotentes e testados.
- Ordem: expandir → migrar → contrair.
- Nunca alterar documento em produção por comando manual sem registro.
- Toda migração tem rollback ou plano de correção.
- A migração `20261125-dados-referencia` (`backend/scripts/migrate/20261125-dados-referencia/`) atualiza loja, categorias, produtos e zonas a partir do plano do seed. Não apaga documentos: desativa os que ficaram fora do plano e preserva pedidos, usuários, sessões e auditoria. Em produção o kill switch `isOpen` atual é preservado.
- Antes de executar em produção: validar catálogo, preços, horários, contato, zonas e taxas com a loja; tirar backup das quatro coleções; conferir `make migrate-reference-check`.
- Execução: `APP_ENV=production MONGO_DB=brocados MIGRATION_CONFIRM=APPLY_VALIDATED_REFERENCE_DATA_20261125 make migrate-reference`. A URI vem de `MONGO_URI`/`MONGODB_URI` no ambiente e nunca entra no repositório.
- Rollback: restaurar o backup das quatro coleções ou aplicar migração corretiva versionada. Documentos desativados só voltam a `active: true` com registro da correção.

### 6.3 Retenção

| Dado | Proposta |
|---|---|
| Pedidos | 24 meses; depois anonimizar PII |
| Perfil de conveniência | 24 meses após o último pedido; TTL e tombstone de opt-out |
| Sessões | 7 dias (TTL) |
| Auditoria | 12 meses |
| Logs | 30 dias |
| Backups | 30 dias, com teste de restore |

A retenção final depende de validação jurídica.

---

## 7. Backup e restauração

### 7.1 Requisitos

- Backup automático diário no mínimo.
- Backup criptografado em repouso.
- Retenção de 30 dias.
- Cópia em região/armazenamento separado, se disponível.
- Monitoramento de falha de backup.
- Restauração testada antes do go-live e a cada trimestre.

### 7.2 Metas propostas

| Métrica | Meta |
|---|---|
| RPO | ≤ 24 h (ideal ≤ 1 h com oplog/point-in-time) |
| RTO | ≤ 4 h |
| Teste de restore | trimestral + antes do go-live |

### 7.3 Procedimento de restore

1. Declarar incidente e congelar escritas, se necessário.
2. Selecionar backup e validar integridade.
3. Restaurar em ambiente isolado.
4. Conferir contagens, índices e pedidos recentes.
5. Promover ou copiar dados conforme runbook.
6. Registrar evidência, tempo e lições.
7. Comunicar operação e clientes afetados, se houver.

Nunca testar restore em produção sem autorização.

---

## 8. Observabilidade

### 8.1 Logs

- JSON via `log/slog`.
- Campos: timestamp, level, requestId, method, path, status, duration, errorCode.
- Nunca: nome, telefone, endereço, observação, cookie, authorization, corpo de checkout/login, URL do WhatsApp.
- Retenção: 30 dias.
- Acesso restrito.

### 8.2 Métricas

- requisições por rota/status;
- latência p50/p95/p99;
- erros 4xx/5xx;
- pedidos criados por status;
- falhas de handoff;
- tempo de resposta do Mongo;
- conexões e saúde;
- sucesso de backup.

### 8.3 Alertas

| Alerta | Condição | Ação |
|---|---|---|
| API indisponível | `/healthz` falha 2 min | Verificar processo/rede |
| Mongo indisponível | `/readyz` falha 2 min | Verificar banco/conexão |
| Erro 5xx alto | > 2% em 5 min | Investigar logs |
| Latência alta | p95 > 1 s em 5 min | Verificar carga/Mongo |
| Fila parada | pedidos `pending` > 15 min | Avisar operação |
| Backup falhou | execução diária falhou | Corrigir e reexecutar |
| Disco alto | > 80% | Expandir/limpar |
| Certificado | < 14 dias | Renovar |

Alertas precisam de destinatário e runbook. Sem alerta acionável, não criar.

### 8.4 SLOs propostos

| SLO | Meta |
|---|---|
| Disponibilidade da API | ≥ 99,5% ao mês |
| Disponibilidade da web | ≥ 99,5% ao mês |
| Latência de leitura p95 | ≤ 300 ms |
| Criação de pedido p95 | ≤ 800 ms |
| Taxa de erro 5xx | < 0,5% |

### 8.5 Monitoramento de erros (Sentry)

- Projetos separados para API e web, com `SENTRY_DSN` e `VITE_SENTRY_DSN` definidos no ambiente/build de cada deploy (D-047). O DSN do web é incorporado ao bundle; por isso usa um projeto sem permissões administrativas.
- `SENTRY_ENVIRONMENT` e `SENTRY_RELEASE` identificam o deploy. Sem DSN, nada é enviado.
- Apenas `5xx`, panics, timeout de rede e erro de carregamento de rota geram evento. `4xx`, health checks e reconexões WebSocket não geram.
- Antes do envio, `BeforeSend` remove `user`/`request`, descarta chaves sensíveis e máscara telefone, e-mail, CEP, link `wa.me`, query e tokens. Não há tracing, Session Replay, source maps ou upload com `SENTRY_AUTH_TOKEN`.
- O shutdown da API executa `sentry.Flush` com `SENTRY_FLUSH_TIMEOUT_SECONDS` (padrão 2 s). Falha do SDK ou do envio nunca altera o pedido nem impede a inicialização.
- O proxy abre `connect-src` para `https://*.sentry.io`; se o projeto contratado usar domínio próprio, o CSP dos dois arquivos Nginx precisa ser ajustado junto.
- `make sentry-check` e `make test-sentry` rodam no CI. Nos painéis, o alerta de 5xx do §8.3 continua sendo o gatilho operacional; o Sentry é a ferramenta de diagnóstico, não mais um destino de PII.

---

## 9. Segurança operacional

- Acesso ao servidor por chave, sem senha compartilhada.
- MFA no provedor/CI quando disponível.
- Princípio do menor privilégio.
- Banco com usuário dedicado.
- Segredos em cofre; nunca em issue/chat.
- Dependências verificadas no CI.
- Revisão de headers e cookies após deploy.
- Backup de configuração.
- Plano de resposta a incidente.
- Acesso de emergência documentado.

### 9.1 Resposta a incidente

| Severidade | Exemplo | Resposta |
|---|---|---|
| S1 | Site fora, pedidos perdidos, vazamento | Ação imediata, rollback, comunicação |
| S2 | Admin fora, erro de pedido frequente | Correção em horas |
| S3 | Bug visual, erro isolado | Próximo ciclo |
| S4 | Melhoria | Backlog |

Fluxo: detectar → conter → corrigir → comunicar → registrar postmortem sem culpa.

---

## 10. LGPD e privacidade

### 10.1 Dados coletados

- nome;
- telefone;
- endereço de entrega;
- observação do pedido;
- dados de autenticação do admin (nome/e-mail/senha com hash);
- preferências do checkout no navegador (entrega, zona, nome, telefone, endereço e pagamento), sem observação nem troco; podem ser apagadas em `/privacidade` (D-037);
- perfil de conveniência no servidor, identificado pelo telefone normalizado, com nome, último endereço, entrega/retirada, zona e pagamento; observação e troco ficam de fora e o tratamento pode ser desativado no checkout (D-038). Com a opção habilitada, o checkout consulta esse perfil pelo telefone informado para preencher o próximo pedido, com limites por IP e telefone e sem devolver histórico ou taxa (D-039);
- token opaco de histórico no navegador e seu hash SHA-256 no pedido; sem cadastro, senha ou consulta por telefone (D-041);
- destino de webhook configurado pelo `owner`, que recebe o telefone do cliente, o WhatsApp da loja e a mensagem do pedido em cada evento habilitado (criação e status selecionados no painel; D-042, D-046, D-048). O envio é best-effort e o operador é responsável pelo destino e pelo tratamento desses dados;
- eventos de erro no Sentry: mensagem sanitizada, rota sem query/identificador, método, status e duração, sem identificação de usuário, headers, corpo ou dados do pedido (D-047). A política completa está em `/privacidade`.

Não coletar CPF, e-mail de cliente ou dados de cartão.

### 10.2 Finalidade

- executar o pedido;
- contatar sobre o pedido;
- entregar/retirar;
- conveniência: reaproveitar escolhas no próximo pedido no navegador e no perfil por telefone, com opt-out no checkout e exclusão pelos canais da loja;
- histórico: consultar novamente, neste navegador, os pedidos feitos aqui por meio do token opaco;
- integração: notificar automações externas configuradas pelo `owner` sobre pedidos criados e atualizados (D-042);
- cumprir obrigação legal;
- segurança e auditoria.

### 10.3 Direitos do titular

- confirmação de tratamento;
- acesso;
- correção;
- anonimização/eliminação quando aplicável;
- informação sobre compartilhamento;
- revogação de consentimento, quando aplicável.

Canal do titular publicado na política de privacidade em 25/11/2026: os canais já existentes em `/contato`.

### 10.4 Retenção e anonimização

- Pedidos: 24 meses; depois remover/mascarar nome, telefone, endereço e observação, preservando estatística.
- Exclusão manual: somente `owner`, pelo painel, em qualquer status; remove o pedido fisicamente e mantém auditoria mínima sem PII (D-051). O perfil de conveniência por telefone não é apagado junto.
- Perfil de conveniência: 24 meses após o último pedido; expira por TTL. No opt-out, dados de conveniência são removidos imediatamente e fica apenas um tombstone com a chave e os prazos.
- Sessões: TTL de 7 dias.
- Auditoria: 12 meses, sem PII.
- Preferências do checkout: até o cliente apagar em `/privacidade` ou limpar os dados do navegador.
- Token de acesso ao histórico: até o cliente apagar em `/privacidade`, limpar os dados do navegador ou trocar de dispositivo; depois disso o histórico local deixa de estar disponível.
- Logs: 30 dias, sem PII.
- Eventos de erro no Sentry: até 30 dias, conforme a configuração da conta contratada, sem PII do pedido.
- Backups: 30 dias; respeitar retenção ao restaurar.

### 10.5 Incidente de dados

1. Conter vazamento.
2. Registrar escopo e dados afetados.
3. Avaliar risco.
4. Comunicar responsável legal e, se necessário, titulares/ANPD.
5. Corrigir causa.
6. Documentar postmortem.

---

## 11. Runbooks

### 11.1 Loja não abre

1. Verificar `GET /api/v1/store`.
2. Conferir `isOpen` e horários no admin.
3. Conferir fuso do servidor.
4. Conferir logs de erro.
5. Se necessário, usar kill switch e comunicar operação.

### 11.2 Pedido não aparece no admin

1. Conferir se o `POST /orders` retornou 201.
2. Buscar por `orderNumber`/`publicId` no banco.
3. Verificar filtro de status/data.
4. Verificar índice e conexão.
5. Se persistiu e não aparece, corrigir query; não recriar pedido.

### 11.3 WhatsApp não abre

1. Confirmar que o pedido foi criado.
2. Copiar a mensagem/link.
3. Abrir manualmente no WhatsApp.
4. Orientar o cliente.
5. Registrar ocorrência e verificar bloqueio de popup.

### 11.4 API indisponível

1. Verificar processo e health.
2. Verificar Mongo/readyz.
3. Verificar proxy/TLS.
4. Rollback se o deploy coincidir.
5. Comunicar operação.

### 11.5 Erro de preço/taxa

1. Identificar pedido e snapshot.
2. Comparar com catálogo no momento.
3. Não alterar pedido antigo sem autorização.
4. Corrigir catálogo/zone.
5. Registrar auditoria e comunicar.

### 11.6 Suspeita de acesso indevido

1. Revogar sessões.
2. Trocar senhas/segredos.
3. Preservar logs.
4. Investigar escopo.
5. Seguir plano de incidente/LGPD.

---

## 12. Go-live checklist

### 12.1 Produto e conteúdo

- [ ] Catálogo, preços, opções e adicionais validados.
- [ ] Zonas, taxas, prazo e horários validados.
- [ ] WhatsApp, endereço, Instagram e pagamentos validados.
- [ ] Fotos, logo, alergênicos e textos aprovados.
- [x] Política de privacidade e termos publicados em 25/11/2026.
- [x] Página de contato e canal do titular publicado em `/contato`.

### 12.2 Técnico

- [ ] `make fmt`, `make lint`, `make test`, `make docs-check`, `make design-check`, `make build` passando.
- [ ] E2E crítico passando.
- [ ] Acessibilidade sem erro grave.
- [ ] Performance dentro do orçamento.
- [ ] Índices criados em produção.
- [ ] Backup automático ativo.
- [ ] Restore testado.
- [ ] TLS e headers configurados.
- [ ] Cookies `Secure`/`HttpOnly`/`SameSite`.
- [ ] Rate limit ativo.
- [ ] Logs sem PII.
- [ ] DSNs configurados em ambientes separados e `make sentry-check` aprovado.
- [ ] CSP com `connect-src` compatível com o domínio de ingestão do Sentry.
- [ ] Alertas com destinatário.
- [ ] Rollback definido.

### 12.3 Operação

- [ ] Owner e operator criados.
- [ ] Treinamento realizado.
- [ ] Runbooks disponíveis.
- [ ] Escala de suporte definida.
- [ ] Comunicação de lançamento pronta.
- [ ] Monitoramento acompanhado no primeiro dia.

### 12.4 Pós-lançamento

- [ ] Smoke test em produção.
- [ ] Primeiro pedido real acompanhado ponta a ponta.
- [ ] Nenhum erro 5xx inexplicado.
- [ ] Feedback da operação coletado.
- [ ] Revisão em 48 h e 7 dias.

---

## 13. Operação diária

| Momento | Ação |
|---|---|
| Abertura | Conferir status, horários, itens esgotados e taxas |
| Pico | Acompanhar fila e alertas |
| Fechamento | Conferir pedidos pendentes e fechar loja |
| Semanal | Revisar erros, pedidos perdidos e backup |
| Mensal | Revisar acessos, retenção e custos |
| Trimestral | Testar restore e revisar LGPD |

---

## 14. Pendências

- Provedor de hospedagem.
- ~~Provedor de CI.~~ Resolvido em 20/11/2026: GitHub Actions (D-016).
- MongoDB gerenciado ou container.
- Ferramenta de métricas/alertas.
- Política de retenção aprovada.
- ~~Canal do titular.~~ Resolvido em 25/11/2026: canais publicados em `/contato` (D-033).
- Domínio e DNS.
- Estratégia de backup/point-in-time.
- Responsável de plantão.
