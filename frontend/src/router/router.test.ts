import { describe, expect, it } from 'vitest'

import router from './index'

describe('router', () => {
  it('registers the home route', () => {
    const route = router.getRoutes().find((candidate) => candidate.name === 'home')

    expect(route?.path).toBe('/')
  })
})
