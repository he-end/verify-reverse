package middleware

import (
	"context"
	"io"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/verify/log"
)

const requestIDKey contextKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := uuid.NewV7()
		requestID := id.String()

		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		log.NewLoggerOnRuntime(log.RegisterRuntime{Key: "request_id", Value: requestID})
		defer log.DeferDeleteRuntimeValue()

		c.Next()
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, err any) {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		log.Error("panic recovered",
			zap.Any("error", err),
			zap.String("stack", string(stack)),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func GetRequestID(c *gin.Context) string {
	if v, ok := c.Request.Context().Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}
