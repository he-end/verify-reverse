package log

import (
	"go.uber.org/zap"
)

func Info(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	logger.Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	logger.Panic(msg, fields...)
}
