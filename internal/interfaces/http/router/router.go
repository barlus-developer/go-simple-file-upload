package router

import (
	"github.com/barlus-developer/go-simple-file-upload/internal/infrastructure/config"
	"github.com/barlus-developer/go-simple-file-upload/internal/interfaces/http/handler"
	"github.com/barlus-developer/go-simple-file-upload/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func New(cfg config.Config, log *zap.Logger, healthHandler *handler.HealthHandler, fileHandler *handler.FileHandler) *gin.Engine {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logger(log))
	engine.MaxMultipartMemory = cfg.Storage.MaxUploadSizeBytes()

	engine.GET("/", healthHandler.Index)

	files := engine.Group("/files")
	{
		files.POST("", fileHandler.Upload)
		files.GET("", fileHandler.List)
		files.GET("/:name", fileHandler.Download)
	}

	return engine
}
