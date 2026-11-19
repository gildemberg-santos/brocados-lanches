# ADR 0003 — Dinheiro em centavos e fuso da operação

- **Status:** aceita
- **Data:** 18/11/2026
- **Decisores:** engenharia
- **Relacionados:** `../../AGENTS.md`, `../05-modelo-de-dados.md`

## Contexto

Preço, adicional, taxa e troco precisam ser exatos. `float` introduz erro de arredondamento e permite divergência entre cliente, servidor e banco. Datas precisam ser consistentes entre servidor, banco, API e interface, mas a operação enxerga o horário de Aquiraz–CE.

## Decisão

- Todo valor monetário é inteiro em **centavos**, tipo `int64`, com sufixo `Cents` (`priceCents`, `feeCents`, `totalCents`).
- Proibido `float`, `float64`, `parseFloat` ou `Number` decimal para dinheiro.
- Persistência e API usam **UTC** em RFC 3339.
- A loja avalia horários em **`America/Fortaleza`**, carregado com `time.LoadLocation`, sem offset fixo.
- A conversão para o fuso local acontece apenas na apresentação.
- Formatação em BRL acontece apenas na UI, com `Intl.NumberFormat`.
- Percentuais futuros, se existirem, usam pontos-base inteiros e regra de arredondamento documentada.

## Alternativas consideradas

| Alternativa | Problema |
|---|---|
| `float64` | Erro de arredondamento e comparação |
| Decimal string | Mais complexo para a V1; centavos bastam |
| Armazenar em reais | Confusão e risco de fração |
| Offset fixo `-03:00` | Quebra com mudança de fuso/regra local |
| Horário local no banco | Ambiguidade e erro em DST/UTC |

## Consequências

Positivas:

- soma e comparação exatas;
- contrato simples;
- testes determinísticos;
- exibição correta para o cliente.

Negativas:

- toda entrada/saída precisa converter na borda;
- formulário em reais exige conversão cuidadosa;
- horários que cruzam meia-noite exigem lógica explícita;
- o desenvolvedor precisa lembrar que `time.Time` é UTC no banco/API.

Mitigações:

- helpers de conversão testados;
- validação de entrada;
- testes de tabela com valores limite;
- relógio injetável para status da loja;
- documentação no modelo de dados.

## Referências

- `../../AGENTS.md`
- `../05-modelo-de-dados.md`
- `../09-qualidade-e-testes.md` (cenários 5.1, 5.17)
