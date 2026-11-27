import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

import { cspNoncePlugin } from './csp.ts'

// No desenvolvimento o Vite roda em outra porta da API. O proxy alinha o
// Origin ao Host repassado para não falsificar a proteção CSRF do backend.
const apiProxy = {
  target: 'http://127.0.0.1:8080',
  changeOrigin: true,
  headers: { Origin: 'http://127.0.0.1:8080' },
  // Os canais de pedido usam WebSocket; sem isso o Vite não repassa o upgrade.
  ws: true,
}

export default defineConfig({
  plugins: [cspNoncePlugin(), vue(), tailwindcss()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    fs: {
      allow: ['..'],
    },
    proxy: {
      '/api': apiProxy,
      '/healthz': apiProxy,
      '/readyz': apiProxy,
    },
  },
  preview: {
    host: '0.0.0.0',
    port: 4173,
  },
})
