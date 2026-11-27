package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeUploadRepository struct {
	stored domain.Upload
	audit  domain.AuditLog
	get    domain.Upload
	err    error
	calls  int
}

func (f *fakeUploadRepository) CreateUpload(_ context.Context, upload domain.Upload, audit domain.AuditLog) error {
	f.calls++
	f.stored = upload
	f.audit = audit
	return f.err
}

func (f *fakeUploadRepository) GetUpload(_ context.Context, _ string) (domain.Upload, error) {
	f.calls++
	return f.get, f.err
}

func pngBytes() []byte {
	return []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
}

func TestNewUploadServiceRequiresRepository(t *testing.T) {
	if _, err := NewUploadService(nil); err == nil {
		t.Fatal("NewUploadService(nil) error = nil, want error")
	}
}

func TestUploadServiceSave(t *testing.T) {
	repo := &fakeUploadRepository{}
	service, err := NewUploadService(repo)
	if err != nil {
		t.Fatalf("NewUploadService() error = %v", err)
	}
	service.now = func() time.Time { return time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC) }
	service.newID = func() (string, error) { return strings.Repeat("a", 24), nil }

	data := pngBytes()
	created, err := service.Save(context.Background(), "actor-1", data)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if created.ID != "aaaaaaaaaaaaaaaaaaaaaaaa" || created.ContentType != domain.UploadContentTypePNG {
		t.Errorf("Save() = %+v", created)
	}
	if created.SizeBytes != int64(len(data)) || !bytes.Equal(created.Data, data) {
		t.Errorf("Save() stored %d bytes / %v", created.SizeBytes, created.Data)
	}
	if repo.calls != 1 {
		t.Fatalf("CreateUpload calls = %d, want 1", repo.calls)
	}
	audit := repo.audit
	if audit.ActorID != "actor-1" || audit.Action != domain.AuditActionCreate ||
		audit.Entity != domain.AuditEntityUpload || audit.EntityID != created.ID {
		t.Errorf("audit = %+v", audit)
	}
	if audit.Before != nil || audit.After["contentType"] != domain.UploadContentTypePNG ||
		audit.After["sizeBytes"] != int64(len(data)) {
		t.Errorf("audit snapshots = %+v / %+v", audit.Before, audit.After)
	}
	if !audit.CreatedAt.Equal(created.CreatedAt) || audit.CreatedAt.Location() != time.UTC {
		t.Errorf("timestamps = %v / %v", audit.CreatedAt, created.CreatedAt)
	}

	data[0] = 0
	if repo.stored.Data[0] != 0x89 {
		t.Error("Save() must copy the bytes instead of keeping the caller slice")
	}
}

func TestUploadServiceSaveRejectsInvalidFiles(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantCode string
	}{
		{name: "vazio", data: nil, wantCode: domain.CodeRequired},
		{name: "muito grande", data: append(pngBytes(), bytes.Repeat([]byte{0}, domain.MaxUploadBytes)...), wantCode: domain.CodeOutOfRange},
		{name: "html", data: []byte("<script>alert(1)</script>"), wantCode: domain.CodeInvalid},
		{name: "gif", data: []byte("GIF89a0000"), wantCode: domain.CodeInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeUploadRepository{}
			service, err := NewUploadService(repo)
			if err != nil {
				t.Fatalf("NewUploadService() error = %v", err)
			}

			_, err = service.Save(context.Background(), "actor-1", test.data)
			var validation *domain.ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("Save() error = %v, want ValidationError", err)
			}
			if validation.Field != "file" || validation.Code != test.wantCode {
				t.Errorf("validation = %+v, want field=file code=%s", validation, test.wantCode)
			}
			if repo.calls != 0 {
				t.Errorf("CreateUpload calls = %d, want 0", repo.calls)
			}
		})
	}
}

func TestUploadServiceGet(t *testing.T) {
	id := "665f1c2a9b1e4a0012a3b4f1"
	want := domain.Upload{ID: id, ContentType: domain.UploadContentTypeWebP, Data: []byte("imagem")}
	repo := &fakeUploadRepository{get: want}
	service, err := NewUploadService(repo)
	if err != nil {
		t.Fatalf("NewUploadService() error = %v", err)
	}

	got, err := service.Get(context.Background(), id)
	if err != nil || got.ID != id {
		t.Fatalf("Get() = %+v, %v", got, err)
	}

	if _, err := service.Get(context.Background(), "curto"); err == nil {
		t.Error("Get() with malformed id error = nil, want error")
	}

	repo.err = domain.ErrNotFound
	if _, err := service.Get(context.Background(), id); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get() error = %v, want domain.ErrNotFound", err)
	}
}

func TestDetectImageContentType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{name: "jpeg", data: []byte{0xff, 0xd8, 0xff, 0xe0}, want: domain.UploadContentTypeJPEG},
		{name: "png", data: pngBytes(), want: domain.UploadContentTypePNG},
		{name: "webp", data: []byte("RIFF0000WEBPVP8 "), want: domain.UploadContentTypeWebP},
		{name: "avif", data: []byte{0, 0, 0, 0x18, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f'}, want: domain.UploadContentTypeAVIF},
		{name: "avis", data: []byte{0, 0, 0, 0x18, 'f', 't', 'y', 'p', 'a', 'v', 'i', 's'}, want: domain.UploadContentTypeAVIF},
		{name: "texto", data: []byte("hello"), want: ""},
		{name: "vazio", data: nil, want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := detectImageContentType(test.data); got != test.want {
				t.Errorf("detectImageContentType() = %q, want %q", got, test.want)
			}
		})
	}
}
