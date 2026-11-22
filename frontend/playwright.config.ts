import { defineConfig, devices } from '@playwright/test'

const API_PORT = 8080
const WEB_PORT = 5173
const API_URL = `http://127.0.0.1:${API_PORT}`
const WEB_URL = `http://127.0.0.1:${WEB_PORT}`

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 2 : 0,
  // O login admin rotaciona todas as sessões do usuário; testes paralelos com a
  // mesma conta se invalidam. Por isso o E2E roda com um worker por vez.
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: WEB_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'desktop',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'mobile',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 375, height: 812 },
        isMobile: true,
        hasTouch: true,
      },
    },
  ],
  webServer: [
    {
      command: `pnpm dev --host 127.0.0.1 --port ${WEB_PORT}`,
      url: WEB_URL,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
    },
    {
      command: 'cd ../backend && go run ./cmd/api',
      url: `${API_URL}/readyz`,
      // O servidor do E2E sempre sobe isolado, com brocados_test.
      reuseExistingServer: false,
      timeout: 120_000,
      env: {
        ...process.env,
        APP_ENV: 'test',
        PORT: String(API_PORT),
        MONGO_URI: process.env.MONGO_URI ?? 'mongodb://127.0.0.1:27017',
        MONGO_DB: process.env.MONGO_DB ?? 'brocados_test',
        // Usuário fictício do E2E: o bootstrap só cria o owner quando a coleção está vazia.
        ADMIN_BOOTSTRAP_NAME: 'Dono de Teste',
        ADMIN_BOOTSTRAP_EMAIL: 'dono@teste.local',
        ADMIN_BOOTSTRAP_PASSWORD: 'senha-de-teste-12345',
      },
    },
  ],
})
