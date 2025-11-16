package logger

import (
	"kei-services/pkg/config"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestInit_StdoutOutput(t *testing.T) {
	loggerCfg := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	appCfg := &config.App{
		Name:        "test-app",
		Environment: config.Dev,
		Version:     "1.0.0",
	}

	logger := Init(loggerCfg, appCfg)

	if logger == nil {
		t.Error("expected non-nil logger")
	}

	// Test that logger works
	logger.Info("test message")
}

func TestInit_FileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	loggerCfg := &Config{
		Level:    "info",
		Format:   "json",
		Output:   "file",
		FilePath: logFile,
	}
	appCfg := &config.App{
		Name:        "test-app",
		Environment: config.Dev,
		Version:     "1.0.0",
	}

	logger := Init(loggerCfg, appCfg)

	if logger == nil {
		t.Error("expected non-nil logger")
	}

	// Write a log message
	logger.Info("test file message")

	// Check that file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("expected log file to be created at %s", logFile)
	}
}

func TestInit_BothOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	loggerCfg := &Config{
		Level:    "info",
		Format:   "json",
		Output:   "both",
		FilePath: logFile,
	}
	appCfg := &config.App{
		Name:        "test-app",
		Environment: config.Dev,
		Version:     "1.0.0",
	}

	logger := Init(loggerCfg, appCfg)

	if logger == nil {
		t.Error("expected non-nil logger")
	}

	// Write a log message
	logger.Info("test both message")

	// Check that file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("expected log file to be created at %s", logFile)
	}
}

func TestInit_ProductionEnvironment(t *testing.T) {
	loggerCfg := &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	appCfg := &config.App{
		Name:        "test-app",
		Environment: config.Prod,
		Version:     "1.0.0",
	}

	logger := Init(loggerCfg, appCfg)

	if logger == nil {
		t.Error("expected non-nil logger")
	}

	// Test that logger works
	logger.Info("test production message")
}

func TestInit_ConsoleFormat(t *testing.T) {
	loggerCfg := &Config{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
	appCfg := &config.App{
		Name:        "test-app",
		Environment: config.Dev,
		Version:     "1.0.0",
	}

	logger := Init(loggerCfg, appCfg)

	if logger == nil {
		t.Error("expected non-nil logger")
	}

	// Test that logger works
	logger.Debug("test console message")
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel zapcore.Level
	}{
		{
			name:          "debug level",
			level:         "debug",
			expectedLevel: zapcore.DebugLevel,
		},
		{
			name:          "info level",
			level:         "info",
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name:          "warn level",
			level:         "warn",
			expectedLevel: zapcore.WarnLevel,
		},
		{
			name:          "error level",
			level:         "error",
			expectedLevel: zapcore.ErrorLevel,
		},
		{
			name:          "invalid level defaults to info",
			level:         "invalid",
			expectedLevel: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Level: tt.level}
			result := getLogLevel(cfg)

			if result != tt.expectedLevel {
				t.Errorf("expected %v, got %v", tt.expectedLevel, result)
			}
		})
	}
}

func TestGetLogEncoder_Json(t *testing.T) {
	cfg := &Config{
		Format:      "json",
		Environment: config.Dev,
	}

	encoder := getLogEncoder(cfg)

	if encoder == nil {
		t.Error("expected non-nil encoder")
	}
}

func TestGetLogEncoder_Console(t *testing.T) {
	cfg := &Config{
		Format:      "console",
		Environment: config.Dev,
	}

	encoder := getLogEncoder(cfg)

	if encoder == nil {
		t.Error("expected non-nil encoder")
	}
}

func TestGetLogEncoder_ConsoleProduction(t *testing.T) {
	cfg := &Config{
		Format:      "console",
		Environment: config.Prod,
	}

	encoder := getLogEncoder(cfg)

	if encoder == nil {
		t.Error("expected non-nil encoder")
	}
}

func TestGetLogEncoder_Invalid(t *testing.T) {
	cfg := &Config{
		Format:      "invalid",
		Environment: config.Dev,
	}

	encoder := getLogEncoder(cfg)

	if encoder != nil {
		t.Error("expected nil encoder for invalid format")
	}
}

func TestInit_InvalidLogDir(t *testing.T) {
	// Use a path that we can't create (in a non-existent parent dir)
	// This test documents the panic behavior
	loggerCfg := &Config{
		Level:    "info",
		Format:   "json",
		Output:   "file",
		FilePath: "/root/nonexistent/test.log", // typically can't write here
	}
	appCfg := &config.App{
		Name:        "test-app",
		Environment: config.Dev,
		Version:     "1.0.0",
	}

	defer func() {
		if r := recover(); r == nil {
			// If we can actually create the directory (unlikely), that's fine
			t.Log("directory creation succeeded unexpectedly, or test is running as root")
		}
	}()

	Init(loggerCfg, appCfg)
}
