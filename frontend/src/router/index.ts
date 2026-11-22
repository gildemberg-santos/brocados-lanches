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
    {
      path: '/checkout',
      name: 'checkout',
      component: () => import('../views/CheckoutView.vue'),
    },
    {
      path: '/pedido/:publicId',
      name: 'order',
      component: () => import('../views/OrderView.vue'),
      props: true,
    },
  ],
})

export default router
