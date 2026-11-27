package domain

// OrderHistoryQuery descreve a página do histórico do cliente. A ordenação é
// fixa (mais recente primeiro) para não aceitar campo arbitrário do cliente.
type OrderHistoryQuery struct {
	Page     int
	PageSize int
}
