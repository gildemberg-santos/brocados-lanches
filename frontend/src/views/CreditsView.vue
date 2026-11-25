<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell info-page">
    <h1>Créditos</h1>

    <aside class="info-page__notice" role="note" aria-label="Situação do texto">
      <p>
        <strong>Página institucional em avaliação.</strong>
        Os créditos de desenvolvimento abaixo usam dados do currículo fornecido pelo autor.
        Logotipo, fotografias e textos definitivos da loja ainda dependem de validação; nenhum
        crédito sobre esses itens é afirmado antes dessa confirmação.
      </p>
    </aside>

    <section class="credit-hero" aria-labelledby="credits-author-title">
      <h2 id="credits-author-title">Gildemberg Santos</h2>
      <p>
        Desenvolvedor de software com mais de uma década de experiência, especializado em Ruby on
        Rails, com conhecimento em Go, Python, RSpec e Docker.
      </p>
      <p class="credit-hero__meta">Aquiraz, CE</p>
      <nav class="credit-links" aria-label="Links profissionais">
        <a
          href="https://www.linkedin.com/in/gildemberg-santos/"
          target="_blank"
          rel="noopener noreferrer"
          >LinkedIn</a
        >
        <a href="https://github.com/gildemberg-santos" target="_blank" rel="noopener noreferrer"
          >GitHub</a
        >
      </nav>
    </section>

    <div class="credit-grid">
      <section class="credit-card" aria-labelledby="credits-career-title">
        <h2 id="credits-career-title">Trajetória</h2>
        <ul>
          <li>
            Desenvolvedor Ruby on Rails na Leadster (out/2021–nov/2026): backend, integrações com
            Meta Ads, WhatsApp e GA4 e soluções com IA generativa para chatbots.
          </li>
          <li>
            Desenvolvedor autônomo (out/2013–set/2021, Aquiraz–CE): sistemas e sites sob demanda.
          </li>
          <li>Instrutor e monitor de informática (2008–2014): Instituto Mix e Iteva.</li>
        </ul>
      </section>

      <section class="credit-card" aria-labelledby="credits-training-title">
        <h2 id="credits-training-title">Formação e projetos</h2>
        <ul>
          <li>
            Bacharelado em Ciência da Computação — Centro Universitário Estácio do Ceará (em
            andamento).
          </li>
          <li>Certificação MTA 98-361 — Microsoft.</li>
          <li>
            Projetos publicados: <code>bitrix24</code> e <code>salesforce</code> em Ruby, e
            <code>webcrawlerurl_v2</code> em Go, com testes e integração contínua.
          </li>
        </ul>
      </section>

      <section class="credit-card" aria-labelledby="credits-implementation-title">
        <h2 id="credits-implementation-title">Implementação</h2>
        <p>
          Backend em Go + Chi, frontend em Vue 3 + Vite + TypeScript e persistência em MongoDB.
          Cardápio, carrinho, checkout e acompanhamento usam preços, taxas e status validados no
          servidor.
        </p>
      </section>
    </div>

    <section aria-labelledby="credits-project-title">
      <h2 id="credits-project-title">Projeto</h2>
      <p>
        Este site apresenta o cardápio, o pedido e o acompanhamento da Brocados Lanches. As
        informações de contato publicadas ficam na
        <RouterLink to="/contato">página de contato</RouterLink>.
      </p>
      <p>
        Logotipo, fotografias e textos definitivos aguardam validação; as imagens usadas no
        desenvolvimento são ilustrativas. Esta página está marcada como <code>noindex</code> nos
        metadados da rota.
      </p>
    </section>

    <nav class="credit-actions" aria-label="Navegação de créditos">
      <RouterLink class="credit-action" to="/">Voltar para Home</RouterLink>
      <RouterLink class="credit-action credit-action--outline" to="/cardapio"
        >Ver cardápio</RouterLink
      >
    </nav>
  </main>
</template>

<style scoped>
.info-page {
  display: grid;
  gap: 20px;
}

.info-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(var(--font-size-h1), 7vw, var(--font-size-display));
  line-height: var(--line-height-display);
}

.info-page__notice {
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.info-page__notice strong {
  color: var(--brand-primary);
}

.info-page section,
.credit-card {
  display: grid;
  gap: 8px;
}

.info-page h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.info-page p,
.info-page li {
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.info-page p {
  max-width: 72ch;
}

.info-page ul {
  display: grid;
  gap: 6px;
  margin: 0;
  padding-left: 20px;
}

.info-page a {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  color: var(--brand-neon);
  font-weight: 700;
}

.credit-hero {
  justify-items: center;
  padding: 20px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
  text-align: center;
}

.credit-hero h2 {
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
  color: var(--brand-primary);
}

.credit-hero p:not(.credit-hero__meta) {
  max-width: 60ch;
  margin-inline: auto;
}

.credit-hero__meta {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.credit-links {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
}

.credit-links a {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  padding: 0 16px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  text-decoration: none;
}

.credit-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px;
}

.credit-card {
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.credit-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.credit-action {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  padding: 0 18px;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-weight: 700;
  text-decoration: none;
}

.credit-action--outline {
  border: 1px solid var(--bg-border);
  background: var(--bg-surface);
  color: var(--text-main);
}

.info-page code {
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
}
</style>
