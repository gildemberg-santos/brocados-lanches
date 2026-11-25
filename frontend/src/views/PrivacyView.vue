<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main id="conteudo" class="page-shell info-page">
    <h1>Privacidade</h1>

    <aside class="info-page__notice" role="note" aria-label="Situação do texto">
      <strong>Texto sujeito a aprovação.</strong>
      Esta é uma página informativa provisória: a política de privacidade definitiva depende de
      aprovação do cliente e do jurídico. Não a apresente como texto vigente.
    </aside>

    <section aria-labelledby="privacy-scope-title">
      <h2 id="privacy-scope-title">Dados usados no pedido</h2>
      <p>
        Para registrar e entregar um pedido, o site solicita nome, telefone, endereço de entrega e
        observação do pedido. O servidor valida esses dados e calcula preços, taxas e total.
      </p>
    </section>

    <section aria-labelledby="privacy-admin-title">
      <h2 id="privacy-admin-title">Acesso administrativo</h2>
      <p>
        O painel usa e-mail e senha com hash para autenticação de usuários autorizados e registra
        auditoria das alterações sem incluir dados pessoais do cliente.
      </p>
    </section>

    <section aria-labelledby="privacy-channel-title">
      <h2 id="privacy-channel-title">Canal do titular e retenção</h2>
      <p>
        O canal oficial do titular e a retenção definitiva dos dados ainda dependem de confirmação
        do cliente e do jurídico. Os canais já publicados pela loja estão na
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
