import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

import HomeView from '../views/HomeView.vue'
import { redirectAuthenticatedFromLogin, requireAdminAuth } from './adminGuard'

const publicRoutes: RouteRecordRaw[] = [
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
]

const adminPlaceholder = () => import('../views/admin/AdminPlaceholderView.vue')

const adminRoutes: RouteRecordRaw = {
  path: '/admin',
  component: () => import('../components/admin/AdminShell.vue'),
  meta: { admin: true },
  beforeEnter: requireAdminAuth,
  children: [
    {
      path: '',
      name: 'admin-dashboard',
      component: () => import('../views/admin/AdminDashboardView.vue'),
      meta: { title: 'Painel' },
    },
    {
      path: 'pedidos',
      name: 'admin-orders',
      component: () => import('../views/admin/AdminOrdersView.vue'),
      meta: { title: 'Pedidos' },
    },
    {
      path: 'pedidos/:id',
      name: 'admin-order-detail',
      component: () => import('../views/admin/AdminOrderDetailView.vue'),
      meta: { title: 'Detalhe do pedido' },
    },
    {
      path: 'produtos',
      name: 'admin-products',
      component: adminPlaceholder,
      meta: { title: 'Produtos' },
    },
    {
      path: 'categorias',
      name: 'admin-categories',
      component: adminPlaceholder,
      meta: { title: 'Categorias' },
    },
    {
      path: 'adicionais',
      name: 'admin-addons',
      component: adminPlaceholder,
      meta: { title: 'Adicionais' },
    },
    {
      path: 'locais',
      name: 'admin-zones',
      component: adminPlaceholder,
      meta: { title: 'Locais e taxas' },
    },
    {
      path: 'loja',
      name: 'admin-store',
      component: adminPlaceholder,
      meta: { title: 'Loja' },
    },
    {
      path: 'conta',
      name: 'admin-account',
      component: adminPlaceholder,
      meta: { title: 'Conta' },
    },
    {
      path: 'usuarios',
      name: 'admin-users',
      component: adminPlaceholder,
      meta: { title: 'Usuários', roles: ['owner'] },
    },
    {
      path: 'auditoria',
      name: 'admin-audit',
      component: adminPlaceholder,
      meta: { title: 'Auditoria', roles: ['owner'] },
    },
    {
      path: 'sem-acesso',
      name: 'admin-forbidden',
      component: () => import('../views/admin/AdminForbiddenView.vue'),
      meta: { title: 'Acesso negado' },
    },
  ],
}

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    ...publicRoutes,
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('../views/admin/AdminLoginView.vue'),
      beforeEnter: redirectAuthenticatedFromLogin,
    },
    adminRoutes,
  ],
})

export default router
