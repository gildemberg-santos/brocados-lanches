import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { fetchHealth } from '../services/health'

export type HealthState = 'loading' | 'online' | 'offline'

export const useHealthStore = defineStore('health', () => {
  const state = ref<HealthState>('loading')
  const mongo = ref('unknown')
  const error = ref('')

  const isReady = computed(() => state.value === 'online')

  async function check(): Promise<void> {
    state.value = 'loading'
    error.value = ''

    try {
      const health = await fetchHealth()
      mongo.value = health.checks?.mongo ?? 'unknown'

      if (health.status === 'ok' && mongo.value === 'ok') {
        state.value = 'online'
      } else {
        state.value = 'offline'
        error.value = 'Serviço indisponível.'
      }
    } catch {
      mongo.value = 'error'
      state.value = 'offline'
      error.value = 'Não foi possível falar com a API.'
    }
  }

  return {
    state,
    mongo,
    error,
    isReady,
    check,
  }
})
