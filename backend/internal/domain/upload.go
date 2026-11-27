package domain

import "time"

// MaxUploadBytes é o teto de uma imagem enviada pelo painel (2 MiB).
const MaxUploadBytes = 2 << 20

// Tipos de imagem aceitos no upload. A leitura pública usa apenas estes valores.
const (
	UploadContentTypeJPEG = "image/jpeg"
	UploadContentTypePNG  = "image/png"
	UploadContentTypeWebP = "image/webp"
	UploadContentTypeAVIF = "image/avif"
)

// Upload é uma imagem enviada pelo painel e servida por identificador opaco.
type Upload struct {
	ID          string
	ContentType string
	SizeBytes   int64
	Data        []byte
	CreatedAt   time.Time
}

// IsAllowedUploadContentType impede que um documento alterado fora da API sirva
// HTML, SVG ou qualquer outro tipo executável pelo navegador.
func IsAllowedUploadContentType(value string) bool {
	switch value {
	case UploadContentTypeJPEG, UploadContentTypePNG, UploadContentTypeWebP, UploadContentTypeAVIF:
		return true
	default:
		return false
	}
}
