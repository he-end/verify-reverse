package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/he-end/verify-reverse/verify/repository"
	authsvc "github.com/he-end/verify-reverse/verify/service/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthenticatedUser struct {
	UserID   uuid.UUID
	UserHash string
}

func AuthMiddleware(jwtSvc *authsvc.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			writeAuthError(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeAuthError(c, "invalid authorization header format")
			return
		}

		claims, err := jwtSvc.ValidateAccessToken(c.Request.Context(), parts[1])
		if err != nil {
			writeAuthError(c, "invalid or expired token")
			return
		}

		user := &AuthenticatedUser{
			UserID:   claims.UserID,
			UserHash: claims.UserHash,
		}

		c.Set(string(UserContextKey), user)
		c.Next()
	}
}

func GetUserFromContext(c *gin.Context) (*AuthenticatedUser, error) {
	user, exists := c.Get(string(UserContextKey))
	if !exists {
		return nil, repository.ErrTokenInvalid
	}
	u, ok := user.(*AuthenticatedUser)
	if !ok {
		return nil, repository.ErrTokenInvalid
	}
	return u, nil
}

func writeAuthError(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
