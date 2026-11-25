<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

import {
  buildMenuJsonLd,
  buildProductJsonLd,
  buildRestaurantJsonLd,
  serializeJsonLd,
} from '../features/seo/structuredData'
import { useMenuStore } from '../stores/menu'
import { useStoreStatusStore } from '../stores/storeStatus'

const SCRIPT_ID = 'brocados-structured-data'
const route = useRoute()
const statusStore = useStoreStatusStore()
const menuStore = useMenuStore()
let requestId = 0

function clear(): void {
  document.getElementById(SCRIPT_ID)?.remove()
}

function inject(content: string): void {
  clear()
  if (content.length === 0) {
    return
  }

  const script = document.createElement('script')
  script.id = SCRIPT_ID
  script.type = 'application/ld+json'
  script.textContent = content
  document.head.appendChild(script)
}

async function apply(): Promise<void> {
  const current = ++requestId
  const name = route.name

  if (name !== 'home' && name !== 'menu' && name !== 'product') {
    clear()
    return
  }

  await Promise.all([statusStore.load(), menuStore.ensureLoaded()])
  if (current !== requestId) {
    return
  }

  const status = statusStore.status
  if (!status) {
    clear()
    return
  }

  if (name === 'product') {
    const slug = route.params.slug
    const product = typeof slug === 'string' ? menuStore.findProductBySlug(slug) : undefined
    inject(product ? serializeJsonLd([buildProductJsonLd(product)]) : '')
    return
  }

  const objects = [buildRestaurantJsonLd(status)]
  if (name === 'menu') {
    const menu = buildMenuJsonLd(menuStore.categories)
    if (menu) {
      objects.push(menu)
    }
  }
  inject(serializeJsonLd(objects))
}

onMounted(() => {
  void apply()
})

watch(
  () => route.fullPath,
  () => {
    void apply()
  },
)
</script>

<template>
  <span hidden aria-hidden="true"></span>
</template>
