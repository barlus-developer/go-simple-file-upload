package file

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/barlus-developer/go-simple-file-upload/internal/domain/file"
)

type fakeStorage struct {
	saved map[string][]byte
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{saved: map[string][]byte{}}
}

func (f *fakeStorage) Save(_ context.Context, name string, _ string, reader io.Reader) (file.File, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return file.File{}, err
	}
	f.saved[name] = data
	return file.File{Name: name, Size: int64(len(data))}, nil
}

func (f *fakeStorage) Open(_ context.Context, name string) (io.ReadCloser, file.File, error) {
	data, ok := f.saved[name]
	if !ok {
		return nil, file.File{}, ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), file.File{Name: name, Size: int64(len(data))}, nil
}

func (f *fakeStorage) List(_ context.Context) ([]file.File, error) {
	files := make([]file.File, 0, len(f.saved))
	for name, data := range f.saved {
		files = append(files, file.File{Name: name, Size: int64(len(data))})
	}
	return files, nil
}

func TestServiceUploadRejectsEmptyName(t *testing.T) {
	svc := NewService(newFakeStorage())

	if _, err := svc.Upload(context.Background(), "", "text/plain", bytes.NewReader(nil)); err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestServiceUploadThenDownloadRoundTrips(t *testing.T) {
	svc := NewService(newFakeStorage())

	if _, err := svc.Upload(context.Background(), "doc.txt", "text/plain", bytes.NewReader([]byte("content"))); err != nil {
		t.Fatalf("upload: %v", err)
	}

	reader, meta, err := svc.Download(context.Background(), "doc.txt")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer reader.Close()

	if meta.Name != "doc.txt" {
		t.Fatalf("expected name doc.txt, got %q", meta.Name)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "content" {
		t.Fatalf("expected content %q, got %q", "content", string(data))
	}
}

func TestServiceDownloadMissingFileReturnsErrNotFound(t *testing.T) {
	svc := NewService(newFakeStorage())

	if _, _, err := svc.Download(context.Background(), "missing.txt"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestServiceListReturnsUploadedFiles(t *testing.T) {
	svc := NewService(newFakeStorage())

	if _, err := svc.Upload(context.Background(), "a.txt", "text/plain", bytes.NewReader([]byte("a"))); err != nil {
		t.Fatalf("upload: %v", err)
	}

	files, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}
