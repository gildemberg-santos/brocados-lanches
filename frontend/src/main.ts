import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { initSentry } from './features/monitoring/sentry'
import router from './router'
import './styles/main.css'

const app = createApp(App)
initSentry(app)
app.use(createPinia()).use(router).mount('#app')
