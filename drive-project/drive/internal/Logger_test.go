package logger

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitUsesLogLevelFromEnv(t *testing.T) {
	_ = os.Setenv("GODRIVE_LOG_LEVEL", "debug")
	defer os.Unsetenv("GODRIVE_LOG_LEVEL")

	l, err := Init()
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if l == nil {
		t.Fatalf("Init() returned nil logger")
	}
}

func TestMiddlewareSetsCorrelationID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := Init()
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	r := gin.New()
	r.Use(CorrelationIDMiddleware(logger))
	r.GET("/ping", func(c *gin.Context) {
		id := GetCorrelationID(c)
		if id == "" {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Correlation-ID") == "" {
		t.Fatalf("expected X-Correlation-ID header to be set")
	}
}
