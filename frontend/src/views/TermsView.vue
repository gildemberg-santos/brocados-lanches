<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell info-page">
    <h1>Termos de uso</h1>

    <aside class="info-page__notice" role="note" aria-label="Versão dos termos">
      <strong>Versão vigente.</strong>
      Estes termos descrevem as regras de uso deste site.
      <span class="info-page__date">Publicados em 25/11/2026.</span>
    </aside>

    <section aria-labelledby="terms-service-title">
      <h2 id="terms-service-title">1. Sobre o serviço</h2>
      <p>
        Este site permite montar e registrar pedidos da Brocados Lanches, lanchonete em Aquiraz–CE,
        e acompanhar o andamento pelo código público do pedido. Não há cadastro para o cliente nem
        cobrança online.
      </p>
      <p>
        Ao pedir neste site, você concorda com estas regras na versão publicada. O painel
        administrativo é restrito a usuários autorizados pela loja.
      </p>
    </section>

    <section aria-labelledby="terms-order-title">
      <h2 id="terms-order-title">2. Registro e acompanhamento do pedido</h2>
      <ul>
        <li>
          o pedido só é considerado registrado depois de gravado no servidor, com número sequencial
          e código público opaco de acompanhamento;
        </li>
        <li>
          o link do WhatsApp é gerado depois da gravação. Se o handoff falhar, o pedido continua
          salvo e o link pode ser reenviado na confirmação, na mesma sessão;
        </li>
        <li>
          a chave de idempotência evita que um mesmo envio gere dois pedidos. O acompanhamento
          mostra apenas o status disponível, sem expor nome, telefone ou endereço de terceiros.
        </li>
      </ul>
    </section>

    <section aria-labelledby="terms-price-title">
      <h2 id="terms-price-title">3. Preços, taxas e disponibilidade</h2>
      <ul>
        <li>
          preços, adicionais, taxa de entrega, total e disponibilidade são validados no servidor; a
          cotação não grava pedido e o checkout recalcula os valores antes de gravar;
        </li>
        <li>
          a loja não aceita pedido com horário fechado, interruptor manual desligado, zona inativa,
          produto indisponível ou adicional inválido;
        </li>
        <li>
          depois do registro, nome, preços, adicionais, taxa e total ficam congelados no pedido;
          mudanças posteriores no catálogo não alteram pedidos antigos.
        </li>
      </ul>
      <p>
        Valores exibidos apenas nesta página não são contratuais; o valor cotado no checkout é
        calculado pela loja.
      </p>
    </section>

    <section aria-labelledby="terms-delivery-title">
      <h2 id="terms-delivery-title">4. Entrega, retirada e pagamento</h2>
      <p>
        Na entrega, a taxa é a da zona escolhida. Na retirada, a taxa de entrega é zero. O pagamento
        acontece na entrega ou na retirada, em Pix, cartão ou dinheiro; não há gateway online e
        nenhum dado de cartão é solicitado.
      </p>
      <p>
        No pagamento em dinheiro, você pode informar o valor para troco. A forma de pagamento e o
        troco informado ficam registrados no pedido para o atendimento.
      </p>
    </section>

    <section aria-labelledby="terms-status-title">
      <h2 id="terms-status-title">5. Status, cancelamento e atendimento</h2>
      <p>
        A loja acompanha e muda o status no painel dentro das transições permitidas: pendente,
        confirmado, em preparo, pronto, a caminho, entregue, retirado ou cancelado. Entregue,
        retirado e cancelado são estados finais.
      </p>
      <p>
        Depois do registro, itens e valores não são editados pela loja no painel. Para corrigir ou
        cancelar um pedido, use os canais da
        <RouterLink to="/contato">página de contato</RouterLink>.
      </p>
    </section>

    <section aria-labelledby="terms-admin-title">
      <h2 id="terms-admin-title">6. Acesso ao painel</h2>
      <ul>
        <li>
          o acesso exige sessão válida; escrita anônima é bloqueada e as permissões dependem do
          papel do usuário;
        </li>
        <li>
          credenciais são pessoais e intransferíveis; toda alteração relevante gera auditoria;
        </li>
        <li>
          a loja pode desligar temporariamente os pedidos pelo interruptor manual e ajustar
          horários, sem apagar pedidos já registrados.
        </li>
      </ul>
    </section>

    <section aria-labelledby="terms-availability-title">
      <h2 id="terms-availability-title">7. Disponibilidade do site</h2>
      <p>
        O site depende de conexão e da infraestrutura da loja e pode ficar indisponível
        temporariamente. Um pedido já gravado permanece no servidor mesmo que a página falhe depois.
      </p>
      <p>
        O acompanhamento depende do código correto; um código inválido não revela dados de outros
        pedidos.
      </p>
    </section>

    <section aria-labelledby="terms-content-title">
      <h2 id="terms-content-title">8. Conteúdo e propriedade intelectual</h2>
      <p>
        Cardápio, descrições, preços, imagens, marca e textos exibidos pertencem à Brocados Lanches
        ou a quem os forneceu e são usados aqui apenas para operar o serviço. Imagens de
        desenvolvimento são ilustrativas até a substituição pelas definitivas.
      </p>
      <p>Não reproduza nem use esse conteúdo para outro fim sem autorização da loja.</p>
    </section>

    <section aria-labelledby="terms-privacy-title">
      <h2 id="terms-privacy-title">9. Privacidade</h2>
      <p>
        O tratamento dos dados informados no pedido e no painel segue a
        <RouterLink to="/privacidade">Política de Privacidade</RouterLink>.
      </p>
    </section>

    <section aria-labelledby="terms-changes-title">
      <h2 id="terms-changes-title">10. Alterações, contato e legislação</h2>
      <p>
        A data no aviso do topo indica a versão vigente destes termos, publicada em 25/11/2026.
        Mudanças relevantes passam a valer depois de publicadas neste site.
      </p>
      <p>
        Para dúvidas sobre pedido, entrega ou pagamento, use os canais já publicados pela loja na
        <RouterLink to="/contato">página de contato</RouterLink>. Aplica-se a legislação brasileira,
        inclusive a LGPD (Lei nº 13.709/2018) no que couber.
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
</style>
