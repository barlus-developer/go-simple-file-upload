package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	appfile "github.com/barlus-developer/go-simple-file-upload/internal/application/file"
	"github.com/barlus-developer/go-simple-file-upload/internal/domain/file"
)

// Local is a Storage adapter that persists files on the local disk under
// a single root directory. It is intentionally the only persistence
// mechanism: there is no database involved.
type Local struct {
	dir string
}

// NewLocal creates a Local storage adapter rooted at dir, creating the
// directory if it does not already exist.
func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	return &Local{dir: dir}, nil
}

// safeName strips any directory components from name and rejects names
// that would otherwise escape the storage root (e.g. "..", empty names).
func safeName(name string) (string, error) {
	base := filepath.Base(filepath.Clean(name))
	if base == "" || base == "." || base == string(filepath.Separator) || strings.Contains(base, "..") {
		return "", appfile.ErrInvalidName
	}
	return base, nil
}

func (l *Local) Save(_ context.Context, name string, contentType string, reader io.Reader) (file.File, error) {
	base, err := safeName(name)
	if err != nil {
		return file.File{}, err
	}

	dest := filepath.Join(l.dir, base)
	out, err := os.Create(dest)
	if err != nil {
		return file.File{}, err
	}
	defer out.Close()

	written, err := io.Copy(out, reader)
	if err != nil {
		return file.File{}, err
	}

	info, err := out.Stat()
	if err != nil {
		return file.File{}, err
	}

	return file.File{
		Name:        base,
		Size:        written,
		ContentType: contentType,
		ModifiedAt:  info.ModTime(),
	}, nil
}

func (l *Local) Open(_ context.Context, name string) (io.ReadCloser, file.File, error) {
	base, err := safeName(name)
	if err != nil {
		return nil, file.File{}, err
	}

	path := filepath.Join(l.dir, base)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, file.File{}, appfile.ErrNotFound
		}
		return nil, file.File{}, err
	}
	if info.IsDir() {
		return nil, file.File{}, appfile.ErrNotFound
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, file.File{}, err
	}

	return f, file.File{
		Name:       base,
		Size:       info.Size(),
		ModifiedAt: info.ModTime(),
	}, nil
}

func (l *Local) List(_ context.Context) ([]file.File, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, err
	}

	files := make([]file.File, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, file.File{
			Name:       entry.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	return files, nil
}
