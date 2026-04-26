package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger based on configuration
func New(level, format string) (*zap.Logger, error) {
	// Parse log level
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	// Determine encoder based on format
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// Build logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Add development mode if needed
	if level == "debug" {
		logger = logger.WithOptions(zap.Development())
	}

	return logger, nil
}

// Must creates a new logger or panics if error occurs
func Must(level, format string) *zap.Logger {
	logger, err := New(level, format)
	if err != nil {
		panic(err)
	}
	return logger
}