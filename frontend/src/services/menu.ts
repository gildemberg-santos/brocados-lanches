import type { Menu } from '../types/catalog'

import { getData } from './http'

export function fetchMenu(): Promise<Menu> {
  return getData<Menu>('/api/v1/menu')
}
