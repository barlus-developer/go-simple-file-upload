package router

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	appfile "github.com/barlus-developer/go-simple-file-upload/internal/application/file"
	apphealth "github.com/barlus-developer/go-simple-file-upload/internal/application/health"
	"github.com/barlus-developer/go-simple-file-upload/internal/infrastructure/config"
	"github.com/barlus-developer/go-simple-file-upload/internal/infrastructure/storage"
	"github.com/barlus-developer/go-simple-file-upload/internal/interfaces/http/handler"
	"go.uber.org/zap/zaptest"
)

func newTestEngine(t *testing.T) http.Handler {
	t.Helper()

	localStorage, err := storage.NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	return New(
		config.Config{
			App:     config.AppConfig{Environment: "test"},
			Storage: config.StorageConfig{MaxUploadSizeMB: 32},
		},
		zaptest.NewLogger(t),
		handler.NewHealthHandler(apphealth.NewService()),
		handler.NewFileHandler(appfile.NewService(localStorage)),
	)
}

func TestRootEndpointReturnsHealthStatus(t *testing.T) {
	engine := newTestEngine(t)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected body status ok, got %q", body["status"])
	}
	if body["message"] != "Hello, World!!!" {
		t.Fatalf("expected hello message, got %q", body["message"])
	}
}

func TestUnknownEndpointReturnsNotFound(t *testing.T) {
	engine := newTestEngine(t)

	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, response.Code)
	}
}

func uploadRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/files", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request
}

func TestUploadThenDownloadFile(t *testing.T) {
	engine := newTestEngine(t)
	content := []byte("hello file upload")

	uploadResponse := httptest.NewRecorder()
	engine.ServeHTTP(uploadResponse, uploadRequest(t, "greeting.txt", content))

	if uploadResponse.Code != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d: %s", http.StatusCreated, uploadResponse.Code, uploadResponse.Body.String())
	}

	var uploaded map[string]any
	if err := json.Unmarshal(uploadResponse.Body.Bytes(), &uploaded); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if uploaded["name"] != "greeting.txt" {
		t.Fatalf("expected uploaded name greeting.txt, got %v", uploaded["name"])
	}

	downloadRequest := httptest.NewRequest(http.MethodGet, "/files/greeting.txt", nil)
	downloadResponse := httptest.NewRecorder()
	engine.ServeHTTP(downloadResponse, downloadRequest)

	if downloadResponse.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, downloadResponse.Code)
	}

	got, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		t.Fatalf("read download body: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("expected downloaded content %q, got %q", content, got)
	}
}

func TestDownloadMissingFileReturnsNotFound(t *testing.T) {
	engine := newTestEngine(t)

	request := httptest.NewRequest(http.MethodGet, "/files/missing.txt", nil)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestListFilesReturnsUploadedFiles(t *testing.T) {
	engine := newTestEngine(t)

	uploadResponse := httptest.NewRecorder()
	engine.ServeHTTP(uploadResponse, uploadRequest(t, "notes.txt", []byte("notes")))
	if uploadResponse.Code != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, uploadResponse.Code)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/files", nil)
	listResponse := httptest.NewRecorder()
	engine.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, listResponse.Code)
	}

	var body struct {
		Files []map[string]any `json:"files"`
	}
	if err := json.Unmarshal(listResponse.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(body.Files) != 1 {
		t.Fatalf("expected 1 file listed, got %d", len(body.Files))
	}
	if body.Files[0]["name"] != "notes.txt" {
		t.Fatalf("expected listed file notes.txt, got %v", body.Files[0]["name"])
	}
}
