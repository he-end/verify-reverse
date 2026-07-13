package middleware

import (
	"context"
	"io"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth/log"
)

const requestIDKey contextKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := uuid.NewV7()
		requestID := id.String()

		ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-Id", requestID)

		c.Next()
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, err any) {
		stack := make([]byte, 4096)
		stack = stack[:runtime.Stack(stack, false)]
		logger := log.CtxLogger(c.Request.Context())
		logger.Error("panic recovered",
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
