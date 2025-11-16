package middleware

import (
	"kei-services/pkg/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestCors_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled: false,
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Normal request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// CORS headers should not be present
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS headers when disabled")
	}
}

func TestCors_Disabled_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled: false,
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Preflight request
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(w, req)

	// Should return 204 No Content for preflight even when disabled
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestCors_NoOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:    true,
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Request without Origin header (same-origin or non-browser)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// CORS headers should not be added for same-origin requests
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS headers for requests without Origin")
	}
}

func TestCors_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:    true,
		AllowOrigins: []string{"http://localhost:3000", "https://example.com"},
		AllowMethods: []string{"GET", "POST", "PUT"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected Allow-Origin header to be 'http://localhost:3000', got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Check Vary headers
	vary := w.Header().Get("Vary")
	if vary == "" {
		t.Error("expected Vary header to be set")
	}
}

func TestCors_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:    true,
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	r.ServeHTTP(w, req)

	// Should return 403 Forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestCors_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:    true,
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT"},
		AllowHeaders: []string{"Content-Type"},
		AllowMaxAge:  300,
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	r.ServeHTTP(w, req)

	// Should return 204 No Content for successful preflight
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Error("expected Allow-Origin header")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Allow-Methods header")
	}

	if w.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("expected Max-Age header")
	}
}

func TestCors_WithCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:        true,
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST"},
		AllowCredentials: true,
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check credentials header
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected Allow-Credentials header to be 'true'")
	}
}

func TestCors_WebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:    true,
		AllowOrigins: []string{"http://localhost:3000"},
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/ws", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Upgrade", "websocket")
	r.ServeHTTP(w, req)

	// WebSocket upgrade should bypass CORS
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCors_MaxAgeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:    true,
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
		AllowMaxAge:  1000, // Set to value > 600
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	// Max age should be capped at 600
	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "600" {
		t.Errorf("expected Max-Age to be capped at 600, got %s", maxAge)
	}
}

func TestCors_ExposeHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Cors{
		IsEnabled:     true,
		AllowOrigins:  []string{"http://localhost:3000"},
		AllowMethods:  []string{"GET"},
		ExposeHeaders: []string{"X-Request-ID", "X-Custom-Header"},
	}
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Cors(cfg, logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	exposeHeaders := w.Header().Get("Access-Control-Expose-Headers")
	if exposeHeaders == "" {
		t.Error("expected Expose-Headers to be set")
	}
}
