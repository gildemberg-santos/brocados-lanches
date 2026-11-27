package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/go-chi/chi/v5"
)

// maxUploadRequestBytes inclui o overhead do multipart além do teto da imagem.
const maxUploadRequestBytes = domain.MaxUploadBytes + (1 << 20)

// UploadService é a dependência de leitura e escrita de imagens do painel.
type UploadService interface {
	Save(ctx context.Context, actorID string, data []byte) (domain.Upload, error)
	Get(ctx context.Context, id string) (domain.Upload, error)
}

type adminUploadResponse struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

func handleAdminCreateUpload(uploads UploadService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if uploads == nil {
			writeUnconfigured(w, r, logger, "upload service")
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadRequestBytes)
		if err := r.ParseMultipartForm(domain.MaxUploadBytes); err != nil {
			writeError(w, r, logger, &domain.ValidationError{
				Field: "file", Code: domain.CodeOutOfRange, Message: "image file exceeds the size limit",
			})
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			writeError(w, r, logger, &domain.ValidationError{
				Field: "file", Code: domain.CodeRequired, Message: "image file is required",
			})
			return
		}
		defer func() {
			_ = file.Close()
		}()

		data, err := io.ReadAll(io.LimitReader(file, domain.MaxUploadBytes+1))
		if err != nil {
			writeError(w, r, logger, &malformedRequestError{field: "file", code: domain.CodeInvalid, message: "could not read image file"})
			return
		}
		if len(data) > domain.MaxUploadBytes {
			writeError(w, r, logger, &domain.ValidationError{
				Field: "file", Code: domain.CodeOutOfRange, Message: "image file exceeds the size limit",
			})
			return
		}

		created, err := uploads.Save(r.Context(), actor.ID, data)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusCreated, adminItemEnvelope[adminUploadResponse]{Data: newAdminUploadResponse(created)})
	}
}

func handleGetUpload(uploads UploadService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if uploads == nil {
			writeUnconfigured(w, r, logger, "upload service")
			return
		}
		record, err := uploads.Get(r.Context(), chi.URLParam(r, "id"))
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		if !domain.IsAllowedUploadContentType(record.ContentType) {
			writeError(w, r, logger, domain.ErrNotFound)
			return
		}

		w.Header().Set("Content-Type", record.ContentType)
		w.Header().Set("Content-Length", strconv.FormatInt(record.SizeBytes, 10))
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(record.Data)
	}
}

func newAdminUploadResponse(upload domain.Upload) adminUploadResponse {
	return adminUploadResponse{
		ID:          upload.ID,
		URL:         uploadPublicPath(upload.ID),
		ContentType: upload.ContentType,
		SizeBytes:   upload.SizeBytes,
	}
}

func uploadPublicPath(id string) string {
	return "/api/v1/uploads/" + id
}
