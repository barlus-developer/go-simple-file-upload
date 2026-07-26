package file

import (
	"context"
	"errors"
	"io"

	"github.com/barlus-developer/go-simple-file-upload/internal/domain/file"
)

// ErrNotFound is returned when a requested file does not exist in storage.
var ErrNotFound = errors.New("file not found")

// ErrInvalidName is returned when a file name is empty or unsafe to store.
var ErrInvalidName = errors.New("invalid file name")

// Storage is the port implemented by infrastructure adapters that
// persist file bytes. The application layer depends only on this
// interface, never on a concrete storage technology.
type Storage interface {
	Save(ctx context.Context, name string, contentType string, reader io.Reader) (file.File, error)
	Open(ctx context.Context, name string) (io.ReadCloser, file.File, error)
	List(ctx context.Context) ([]file.File, error)
}

// Service is the application-facing use case surface for uploading and
// retrieving files.
type Service interface {
	Upload(ctx context.Context, name string, contentType string, reader io.Reader) (file.File, error)
	Download(ctx context.Context, name string) (io.ReadCloser, file.File, error)
	List(ctx context.Context) ([]file.File, error)
}

type service struct {
	storage Storage
}

// NewService builds a file Service backed by the given Storage adapter.
func NewService(storage Storage) Service {
	return &service{storage: storage}
}

func (s *service) Upload(ctx context.Context, name string, contentType string, reader io.Reader) (file.File, error) {
	if name == "" {
		return file.File{}, ErrInvalidName
	}
	return s.storage.Save(ctx, name, contentType, reader)
}

func (s *service) Download(ctx context.Context, name string) (io.ReadCloser, file.File, error) {
	if name == "" {
		return nil, file.File{}, ErrInvalidName
	}
	return s.storage.Open(ctx, name)
}

func (s *service) List(ctx context.Context) ([]file.File, error) {
	return s.storage.List(ctx)
}
