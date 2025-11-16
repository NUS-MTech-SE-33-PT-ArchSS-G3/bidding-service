package config

import (
	"testing"
)

func TestEnvironmentConstants(t *testing.T) {
	tests := []struct {
		name     string
		env      Environment
		expected string
	}{
		{
			name:     "Development environment",
			env:      Dev,
			expected: "development",
		},
		{
			name:     "Production environment",
			env:      Prod,
			expected: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.env) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.env))
			}
		})
	}
}

func TestAppStruct(t *testing.T) {
	app := App{
		Name:        "test-app",
		Environment: Dev,
		Version:     "1.0.0",
	}

	if app.Name != "test-app" {
		t.Errorf("expected Name to be 'test-app', got %s", app.Name)
	}

	if app.Environment != Dev {
		t.Errorf("expected Environment to be Dev, got %s", app.Environment)
	}

	if app.Version != "1.0.0" {
		t.Errorf("expected Version to be '1.0.0', got %s", app.Version)
	}
}
