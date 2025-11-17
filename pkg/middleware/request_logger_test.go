package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create observable logger for testing
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(RequestLogger(logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	req.Header.Set("User-Agent", "test-agent")
	r.ServeHTTP(w, req)

	// Check that log was recorded
	if logs.Len() == 0 {
		t.Error("expected at least one log entry")
	}

	entry := logs.All()[0]

	// Verify log fields
	fields := entry.ContextMap()

	if fields["requestID"] != "test-request-id" {
		t.Errorf("expected requestID 'test-request-id', got %v", fields["requestID"])
	}

	if fields["method"] != "GET" {
		t.Errorf("expected method 'GET', got %v", fields["method"])
	}

	if fields["status"] != int64(200) {
		t.Errorf("expected status 200, got %v", fields["status"])
	}

	if fields["userAgent"] != "test-agent" {
		t.Errorf("expected userAgent 'test-agent', got %v", fields["userAgent"])
	}
}

func TestRequestLogger_4xxStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.WarnLevel)
	logger := zap.New(core)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(RequestLogger(logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	r.ServeHTTP(w, req)

	// Should log at WARN level for 4xx
	if logs.Len() == 0 {
		t.Error("expected at least one log entry")
	}

	entry := logs.All()[0]
	if entry.Level != zap.WarnLevel {
		t.Errorf("expected WARN level for 4xx, got %v", entry.Level)
	}

	if !strings.Contains(entry.Message, "4XX") {
		t.Errorf("expected message to contain '4XX', got %s", entry.Message)
	}
}

func TestRequestLogger_5xxStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(RequestLogger(logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	r.ServeHTTP(w, req)

	// Should log at ERROR level for 5xx
	if logs.Len() == 0 {
		t.Error("expected at least one log entry")
	}

	entry := logs.All()[0]
	if entry.Level != zap.ErrorLevel {
		t.Errorf("expected ERROR level for 5xx, got %v", entry.Level)
	}

	if !strings.Contains(entry.Message, "5XX") {
		t.Errorf("expected message to contain '5XX', got %s", entry.Message)
	}
}

func TestRequestLogger_WithUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(RequestLogger(logger))
	r.GET("/test", func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	r.ServeHTTP(w, req)

	if logs.Len() == 0 {
		t.Error("expected at least one log entry")
	}

	entry := logs.All()[0]
	fields := entry.ContextMap()

	if fields["user"] != "testuser" {
		t.Errorf("expected user 'testuser', got %v", fields["user"])
	}
}

func TestRequestLogger_Latency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(RequestLogger(logger))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	r.ServeHTTP(w, req)

	if logs.Len() == 0 {
		t.Error("expected at least one log entry")
	}

	entry := logs.All()[0]
	fields := entry.ContextMap()

	// Check that latency field exists (value will be very small)
	if _, exists := fields["latency"]; !exists {
		t.Error("expected latency field to be present")
	}
}
