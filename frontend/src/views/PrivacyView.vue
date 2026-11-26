<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import { clearCheckoutPreferences, hasCheckoutPreferences } from '../features/checkout/preferences'

const preferencesSaved = ref(false)
const preferencesFeedback = ref('')

function forgetPreferences(): void {
  clearCheckoutPreferences()
  preferencesSaved.value = false
  preferencesFeedback.value = 'Preferências apagadas deste navegador.'
}

onMounted(() => {
  preferencesSaved.value = hasCheckoutPreferences()
})
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell info-page">
    <h1>Privacidade</h1>

    <aside class="info-page__notice" role="note" aria-label="Versão da política">
      <strong>Versão vigente.</strong>
      Esta política descreve o tratamento de dados realizado pela Brocados Lanches.
      <span class="info-page__date">Publicada em 25/11/2026. Atualizada em 26/11/2026.</span>
    </aside>

    <section aria-labelledby="privacy-controller-title">
      <h2 id="privacy-controller-title">1. Controlador e abrangência</h2>
      <p>
        Esta política descreve o tratamento de dados feito pela Brocados Lanches, lanchonete em
        Aquiraz–CE, no site de pedidos e no painel administrativo. Ela não se aplica a serviços de
        terceiros que você abre por conta própria, como o WhatsApp.
      </p>
      <p>
        Os canais de atendimento da loja, usados também para solicitações de dados, estão na
        <RouterLink to="/contato">página de contato</RouterLink>.
      </p>
    </section>

    <section aria-labelledby="privacy-data-title">
      <h2 id="privacy-data-title">2. Dados que coletamos</h2>

      <h3>Do cliente</h3>
      <ul>
        <li>nome, telefone, endereço de entrega e observação do pedido, informados no checkout;</li>
        <li>produtos, quantidades, opções e adicionais escolhidos;</li>
        <li>
          forma de pagamento (Pix, cartão ou dinheiro) e, quando pago em dinheiro, o valor informado
          para troco;
        </li>
        <li>
          perfil de conveniência no servidor, identificado pelo telefone: nome, último endereço de
          entrega, entrega/retirada, zona e forma de pagamento usados no pedido. Observação e valor
          de troco não entram nesse perfil.
        </li>
      </ul>
      <p>
        O endereço é solicitado apenas na entrega; na retirada, não é pedido. Não solicitamos CPF,
        e-mail de cliente nem dados de cartão. Nenhum pagamento é processado neste site.
      </p>

      <h3>Do painel administrativo</h3>
      <ul>
        <li>
          nome, e-mail e senha de usuários autorizados; a senha é armazenada somente como hash
          Argon2id;
        </li>
        <li>
          sessão de acesso: um token opaco, guardado no servidor apenas como hash SHA-256 e enviado
          ao navegador em cookie HttpOnly;
        </li>
        <li>
          registros de auditoria das alterações feitas no painel, sem dados pessoais dos clientes.
        </li>
      </ul>

      <h3>Da navegação</h3>
      <ul>
        <li>
          endereço IP, usado somente para limitar requisições e proteger o serviço contra abuso;
        </li>
        <li>
          carrinho salvo no navegador com identificadores e escolhas dos itens; os preços confiáveis
          ficam no servidor.
        </li>
        <li>
          preferências do checkout salvas neste navegador após um pedido concluído: forma de
          entrega, zona, nome, telefone, endereço e forma de pagamento. Observação e valor de troco
          não são salvos. Você pode apagá-las na seção 5.
        </li>
      </ul>
      <p>Não usamos analytics, pixels de rastreamento, mapas de calor nem publicidade.</p>
    </section>

    <section aria-labelledby="privacy-purpose-title">
      <h2 id="privacy-purpose-title">3. Finalidades e bases legais</h2>
      <ul>
        <li>
          execução do pedido: registrar, preparar, entregar ou disponibilizar para retirada, além de
          contatar você sobre o pedido;
        </li>
        <li>
          conveniência: reaproveitar no próximo pedido as escolhas já informadas, no navegador e em
          um perfil no servidor identificado pelo telefone. Esse tratamento pode ser desativado no
          checkout; nesse caso o perfil existente é desativado e seus dados de conveniência são
          removidos;
        </li>
        <li>
          legítimo interesse: segurança, prevenção a abuso e auditoria das alterações no painel;
        </li>
        <li>cumprimento de obrigação legal aplicável à operação da lanchonete.</li>
      </ul>
    </section>

    <section aria-labelledby="privacy-sharing-title">
      <h2 id="privacy-sharing-title">4. Compartilhamento</h2>
      <ul>
        <li>
          equipe da Brocados Lanches: o pedido fica gravado no servidor e aparece no painel para
          preparo e acompanhamento;
        </li>
        <li>
          WhatsApp: ao confirmar, você recebe um link wa.me para o número da loja. A mensagem traz
          número do pedido, itens, observações, nome, telefone e endereço quando há entrega. O envio
          depende de você abrir o link, e o uso do aplicativo segue a política do WhatsApp;
        </li>
        <li>
          infraestrutura de hospedagem e banco de dados necessária para servir o site e armazenar o
          pedido.
        </li>
      </ul>
      <p>
        Não vendemos dados pessoais. Não usamos gateway de pagamento, geocodificação, notificações
        push, inteligência artificial nem serviços de analytics.
      </p>
    </section>

    <section aria-labelledby="privacy-retention-title">
      <h2 id="privacy-retention-title">5. Retenção e eliminação</h2>
      <p>Os prazos abaixo são a proposta do projeto e serão confirmados na versão definitiva:</p>
      <ul>
        <li>
          pedidos: 24 meses; depois disso, nome, telefone, endereço e observação são removidos ou
          mascarados, preservando apenas dados agregados;
        </li>
        <li>
          sessões administrativas: no máximo 7 dias, com expiração por inatividade em 8 horas;
        </li>
        <li>
          perfil de conveniência no servidor: 24 meses, contados do último pedido; depois disso é
          removido automaticamente. Você também pode desativá-lo no checkout ou pedir a exclusão
          pelos canais da loja;
        </li>
        <li>auditoria: 12 meses, sem dados pessoais dos clientes;</li>
        <li>logs: 30 dias, sem nome, telefone, endereço ou observação do pedido;</li>
        <li>backups: 30 dias; ao restaurar, os prazos de retenção são reaplicados.</li>
        <li>
          preferências do checkout: até você apagar nesta página ou limpar os dados deste navegador.
        </li>
      </ul>
      <div class="info-page__preferences">
        <BaseButton variant="secondary" :disabled="!preferencesSaved" @click="forgetPreferences">
          Apagar preferências salvas neste navegador
        </BaseButton>
        <p v-if="preferencesFeedback" class="info-page__feedback" role="status">
          {{ preferencesFeedback }}
        </p>
      </div>
    </section>

    <section aria-labelledby="privacy-rights-title">
      <h2 id="privacy-rights-title">6. Direitos do titular</h2>
      <p>
        Você pode pedir confirmação de tratamento, acesso, correção, portabilidade, anonimização ou
        eliminação dos seus dados, além de informação sobre o compartilhamento e revogação de
        consentimento quando aplicável.
      </p>
      <p>
        O pedido deve ser feito pelos canais já publicados pela loja na
        <RouterLink to="/contato">página de contato</RouterLink>. As solicitações são respondidas
        nos prazos da LGPD (Lei nº 13.709/2018).
      </p>
    </section>

    <section aria-labelledby="privacy-security-title">
      <h2 id="privacy-security-title">7. Segurança</h2>
      <p>
        O site usa TLS, senhas com hash Argon2id, sessões opacas com hash no servidor, cookie
        HttpOnly (Secure em produção), proteção contra CSRF por Origem/Referer e limite de
        requisições. Logs e mensagens de erro não incluem dados pessoais. Nenhum dado de cartão é
        coletado ou armazenado.
      </p>
    </section>

    <section aria-labelledby="privacy-changes-title">
      <h2 id="privacy-changes-title">8. Alterações desta política</h2>
      <p>
        Esta é a versão vigente, publicada em 25/11/2026 e atualizada em 26/11/2026 para descrever
        as preferências do checkout e o perfil de conveniência no servidor. Mudanças relevantes
        serão comunicadas pelos canais da loja nesta página.
      </p>
    </section>
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

.info-page__date {
  display: block;
  margin-top: 8px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.info-page section {
  display: grid;
  gap: 8px;
}

.info-page h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.info-page h3 {
  margin: 8px 0 0;
  font-size: var(--font-size-body);
}

.info-page p,
.info-page li {
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.info-page ul {
  display: grid;
  gap: 6px;
  margin: 0;
  padding-left: 20px;
}

.info-page a {
  color: var(--brand-neon);
  font-weight: 700;
}

.info-page__preferences {
  display: grid;
  justify-items: start;
  gap: 8px;
  margin-top: 8px;
}

.info-page__feedback {
  color: var(--text-main);
}
</style>
