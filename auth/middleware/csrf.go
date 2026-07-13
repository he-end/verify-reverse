package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth/conf"
	"github.com/he-end/verify-reverse/auth/log"
)

const (
	csrfCookieName     = "csrf_token"
	csrfCookiePath     = "/api/v1.0"
	csrfCookieSameSite = http.SameSiteStrictMode
	csrfTokenLength    = 32
	csrfMaxAge         = 86400
	csrfHeaderName     = "X-CSRF-Token"
)

func isProduction(env string) bool {
	return env != "dev" && env != "development"
}

func CSRFMiddleware(cfg *conf.Conf) gin.HandlerFunc {
	if !isProduction(cfg.AppEnv) {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			c.Next()
			return
		}

		cookie, err := c.Cookie(csrfCookieName)
		if err != nil {
			logger := log.CtxLogger(c.Request.Context())
			logger.Warn("csrf token missing from cookie", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token is missing"})
			return
		}

		header := c.GetHeader(csrfHeaderName)
		if header == "" {
			logger := log.CtxLogger(c.Request.Context())
			logger.Warn("csrf token missing from header", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token is missing"})
			return
		}

		if cookie != header {
			logger := log.CtxLogger(c.Request.Context())
			logger.Warn("csrf token mismatch", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
			return
		}

		c.Next()
	}
}

func CSRFTokenHandler(cfg *conf.Conf) gin.HandlerFunc {
	return func(c *gin.Context) {
		bytes := make([]byte, csrfTokenLength)
		if _, err := rand.Read(bytes); err != nil {
			logger := log.CtxLogger(c.Request.Context())
			logger.Error("failed to generate csrf token", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate CSRF token"})
			return
		}

		token := hex.EncodeToString(bytes)

		secure := c.Request.TLS != nil
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     csrfCookieName,
			Value:    token,
			Path:     csrfCookiePath,
			MaxAge:   csrfMaxAge,
			HttpOnly: false,
			Secure:   secure,
			SameSite: csrfCookieSameSite,
		})

		expiresAt := time.Now().Add(csrfMaxAge * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
			"expires_at": expiresAt,
		})
	}
}
