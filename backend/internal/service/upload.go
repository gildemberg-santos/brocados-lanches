package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

// UploadRepository persiste imagens enviadas pelo painel e sua auditoria.
type UploadRepository interface {
	CreateUpload(ctx context.Context, upload domain.Upload, audit domain.AuditLog) error
	GetUpload(ctx context.Context, id string) (domain.Upload, error)
}

// UploadService valida o conteúdo da imagem antes de gravar e expor a URL.
type UploadService struct {
	repo  UploadRepository
	now   func() time.Time
	newID func() (string, error)
}

// NewUploadService valida as dependências do serviço de imagens.
func NewUploadService(repo UploadRepository) (*UploadService, error) {
	if repo == nil {
		return nil, fmt.Errorf("upload repository is required")
	}
	return &UploadService{
		repo:  repo,
		now:   time.Now,
		newID: newObjectIDHex,
	}, nil
}

// Save grava a imagem e registra auditoria sem incluir bytes nem nome de arquivo.
func (s *UploadService) Save(ctx context.Context, actorID string, data []byte) (domain.Upload, error) {
	if len(data) == 0 {
		return domain.Upload{}, invalidField("file", domain.CodeRequired, "image file is required")
	}
	if len(data) > domain.MaxUploadBytes {
		return domain.Upload{}, invalidField("file", domain.CodeOutOfRange, "image file exceeds the size limit")
	}

	contentType := detectImageContentType(data)
	if !domain.IsAllowedUploadContentType(contentType) {
		return domain.Upload{}, invalidField("file", domain.CodeInvalid, "only JPEG, PNG, WebP and AVIF images are supported")
	}

	id, err := s.newID()
	if err != nil {
		return domain.Upload{}, fmt.Errorf("generate upload id: %w", err)
	}
	now := s.now().UTC()
	upload := domain.Upload{
		ID:          id,
		ContentType: contentType,
		SizeBytes:   int64(len(data)),
		Data:        append([]byte(nil), data...),
		CreatedAt:   now,
	}
	audit := domain.AuditLog{
		ActorID:   actorID,
		Action:    domain.AuditActionCreate,
		Entity:    domain.AuditEntityUpload,
		EntityID:  id,
		After:     map[string]any{"contentType": contentType, "sizeBytes": upload.SizeBytes},
		CreatedAt: now,
	}
	if err := s.repo.CreateUpload(ctx, upload, audit); err != nil {
		return domain.Upload{}, err
	}
	return upload, nil
}

// Get devolve a imagem por identificador hexadecimal.
func (s *UploadService) Get(ctx context.Context, id string) (domain.Upload, error) {
	if err := validateObjectID("id", id); err != nil {
		return domain.Upload{}, err
	}
	return s.repo.GetUpload(ctx, id)
}

// detectImageContentType usa a assinatura do arquivo, nunca o Content-Type
// informado pelo cliente, para decidir o tipo servido ao navegador.
func detectImageContentType(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		return domain.UploadContentTypeJPEG
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return domain.UploadContentTypePNG
	case len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return domain.UploadContentTypeWebP
	case len(data) >= 12 && bytes.Equal(data[4:8], []byte("ftyp")) &&
		(bytes.Equal(data[8:12], []byte("avif")) || bytes.Equal(data[8:12], []byte("avis"))):
		return domain.UploadContentTypeAVIF
	default:
		return ""
	}
}
