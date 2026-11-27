import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

import HomeView from '../views/HomeView.vue'
import { redirectAuthenticatedFromLogin, requireAdminAuth } from './adminGuard'
import { applyRouteMeta } from './meta'

const publicRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: HomeView,
    meta: { title: 'Início', description: 'Monte seu lanche do seu jeito e acompanhe seu pedido.' },
  },
  {
    path: '/cardapio',
    name: 'menu',
    component: () => import('../views/MenuView.vue'),
    meta: { title: 'Cardápio', description: 'Veja as opções do cardápio e escolha seu lanche.' },
  },
  {
    path: '/produto/:slug',
    name: 'product',
    component: () => import('../views/ProductView.vue'),
    meta: { title: 'Detalhes do produto', noindex: true },
  },
  {
    path: '/carrinho',
    name: 'cart',
    component: () => import('../views/CartView.vue'),
    meta: { title: 'Carrinho', noindex: true },
  },
  {
    path: '/checkout',
    name: 'checkout',
    component: () => import('../views/CheckoutView.vue'),
    meta: { title: 'Finalizar pedido', noindex: true },
  },
  {
    path: '/pedido/:publicId',
    name: 'order',
    component: () => import('../views/OrderView.vue'),
    props: true,
    meta: { title: 'Acompanhar pedido', noindex: true },
  },
  {
    path: '/meus-pedidos',
    name: 'my-orders',
    component: () => import('../views/MyOrdersView.vue'),
    meta: { title: 'Meus pedidos', noindex: true },
  },
  {
    path: '/contato',
    name: 'contact',
    component: () => import('../views/ContactView.vue'),
    meta: {
      title: 'Contato',
      description: 'Canais de contato, endereço e horários publicados pela Brocados Lanches.',
    },
  },
  {
    path: '/privacidade',
    name: 'privacy',
    component: () => import('../views/PrivacyView.vue'),
    meta: {
      title: 'Privacidade',
      description: 'Política de privacidade da Brocados Lanches, publicada em 25/11/2026.',
    },
  },
  {
    path: '/termos',
    name: 'terms',
    component: () => import('../views/TermsView.vue'),
    meta: {
      title: 'Termos de uso',
      description: 'Regras de uso do site da Brocados Lanches, publicadas em 25/11/2026.',
    },
  },
  {
    path: '/creditos',
    name: 'credits',
    component: () => import('../views/CreditsView.vue'),
    // Página institucional publicada em 25/11/2026; mídia da loja segue pendente (D-033).
    meta: {
      title: 'Créditos',
      description: 'Créditos de desenvolvimento e informações institucionais da Brocados Lanches.',
    },
  },
]

const adminRoutes: RouteRecordRaw = {
  path: '/admin',
  component: () => import('../components/admin/AdminShell.vue'),
  meta: { admin: true, noindex: true },
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
      component: () => import('../views/admin/AdminProductsView.vue'),
      meta: { title: 'Produtos' },
    },
    {
      path: 'categorias',
      name: 'admin-categories',
      component: () => import('../views/admin/AdminCategoriesView.vue'),
      meta: { title: 'Categorias' },
    },
    {
      path: 'adicionais',
      name: 'admin-addons',
      component: () => import('../views/admin/AdminAddonsView.vue'),
      meta: { title: 'Adicionais' },
    },
    {
      path: 'locais',
      name: 'admin-zones',
      component: () => import('../views/admin/AdminZonesView.vue'),
      meta: { title: 'Locais e taxas' },
    },
    {
      path: 'loja',
      name: 'admin-store',
      component: () => import('../views/admin/AdminStoreView.vue'),
      meta: { title: 'Loja' },
    },
    {
      path: 'integracoes/webhook',
      name: 'admin-webhook',
      component: () => import('../views/admin/AdminWebhookView.vue'),
      meta: { title: 'Webhook', roles: ['owner'] },
    },
    {
      path: 'conta',
      name: 'admin-account',
      component: () => import('../views/admin/AdminAccountView.vue'),
      meta: { title: 'Conta' },
    },
    {
      path: 'usuarios',
      name: 'admin-users',
      component: () => import('../views/admin/AdminUsersView.vue'),
      meta: { title: 'Usuários', roles: ['owner'] },
    },
    {
      path: 'auditoria',
      name: 'admin-audit',
      component: () => import('../views/admin/AdminAuditView.vue'),
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
      meta: { title: 'Acessar painel', noindex: true },
    },
    adminRoutes,
    // Catch-all por último: qualquer URL desconhecida cai na página 404.
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/NotFoundView.vue'),
      meta: {
        title: 'Página não encontrada',
        description: 'Página não encontrada.',
        noindex: true,
      },
    },
  ],
})

router.afterEach((to) => {
  applyRouteMeta(to)
})

export default router
