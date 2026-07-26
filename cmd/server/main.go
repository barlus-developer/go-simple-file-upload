package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/barlus-developer/go-simple-file-upload/internal/bootstrap"
)

func main() {
	app, err := bootstrap.New()
	if err != nil {
		panic(err)
	}
	defer app.Logger.Sync()

	server := &http.Server{
		Addr:              app.Config.Server.Address(),
		Handler:           app.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		app.Logger.Info("http server started", app.Config.Server.ZapFields()...)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Fatal("http server failed", app.Config.Server.ErrorField(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Fatal("http server shutdown failed", app.Config.Server.ErrorField(err))
	}

	app.Logger.Info("http server stopped")
}
