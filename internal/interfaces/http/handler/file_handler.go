package handler

import (
	"errors"
	"fmt"
	"net/http"

	appfile "github.com/barlus-developer/go-simple-file-upload/internal/application/file"
	"github.com/gin-gonic/gin"
)

// FileHandler exposes HTTP endpoints for uploading and retrieving files.
type FileHandler struct {
	service appfile.Service
}

// NewFileHandler builds a FileHandler backed by the given file service.
func NewFileHandler(service appfile.Service) *FileHandler {
	return &FileHandler{service: service}
}

// Upload handles POST /files. It expects a multipart form with a "file"
// field and stores the uploaded content.
func (h *FileHandler) Upload(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing \"file\" form field"})
		return
	}

	src, err := header.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read uploaded file"})
		return
	}
	defer src.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	stored, err := h.service.Upload(c.Request.Context(), header.Filename, contentType, src)
	if err != nil {
		if errors.Is(err, appfile.ErrInvalidName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store file"})
		return
	}

	c.JSON(http.StatusCreated, stored)
}

// Download handles GET /files/:name and streams the stored file back to
// the client.
func (h *FileHandler) Download(c *gin.Context) {
	name := c.Param("name")

	reader, meta, err := h.service.Download(c.Request.Context(), name)
	if err != nil {
		switch {
		case errors.Is(err, appfile.ErrInvalidName):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
		case errors.Is(err, appfile.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read file"})
		}
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, meta.Name))
	c.DataFromReader(http.StatusOK, meta.Size, "application/octet-stream", reader, nil)
}

// List handles GET /files and returns metadata for every stored file.
func (h *FileHandler) List(c *gin.Context) {
	files, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}
