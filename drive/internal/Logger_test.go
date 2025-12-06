package logger

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitUsesLogLevelFromEnv(t *testing.T) {
	_ = os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

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

	r := gin.New()
	r.Use(Middleware())
	r.GET("/ping", func(c *gin.Context) {
		id := CorrelationID(c)
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
	if rr.Header().Get(CorrelationIDHeader) == "" {
		t.Fatalf("expected %s header to be set", CorrelationIDHeader)
	}
}
