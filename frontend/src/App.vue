<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import SiteHeader from './components/SiteHeader.vue'

const route = useRoute()
const isAdminRoute = computed(() => route.meta.admin === true)
</script>

<template>
  <a class="skip-link" href="#conteudo">Pular para o conteúdo</a>
  <SiteHeader v-if="!isAdminRoute" />
  <RouterView />
</template>
