package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func VerifyMetaWebhook(appSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signatureHeader := c.GetHeader("X-Hub-Signature-256")
		if signatureHeader == "" || !strings.HasPrefix(signatureHeader, "sha256=") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		expectedSignature := strings.TrimPrefix(signatureHeader, "sha256=")

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		mac := hmac.New(sha256.New, []byte(appSecret))
		mac.Write(bodyBytes)
		calculatedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(calculatedSignature), []byte(expectedSignature)) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
