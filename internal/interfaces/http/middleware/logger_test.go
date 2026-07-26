package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerWritesRequestFields(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Logger(logger))
	engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if logs.Len() != 1 {
		t.Fatalf("expected one request log, got %d", logs.Len())
	}

	fields := logs.All()[0].ContextMap()
	if fields["method"] != http.MethodGet {
		t.Fatalf("expected logged method GET, got %#v", fields["method"])
	}
	if fields["path"] != "/" {
		t.Fatalf("expected logged path /, got %#v", fields["path"])
	}
	if fields["status"] != int64(http.StatusOK) {
		t.Fatalf("expected logged status 200, got %#v", fields["status"])
	}
	if fields["body_size"] != int64(2) {
		t.Fatalf("expected logged body_size 2, got %#v", fields["body_size"])
	}
	if fields["client_ip"] == "" {
		t.Fatal("expected client_ip to be logged")
	}
}
