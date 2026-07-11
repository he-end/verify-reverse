package log

import (
	"go.uber.org/zap"
)

func Info(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	rt := GetLoggerRuntimeStore()
	if rt != nil {
		fields = append(fields, zap.String(rt.Key, rt.Value))
	}
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	rt := GetLoggerRuntimeStore()
	if rt != nil {
		fields = append(fields, zap.String(rt.Key, rt.Value))
	}
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	rt := GetLoggerRuntimeStore()
	if rt != nil {
		fields = append(fields, zap.String(rt.Key, rt.Value))
	}
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	rt := GetLoggerRuntimeStore()
	if rt != nil {
		fields = append(fields, zap.String(rt.Key, rt.Value))
	}
	logger.Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger := GetLogger().WithOptions(zap.AddCallerSkip(1))
	rt := GetLoggerRuntimeStore()
	if rt != nil {
		fields = append(fields, zap.String(rt.Key, rt.Value))
	}
	logger.Panic(msg, fields...)
}
