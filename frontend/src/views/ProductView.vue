<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import AddonRow from '../components/AddonRow.vue'
import BaseButton from '../components/BaseButton.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import OptionGroup from '../components/OptionGroup.vue'
import PriceText from '../components/PriceText.vue'
import ProductImage from '../components/ProductImage.vue'
import {
  calculateConfiguredPriceCents,
  createOptionSelections,
  isOptionSelectionValid,
  validateOptionGroups,
} from '../features/catalog/customization'
import type { OptionSelections } from '../features/catalog/customization'
import { useCartStore } from '../stores/cart'
import { useMenuStore } from '../stores/menu'
import type { MenuProduct } from '../types/catalog'

const MAX_NOTE_LENGTH = 140

const menu = useMenuStore()
const cart = useCartStore()
const route = useRoute()

const slug = computed(() => (typeof route.params.slug === 'string' ? route.params.slug : ''))
const product = computed(() => menu.findProductBySlug(slug.value))
const isLoading = computed(() => menu.state === 'loading' || menu.state === 'idle')

const optionGroups = computed(() => product.value?.optionGroups ?? [])
const addons = computed(() => product.value?.addons ?? [])
const selections = reactive<OptionSelections>({})
const selectedAddonIds = ref<string[]>([])
const note = ref('')
const submitAttempted = ref(false)
const added = ref(false)
const addFailed = ref(false)

const validationErrors = computed(() => validateOptionGroups(optionGroups.value, selections))
const hasValidationErrors = computed(() => Object.keys(validationErrors.value).length > 0)
const isConfigurationValid = computed(
  () => product.value?.available === true && isOptionSelectionValid(optionGroups.value, selections),
)

const totalCents = computed(() => {
  if (!product.value) {
    return 0
  }

  return calculateConfiguredPriceCents(
    product.value,
    optionGroups.value,
    selections,
    addons.value,
    selectedAddonIds.value,
  )
})

function resetConfiguration(item: MenuProduct | undefined): void {
  for (const groupId of Object.keys(selections)) {
    delete selections[groupId]
  }

  Object.assign(selections, createOptionSelections(item?.optionGroups ?? []))
  selectedAddonIds.value = []
  note.value = ''
  submitAttempted.value = false
  added.value = false
  addFailed.value = false
}

function setSelection(groupId: string, ids: string[]): void {
  selections[groupId] = ids
  added.value = false
}

function toggleAddon(addonId: string, selected: boolean): void {
  const next = new Set(selectedAddonIds.value)

  if (selected) {
    next.add(addonId)
  } else {
    next.delete(addonId)
  }

  selectedAddonIds.value = [...next]
  added.value = false
}

async function focusFirstInvalidGroup(): Promise<void> {
  const [firstGroupId] = Object.keys(validationErrors.value)

  if (!firstGroupId) {
    return
  }

  await nextTick()
  document.getElementById(`option-group-${firstGroupId}`)?.focus()
}

async function handleAdd(): Promise<void> {
  if (!product.value) {
    return
  }

  submitAttempted.value = true
  added.value = false
  addFailed.value = false

  if (!isConfigurationValid.value) {
    await focusFirstInvalidGroup()
    return
  }

  const addedToCart = cart.add(product.value, selections, selectedAddonIds.value, note.value, 1)

  if (!addedToCart) {
    addFailed.value = true
    return
  }

  added.value = true
}

watch(product, (item) => resetConfiguration(item), { immediate: true })

onMounted(() => {
  void menu.ensureLoaded()
})
</script>

<template>
  <main id="conteudo" class="page-shell product-page">
    <RouterLink class="back-link" :to="{ name: 'menu' }">← Voltar ao cardápio</RouterLink>

    <LoadingSkeleton v-if="isLoading" label="Carregando produto" :count="2" />

    <ErrorState
      v-else-if="menu.state === 'error'"
      :message="menu.error"
      @retry="menu.loadMenu(true)"
    />

    <EmptyState
      v-else-if="!product"
      title="Produto não encontrado"
      description="Não encontramos este produto no cardápio. Ele pode ter sido removido ou o link está incorreto."
    >
      <template #actions>
        <RouterLink class="text-link" :to="{ name: 'menu' }">Ver cardápio</RouterLink>
      </template>
    </EmptyState>

    <template v-else>
      <div class="product-layout">
        <section class="product-summary" aria-labelledby="product-name">
          <div class="product-summary__media">
            <ProductImage
              :src="product.imageUrl"
              :alt="product.imageAlt.length > 0 ? product.imageAlt : product.name"
              :width="720"
              :height="540"
              eager
            />
          </div>

          <div class="product-summary__copy">
            <p class="eyebrow">{{ product.categoryName }}</p>
            <h1 id="product-name">{{ product.name }}</h1>
            <p class="product-summary__description">{{ product.description }}</p>

            <p v-if="product.allergens.length > 0" class="product-summary__allergens">
              <strong>Alergênicos:</strong> {{ product.allergens.join(', ') }}
            </p>

            <div class="base-price">
              <span>Preço base</span>
              <PriceText :value-cents="product.priceCents" />
            </div>

            <p v-if="!product.available" class="sold-out-message" role="status">
              Este item está esgotado. Escolha outro para continuar.
            </p>
          </div>
        </section>

        <form class="product-config" @submit.prevent="handleAdd">
          <section v-if="optionGroups.length > 0" class="config-section">
            <h2>Personalize</h2>
            <div class="config-section__body">
              <OptionGroup
                v-for="group in optionGroups"
                :id="`option-group-${group.id}`"
                :key="group.id"
                :group="group"
                :selected-ids="selections[group.id] ?? []"
                :error="submitAttempted ? validationErrors[group.id] : undefined"
                tabindex="-1"
                @update:selected-ids="setSelection(group.id, $event)"
              />
            </div>
          </section>

          <section v-if="addons.length > 0" class="config-section">
            <h2>Adicionais</h2>
            <div class="config-section__body addons-list">
              <AddonRow
                v-for="addon in addons"
                :key="addon.id"
                :addon="addon"
                :selected="selectedAddonIds.includes(addon.id)"
                @update:selected="toggleAddon(addon.id, $event)"
              />
            </div>
          </section>

          <section class="config-section">
            <label class="note-label" for="product-note"> Observação <span>opcional</span> </label>
            <textarea
              id="product-note"
              v-model="note"
              class="note-field"
              :maxlength="MAX_NOTE_LENGTH"
              rows="3"
              placeholder="Ex.: sem cebola, molho à parte..."
              aria-describedby="product-note-counter"
            ></textarea>
            <p id="product-note-counter" class="note-counter">
              {{ note.length }}/{{ MAX_NOTE_LENGTH }} caracteres
            </p>
          </section>

          <p v-if="hasValidationErrors && submitAttempted" class="form-error" role="alert">
            Revise as opções obrigatórias antes de adicionar.
          </p>

          <p v-if="addFailed" class="form-error" role="alert">
            Não foi possível adicionar este item. Confira as opções e tente novamente.
          </p>

          <p v-if="added" class="config-success" role="status">
            Adicionado ao carrinho.
            <RouterLink class="config-success__link" :to="{ name: 'cart' }">
              Ver carrinho
            </RouterLink>
          </p>

          <div class="product-action">
            <div class="product-action__total">
              <span class="product-action__label">Total estimado</span>
              <PriceText :value-cents="totalCents" />
            </div>
            <BaseButton type="submit" :disabled="!product.available">
              Adicionar ao carrinho
            </BaseButton>
          </div>

          <p class="server-note">
            O valor final é confirmado pelo servidor no carrinho e no checkout.
          </p>
        </form>
      </div>
    </template>
  </main>
</template>

<style scoped>
.product-page {
  display: grid;
  gap: 16px;
}

.back-link {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.back-link:hover {
  color: var(--brand-neon);
}

.product-layout {
  display: grid;
  gap: 18px;
}

.product-summary {
  display: grid;
  gap: 16px;
  align-content: start;
}

.product-summary__media {
  aspect-ratio: 4 / 3;
  overflow: hidden;
  border: 1px solid var(--bg-border);
  border-radius: 20px;
  background: var(--bg-input);
}

.product-summary__copy {
  display: grid;
  gap: 10px;
}

.eyebrow {
  margin: 0;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.product-summary h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  font-weight: 400;
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.product-summary__description {
  margin: 0;
  color: var(--text-sub);
  line-height: 1.65;
}

.product-summary__allergens {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.product-summary__allergens strong {
  color: var(--text-main);
}

.base-price {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  padding-top: 4px;
}

.base-price > span {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.sold-out-message {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid var(--status-error);
  border-radius: 12px;
  color: var(--status-error);
  font-size: var(--font-size-caption);
  font-weight: 700;
  line-height: 1.5;
}

.product-config {
  display: grid;
  gap: 14px;
}

.config-section {
  display: grid;
  gap: 10px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.config-section h2 {
  margin: 0;
  font-size: var(--font-size-h1);
}

.config-section__body {
  display: grid;
  gap: 10px;
}

.addons-list {
  gap: 2px;
}

.note-label {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 4px 12px;
  font-size: var(--font-size-h2);
  font-weight: 800;
}

.note-label span {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 600;
}

.note-field {
  width: 100%;
  min-height: 96px;
  resize: vertical;
  padding: 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  line-height: 1.5;
}

.note-field::placeholder {
  color: var(--text-sub);
}

.note-field:focus {
  border-color: var(--brand-neon);
  outline: 0;
}

.note-counter {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  text-align: right;
}

.form-error,
.config-success {
  margin: 0;
  padding: 12px 14px;
  border-radius: 12px;
  font-size: var(--font-size-caption);
  font-weight: 700;
  line-height: 1.5;
}

.form-error {
  border: 1px solid var(--status-error);
  color: var(--status-error);
}

.config-success {
  border: 1px solid var(--status-open);
  color: var(--status-open);
}

.config-success__link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  margin-left: 8px;
  color: var(--brand-primary);
  font-weight: 800;
  text-decoration: none;
}

.product-action {
  position: sticky;
  bottom: 0;
  z-index: 30;
  display: grid;
  gap: 10px;
  margin: 2px -16px -16px;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
  border-top: 1px solid var(--bg-border);
  background: var(--bg-primary);
}

.product-action__total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.product-action__label {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.product-action .base-button {
  width: 100%;
}

.server-note {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
  text-align: center;
}

.text-link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-weight: 800;
  text-decoration: none;
}

.text-link:hover {
  color: var(--brand-neon);
}

@media (min-width: 900px) {
  .product-layout {
    grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
    align-items: start;
    gap: 28px;
  }

  .product-summary {
    position: sticky;
    top: 88px;
  }

  .product-action {
    position: static;
    margin: 0;
    padding: 0;
    border-top: 0;
    background: transparent;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .product-action .base-button {
    width: auto;
  }

  .server-note {
    text-align: left;
  }
}
</style>
