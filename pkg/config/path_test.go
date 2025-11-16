package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPath_WithConfigPathEnv(t *testing.T) {
	// Save original env and restore after test
	originalEnv := os.Getenv("CONFIG_PATH")
	defer func() {
		if originalEnv != "" {
			os.Setenv("CONFIG_PATH", originalEnv)
		} else {
			os.Unsetenv("CONFIG_PATH")
		}
	}()

	// Set CONFIG_PATH environment variable
	expectedPath := "/custom/config.json"
	os.Setenv("CONFIG_PATH", expectedPath)

	result := Path()

	if result != expectedPath {
		t.Errorf("expected %s, got %s", expectedPath, result)
	}
}

func TestPath_WithFlag(t *testing.T) {
	// Clear CONFIG_PATH env
	os.Unsetenv("CONFIG_PATH")

	// Reset flag set for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	configFlag := flag.String("config", "/flag/config.json", "config path")
	flag.Parse()

	result := Path()

	if result != *configFlag {
		t.Errorf("expected %s, got %s", *configFlag, result)
	}

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestPath_ContainerDefault(t *testing.T) {
	// Clear CONFIG_PATH env
	os.Unsetenv("CONFIG_PATH")

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Create temporary container config file
	tmpFile := "/tmp/test-app-config.json"
	if err := os.WriteFile(tmpFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Mock container path by creating symbolic link if possible
	// This test is primarily for documentation as we can't easily mock /app/config.json
	result := Path()

	// The result should be a valid path (either container default or fallback)
	if result == "" {
		t.Error("expected non-empty path")
	}
}

func TestPath_ExecutableDir(t *testing.T) {
	// Clear CONFIG_PATH env
	os.Unsetenv("CONFIG_PATH")

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	result := Path()

	// Should return a path ending with config.json
	if !strings.HasSuffix(result, "config.json") {
		t.Errorf("expected path to end with 'config.json', got %s", result)
	}

	// Should be a valid path format
	if !filepath.IsAbs(result) && result != "config.json" {
		t.Logf("path is relative: %s", result)
	}
}

func TestPath_Fallback(t *testing.T) {
	// Clear CONFIG_PATH env
	os.Unsetenv("CONFIG_PATH")

	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Ensure /app/config.json doesn't exist (it shouldn't in test environment)
	// The function should fall back to either executable dir or "config.json"
	
	result := Path()

	// Result should not be empty
	if result == "" {
		t.Error("expected non-empty path")
	}

	// Result should end with config.json
	if !strings.HasSuffix(result, "config.json") {
		t.Errorf("expected result to end with 'config.json', got %s", result)
	}
}
