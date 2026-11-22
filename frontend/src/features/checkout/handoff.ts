const HANDOFF_PREFIX = 'brocados:handoff:'

/** Guarda o link de handoff por pedido para reenvio após recarregar a página. */
export function saveHandoffUrl(publicId: string, url: string): void {
  if (!publicId || !url || typeof sessionStorage === 'undefined') {
    return
  }

  try {
    sessionStorage.setItem(HANDOFF_PREFIX + publicId, url)
  } catch {
    // Armazenamento indisponível não impede a confirmação do pedido.
  }
}

export function readHandoffUrl(publicId: string): string {
  if (!publicId || typeof sessionStorage === 'undefined') {
    return ''
  }

  try {
    return sessionStorage.getItem(HANDOFF_PREFIX + publicId) ?? ''
  } catch {
    return ''
  }
}
