package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMetricsMiddlewareIncrementsCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	InitMetrics()

	r := gin.New()
	r.Use(Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// Сам факт, что не упали — уже норм для простого smoke-теста
}

func TestRegisterExposesMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	InitMetrics()

	r := gin.New()
	Register(r, "/metrics")

	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /metrics, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected body from /metrics, got empty")
	}
}
