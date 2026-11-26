import { beforeEach, describe, expect, it, vi } from 'vitest'

import { postData } from './http'
import { fetchCustomerProfile } from './customers'

vi.mock('./http', () => ({ postData: vi.fn() }))

const mockedPostData = vi.mocked(postData)

describe('customers service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('busca o perfil pelo telefone sem expor o número na URL', async () => {
    const response = { found: false }
    mockedPostData.mockResolvedValue(response)

    await expect(fetchCustomerProfile('(85) 99999-9999')).resolves.toEqual(response)
    expect(mockedPostData).toHaveBeenCalledWith('/api/v1/customers/profile', {
      phone: '(85) 99999-9999',
    })
  })
})
