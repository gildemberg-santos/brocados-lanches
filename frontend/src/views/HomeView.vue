<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import CategoryPill from '../components/CategoryPill.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import PriceText from '../components/PriceText.vue'
import ProductCardCompact from '../components/ProductCardCompact.vue'
import ProductImage from '../components/ProductImage.vue'
import { useMenuStore } from '../stores/menu'
import { useStoreStatusStore } from '../stores/storeStatus'

const menu = useMenuStore()
const statusStore = useStoreStatusStore()
const router = useRouter()

const heroProduct = computed(
  () => menu.featuredProducts[0] ?? menu.products.find((product) => product.available),
)

const isLoading = computed(() => menu.state === 'loading' || menu.state === 'idle')

const currentStatus = computed(() => (statusStore.state === 'success' ? statusStore.status : null))
const statusLabel = computed(() => {
  if (currentStatus.value?.status === 'open') {
    return 'Loja aberta'
  }
  if (currentStatus.value?.status === 'closed') {
    return 'Loja fechada'
  }
  return ''
})
const hasStatus = computed(() => statusLabel.value.length > 0)
const deliveryTime = computed(() => currentStatus.value?.deliveryTime.trim() ?? '')
const whatsappUrl = computed(() => {
  const digits = (currentStatus.value?.whatsapp ?? '').replace(/\D/g, '')
  return digits ? `https://wa.me/${digits}` : ''
})
const instagram = computed(() => currentStatus.value?.instagram.trim() ?? '')
const address = computed(() => currentStatus.value?.address.trim() ?? '')
const hasContacts = computed(
  () => whatsappUrl.value.length > 0 || instagram.value.length > 0 || address.value.length > 0,
)
const nextOpening = computed(() => {
  if (currentStatus.value?.status !== 'closed' || !currentStatus.value.nextOpeningAt) {
    return ''
  }
  const date = new Date(currentStatus.value.nextOpeningAt)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
    timeZone: 'America/Fortaleza',
  }).format(date)
})

function openCategory(slug: string): void {
  void router.push({ name: 'menu', query: { categoria: slug } })
}

function retry(): void {
  void menu.loadMenu(true)
}

function retryStatus(): void {
  void statusStore.refresh()
}

onMounted(() => {
  void menu.ensureLoaded()
  void statusStore.load()
})
</script>

<template>
  <main id="conteudo" class="page-shell home">
    <section class="home-hero" :class="{ 'home-hero--with-feature': heroProduct }">
      <div class="home-hero__copy">
        <p class="eyebrow">Brocados Lanches</p>
        <h1>Seu lanche, do seu jeito.</h1>
        <p class="home-hero__lead">
          Veja o cardápio, escolha seus favoritos e personalize cada item antes de pedir.
        </p>
        <RouterLink class="home-hero__cta" :to="{ name: 'menu' }">Abrir cardápio</RouterLink>
      </div>

      <RouterLink
        v-if="heroProduct"
        class="home-hero__spotlight"
        :to="{ name: 'product', params: { slug: heroProduct.slug } }"
      >
        <span class="home-hero__media">
          <ProductImage
            :src="heroProduct.imageUrl"
            :alt="heroProduct.imageAlt.length > 0 ? heroProduct.imageAlt : heroProduct.name"
            :width="480"
            :height="360"
            eager
          />
        </span>
        <span class="home-hero__spotlight-body">
          <span class="home-hero__spotlight-label">Destaque</span>
          <strong>{{ heroProduct.name }}</strong>
          <PriceText :value-cents="heroProduct.priceCents" />
        </span>
      </RouterLink>
    </section>

    <section
      class="home-status"
      data-testid="store-status-banner"
      aria-labelledby="home-status-title"
    >
      <template v-if="statusStore.state === 'idle' || statusStore.state === 'loading'">
        <h2 id="home-status-title" class="home-status__title">Consultando status da loja...</h2>
      </template>

      <template v-else-if="hasStatus">
        <h2
          id="home-status-title"
          class="home-status__title"
          :class="currentStatus?.status === 'open' ? 'is-open' : 'is-closed'"
        >
          {{ statusLabel }}
        </h2>
        <p v-if="deliveryTime" class="home-status__line">Entrega: {{ deliveryTime }}</p>
        <p v-if="nextOpening" class="home-status__line">Próxima abertura: {{ nextOpening }}</p>
        <ul v-if="hasContacts" class="home-status__contacts">
          <li v-if="whatsappUrl">
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer">
              WhatsApp: {{ currentStatus?.whatsapp }}
            </a>
          </li>
          <li v-if="instagram">Instagram: {{ instagram }}</li>
          <li v-if="address">Endereço: {{ address }}</li>
        </ul>
      </template>

      <template v-else>
        <h2 id="home-status-title" class="home-status__title">Status da loja indisponível.</h2>
        <BaseButton variant="secondary" @click="retryStatus">Tentar novamente</BaseButton>
      </template>
    </section>

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
      <section class="home-section" aria-labelledby="home-categories-title">
        <div class="section-heading">
          <h2 id="home-categories-title">Categorias</h2>
          <RouterLink class="section-heading__link" :to="{ name: 'menu' }">
            Ver cardápio completo
          </RouterLink>
        </div>

        <div class="category-row">
          <CategoryPill
            v-for="category in menu.categories"
            :key="category.id"
            :label="category.name"
            :active="false"
            @select="openCategory(category.slug)"
          />
        </div>
      </section>

      <section
        v-if="menu.featuredProducts.length > 0"
        class="home-section"
        aria-labelledby="home-featured-title"
      >
        <div class="section-heading">
          <h2 id="home-featured-title">Mais pedidos</h2>
          <RouterLink class="section-heading__link" :to="{ name: 'menu' }">Ver todos</RouterLink>
        </div>

        <div class="featured-grid">
          <ProductCardCompact
            v-for="product in menu.featuredProducts"
            :key="product.id"
            :product="product"
          />
        </div>
      </section>
    </template>
  </main>
</template>

<style scoped>
.home {
  display: grid;
  gap: 28px;
}

.home-hero {
  display: grid;
  gap: 18px;
  padding: clamp(24px, 6vw, 48px);
  border: 1px solid var(--bg-border);
  border-radius: 24px;
  background: var(--bg-surface);
}

.home-hero__copy {
  display: grid;
  align-content: center;
  justify-items: start;
  gap: 14px;
}

.eyebrow {
  margin: 0;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.home-hero h1 {
  max-width: 14ch;
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(var(--font-size-h1), 8vw, var(--font-size-display));
  font-weight: 400;
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.home-hero__lead {
  max-width: 48ch;
  margin: 0;
  color: var(--text-sub);
  line-height: 1.65;
}

.home-hero__cta {
  display: inline-flex;
  min-height: 48px;
  align-items: center;
  justify-content: center;
  padding: 12px 24px;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-weight: 800;
  text-decoration: none;
}

.home-hero__cta:hover {
  background: var(--brand-hover);
}

.home-hero__spotlight {
  display: grid;
  align-self: stretch;
  overflow: hidden;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-input);
  color: inherit;
  text-decoration: none;
}

.home-hero__spotlight:hover {
  border-color: var(--brand-primary);
}

.home-hero__media {
  display: block;
  aspect-ratio: 4 / 3;
  overflow: hidden;
}

.home-hero__spotlight-body {
  display: grid;
  gap: 4px;
  padding: 14px;
}

.home-hero__spotlight-label {
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.home-hero__spotlight strong {
  font-size: var(--font-size-h1);
  line-height: 1.2;
}

.home-status {
  display: grid;
  gap: 8px;
  padding: 16px 20px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.home-status__title {
  margin: 0;
  font-size: var(--font-size-h2);
  line-height: 1.3;
}

.home-status__title.is-open {
  color: var(--status-open);
}

.home-status__title.is-closed {
  color: var(--text-main);
}

.home-status__line {
  margin: 0;
  color: var(--text-sub);
  line-height: 1.5;
}

.home-status__contacts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
  margin: 0;
  padding: 0;
  color: var(--text-sub);
  list-style: none;
  line-height: 1.5;
}

.home-status__contacts a {
  color: var(--brand-neon);
  font-weight: 700;
}

.home-section {
  display: grid;
  gap: 14px;
}

.section-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px 16px;
}

.section-heading h2 {
  margin: 0;
  font-size: var(--font-size-h1);
}

.section-heading__link {
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.section-heading__link:hover {
  color: var(--brand-neon);
}

.category-row {
  display: flex;
  min-width: 0;
  max-width: 100%;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: none;
}

.category-row::-webkit-scrollbar {
  display: none;
}

.featured-grid {
  display: grid;
  gap: 10px;
}

@media (min-width: 640px) {
  .home-hero--with-feature {
    grid-template-columns: minmax(0, 1.15fr) minmax(220px, 0.85fr);
    align-items: center;
  }

  .home-hero__spotlight {
    grid-template-columns: 0.9fr 1.1fr;
  }

  .home-hero__media {
    aspect-ratio: auto;
    min-height: 180px;
  }

  .home-hero__spotlight-body {
    align-content: center;
    padding: 16px;
  }

  .featured-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 960px) {
  .home {
    gap: 36px;
  }

  .featured-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
