package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMaxBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		bodySize       int
		limit          int64
		contentLength  string
		expectStatus   int
		checkHeader    bool
		headerPresent  bool
	}{
		{
			name:          "GET request bypasses check",
			method:        http.MethodGet,
			limit:         100,
			expectStatus:  http.StatusOK,
			checkHeader:   true,
			headerPresent: false,
		},
		{
			name:          "HEAD request bypasses check",
			method:        http.MethodHead,
			limit:         100,
			expectStatus:  http.StatusOK,
			checkHeader:   true,
			headerPresent: false,
		},
		{
			name:          "POST within limit succeeds",
			method:        http.MethodPost,
			bodySize:      50,
			limit:         100,
			expectStatus:  http.StatusOK,
			checkHeader:   true,
			headerPresent: true,
		},
		{
			name:          "POST exceeds limit with Content-Length",
			method:        http.MethodPost,
			limit:         100,
			contentLength: "150",
			expectStatus:  http.StatusRequestEntityTooLarge,
			checkHeader:   false,
		},
		{
			name:          "PUT within limit succeeds",
			method:        http.MethodPut,
			bodySize:      50,
			limit:         100,
			expectStatus:  http.StatusOK,
			checkHeader:   true,
			headerPresent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// Setup middleware and handler
			r.Use(MaxBody(tt.limit))
			r.Any("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Create request
			var body string
			if tt.bodySize > 0 {
				body = strings.Repeat("a", tt.bodySize)
			}
			req := httptest.NewRequest(tt.method, "/test", strings.NewReader(body))
			
			if tt.contentLength != "" {
				req.Header.Set("Content-Length", tt.contentLength)
			}

			// Execute request
			r.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, w.Code)
			}

			// Check X-Max-Body-Bytes header
			if tt.checkHeader {
				header := w.Header().Get("X-Max-Body-Bytes")
				if tt.headerPresent {
					if header == "" {
						t.Error("expected X-Max-Body-Bytes header to be set")
					}
				} else {
					if header != "" {
						t.Error("expected X-Max-Body-Bytes header to not be set")
					}
				}
			}

			// Check Connection: close header for rejected requests
			if tt.expectStatus == http.StatusRequestEntityTooLarge {
				if w.Header().Get("Connection") != "close" {
					t.Error("expected Connection: close header for rejected request")
				}
			}
		})
	}
}

func TestMaxBody_WebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(MaxBody(100))
	r.GET("/ws", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Upgrade", "websocket")

	r.ServeHTTP(w, req)

	// WebSocket upgrade should bypass the check
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should not have X-Max-Body-Bytes header
	if w.Header().Get("X-Max-Body-Bytes") != "" {
		t.Error("expected no X-Max-Body-Bytes header for websocket")
	}
}
