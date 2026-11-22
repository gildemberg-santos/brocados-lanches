import { createRouter, createWebHistory } from 'vue-router'

import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/cardapio',
      name: 'menu',
      component: () => import('../views/MenuView.vue'),
    },
    {
      path: '/produto/:slug',
      name: 'product',
      component: () => import('../views/ProductView.vue'),
    },
    {
      path: '/carrinho',
      name: 'cart',
      component: () => import('../views/CartView.vue'),
    },
  ],
})

export default router
