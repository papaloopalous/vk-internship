package logger

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

// InitLogger инициализирует логгер
func InitLogger(logDir string) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "service.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)

	core := zapcore.NewTee(consoleCore, fileCore)
	logger = zap.New(core)
}

func Info(service, message string, metadata map[string]string) {
	logWithLevel(zap.InfoLevel, service, message, metadata)
}

func Error(service, message string, metadata map[string]string) {
	logWithLevel(zap.ErrorLevel, service, message, metadata)
}

func logWithLevel(level zapcore.Level, service, message string, metadata map[string]string) {
	fields := []zap.Field{zap.String("service", service)}
	for k, v := range metadata {
		fields = append(fields, zap.String(k, v))
	}

	switch level {
	case zap.InfoLevel:
		logger.Info(message, fields...)
	case zap.ErrorLevel:
		logger.Error(message, fields...)
	case zap.DebugLevel:
		logger.Debug(message, fields...)
	default:
		logger.Warn("unknown log level", fields...)
	}
}

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
