package log

import (
	"context"

	"go.uber.org/zap"
)

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

func CtxLogger(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return l
	}
	return GetLogger()
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
