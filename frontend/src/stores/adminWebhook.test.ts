import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from '../services/http'
import * as api from '../services/adminWebhook'
import { useAdminWebhookStore } from './adminWebhook'

describe('store de webhook administrativo', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('carrega a configuração', async () => {
    const fetchSpy = vi.spyOn(api, 'fetchAdminWebhook').mockResolvedValue({
      url: 'https://hooks.exemplo.invalid/path',
      active: true,
      events: ['order.created'],
      createdAt: '2026-11-26T12:00:00Z',
      updatedAt: '2026-11-26T12:00:00Z',
    })
    const store = useAdminWebhookStore()

    await store.load()

    expect(store.state).toBe('success')
    expect(store.settings?.active).toBe(true)
    fetchSpy.mockRestore()
  })

  it('transforma erro de permissão em mensagem amigável', async () => {
    const fetchSpy = vi
      .spyOn(api, 'fetchAdminWebhook')
      .mockRejectedValue(new ApiError(403, 'FORBIDDEN', 'negado'))
    const store = useAdminWebhookStore()

    await store.load()

    expect(store.state).toBe('error')
    expect(store.error).toContain('permissão')
    fetchSpy.mockRestore()
  })

  it('salva a configuração e mantém o estado de sucesso', async () => {
    const updateSpy = vi.spyOn(api, 'updateAdminWebhook').mockResolvedValue({
      url: 'http://127.0.0.1:5678/hook',
      active: true,
      events: ['order.created'],
      createdAt: '2026-11-26T12:00:00Z',
      updatedAt: '2026-11-26T12:00:00Z',
    })
    const store = useAdminWebhookStore()

    const saved = await store.save({
      url: 'http://127.0.0.1:5678/hook',
      active: true,
      events: ['order.created'],
    })

    expect(saved).toBe(true)
    expect(updateSpy).toHaveBeenCalledWith({
      url: 'http://127.0.0.1:5678/hook',
      active: true,
      events: ['order.created'],
    })
    expect(store.saveSuccess).toContain('sucesso')
    expect(store.settings?.url).toBe('http://127.0.0.1:5678/hook')
    updateSpy.mockRestore()
  })
})
