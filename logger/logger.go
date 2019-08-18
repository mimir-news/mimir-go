package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetLogger creates a named logger for internal application logs.
func GetLogger(name string, level zapcore.Level) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		return logger, err
	}
	return logger.With(zap.String("logger", name)), nil
}

// MustGetLogger creates a names logger and panics on failure.
func MustGetLogger(name string, level zapcore.Level) *zap.Logger {
	logger, err := GetLogger(name, level)
	if err != nil {
		log.Fatalln("Failed to get zap.Logger "+name, err)
	}

	return logger
}

// GetDefaultLogger gets a named logger using the default implementation.
func GetDefaultLogger(name string) *zap.Logger {
	return MustGetLogger(name, zap.DebugLevel)
}
