<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import MenuCategoryNav from '../components/MenuCategoryNav.vue'
import ProductCard from '../components/ProductCard.vue'
import SearchInput from '../components/SearchInput.vue'
import { categorySectionId } from '../features/catalog/categorySections'
import { filterProducts } from '../features/catalog/filter'
import { useCategoryScrollSpy } from '../features/catalog/useCategoryScrollSpy'
import { useMenuStore } from '../stores/menu'

const menu = useMenuStore()
const route = useRoute()
const router = useRouter()

const search = ref('')
const isSearching = computed(() => search.value.trim().length > 0)
const spyEnabled = computed(() => menu.state === 'success' && !isSearching.value)

const { activeSlug, scrollToCategory, remeasure } = useCategoryScrollSpy(
  computed(() => menu.categories),
  spyEnabled,
)

const sections = computed(() =>
  menu.categories
    .map((category) => ({
      category,
      products: menu.products.filter((product) => product.categorySlug === category.slug),
    }))
    .filter((section) => section.products.length > 0),
)

const searchResults = computed(() => filterProducts(menu.products, '', search.value))
const filteredProducts = computed(() => (isSearching.value ? searchResults.value : menu.products))

const isLoading = computed(() => menu.state === 'loading' || menu.state === 'idle')

const resultLabel = computed(() => {
  const count = filteredProducts.value.length
  return `${count} ${count === 1 ? 'produto encontrado' : 'produtos encontrados'}`
})

function categoryFromQuery(): string {
  const value = route.query.categoria
  if (typeof value !== 'string') {
    return ''
  }
  return menu.categories.some((category) => category.slug === value) ? value : ''
}

function updateCategoryQuery(slug: string): void {
  const query = { ...route.query }
  if (slug.length > 0) {
    query.categoria = slug
  } else {
    delete query.categoria
  }
  void router.replace({ query })
}

function selectCategory(slug: string): void {
  scrollToCategory(slug)
  updateCategoryQuery(slug)
}

async function clearFilters(): Promise<void> {
  search.value = ''
  updateCategoryQuery('')
  await nextTick()
  remeasure()
  scrollToCategory('')
}

function retry(): void {
  void menu.loadMenu(true)
}

watch(
  [() => route.query.categoria, () => menu.state, () => menu.categories.length] as const,
  async () => {
    if (menu.state !== 'success' || isSearching.value) {
      return
    }

    const slug = categoryFromQuery()
    // Sem categoria na URL, o spy já destaca a primeira no carregamento:
    // não é preciso rolar nem alterar a posição do usuário.
    if (slug.length > 0 && slug !== activeSlug.value) {
      await nextTick()
      remeasure()
      scrollToCategory(slug)
    }
  },
  { immediate: true },
)

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
        Navegue pelas categorias ou busque pelo nome ou descrição. Itens esgotados continuam
        visíveis para consulta.
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
      <MenuCategoryNav
        v-if="!isSearching"
        :categories="menu.categories"
        :active-slug="activeSlug"
        @select="selectCategory"
      />

      <section class="menu-tools" aria-label="Busca do cardápio">
        <SearchInput v-model="search" />
      </section>

      <p class="result-count" aria-live="polite">{{ resultLabel }}</p>

      <EmptyState
        v-if="filteredProducts.length === 0"
        title="Nenhum produto encontrado"
        description="Tente outro termo ou limpe a busca para ver o cardápio completo."
      >
        <template #actions>
          <BaseButton variant="secondary" @click="clearFilters">Limpar busca</BaseButton>
        </template>
      </EmptyState>

      <div v-else-if="isSearching" class="product-grid">
        <ProductCard v-for="product in filteredProducts" :key="product.id" :product="product" />
      </div>

      <div v-else class="category-sections">
        <section
          v-for="section in sections"
          :id="categorySectionId(section.category.slug)"
          :key="section.category.id"
          class="category-section"
          :data-category-section="section.category.slug"
          :aria-labelledby="`titulo-${section.category.slug}`"
        >
          <h2 :id="`titulo-${section.category.slug}`" class="category-section__title">
            {{ section.category.name }}
          </h2>
          <div class="product-grid">
            <ProductCard v-for="product in section.products" :key="product.id" :product="product" />
          </div>
        </section>
      </div>
    </template>
  </main>
</template>

<style scoped>
.menu-page {
  display: grid;
  /* A faixa de categorias é rolável: sem isso a coluna cresce até o
     min-content dos chips e cria rolagem horizontal na página. */
  grid-template-columns: minmax(0, 1fr);
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

.result-count {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.category-sections {
  display: grid;
  gap: 32px;
}

.category-section {
  scroll-margin-top: 132px;
}

.category-section__title {
  margin: 0 0 12px;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  font-weight: 400;
  letter-spacing: 0.02em;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

@media (max-width: 767px) {
  .category-section {
    scroll-margin-top: 124px;
  }
}

@media (min-width: 900px) {
  .product-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
  }
}
</style>
