<script setup lang="ts">
import { RouterLink } from 'vue-router'

import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell not-found-page">
    <EmptyState
      title="Página não encontrada"
      description="O endereço acessado não existe ou foi movido. Escolha uma das opções abaixo."
    >
      <template #actions>
        <RouterLink class="not-found-page__link" to="/">Voltar para o início</RouterLink>
        <RouterLink class="not-found-page__link not-found-page__link--secondary" to="/cardapio">
          Ver cardápio
        </RouterLink>
      </template>
    </EmptyState>
  </main>
</template>

<style scoped>
.not-found-page {
  display: grid;
  place-items: center;
  min-height: 60vh;
}

.not-found-page__link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  padding: 12px 22px;
  border: 1px solid transparent;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-weight: 700;
  text-decoration: none;
}

.not-found-page__link:hover {
  background: var(--brand-hover);
}

.not-found-page__link--secondary {
  border-color: var(--bg-border);
  background: var(--bg-input);
  color: var(--text-main);
}

.not-found-page__link--secondary:hover {
  border-color: var(--brand-primary);
  background: var(--bg-input);
}
</style>
