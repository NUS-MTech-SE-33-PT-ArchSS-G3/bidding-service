package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		existingRequestID string
		expectGenerated  bool
	}{
		{
			name:             "generates new request ID when not provided",
			existingRequestID: "",
			expectGenerated:  true,
		},
		{
			name:             "uses existing request ID",
			existingRequestID: "existing-request-id-123",
			expectGenerated:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set(RequestIDHeader, tt.existingRequestID)
			}
			c.Request = req

			var requestID string
			middleware := RequestID()
			middleware(c)

			// Check that request ID is set in context
			requestID = c.GetString(RequestIDKey)
			if requestID == "" {
				t.Error("expected request ID to be set in context")
			}

			// Check response header
			responseID := w.Header().Get(RequestIDHeader)
			if responseID == "" {
				t.Error("expected request ID in response header")
			}

			if tt.expectGenerated {
				// Should be a UUID format
				if len(requestID) < 30 {
					t.Errorf("expected generated UUID, got %s", requestID)
				}
			} else {
				if requestID != tt.existingRequestID {
					t.Errorf("expected %s, got %s", tt.existingRequestID, requestID)
				}
			}

			// Verify request and response IDs match
			if requestID != responseID {
				t.Errorf("request ID mismatch: context=%s, header=%s", requestID, responseID)
			}
		})
	}
}

func TestWithRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test logger
	logger := zap.NewNop()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set request ID first
	c.Set(RequestIDKey, "test-request-id")

	middleware := WithRequestLogger(logger)
	middleware(c)

	// Check that logger is set in gin context
	loggerFromGin, exists := c.Get("logger")
	if !exists {
		t.Error("expected logger to be set in gin context")
	}

	if loggerFromGin == nil {
		t.Error("expected non-nil logger in gin context")
	}

	// Check that logger is also in request context
	loggerFromCtx := LoggerFrom(c.Request.Context(), logger)
	if loggerFromCtx == nil {
		t.Error("expected non-nil logger from context")
	}
}

func TestLoggerFrom(t *testing.T) {
	fallbackLogger := zap.NewNop()

	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectFallback bool
	}{
		{
			name: "returns logger from context",
			setupContext: func() context.Context {
				logger := zap.NewNop()
				return context.WithValue(context.Background(), loggerKey, logger)
			},
			expectFallback: false,
		},
		{
			name: "returns fallback when no logger in context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectFallback: true,
		},
		{
			name: "returns fallback when nil logger in context",
			setupContext: func() context.Context {
				var nilLogger *zap.Logger
				return context.WithValue(context.Background(), loggerKey, nilLogger)
			},
			expectFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			result := LoggerFrom(ctx, fallbackLogger)

			if result == nil {
				t.Error("expected non-nil logger")
			}

			if tt.expectFallback && result != fallbackLogger {
				t.Error("expected fallback logger")
			}
		})
	}
}
