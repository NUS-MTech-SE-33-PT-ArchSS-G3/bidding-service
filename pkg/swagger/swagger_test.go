package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestServe_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &Config{
		IsEnabled:   false,
		Title:       "Test API",
		OpenApiName: "test-api",
	}

	logger := zap.NewNop()
	r := gin.New()

	getSwagger := func() (*openapi3.T, error) {
		return &openapi3.T{}, nil
	}

	Serve(getSwagger, r, cfg, logger)

	// Should not have any swagger routes when disabled
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/test-api/", nil)
	r.ServeHTTP(w, req)

	// Should return 404 since routes weren't registered
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestServe_Enabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temporary swagger directory and file
	tmpDir := t.TempDir()
	swaggerDir := filepath.Join(tmpDir, "swagger")
	err := os.MkdirAll(swaggerDir, 0755)
	if err != nil {
		t.Fatalf("failed to create swagger dir: %v", err)
	}

	yamlFile := filepath.Join(swaggerDir, "test-api.yaml")
	err = os.WriteFile(yamlFile, []byte("openapi: 3.0.0\ninfo:\n  title: Test\n  version: 1.0.0\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}

	// Change to temp directory so StaticFile can find the file
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	cfg := &Config{
		IsEnabled:   true,
		Title:       "Test API",
		OpenApiName: "test-api",
	}

	logger := zap.NewNop()
	r := gin.New()

	getSwagger := func() (*openapi3.T, error) {
		return &openapi3.T{
			OpenAPI: "3.0.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
		}, nil
	}

	Serve(getSwagger, r, cfg, logger)

	// Test JSON spec endpoint
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger-spec/test-api.json", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for JSON spec, got %d", w.Code)
	}
}

func TestServe_SwaggerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &Config{
		IsEnabled:   true,
		Title:       "Test API",
		OpenApiName: "test-api",
	}

	logger := zap.NewNop()
	r := gin.New()

	getSwagger := func() (*openapi3.T, error) {
		return nil, http.ErrServerClosed
	}

	Serve(getSwagger, r, cfg, logger)

	// Test JSON spec endpoint with error
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger-spec/test-api.json", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 for error, got %d", w.Code)
	}
}

func TestConfig_Structure(t *testing.T) {
	cfg := Config{
		IsEnabled:   true,
		Title:       "My API",
		OpenApiName: "my-api",
	}

	if !cfg.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}

	if cfg.Title != "My API" {
		t.Errorf("expected Title to be 'My API', got %s", cfg.Title)
	}

	if cfg.OpenApiName != "my-api" {
		t.Errorf("expected OpenApiName to be 'my-api', got %s", cfg.OpenApiName)
	}
}
