package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth"
	"github.com/he-end/verify-reverse/auth/conf"
	"github.com/he-end/verify-reverse/auth/log"
	"github.com/he-end/verify-reverse/auth/middleware"
)

func main() {
	log.InitLogger("dev", "info")
	defer log.Sync()

	cfg := conf.GetEnv()

	logger, err := log.InitLogger(cfg.AppEnv, cfg.LogLevel)
	if err != nil {
		fmt.Printf("failed to reconfigure logger: %v\n", err)
		return
	}
	_ = logger

	log.Info("starting server", zap.String("env", cfg.AppEnv))

	container := auth.NewContainer(context.Background(), cfg)
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.Use(middleware.RequestIDMiddleware())
	router.Use(func(c *gin.Context) {
		requestID := middleware.GetRequestID(c)
		logger := log.GetLogger()
		if requestID != "" {
			logger = logger.With(zap.String("request_id", requestID))
		}
		ctx := log.WithLogger(c.Request.Context(), logger)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.Use(middleware.RecoveryMiddleware())
	container.RegisterRoutes(router)

	logger.Info("server starting", zap.String("port", "8080"))
	router.Run(":8080")
}
