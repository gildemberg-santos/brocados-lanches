<script setup lang="ts">
import { ref, watch } from 'vue'

import BaseButton from './BaseButton.vue'
import PriceText from './PriceText.vue'
import { generatePixQrDataUri } from '../features/pix/qrcode'

const props = defineProps<{
  brCode: string
  amountCents: number
  merchantName: string
  merchantCity: string
}>()

const qrDataUri = ref('')
const qrError = ref('')
const copyMessage = ref('')
const codeField = ref<HTMLTextAreaElement | null>(null)

watch(
  () => props.brCode,
  async (brCode) => {
    qrDataUri.value = ''
    qrError.value = ''
    copyMessage.value = ''
    if (!brCode) {
      return
    }
    try {
      const uri = await generatePixQrDataUri(brCode)
      if (brCode === props.brCode) {
        qrDataUri.value = uri
      }
    } catch {
      if (brCode === props.brCode) {
        qrError.value = 'Não foi possível gerar a imagem do QR Code. Use o Pix Copia e Cola abaixo.'
      }
    }
  },
  { immediate: true },
)

async function copyBrCode(): Promise<void> {
  copyMessage.value = ''
  try {
    await navigator.clipboard.writeText(props.brCode)
    copyMessage.value = 'Código Pix copiado. Cole no aplicativo do seu banco.'
  } catch {
    codeField.value?.select()
    copyMessage.value = 'Selecione o código destacado e copie para o aplicativo do seu banco.'
  }
}
</script>

<template>
  <section class="pix-panel" aria-labelledby="pix-panel-title" data-testid="pix-payment-panel">
    <h3 id="pix-panel-title">Pague com Pix</h3>
    <p class="pix-panel__lead">
      QR Code estático de <PriceText :value-cents="amountCents" />, gerado com o total gravado no
      pedido. A imagem representa o mesmo código textual abaixo.
    </p>

    <div class="pix-panel__content">
      <div class="pix-panel__qr">
        <img
          v-if="qrDataUri"
          :src="qrDataUri"
          width="360"
          height="360"
          alt="QR Code do Pix para pagamento do pedido"
        />
        <p v-else-if="qrError" class="pix-panel__error" role="alert">{{ qrError }}</p>
        <p v-else class="pix-panel__hint" role="status">Gerando QR Code…</p>
      </div>

      <div class="pix-panel__code">
        <label for="pix-brcode">Pix Copia e Cola</label>
        <textarea
          id="pix-brcode"
          ref="codeField"
          :value="brCode"
          rows="4"
          readonly
          data-testid="pix-brcode"
        ></textarea>
        <BaseButton type="button" @click="copyBrCode">Copiar código Pix</BaseButton>
        <p class="pix-panel__hint" aria-live="polite">{{ copyMessage }}</p>
      </div>
    </div>

    <p class="pix-panel__instructions">
      Recebedor: <strong>{{ merchantName }}</strong> · {{ merchantCity }}. Após pagar, envie o
      comprovante pelo WhatsApp desta confirmação. Este QR não confirma o pagamento automaticamente.
    </p>
  </section>
</template>

<style scoped>
.pix-panel {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--brand-primary);
  border-radius: 16px;
  background: var(--bg-surface);
}

.pix-panel h3 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.pix-panel__lead,
.pix-panel__hint,
.pix-panel__instructions {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.pix-panel__error {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.pix-panel__content {
  display: grid;
  gap: 16px;
  align-items: start;
}

.pix-panel__qr {
  display: grid;
  gap: 8px;
  justify-items: center;
}

.pix-panel__qr img {
  width: 100%;
  max-width: 240px;
  height: auto;
  border-radius: 12px;
}

.pix-panel__code {
  display: grid;
  gap: 8px;
}

.pix-panel__code label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.pix-panel__code textarea {
  width: 100%;
  min-height: 96px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 10px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  word-break: break-all;
}

@media (min-width: 640px) {
  .pix-panel__content {
    grid-template-columns: minmax(200px, 240px) minmax(0, 1fr);
  }
}
</style>
