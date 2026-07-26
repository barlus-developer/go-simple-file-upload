package handler

import (
	"net/http"

	"github.com/barlus-developer/go-simple-file-upload/internal/application/health"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	service health.Service
}

func NewHealthHandler(service health.Service) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) Index(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.Status())
}
