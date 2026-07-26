package storage

import (
	"context"
	"io"
	"strings"
	"testing"

	appfile "github.com/barlus-developer/go-simple-file-upload/internal/application/file"
)

func TestLocalSaveAndOpen(t *testing.T) {
	local, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	saved, err := local.Save(context.Background(), "note.txt", "text/plain", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("save file: %v", err)
	}
	if saved.Name != "note.txt" {
		t.Fatalf("expected saved name note.txt, got %q", saved.Name)
	}
	if saved.Size != 5 {
		t.Fatalf("expected saved size 5, got %d", saved.Size)
	}

	reader, meta, err := local.Open(context.Background(), "note.txt")
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer reader.Close()

	if meta.Size != 5 {
		t.Fatalf("expected opened size 5, got %d", meta.Size)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read file content: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("expected content %q, got %q", "hello", string(content))
	}
}

func TestLocalOpenMissingFileReturnsErrNotFound(t *testing.T) {
	local, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	if _, _, err := local.Open(context.Background(), "missing.txt"); err != appfile.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalSaveRejectsPathTraversal(t *testing.T) {
	local, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	if _, err := local.Save(context.Background(), "..", "text/plain", strings.NewReader("x")); err != appfile.ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestLocalSaveStripsDirectoryComponents(t *testing.T) {
	local, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	saved, err := local.Save(context.Background(), "../../escape.txt", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("save file: %v", err)
	}
	if saved.Name != "escape.txt" {
		t.Fatalf("expected sanitized name escape.txt, got %q", saved.Name)
	}
}

func TestLocalListReturnsSavedFiles(t *testing.T) {
	local, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	if _, err := local.Save(context.Background(), "b.txt", "text/plain", strings.NewReader("b")); err != nil {
		t.Fatalf("save b.txt: %v", err)
	}
	if _, err := local.Save(context.Background(), "a.txt", "text/plain", strings.NewReader("a")); err != nil {
		t.Fatalf("save a.txt: %v", err)
	}

	files, err := local.List(context.Background())
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Name != "a.txt" || files[1].Name != "b.txt" {
		t.Fatalf("expected sorted names [a.txt b.txt], got [%s %s]", files[0].Name, files[1].Name)
	}
}
