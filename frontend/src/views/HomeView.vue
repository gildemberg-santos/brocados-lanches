<script setup lang="ts">
import { onMounted } from 'vue'

import { useHealthStore } from '../stores/health'

const health = useHealthStore()

const statusLabel = {
  loading: 'API: verificando',
  online: 'API: conectada',
  offline: 'API: indisponível',
} as const

onMounted(() => {
  void health.check()
})
</script>

<template>
  <main class="app-shell">
    <section class="card">
      <p class="eyebrow">Brocados Lanches</p>
      <h1>Estrutura inicial</h1>
      <p class="lead">
        Base do projeto com Go, Vue e MongoDB. O próximo passo é implementar o cardápio e o fluxo de
        pedidos.
      </p>

      <div class="status" :class="`status--${health.state}`" aria-live="polite">
        <span class="status__dot" aria-hidden="true"></span>
        <div>
          <strong>{{ statusLabel[health.state] }}</strong>
          <p>MongoDB: {{ health.mongo }}</p>
        </div>
      </div>

      <button class="button" type="button" @click="health.check()">Verificar novamente</button>
      <p v-if="health.error" class="error" role="alert">{{ health.error }}</p>
    </section>
  </main>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background: radial-gradient(circle at top, var(--bg-surface), var(--bg-primary) 62%);
}

.card {
  width: min(720px, 100%);
  padding: clamp(24px, 6vw, 56px);
  border: 1px solid var(--bg-border);
  border-radius: 24px;
  background: var(--bg-surface);
}

.eyebrow {
  margin: 0 0 8px;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(var(--font-size-h1), 8vw, var(--font-size-display));
  font-weight: 400;
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.lead {
  max-width: 54ch;
  margin: 16px 0 0;
  color: var(--text-sub);
  line-height: 1.65;
}

.status {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 28px 0 20px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-input);
}

.status__dot {
  width: 12px;
  height: 12px;
  flex: 0 0 auto;
  border-radius: 999px;
  background: var(--text-sub);
}

.status--online .status__dot {
  background: var(--status-open);
}

.status--offline .status__dot {
  background: var(--status-error);
}

.status--loading .status__dot {
  background: var(--brand-neon);
}

.status p {
  margin: 4px 0 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.button {
  min-height: 44px;
  padding: 12px 20px;
  border: 0;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-weight: 700;
  cursor: pointer;
}

.button:hover {
  background: var(--brand-hover);
}

.error {
  margin: 16px 0 0;
  color: var(--status-error);
}
</style>
