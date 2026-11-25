<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <footer class="site-footer">
    <nav class="site-footer__inner" aria-label="Links institucionais">
      <RouterLink class="site-footer__link" to="/contato">Contato</RouterLink>
      <RouterLink class="site-footer__link" to="/privacidade">Privacidade</RouterLink>
      <RouterLink class="site-footer__link" to="/termos">Termos</RouterLink>
      <!-- Créditos publicados em 25/11/2026 (D-032/D-033). -->
      <RouterLink class="site-footer__link" to="/creditos">Créditos</RouterLink>
    </nav>
  </footer>
</template>

<style scoped>
.site-footer {
  border-top: 1px solid var(--bg-border);
  background: var(--bg-surface);
}

.site-footer__inner {
  display: flex;
  width: min(1120px, 100%);
  flex-wrap: wrap;
  gap: 4px 12px;
  margin: 0 auto;
  padding: 12px 16px;
}

.site-footer__link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  padding: 8px 10px;
  border-radius: 999px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  text-decoration: none;
}

.site-footer__link:hover,
.site-footer__link.router-link-active {
  color: var(--brand-primary);
}

@media (min-width: 768px) {
  .site-footer__inner {
    padding-inline: 24px;
  }
}
</style>
