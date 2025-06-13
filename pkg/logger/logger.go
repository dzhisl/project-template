package logger

import (
	"strings"

	"example.com/m/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// skipLevel is the number of stack frames to ascend to report the correct caller
const skipLevel = 1

// InitLogger sets up the global logger based on the environment
func InitLogger() {
	var cfg zap.Config
	environment := config.AppConfig.StageLevel

	// Use JSON logger for production, console logger for development
	if environment == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.MessageKey = "message"
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Disable stacktrace to reduce verbosity
	cfg.EncoderConfig.StacktraceKey = ""

	// Set log level from configuration
	levelStr := strings.ToLower(config.AppConfig.LogLevel)
	switch levelStr {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		// If we can't build the logger, use a default logger to report the error
		zap.NewExample().Fatal("Error building logger", zap.Error(err))
	}

	// Replace the global logger
	zap.ReplaceGlobals(logger)
}

// Info logs an info message with optional fields
func Info(msg string, fields ...zap.Field) {
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Info(msg, fields...)
}

// Error logs an info message with optional fields
func Error(msg string, fields ...zap.Field) {
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Error(msg, fields...)
}

// Debug logs a debug message with optional fields
func Debug(msg string, fields ...zap.Field) {
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Debug(msg, fields...)
}

// Warn logs a warning message with optional fields
func Warn(msg string, fields ...zap.Field) {
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Warn(msg, fields...)
}

// Fatal logs a fatal message with optional fields and then exits
func Fatal(msg string, fields ...zap.Field) {
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Fatal(msg, fields...)
}
