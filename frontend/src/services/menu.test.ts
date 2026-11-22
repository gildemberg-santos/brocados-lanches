import { beforeEach, describe, expect, it, vi } from 'vitest'

import { makeMenu } from '../test/fixtures/catalog'
import { getData } from './http'
import { fetchMenu } from './menu'

vi.mock('./http', () => ({
  getData: vi.fn(),
}))

const mockedGetData = vi.mocked(getData)

describe('menu service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads the public menu from the API', async () => {
    const menu = makeMenu()
    mockedGetData.mockResolvedValue(menu)

    await expect(fetchMenu()).resolves.toEqual(menu)
    expect(mockedGetData).toHaveBeenCalledWith('/api/v1/menu')
  })
})
