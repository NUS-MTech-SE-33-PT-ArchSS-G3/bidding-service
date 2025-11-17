package config

import (
	"testing"
)

func TestCorsStruct(t *testing.T) {
	cors := Cors{
		IsEnabled:        true,
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		AllowMaxAge:      300,
	}

	if !cors.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}

	if len(cors.AllowOrigins) != 2 {
		t.Errorf("expected 2 AllowOrigins, got %d", len(cors.AllowOrigins))
	}

	if len(cors.AllowMethods) != 4 {
		t.Errorf("expected 4 AllowMethods, got %d", len(cors.AllowMethods))
	}

	if !cors.AllowCredentials {
		t.Error("expected AllowCredentials to be true")
	}

	if cors.AllowMaxAge != 300 {
		t.Errorf("expected AllowMaxAge to be 300, got %d", cors.AllowMaxAge)
	}
}
