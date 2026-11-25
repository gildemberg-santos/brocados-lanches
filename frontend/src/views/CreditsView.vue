<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell info-page">
    <h1>Créditos</h1>

    <aside class="info-page__notice" role="note" aria-label="Situação do texto">
      <strong>Página institucional em avaliação.</strong>
      A autoria, as imagens e os direitos de uso dependem de confirmação do cliente. Nenhum crédito
      é afirmado antes dessa confirmação.
    </aside>

    <section aria-labelledby="credits-project-title">
      <h2 id="credits-project-title">Projeto</h2>
      <p>
        Este site apresenta o cardápio, o pedido e o acompanhamento da Brocados Lanches. As
        informações de contato publicadas ficam na
        <RouterLink to="/contato">página de contato</RouterLink>.
      </p>
    </section>

    <section aria-labelledby="credits-pending-title">
      <h2 id="credits-pending-title">Pendências</h2>
      <p>
        Logotipo, fotografias, textos e demais créditos definitivos ainda aguardam validação. Esta
        página é simples e está marcada como <code>noindex</code> nos metadados da rota.
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

.info-page code {
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
}
</style>
