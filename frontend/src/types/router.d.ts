import 'vue-router'

import type { AdminRole } from './admin'

declare module 'vue-router' {
  interface RouteMeta {
    admin?: boolean
    roles?: AdminRole[]
    title?: string
  }
}
