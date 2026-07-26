package bootstrap

import (
	appfile "github.com/barlus-developer/go-simple-file-upload/internal/application/file"
	"github.com/barlus-developer/go-simple-file-upload/internal/application/health"
	"github.com/barlus-developer/go-simple-file-upload/internal/infrastructure/config"
	"github.com/barlus-developer/go-simple-file-upload/internal/infrastructure/logger"
	"github.com/barlus-developer/go-simple-file-upload/internal/infrastructure/storage"
	httpHandler "github.com/barlus-developer/go-simple-file-upload/internal/interfaces/http/handler"
	httpRouter "github.com/barlus-developer/go-simple-file-upload/internal/interfaces/http/router"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Config config.Config
	Logger *zap.Logger
	Router *gin.Engine
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log, err := logger.New(cfg.App.Environment)
	if err != nil {
		return nil, err
	}

	healthService := health.NewService()
	healthHandler := httpHandler.NewHealthHandler(healthService)

	localStorage, err := storage.NewLocal(cfg.Storage.Dir)
	if err != nil {
		return nil, err
	}
	fileService := appfile.NewService(localStorage)
	fileHandler := httpHandler.NewFileHandler(fileService)

	router := httpRouter.New(cfg, log, healthHandler, fileHandler)

	return &App{
		Config: cfg,
		Logger: log,
		Router: router,
	}, nil
}
