<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import SiteFooter from './components/SiteFooter.vue'
import SiteHeader from './components/SiteHeader.vue'
import StructuredData from './components/StructuredData.vue'

const route = useRoute()
// O login também faz parte do painel, embora não use o meta `admin`.
const isAdminRoute = computed(() => route.meta.admin === true || route.path.startsWith('/admin'))
</script>

<template>
  <a class="skip-link" href="#conteudo">Pular para o conteúdo</a>
  <SiteHeader v-if="!isAdminRoute" />
  <StructuredData v-if="!isAdminRoute" />
  <RouterView />
  <SiteFooter v-if="!isAdminRoute" />
</template>
