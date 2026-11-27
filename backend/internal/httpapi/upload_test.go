package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeUploadService struct {
	saved   domain.Upload
	get     domain.Upload
	actorID string
	data    []byte
	err     error
	calls   int
}

func (f *fakeUploadService) Save(_ context.Context, actorID string, data []byte) (domain.Upload, error) {
	f.calls++
	f.actorID = actorID
	f.data = append([]byte(nil), data...)
	return f.saved, f.err
}

func (f *fakeUploadService) Get(_ context.Context, _ string) (domain.Upload, error) {
	f.calls++
	return f.get, f.err
}

func newUploadTestRouter(auth AdminAuthService, uploads UploadService) http.Handler {
	return NewRouter(Dependencies{
		AdminAuth:   auth,
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Uploads:     uploads,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func multipartUploadRequest(t *testing.T, filename string, data []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/uploads", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Origin", "http://example.com")
	request.Host = "example.com"
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
	return request
}

func TestAdminUploadRequiresSessionRoleAndCSRF(t *testing.T) {
	uploads := &fakeUploadService{}
	router := newUploadTestRouter(newAdminOperatorAuth(), uploads)

	anonymous := multipartUploadRequest(t, "burger.png", []byte("png"))
	anonymous.AddCookie(&http.Cookie{Name: "brocados_admin", Value: ""})
	anonymous.Header.Del("Cookie")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, anonymous)
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("anonymous status = %d, want 401", recorder.Code)
	}

	badOrigin := multipartUploadRequest(t, "burger.png", []byte("png"))
	badOrigin.Header.Set("Origin", "https://atacante.example")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, badOrigin)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("foreign Origin status = %d, want 403", recorder.Code)
	}

	if uploads.calls != 0 {
		t.Errorf("Save calls = %d, want 0", uploads.calls)
	}
}

func TestAdminUploadCreatesImage(t *testing.T) {
	id := "665f1c2a9b1e4a0012a3b4f1"
	image := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	uploads := &fakeUploadService{saved: domain.Upload{
		ID: id, ContentType: domain.UploadContentTypePNG, SizeBytes: int64(len(image)),
	}}
	router := newUploadTestRouter(newAdminOperatorAuth(), uploads)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, multipartUploadRequest(t, "burger.png", image))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data adminUploadResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.URL != "/api/v1/uploads/"+id || response.Data.ContentType != domain.UploadContentTypePNG {
		t.Errorf("response = %+v", response.Data)
	}
	if uploads.actorID != testActorIDForHTTP || !bytes.Equal(uploads.data, image) {
		t.Errorf("Save received actor=%q data=%v", uploads.actorID, uploads.data)
	}
}

func TestAdminUploadRejectsMissingAndOversizeFiles(t *testing.T) {
	t.Run("sem arquivo", func(t *testing.T) {
		uploads := &fakeUploadService{}
		router := newUploadTestRouter(newAdminOperatorAuth(), uploads)

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if err := writer.WriteField("outro", "valor"); err != nil {
			t.Fatalf("WriteField() error = %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("writer.Close() error = %v", err)
		}
		request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/uploads", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("Origin", "http://example.com")
		request.Host = "example.com"
		request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), "file") {
			t.Errorf("body = %s, want file detail", recorder.Body.String())
		}
		if uploads.calls != 0 {
			t.Errorf("Save calls = %d, want 0", uploads.calls)
		}
	})

	t.Run("acima do limite", func(t *testing.T) {
		uploads := &fakeUploadService{}
		router := newUploadTestRouter(newAdminOperatorAuth(), uploads)
		data := append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, bytes.Repeat([]byte{0}, domain.MaxUploadBytes)...)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, multipartUploadRequest(t, "grande.png", data))

		if recorder.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, want 422", recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), "file") {
			t.Errorf("body = %s, want file detail", recorder.Body.String())
		}
		if uploads.calls != 0 {
			t.Errorf("Save calls = %d, want 0", uploads.calls)
		}
	})
}

func TestGetUploadServesImage(t *testing.T) {
	image := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	uploads := &fakeUploadService{get: domain.Upload{
		ID: "665f1c2a9b1e4a0012a3b4f1", ContentType: domain.UploadContentTypePNG,
		SizeBytes: int64(len(image)), Data: image,
	}}
	router := newUploadTestRouter(nil, uploads)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/uploads/665f1c2a9b1e4a0012a3b4f1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != domain.UploadContentTypePNG {
		t.Errorf("Content-Type = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("Cache-Control = %q", got)
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q", recorder.Header().Get("X-Content-Type-Options"))
	}
	if !bytes.Equal(recorder.Body.Bytes(), image) {
		t.Errorf("body = %v, want %v", recorder.Body.Bytes(), image)
	}
}

func TestGetUploadRejectsMissingAndUnsafeContent(t *testing.T) {
	tests := []struct {
		name    string
		uploads *fakeUploadService
	}{
		{name: "inexistente", uploads: &fakeUploadService{err: domain.ErrNotFound}},
		{name: "tipo inseguro", uploads: &fakeUploadService{get: domain.Upload{
			ID: "665f1c2a9b1e4a0012a3b4f1", ContentType: "text/html", Data: []byte("<script>"),
		}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := newUploadTestRouter(nil, test.uploads)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/uploads/665f1c2a9b1e4a0012a3b4f1", nil))

			if recorder.Code != http.StatusNotFound {
				t.Errorf("status = %d, want 404", recorder.Code)
			}
			if strings.Contains(recorder.Body.String(), "<script>") {
				t.Error("unsafe content leaked in the response")
			}
		})
	}
}
