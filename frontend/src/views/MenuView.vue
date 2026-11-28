<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import CategoryPill from '../components/CategoryPill.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import ProductCard from '../components/ProductCard.vue'
import SearchInput from '../components/SearchInput.vue'
import { filterProducts } from '../features/catalog/filter'
import { useMenuStore } from '../stores/menu'

const menu = useMenuStore()
const route = useRoute()
const router = useRouter()

const search = ref('')
const selectedCategory = ref<string>(
  typeof route.query.categoria === 'string' ? route.query.categoria : '',
)

const isLoading = computed(() => menu.state === 'loading' || menu.state === 'idle')
const hasFilters = computed(
  () => search.value.trim().length > 0 || selectedCategory.value.length > 0,
)

const filteredProducts = computed(() =>
  filterProducts(menu.products, selectedCategory.value, search.value),
)

const resultLabel = computed(() => {
  const count = filteredProducts.value.length
  return `${count} ${count === 1 ? 'produto encontrado' : 'produtos encontrados'}`
})

watch(
  () => route.query.categoria,
  (value) => {
    selectedCategory.value = typeof value === 'string' ? value : ''
  },
)

function selectCategory(slug: string): void {
  selectedCategory.value = slug

  const query = { ...route.query }
  if (slug.length > 0) {
    query.categoria = slug
  } else {
    delete query.categoria
  }

  void router.replace({ query })
}

function clearFilters(): void {
  search.value = ''
  selectedCategory.value = ''
  void router.replace({ query: {} })
}

function retry(): void {
  void menu.loadMenu(true)
}

onMounted(() => {
  void menu.ensureLoaded()
})
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell menu-page">
    <header class="menu-header">
      <p class="eyebrow">Cardápio digital</p>
      <h1>Escolha seu próximo lanche</h1>
      <p class="menu-header__lead">
        Busque pelo nome ou descrição e filtre por categoria. Itens esgotados continuam visíveis
        para consulta.
      </p>
    </header>

    <LoadingSkeleton v-if="isLoading" label="Carregando cardápio" />

    <ErrorState v-else-if="menu.state === 'error'" :message="menu.error" @retry="retry" />

    <EmptyState
      v-else-if="menu.isEmpty"
      title="Cardápio em preparação"
      description="Ainda não há produtos disponíveis. Tente novamente em instantes."
    >
      <template #actions>
        <BaseButton variant="secondary" @click="retry">Tentar novamente</BaseButton>
      </template>
    </EmptyState>

    <template v-else>
      <section class="menu-tools" aria-label="Busca e filtros do cardápio">
        <SearchInput v-model="search" />

        <div class="category-row">
          <CategoryPill
            label="Todos"
            :active="selectedCategory.length === 0"
            @select="selectCategory('')"
          />
          <CategoryPill
            v-for="category in menu.categories"
            :key="category.id"
            :label="category.name"
            :active="selectedCategory === category.slug"
            @select="selectCategory(category.slug)"
          />
        </div>
      </section>

      <p class="result-count" aria-live="polite">{{ resultLabel }}</p>

      <EmptyState
        v-if="filteredProducts.length === 0"
        title="Nenhum produto encontrado"
        description="Tente outro termo ou remova os filtros para ver o cardápio completo."
      >
        <template #actions>
          <BaseButton v-if="hasFilters" variant="secondary" @click="clearFilters">
            Limpar filtros
          </BaseButton>
        </template>
      </EmptyState>

      <div v-else class="product-grid">
        <ProductCard v-for="product in filteredProducts" :key="product.id" :product="product" />
      </div>
    </template>
  </main>
</template>

<style scoped>
.menu-page {
  display: grid;
  gap: 22px;
}

.menu-header {
  display: grid;
  gap: 8px;
}

.eyebrow {
  margin: 0;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.menu-header h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  font-weight: 400;
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.menu-header__lead {
  max-width: 58ch;
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.menu-tools {
  display: grid;
  gap: 12px;
}

.category-row {
  display: flex;
  min-width: 0;
  max-width: 100%;
  gap: 8px;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  padding-bottom: 4px;
  scroll-snap-type: x proximity;
  scrollbar-width: none;
}

.category-row::-webkit-scrollbar {
  display: none;
}

.result-count {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

@media (min-width: 900px) {
  .product-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
  }
}
</style>
