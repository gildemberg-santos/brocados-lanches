<script setup lang="ts">
import type { MenuCategory } from '../types/catalog'
import CategoryPill from './CategoryPill.vue'

defineProps<{
  categories: MenuCategory[]
  activeSlug: string
}>()

defineEmits<{
  select: [slug: string]
}>()
</script>

<template>
  <nav class="category-nav" aria-label="Categorias do cardápio">
    <div class="category-nav__inner">
      <CategoryPill
        v-for="category in categories"
        :key="category.id"
        :label="category.name"
        :active="activeSlug === category.slug"
        @select="$emit('select', category.slug)"
      />
    </div>
  </nav>
</template>

<style scoped>
.category-nav {
  position: sticky;
  /* 64 px do header + 1 px da borda inferior, mais o safe-area. */
  top: calc(65px + env(safe-area-inset-top));
  z-index: 30;
  min-width: 0;
  border-bottom: 1px solid var(--bg-border);
  background: var(--bg-primary);
}

.category-nav__inner {
  display: flex;
  min-width: 0;
  max-width: 100%;
  gap: 8px;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  padding: 8px 0;
  scroll-snap-type: x proximity;
  scrollbar-width: none;
}

.category-nav__inner::-webkit-scrollbar {
  display: none;
}

@media (max-width: 767px) {
  .category-nav {
    /* 56 px do header + 1 px da borda inferior, mais o safe-area. */
    top: calc(57px + env(safe-area-inset-top));
  }
}
</style>
