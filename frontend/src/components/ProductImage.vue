<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    src: string
    alt: string
    width: number
    height: number
    eager?: boolean
  }>(),
  {
    eager: false,
  },
)

const failed = ref(false)

const initial = computed(() => {
  const trimmed = props.alt.trim()
  return trimmed.length > 0 ? trimmed.charAt(0).toUpperCase() : 'B'
})

watch(
  () => props.src,
  () => {
    failed.value = false
  },
)
</script>

<template>
  <span class="product-image">
    <img
      v-if="src.length > 0 && !failed"
      class="product-image__asset"
      :src="src"
      :alt="alt"
      :width="width"
      :height="height"
      :loading="eager ? 'eager' : 'lazy'"
      @error="failed = true"
    />
    <span v-else class="product-image__fallback" role="img" :aria-label="alt">
      <span class="product-image__initial" aria-hidden="true">{{ initial }}</span>
      <span class="product-image__caption" aria-hidden="true">Sem foto</span>
    </span>
  </span>
</template>

<style scoped>
.product-image {
  display: block;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: linear-gradient(135deg, var(--bg-input), var(--bg-surface));
}

.product-image__asset {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-image__fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-content: center;
  justify-items: center;
  gap: 2px;
  color: var(--text-sub);
}

.product-image__initial {
  color: var(--brand-primary);
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
}

.product-image__caption {
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
</style>
