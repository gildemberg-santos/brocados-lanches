<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  valueCents: number
}>()

const currencyFormatter = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
})

const formattedPrice = computed(() => currencyFormatter.format(props.valueCents / 100))
</script>

<template>
  <span class="price-text">{{ formattedPrice }}</span>
</template>

<style scoped>
.price-text {
  color: var(--brand-neon);
  font-size: var(--font-size-price);
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}
</style>
