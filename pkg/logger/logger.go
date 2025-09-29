package logger

import (
	"kei-services/pkg/config"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(loggerCfg *config.Logger, appCfg *config.App) *zap.Logger {
	level := getLogLevel(loggerCfg)
	encoder := getLogEncoder(loggerCfg)

	var cores []zapcore.Core

	// stdout writer
	if loggerCfg.Output == "stdout" || loggerCfg.Output == "both" {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// file writer
	if loggerCfg.Output == "file" || loggerCfg.Output == "both" {
		logPath := loggerCfg.FilePath
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			panic("Failed to create log directory: " + err.Error())
		}

		file, err := os.OpenFile(loggerCfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(file),
			level,
		))
	}

	if appCfg.Environment == config.Prod {
		return zap.New(
			zapcore.NewTee(cores...),
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		return zap.New(
			zapcore.NewTee(cores...),
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
			zap.Development())
	}
}

func getLogLevel(cfg *config.Logger) zapcore.Level {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return zapcore.InfoLevel
	}
	return level
}

func getLogEncoder(cfg *config.Logger) zapcore.Encoder {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.CallerKey = "caller"
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "msg"

	if strings.ToLower(cfg.Format) == "json" {
		return zapcore.NewJSONEncoder(encoderCfg)
	} else if strings.ToLower(cfg.Format) == "console" {
		if cfg.Environment == config.Dev {
			encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return zapcore.NewConsoleEncoder(encoderCfg)
	}

	return nil
}
