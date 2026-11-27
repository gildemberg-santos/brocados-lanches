<script setup lang="ts">
import { ref, watch } from 'vue'

import BaseButton from '../BaseButton.vue'
import {
  allergensFromText,
  allergensToText,
  moneyToCents,
  newLocalId,
  slugify,
  validateProductInput,
} from '../../features/admin-catalog/catalogForm'
import type { FieldErrors } from '../../features/admin-catalog/catalogForm'
import { IMAGE_UPLOAD_ACCEPT, validateImageFile } from '../../features/admin-catalog/imageUpload'
import { uploadProductImage } from '../../services/adminCatalog'
import type {
  AdminAddon,
  AdminCategory,
  AdminOptionGroup,
  AdminProduct,
  AdminProductInput,
} from '../../types/adminCatalog'

const props = defineProps<{
  item: AdminProduct | null
  categories: AdminCategory[]
  addons: AdminAddon[]
  saving: boolean
  error: string
  success: string
}>()

const emit = defineEmits<{ submit: [value: AdminProductInput]; cancel: [] }>()

const categoryId = ref('')
const name = ref('')
const slug = ref('')
const description = ref('')
const price = ref('')
const imageUrl = ref('')
const imageAlt = ref('')
const uploadingImage = ref(false)
const imageUploadError = ref('')
const imageUploadSuccess = ref('')
const allergensText = ref('')
const available = ref(true)
const featured = ref(false)
const active = ref(true)
const sortOrder = ref(0)
const addonIds = ref<string[]>([])
const optionGroups = ref<AdminOptionGroup[]>([])
const optionAmounts = ref<string[][]>([])
const errors = ref<FieldErrors>({})

function errorId(key: string): string {
  return `product-${key.replace(/\./g, '-')}-error`
}

function errorFor(key: string): string {
  return errors.value[key] ?? ''
}

function errorAttrs(key: string) {
  const message = errorFor(key)
  return {
    'aria-invalid': message ? ('true' as const) : undefined,
    'aria-describedby': message ? errorId(key) : undefined,
  }
}

async function onImageChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }

  const problem = validateImageFile(file)
  imageUploadSuccess.value = ''
  if (problem) {
    imageUploadError.value = problem
    input.value = ''
    return
  }

  uploadingImage.value = true
  try {
    const uploaded = await uploadProductImage(file)
    imageUrl.value = uploaded.url
    imageUploadError.value = ''
    imageUploadSuccess.value = 'Imagem enviada. Descreva o conteúdo dela no texto alternativo.'
  } catch {
    imageUploadError.value = 'Não foi possível enviar a imagem. Tente novamente.'
  } finally {
    uploadingImage.value = false
    input.value = ''
  }
}

function centsToInput(cents: number): string {
  return (cents / 100).toFixed(2).replace('.', ',')
}

watch(
  () => props.item,
  (item) => {
    errors.value = {}
    imageUploadError.value = ''
    imageUploadSuccess.value = ''
    categoryId.value = item?.categoryId ?? ''
    name.value = item?.name ?? ''
    slug.value = item?.slug ?? ''
    description.value = item?.description ?? ''
    price.value = centsToInput(item?.priceCents ?? 0)
    imageUrl.value = item?.imageUrl ?? ''
    imageAlt.value = item?.imageAlt ?? ''
    allergensText.value = allergensToText(item?.allergens ?? [])
    available.value = item?.available ?? true
    featured.value = item?.featured ?? false
    active.value = item?.active ?? true
    sortOrder.value = item?.sortOrder ?? 0
    addonIds.value = [...(item?.addonIds ?? [])]
    optionGroups.value = (item?.optionGroups ?? []).map((group) => ({
      ...group,
      options: group.options.map((option) => ({ ...option })),
    }))
    optionAmounts.value = optionGroups.value.map((group) =>
      group.options.map((option) => centsToInput(option.priceCents)),
    )
  },
  { immediate: true },
)

function addGroup(): void {
  optionGroups.value.push({
    id: newLocalId('grupo'),
    name: '',
    required: false,
    minSelections: 0,
    maxSelections: 0,
    options: [],
  })
  optionAmounts.value.push([])
}

function removeGroup(index: number): void {
  optionGroups.value.splice(index, 1)
  optionAmounts.value.splice(index, 1)
}

function addOption(groupIndex: number): void {
  const group = optionGroups.value[groupIndex]
  if (!group) {
    return
  }
  group.options.push({
    id: newLocalId('opcao'),
    name: '',
    priceCents: 0,
    available: true,
    sortOrder: group.options.length,
  })
  const amounts = optionAmounts.value[groupIndex]
  if (amounts) {
    amounts.push('0,00')
  }
}

function removeOption(groupIndex: number, optionIndex: number): void {
  optionGroups.value[groupIndex]?.options.splice(optionIndex, 1)
  optionAmounts.value[groupIndex]?.splice(optionIndex, 1)
}

function toggleAddon(id: string): void {
  addonIds.value = addonIds.value.includes(id)
    ? addonIds.value.filter((item) => item !== id)
    : [...addonIds.value, id]
}

function submit(): void {
  const priceCents = moneyToCents(price.value)
  const optionCents = optionGroups.value.map((group, groupIndex) =>
    group.options.map((_, optionIndex) =>
      moneyToCents(optionAmounts.value[groupIndex]?.[optionIndex] ?? ''),
    ),
  )
  const groups = optionGroups.value.map((group, groupIndex) => ({
    ...group,
    options: group.options.map((option, optionIndex) => ({
      ...option,
      priceCents: optionCents[groupIndex]?.[optionIndex] ?? 0,
    })),
  }))

  const input: AdminProductInput = {
    categoryId: categoryId.value,
    name: name.value.trim(),
    slug: slug.value.trim(),
    description: description.value.trim(),
    priceCents: priceCents ?? 0,
    imageUrl: imageUrl.value.trim(),
    imageAlt: imageAlt.value.trim(),
    allergens: allergensFromText(allergensText.value),
    available: available.value,
    featured: featured.value,
    active: active.value,
    sortOrder: Number(sortOrder.value),
    optionGroups: groups,
    addonIds: addonIds.value,
  }

  errors.value = validateProductInput(input)
  if (priceCents === null) {
    errors.value.priceCents = 'Informe um preço válido em reais.'
  }
  optionCents.forEach((amounts, groupIndex) => {
    amounts.forEach((cents, optionIndex) => {
      if (cents === null) {
        errors.value[`optionGroups.${groupIndex}.options.${optionIndex}.priceCents`] =
          'Informe um preço válido em reais.'
      }
    })
  })
  if (Object.keys(errors.value).length > 0) {
    return
  }
  emit('submit', input)
}
</script>

<template>
  <form class="product-form" aria-label="Formulário de produto" @submit.prevent="submit">
    <h2>{{ props.item ? 'Editar produto' : 'Novo produto' }}</h2>

    <div class="product-form__grid">
      <div class="product-form__field">
        <label for="product-category">Categoria</label>
        <select id="product-category" v-model="categoryId" v-bind="errorAttrs('categoryId')">
          <option value="">Selecione</option>
          <option v-for="category in props.categories" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
        <p
          v-if="errorFor('categoryId')"
          :id="errorId('categoryId')"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor('categoryId') }}
        </p>
      </div>

      <div class="product-form__field">
        <label for="product-name">Nome</label>
        <input
          id="product-name"
          v-model="name"
          type="text"
          maxlength="120"
          v-bind="errorAttrs('name')"
        />
        <p v-if="errorFor('name')" :id="errorId('name')" class="product-form__error" role="alert">
          {{ errorFor('name') }}
        </p>
      </div>

      <div class="product-form__field">
        <label for="product-slug">Slug</label>
        <div class="product-form__inline">
          <input
            id="product-slug"
            v-model="slug"
            type="text"
            maxlength="80"
            v-bind="errorAttrs('slug')"
          />
          <button type="button" class="product-form__small" @click="slug = slugify(name)">
            Gerar
          </button>
        </div>
        <p v-if="errorFor('slug')" :id="errorId('slug')" class="product-form__error" role="alert">
          {{ errorFor('slug') }}
        </p>
      </div>

      <div class="product-form__field">
        <label for="product-price">Preço (R$)</label>
        <input
          id="product-price"
          v-model="price"
          type="text"
          inputmode="decimal"
          placeholder="0,00"
          v-bind="errorAttrs('priceCents')"
        />
        <p
          v-if="errorFor('priceCents')"
          :id="errorId('priceCents')"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor('priceCents') }}
        </p>
      </div>

      <div class="product-form__field product-form__field--wide">
        <label for="product-description">Descrição</label>
        <textarea
          id="product-description"
          v-model="description"
          rows="3"
          maxlength="2000"
          v-bind="errorAttrs('description')"
        />
        <p
          v-if="errorFor('description')"
          :id="errorId('description')"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor('description') }}
        </p>
      </div>

      <div class="product-form__field">
        <label for="product-image">URL ou caminho da imagem</label>
        <input id="product-image" v-model="imageUrl" type="text" maxlength="500" />
        <p class="product-form__hint">Caminhos como /products/burger.png também são aceitos.</p>
        <img
          v-if="imageUrl"
          class="product-form__preview"
          :src="imageUrl"
          :alt="imageAlt || 'Prévia da imagem do produto'"
          width="160"
          height="120"
          loading="lazy"
        />
      </div>

      <div class="product-form__field">
        <label for="product-image-file">Enviar imagem</label>
        <input
          id="product-image-file"
          type="file"
          :accept="IMAGE_UPLOAD_ACCEPT"
          :disabled="uploadingImage"
          :aria-invalid="imageUploadError ? 'true' : undefined"
          :aria-describedby="imageUploadError ? 'product-image-file-error' : undefined"
          @change="onImageChange"
        />
        <p class="product-form__hint">
          JPEG, PNG, WebP ou AVIF de até 2 MB. O arquivo é salvo no servidor.
        </p>
        <p
          v-if="imageUploadError"
          id="product-image-file-error"
          class="product-form__error"
          role="alert"
        >
          {{ imageUploadError }}
        </p>
        <p v-else-if="imageUploadSuccess" class="product-form__message" role="status">
          {{ imageUploadSuccess }}
        </p>
        <p v-if="uploadingImage" class="product-form__hint" role="status">Enviando imagem...</p>
      </div>

      <div class="product-form__field">
        <label for="product-image-alt">Texto alternativo</label>
        <input
          id="product-image-alt"
          v-model="imageAlt"
          type="text"
          maxlength="200"
          v-bind="errorAttrs('imageAlt')"
        />
        <p
          v-if="errorFor('imageAlt')"
          :id="errorId('imageAlt')"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor('imageAlt') }}
        </p>
      </div>

      <div class="product-form__field">
        <label for="product-allergens">Alergênicos (separados por vírgula)</label>
        <input
          id="product-allergens"
          v-model="allergensText"
          type="text"
          v-bind="errorAttrs('allergens')"
        />
        <p
          v-if="errorFor('allergens')"
          :id="errorId('allergens')"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor('allergens') }}
        </p>
      </div>

      <div class="product-form__field">
        <label for="product-sort">Ordem</label>
        <input
          id="product-sort"
          v-model.number="sortOrder"
          type="number"
          min="0"
          v-bind="errorAttrs('sortOrder')"
        />
        <p
          v-if="errorFor('sortOrder')"
          :id="errorId('sortOrder')"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor('sortOrder') }}
        </p>
      </div>
    </div>

    <div class="product-form__checks">
      <label><input v-model="available" type="checkbox" /> Disponível</label>
      <label><input v-model="featured" type="checkbox" /> Destaque</label>
      <label><input v-model="active" type="checkbox" /> Ativo</label>
    </div>

    <fieldset
      class="product-form__section"
      :aria-describedby="errorFor('addonIds') ? errorId('addonIds') : undefined"
    >
      <legend>Adicionais permitidos</legend>
      <label v-for="addon in props.addons" :key="addon.id" class="product-form__check">
        <input
          type="checkbox"
          :checked="addonIds.includes(addon.id)"
          @change="toggleAddon(addon.id)"
        />
        {{ addon.name }}
      </label>
      <p v-if="props.addons.length === 0" class="product-form__hint">
        Cadastre adicionais antes de vincular.
      </p>
      <p
        v-if="errorFor('addonIds')"
        :id="errorId('addonIds')"
        class="product-form__error"
        role="alert"
      >
        {{ errorFor('addonIds') }}
      </p>
    </fieldset>

    <fieldset class="product-form__section">
      <legend>Grupos de opções</legend>
      <div v-for="(group, groupIndex) in optionGroups" :key="group.id" class="product-form__group">
        <div class="product-form__group-head">
          <input
            :id="`product-group-${groupIndex}-name`"
            v-model="group.name"
            type="text"
            placeholder="Nome do grupo"
            maxlength="120"
            v-bind="errorAttrs(`optionGroups.${groupIndex}.name`)"
          />
          <button type="button" class="product-form__small" @click="removeGroup(groupIndex)">
            Remover grupo
          </button>
        </div>
        <p
          v-if="errorFor(`optionGroups.${groupIndex}.name`)"
          :id="errorId(`optionGroups.${groupIndex}.name`)"
          class="product-form__error"
          role="alert"
        >
          {{ errorFor(`optionGroups.${groupIndex}.name`) }}
        </p>

        <div class="product-form__inline">
          <label>
            <input v-model="group.required" type="checkbox" />
            Obrigatório
          </label>
          <label>
            Mínimo
            <input
              :id="`product-group-${groupIndex}-min`"
              v-model.number="group.minSelections"
              type="number"
              min="0"
              v-bind="errorAttrs(`optionGroups.${groupIndex}.minSelections`)"
            />
            <span
              v-if="errorFor(`optionGroups.${groupIndex}.minSelections`)"
              :id="errorId(`optionGroups.${groupIndex}.minSelections`)"
              class="product-form__error"
              role="alert"
            >
              {{ errorFor(`optionGroups.${groupIndex}.minSelections`) }}
            </span>
          </label>
          <label>
            Máximo
            <input
              :id="`product-group-${groupIndex}-max`"
              v-model.number="group.maxSelections"
              type="number"
              min="0"
              v-bind="errorAttrs(`optionGroups.${groupIndex}.maxSelections`)"
            />
            <span
              v-if="errorFor(`optionGroups.${groupIndex}.maxSelections`)"
              :id="errorId(`optionGroups.${groupIndex}.maxSelections`)"
              class="product-form__error"
              role="alert"
            >
              {{ errorFor(`optionGroups.${groupIndex}.maxSelections`) }}
            </span>
          </label>
        </div>

        <div
          v-for="(option, optionIndex) in group.options"
          :key="option.id"
          class="product-form__option"
        >
          <input
            :id="`product-option-${groupIndex}-${optionIndex}-name`"
            v-model="option.name"
            type="text"
            placeholder="Nome da opção"
            maxlength="120"
            v-bind="errorAttrs(`optionGroups.${groupIndex}.options.${optionIndex}.name`)"
          />
          <input
            :id="`product-option-${groupIndex}-${optionIndex}-price`"
            v-model="optionAmounts[groupIndex][optionIndex]"
            type="text"
            inputmode="decimal"
            placeholder="0,00"
            :aria-label="`Preço de ${option.name || 'opção'}`"
            v-bind="errorAttrs(`optionGroups.${groupIndex}.options.${optionIndex}.priceCents`)"
          />
          <label><input v-model="option.available" type="checkbox" /> Disponível</label>
          <button
            type="button"
            class="product-form__small"
            @click="removeOption(groupIndex, optionIndex)"
          >
            Remover
          </button>
          <p
            v-if="errorFor(`optionGroups.${groupIndex}.options.${optionIndex}.name`)"
            :id="errorId(`optionGroups.${groupIndex}.options.${optionIndex}.name`)"
            class="product-form__error"
            role="alert"
          >
            {{ errorFor(`optionGroups.${groupIndex}.options.${optionIndex}.name`) }}
          </p>
          <p
            v-if="errorFor(`optionGroups.${groupIndex}.options.${optionIndex}.priceCents`)"
            :id="errorId(`optionGroups.${groupIndex}.options.${optionIndex}.priceCents`)"
            class="product-form__error"
            role="alert"
          >
            {{ errorFor(`optionGroups.${groupIndex}.options.${optionIndex}.priceCents`) }}
          </p>
        </div>

        <button type="button" class="product-form__small" @click="addOption(groupIndex)">
          Adicionar opção
        </button>
      </div>

      <button type="button" class="product-form__small" @click="addGroup">Adicionar grupo</button>
    </fieldset>

    <p v-if="props.error" class="product-form__message" role="alert">{{ props.error }}</p>
    <p v-if="props.success" class="product-form__message" role="status">{{ props.success }}</p>

    <div class="product-form__actions">
      <BaseButton type="submit" :loading="props.saving" :disabled="props.saving">Salvar</BaseButton>
      <BaseButton type="button" variant="ghost" @click="emit('cancel')">Cancelar</BaseButton>
    </div>
  </form>
</template>

<style scoped>
.product-form {
  display: grid;
  gap: 18px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.product-form h2 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
}

.product-form__grid {
  display: grid;
  gap: 14px;
}

.product-form__field {
  display: grid;
  gap: 8px;
}

.product-form__field label,
.product-form__check {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.product-form input,
.product-form select,
.product-form textarea,
.product-form__small {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.product-form__small {
  width: fit-content;
  font-weight: 700;
}

.product-form input[type='file']::file-selector-button {
  padding: 8px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--text-main);
  font-family: var(--font-sans);
  cursor: pointer;
}

.product-form__preview {
  width: 160px;
  height: 120px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  object-fit: cover;
}

.product-form__inline,
.product-form__checks {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.product-form__checks label,
.product-form__check {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
}

.product-form__section {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
}

.product-form__section legend {
  padding: 0 8px;
  font-weight: 800;
}

.product-form__group {
  display: grid;
  gap: 10px;
  padding: 14px;
  border: 1px solid var(--bg-border);
  border-radius: 14px;
}

.product-form__group-head,
.product-form__option {
  display: grid;
  gap: 8px;
}

.product-form__error {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.product-form__hint,
.product-form__message {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.product-form__message {
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  color: var(--text-main);
}

.product-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (min-width: 768px) {
  .product-form__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .product-form__field--wide {
    grid-column: 1 / -1;
  }

  .product-form__group-head,
  .product-form__option {
    grid-template-columns: 2fr 1fr auto auto;
    align-items: center;
  }

  .product-form__option .product-form__error {
    grid-column: 1 / -1;
  }
}
</style>
