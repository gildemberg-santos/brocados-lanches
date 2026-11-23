import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchDeliveryZones } from '../services/deliveryZones'
import type { DeliveryZone } from '../types/delivery'

export type DeliveryZonesState = 'idle' | 'loading' | 'success' | 'error'

const LOAD_ERROR_MESSAGE = 'Não foi possível carregar as taxas de entrega. Tente novamente.'

export const useDeliveryZonesStore = defineStore('deliveryZones', () => {
  const state = ref<DeliveryZonesState>('idle')
  const zones = ref<DeliveryZone[]>([])
  const error = ref('')

  const cheapestFeeCents = computed<number | null>(() => {
    if (state.value !== 'success' || zones.value.length === 0) {
      return null
    }

    return Math.min(...zones.value.map((zone) => zone.feeCents))
  })

  async function loadZones(force = false): Promise<void> {
    if (state.value === 'loading' || (!force && state.value === 'success')) {
      return
    }

    state.value = 'loading'
    error.value = ''

    try {
      zones.value = await fetchDeliveryZones()
      state.value = 'success'
    } catch {
      zones.value = []
      state.value = 'error'
      error.value = LOAD_ERROR_MESSAGE
    }
  }

  function ensureLoaded(): Promise<void> {
    if (state.value === 'success' || state.value === 'loading') {
      return Promise.resolve()
    }

    return loadZones()
  }

  return {
    state,
    zones,
    error,
    cheapestFeeCents,
    loadZones,
    ensureLoaded,
  }
})
