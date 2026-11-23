import { ApiError } from '../../services/http'

/** Erro de API transformado em mensagem operacional, sem expor detalhes internos. */
export function adminCatalogErrorMessage(error: unknown, action: 'load' | 'save'): string {
  const fallback =
    action === 'load'
      ? 'Não foi possível carregar os dados. Tente novamente.'
      : 'Não foi possível salvar. Tente novamente.'

  if (!(error instanceof ApiError)) {
    return fallback
  }
  if (error.status === 401) {
    return 'Sua sessão expirou. Entre novamente no painel.'
  }
  if (error.status === 403) {
    return 'Você não tem permissão para esta ação.'
  }
  if (error.status === 404) {
    return 'Registro não encontrado. Atualize a lista e tente novamente.'
  }
  if (error.status === 409) {
    return 'Já existe um registro com este slug ou ele foi alterado por outra pessoa. Revise e tente novamente.'
  }
  if (error.code === 'VALIDATION_ERROR') {
    const detail = error.details.find((item) => item.message?.trim())
    return detail?.message ?? 'Revise os campos informados e tente novamente.'
  }
  return fallback
}
