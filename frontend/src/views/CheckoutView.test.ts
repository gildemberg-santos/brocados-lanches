import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { CHECKOUT_PREFERENCES_KEY, loadCheckoutPreferences } from '../features/checkout/preferences'
import { fetchDeliveryZones } from '../services/deliveryZones'
import { fetchMenu } from '../services/menu'
import { createOrder, fetchOrderQuote } from '../services/orders'
import { useCartStore } from '../stores/cart'
import {
  makeCatalogProduct,
  makeCategory,
  makeMenu,
  makeOption,
  makeOptionGroup,
} from '../test/fixtures/catalog'
import CheckoutView from './CheckoutView.vue'

vi.mock('../services/menu', () => ({ fetchMenu: vi.fn() }))
vi.mock('../services/deliveryZones', () => ({ fetchDeliveryZones: vi.fn() }))
vi.mock('../services/orders', () => ({
  fetchOrderQuote: vi.fn(),
  createOrder: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)
const mockedFetchDeliveryZones = vi.mocked(fetchDeliveryZones)
const mockedFetchOrderQuote = vi.mocked(fetchOrderQuote)
const mockedCreateOrder = vi.mocked(createOrder)

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })
const product = makeCatalogProduct({
  id: 'product-1',
  optionGroups: [
    makeOptionGroup({ id: 'group-1', options: [makeOption({ id: 'option-1', name: 'Ao ponto' })] }),
  ],
})
const menu = makeMenu({ categories: [makeCategory({ products: [product] })] })
const quote = {
  subtotalCents: 2300,
  deliveryFeeCents: 800,
  totalCents: 3100,
  currency: 'BRL',
  items: [],
  warnings: [],
}
const created = {
  orderNumber: 1042,
  publicId: 'public-123',
  status: 'pending',
  fulfillment: 'delivery' as const,
  subtotalCents: 2300,
  deliveryFeeCents: 800,
  totalCents: 3100,
  createdAt: '2026-11-22T22:00:00Z',
  whatsappUrl: 'https://wa.me/5585999999999?text=pedido',
}

interface MountOptions {
  seed?: boolean
  quoteError?: boolean
  createError?: boolean
}

async function mountCheckout(options: MountOptions = {}) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const cart = useCartStore()
  if (options.seed) {
    cart.add(product, { 'group-1': ['option-1'] }, [], '', 1)
  }

  mockedFetchMenu.mockResolvedValue(menu)
  mockedFetchDeliveryZones.mockResolvedValue([
    { id: 'zona-1', name: 'Centro', feeCents: 800, sortOrder: 1 },
  ])
  if (options.quoteError) {
    mockedFetchOrderQuote.mockRejectedValue(new Error('falha'))
  } else {
    mockedFetchOrderQuote.mockResolvedValue(quote)
  }
  if (options.createError) {
    mockedCreateOrder.mockRejectedValue(new Error('falha'))
  } else {
    mockedCreateOrder.mockResolvedValue(created)
  }

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/carrinho', name: 'cart', component: StubView },
      { path: '/checkout', name: 'checkout', component: StubView },
      { path: '/privacidade', name: 'privacy', component: StubView },
      { path: '/pedido/:publicId', name: 'order', component: StubView },
    ],
  })
  await router.push('/checkout')
  await router.isReady()

  const wrapper = mount(CheckoutView, { global: { plugins: [pinia, router] } })
  await flushPromises()

  return { wrapper, cart, router }
}

type CheckoutWrapper = Awaited<ReturnType<typeof mountCheckout>>['wrapper']

async function chooseDeliveryZone(wrapper: CheckoutWrapper): Promise<void> {
  await wrapper.get('select[name="zoneId"]').setValue('zona-1')
  await flushPromises()
  vi.advanceTimersByTime(300)
  await flushPromises()
}

function seedPreferences(preferences: Record<string, unknown>): void {
  localStorage.setItem(CHECKOUT_PREFERENCES_KEY, JSON.stringify({ version: 1, preferences }))
}

async function fillForm(wrapper: CheckoutWrapper): Promise<void> {
  await wrapper.get('select[name="zoneId"]').setValue('zona-1')
  await wrapper.get('input[name="customerName"]').setValue('Maria Silva')
  await wrapper.get('input[name="phone"]').setValue('85999999999')
  await wrapper.get('input[name="street"]').setValue('Rua das Flores')
  await wrapper.get('input[name="number"]').setValue('450')
  await wrapper.get('input[name="district"]').setValue('Tapera')
  await wrapper.get('input[name="zipCode"]').setValue('61700000')
  await wrapper.get('input[name="changeFor"]').setValue('50,00')
  await wrapper.get('textarea[name="note"]').setValue('Sem cebola')
  await flushPromises()
  vi.advanceTimersByTime(300)
  await flushPromises()
}

describe('CheckoutView', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows the empty state when the cart has no items', async () => {
    const { wrapper } = await mountCheckout()

    expect(wrapper.text()).toContain('Seu carrinho está vazio')
    expect(mockedFetchOrderQuote).not.toHaveBeenCalled()
  })

  it('não cota o pedido após recarregar sem zona de entrega', async () => {
    const { wrapper } = await mountCheckout({ seed: true })

    expect(mockedFetchOrderQuote).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escolha uma zona de entrega para calcular o total.')
    expect(wrapper.text()).not.toContain('Não foi possível')
  })

  it('preenche o checkout com as preferências salvas no próximo pedido', async () => {
    seedPreferences({
      fulfillment: 'pickup',
      zoneId: 'zona-antiga',
      customerName: 'João Souza',
      phone: '85988888888',
      paymentMethod: 'pix',
    })

    const { wrapper } = await mountCheckout({ seed: true })

    expect(
      (wrapper.get('input[name="fulfillment"][value="pickup"]').element as HTMLInputElement)
        .checked,
    ).toBe(true)
    expect((wrapper.get('input[name="customerName"]').element as HTMLInputElement).value).toBe(
      'João Souza',
    )
    expect((wrapper.get('input[name="phone"]').element as HTMLInputElement).value).toBe(
      '85988888888',
    )
    expect(
      (wrapper.get('input[name="paymentMethod"][value="pix"]').element as HTMLInputElement).checked,
    ).toBe(true)
    expect(mockedFetchOrderQuote).toHaveBeenCalledTimes(1)
  })

  it('descarta uma zona salva que não está mais ativa', async () => {
    seedPreferences({ fulfillment: 'delivery', zoneId: 'zona-inexistente' })

    const { wrapper } = await mountCheckout({ seed: true })

    expect((wrapper.get('select[name="zoneId"]').element as HTMLSelectElement).value).toBe('')
    expect(mockedFetchOrderQuote).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escolha uma zona de entrega para calcular o total.')
  })

  it('não reapresenta zona inativa salva na retirada ao voltar para entrega', async () => {
    seedPreferences({ fulfillment: 'pickup', zoneId: 'zona-inexistente' })

    const { wrapper } = await mountCheckout({ seed: true })
    mockedFetchOrderQuote.mockClear()
    await wrapper.get('input[name="fulfillment"][value="delivery"]').setValue(true)
    await flushPromises()

    expect((wrapper.get('select[name="zoneId"]').element as HTMLSelectElement).value).toBe('')
    expect(mockedFetchOrderQuote).not.toHaveBeenCalled()
  })

  it('avisa que as escolhas ficam salvas neste navegador', async () => {
    const { wrapper } = await mountCheckout({ seed: true })

    expect(wrapper.text()).toContain('Suas escolhas ficam salvas neste navegador')
    expect(wrapper.find('a[href="/privacidade"]').exists()).toBe(true)
  })

  it('shows server totals and catalog item names', async () => {
    const { wrapper } = await mountCheckout({ seed: true })
    await chooseDeliveryZone(wrapper)

    expect(wrapper.text()).toContain('23,00')
    expect(wrapper.text()).toContain('8,00')
    expect(wrapper.text()).toContain('31,00')
    expect(wrapper.text()).toContain('Brocado Cheddar')
    expect(mockedFetchOrderQuote).toHaveBeenCalledTimes(1)
  })

  it('shows a quote error with retry', async () => {
    const { wrapper } = await mountCheckout({ seed: true, quoteError: true })
    await chooseDeliveryZone(wrapper)
    const review = wrapper.get('section[aria-labelledby="review-title"]')

    expect(wrapper.text()).toContain('Não foi possível')
    expect(wrapper.text()).toContain('Tentar novamente')
    expect(review.text()).toContain('Aguardando cotação')
    expect(review.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('mantém a revisão e o envio como última etapa com o total cotado', async () => {
    const { wrapper } = await mountCheckout({ seed: true })
    await chooseDeliveryZone(wrapper)
    const review = wrapper.get('section[aria-labelledby="review-title"]')

    expect(wrapper.get('#review-title').text()).toContain('5. Revisão e envio')
    expect(review.get('[data-testid="review-total"]').text()).toContain('31,00')
    expect(review.text()).toContain('Enviar pedido')
    expect(review.find('button[type="submit"]').exists()).toBe(true)

    await wrapper.get('input[name="fulfillment"][value="pickup"]').setValue(true)
    await flushPromises()
    vi.advanceTimersByTime(300)
    await flushPromises()

    expect(wrapper.get('#review-title').text()).toContain('4. Revisão e envio')
    expect(wrapper.find('section[aria-labelledby="address-title"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="review-total"]').text()).toContain('31,00')
  })

  it('validates the form before submitting', async () => {
    const { wrapper } = await mountCheckout({ seed: true })
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Informe seu nome completo.')
    expect(mockedCreateOrder).not.toHaveBeenCalled()
  })

  it('associa as seções aos títulos existentes', async () => {
    const { wrapper } = await mountCheckout({ seed: true })

    const sections = wrapper.findAll('section[aria-labelledby]')
    expect(sections.length).toBeGreaterThan(0)
    for (const section of sections) {
      const titleId = section.attributes('aria-labelledby')
      expect(titleId).toBeTruthy()
      expect(wrapper.find(`#${titleId}`).exists()).toBe(true)
    }
    expect(wrapper.find('#customer-title').exists()).toBe(true)
    expect(wrapper.find('#payment-title').exists()).toBe(true)
  })

  it('associa cada erro de campo ao input correspondente', async () => {
    const { wrapper } = await mountCheckout({ seed: true })
    await wrapper.get('input[name="city"]').setValue('')
    await wrapper.get('input[name="state"]').setValue('')
    await wrapper.get('textarea[name="note"]').setValue('x'.repeat(281))
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.get('[name="zoneId"]').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#checkout-zoneId-error').text()).not.toBe('')

    await chooseDeliveryZone(wrapper)
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    const fields: ReadonlyArray<readonly [name: string, id: string]> = [
      ['street', 'checkout-street'],
      ['number', 'checkout-number'],
      ['district', 'checkout-district'],
      ['zipCode', 'checkout-zipCode'],
      ['city', 'checkout-city'],
      ['state', 'checkout-state'],
      ['customerName', 'checkout-customerName'],
      ['phone', 'checkout-phone'],
      ['changeFor', 'checkout-changeFor'],
      ['note', 'checkout-note'],
    ]

    for (const [name, id] of fields) {
      const input = wrapper.get(`[name="${name}"]`)
      const error = wrapper.get(`#${id}-error`)

      expect(input.attributes('id')).toBe(id)
      expect(input.attributes('aria-describedby')).toBe(`${id}-error`)
      expect(input.attributes('aria-invalid')).toBe('true')
      expect(error.text()).not.toBe('')
    }

    for (const error of wrapper.findAll('small.field-error')) {
      const errorId = error.attributes('id')
      expect(errorId).toBeTruthy()
      expect(wrapper.get(`[aria-describedby="${errorId}"]`).attributes('aria-invalid')).toBe('true')
    }
  })

  it('clears the cart and navigates after the order is created', async () => {
    const { wrapper, cart, router } = await mountCheckout({ seed: true })
    await fillForm(wrapper)

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mockedCreateOrder).toHaveBeenCalledTimes(1)
    const [request, key] = mockedCreateOrder.mock.calls[0]
    expect(request.payment.changeForCents).toBe(5000)
    expect(key.length).toBeGreaterThanOrEqual(16)
    expect(cart.isEmpty).toBe(true)
    expect(router.currentRoute.value.name).toBe('order')
    expect(sessionStorage.getItem('brocados:handoff:public-123')).toContain('wa.me')

    const preferences = loadCheckoutPreferences()
    expect(preferences).toMatchObject({
      fulfillment: 'delivery',
      zoneId: 'zona-1',
      customerName: 'Maria Silva',
      phone: '(85) 99999-9999',
      street: 'Rua das Flores',
      number: '450',
      district: 'Tapera',
      city: 'Aquiraz',
      state: 'CE',
      zipCode: '61700000',
      paymentMethod: 'cash',
    })
    const stored = localStorage.getItem(CHECKOUT_PREFERENCES_KEY) ?? ''
    expect(stored).not.toContain('Sem cebola')
    expect(stored).not.toContain('50,00')
  })

  it('keeps the cart when the order fails', async () => {
    const { wrapper, cart } = await mountCheckout({ seed: true, createError: true })
    await fillForm(wrapper)

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(cart.isEmpty).toBe(false)
    expect(wrapper.text()).toContain('Não foi possível enviar o pedido')
    expect(loadCheckoutPreferences()).toBeNull()
  })

  it('reuses the idempotency key when retrying the same request', async () => {
    const { wrapper } = await mountCheckout({ seed: true })
    mockedCreateOrder.mockRejectedValueOnce(new Error('falha de rede'))
    await fillForm(wrapper)

    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mockedCreateOrder).toHaveBeenCalledTimes(2)
    expect(mockedCreateOrder.mock.calls[0][1]).toBe(mockedCreateOrder.mock.calls[1][1])
  })

  it('renews the idempotency key when the request changes after an error', async () => {
    const { wrapper } = await mountCheckout({ seed: true })
    mockedCreateOrder.mockRejectedValueOnce(new Error('falha de rede'))
    await fillForm(wrapper)

    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await wrapper.get('input[name="phone"]').setValue('85988888888')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mockedCreateOrder).toHaveBeenCalledTimes(2)
    expect(mockedCreateOrder.mock.calls[0][1]).not.toBe(mockedCreateOrder.mock.calls[1][1])
  })
})
