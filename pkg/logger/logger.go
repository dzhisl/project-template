package logger

import (
	"context"
	"strings"
	"sync"

	"example.com/m/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// skipLevel is the number of stack frames to ascend to report the correct caller
const skipLevel = 1

var (
	once sync.Once
)

// InitLogger sets up the global logger based on the environment
func InitLogger() {
	var cfg zap.Config
	environment := config.Get().Logging.Stage

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
	levelStr := strings.ToLower(config.Get().Logging.Level)
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

// withReqID attaches request_id from ctx if present
func withReqID(ctx context.Context, fields ...zap.Field) []zap.Field {
	if reqID := extractReqIdFromCtx(ctx); reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	return fields
}

// Info logs an info message with optional fields
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	once.Do(func() {
		InitLogger()
	})
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Info(msg, withReqID(ctx, fields...)...)
}

// Error logs an info message with optional fields
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	once.Do(func() {
		InitLogger()
	})
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Error(msg, withReqID(ctx, fields...)...)
}

// Debug logs a debug message with optional fields
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	once.Do(func() {
		InitLogger()
	})
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Debug(msg, withReqID(ctx, fields...)...)
}

// Warn logs a warning message with optional fields
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	once.Do(func() {
		InitLogger()
	})
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Warn(msg, withReqID(ctx, fields...)...)
}

// Fatal logs a fatal message with optional fields and then exits
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	once.Do(func() {
		InitLogger()
	})
	zap.L().WithOptions(zap.AddCallerSkip(skipLevel)).Fatal(msg, withReqID(ctx, fields...)...)
}

func extractReqIdFromCtx(ctx context.Context) string {
	if ctx != nil {
		val := ctx.Value("request_id")
		if reqID, ok := val.(string); ok {
			return reqID
		}
	}
	return ""
}
