<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import CartLine from '../components/CartLine.vue'
import CartSummary from '../components/CartSummary.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import { buildCartLines, calculateSubtotalCents } from '../features/cart/cart'
import { useCartStore } from '../stores/cart'
import { useDeliveryZonesStore } from '../stores/deliveryZones'
import { useMenuStore } from '../stores/menu'

const cart = useCartStore()
const menu = useMenuStore()
const zones = useDeliveryZonesStore()
const router = useRouter()

const isLoading = computed(() => menu.state === 'idle' || menu.state === 'loading')
const lines = computed(() => buildCartLines(cart.items, menu.products).lines)
const subtotalCents = computed(() => calculateSubtotalCents(lines.value))

onMounted(async () => {
  const menuLoad = menu.ensureLoaded()
  const zonesLoad = zones.ensureLoaded()

  await menuLoad

  if (menu.state === 'success') {
    const { unavailableKeys } = buildCartLines(cart.items, menu.products)
    cart.removeUnavailable(unavailableKeys)
  }

  await zonesLoad
})
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell cart-page">
    <RouterLink class="back-link" :to="{ name: 'menu' }">← Continuar comprando</RouterLink>
    <h1>Seu carrinho</h1>

    <div v-if="cart.notice.length > 0" class="cart-notice" role="status">
      <span>{{ cart.notice }}</span>
      <button type="button" class="cart-notice__dismiss" @click="cart.dismissNotice()">
        Dispensar
      </button>
    </div>

    <EmptyState
      v-if="cart.isEmpty"
      title="Seu carrinho está vazio"
      description="Escolha um lanche no cardápio para começar o seu pedido."
    >
      <template #actions>
        <RouterLink class="text-link" :to="{ name: 'menu' }">Ver cardápio</RouterLink>
      </template>
    </EmptyState>

    <template v-else>
      <LoadingSkeleton v-if="isLoading" label="Carregando carrinho" :count="2" />

      <ErrorState
        v-else-if="menu.state === 'error'"
        :message="menu.error"
        @retry="menu.loadMenu(true)"
      />

      <div v-else class="cart-layout">
        <section class="cart-lines" aria-labelledby="cart-lines-title">
          <div class="cart-lines__header">
            <h2 id="cart-lines-title">Itens do carrinho</h2>
            <BaseButton variant="ghost" class="cart-lines__clear" @click="cart.clear()">
              Limpar carrinho
            </BaseButton>
          </div>

          <CartLine
            v-for="line in lines"
            :key="line.key"
            :line="line"
            @update:quantity="cart.setQuantity(line.key, $event)"
            @remove="cart.remove(line.key)"
          />

          <p v-if="lines.length === 0" class="cart-lines__empty" role="status">
            Os itens salvos não estão mais disponíveis. Escolha outros produtos no cardápio.
          </p>
        </section>

        <CartSummary
          :subtotal-cents="subtotalCents"
          :delivery-fee-cents="zones.cheapestFeeCents"
          :item-count="cart.count"
        >
          <template #action>
            <BaseButton :disabled="lines.length === 0" @click="router.push({ name: 'checkout' })">
              Continuar
            </BaseButton>
            <p class="cart-cta-note">O total é calculado no servidor no próximo passo.</p>
          </template>
        </CartSummary>
      </div>
    </template>
  </main>
</template>

<style scoped>
.cart-page {
  display: grid;
  gap: 16px;
}

.cart-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  font-weight: 400;
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.back-link,
.text-link {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.back-link:hover,
.text-link:hover {
  color: var(--brand-neon);
}

.cart-notice {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
  padding: 10px 14px;
  border: 1px solid var(--status-open);
  border-radius: 12px;
  color: var(--text-main);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.cart-notice__dismiss {
  min-height: 44px;
  padding: 0 8px;
  border: 0;
  background: transparent;
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
  font-weight: 800;
  cursor: pointer;
}

.cart-layout {
  display: grid;
  gap: 16px;
  align-items: start;
}

.cart-lines {
  display: grid;
  gap: 10px;
}

.cart-lines__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.cart-lines__header h2 {
  margin: 0;
  font-size: var(--font-size-h1);
}

.cart-lines__clear {
  min-height: 44px;
  padding-inline: 12px;
  font-size: var(--font-size-caption);
}

.cart-lines__empty {
  margin: 0;
  padding: 16px;
  border: 1px dashed var(--bg-border);
  border-radius: 16px;
  color: var(--text-sub);
  line-height: 1.6;
}

.cart-cta-note {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
  text-align: center;
}

@media (min-width: 900px) {
  .cart-layout {
    grid-template-columns: minmax(0, 1.5fr) minmax(280px, 1fr);
    gap: 24px;
  }
}
</style>
