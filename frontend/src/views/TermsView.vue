<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell info-page">
    <h1>Termos de uso</h1>

    <aside class="info-page__notice" role="note" aria-label="Situação do texto">
      <strong>Texto sujeito a aprovação.</strong>
      Estes termos são informativos e provisórios: o texto definitivo depende de aprovação do
      cliente e do jurídico. Não substituem um termo aprovado.
    </aside>

    <section aria-labelledby="terms-order-title">
      <h2 id="terms-order-title">Pedidos</h2>
      <p>
        O pedido só é considerado registrado depois de gravado no servidor. O acompanhamento é feito
        por um código público opaco, sem exposição de dados pessoais.
      </p>
    </section>

    <section aria-labelledby="terms-price-title">
      <h2 id="terms-price-title">Preços, taxas e disponibilidade</h2>
      <p>
        Preços, adicionais, taxa de entrega, total e disponibilidade são validados no servidor.
        Valores exibidos apenas nesta página não são contratuais; o valor cotado no checkout é
        calculado pela loja.
      </p>
    </section>

    <section aria-labelledby="terms-payment-title">
      <h2 id="terms-payment-title">Entrega e pagamento</h2>
      <p>
        A retirada zera a taxa de entrega e a entrega usa a taxa da zona escolhida. O pagamento
        acontece na entrega ou na retirada; nenhum dado de cartão é solicitado.
      </p>
      <p>
        Para dúvidas sobre o pedido, use os canais da
        <RouterLink to="/contato">página de contato</RouterLink>.
      </p>
    </section>
  </main>
</template>

<style scoped>
.info-page {
  display: grid;
  gap: 20px;
}

.info-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(var(--font-size-h1), 7vw, var(--font-size-display));
  line-height: var(--line-height-display);
}

.info-page__notice {
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.info-page__notice strong {
  color: var(--brand-primary);
}

.info-page section {
  display: grid;
  gap: 8px;
}

.info-page h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.info-page p {
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.info-page a {
  color: var(--brand-neon);
  font-weight: 700;
}
</style>
