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

const sources = computed(() => {
  const match = /^(.*)\.(webp|jpe?g|png)$/i.exec(props.src.trim())
  if (!match) {
    return { avif: '', webp: '' }
  }

  const base = match[1]
  const extension = match[2].toLowerCase()
  return {
    avif: `${base}.avif`,
    webp: extension === 'webp' ? '' : `${base}.webp`,
  }
})

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
    <picture v-if="src.length > 0 && !failed">
      <source v-if="sources.avif.length > 0" :srcset="sources.avif" type="image/avif" />
      <source v-if="sources.webp.length > 0" :srcset="sources.webp" type="image/webp" />
      <img
        class="product-image__asset"
        :src="src"
        :alt="alt"
        :width="width"
        :height="height"
        :loading="eager ? 'eager' : 'lazy'"
        @error="failed = true"
      />
    </picture>
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
