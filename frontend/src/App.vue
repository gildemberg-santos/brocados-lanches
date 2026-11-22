<script setup lang="ts">
import { RouterView } from 'vue-router'

import SiteHeader from './components/SiteHeader.vue'
</script>

<template>
  <a class="skip-link" href="#conteudo">Pular para o conteúdo</a>
  <SiteHeader />
  <RouterView />
</template>
