package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth/log"
	"github.com/he-end/verify-reverse/auth/middleware"
	"github.com/he-end/verify-reverse/auth/repository"
	authrepo "github.com/he-end/verify-reverse/auth/repository/auth"
	"github.com/he-end/verify-reverse/auth/response"
	authsvc "github.com/he-end/verify-reverse/auth/service/auth"
)

type LoginReqBody struct {
	Email  *string `json:"email"`
	Number *string `json:"number"`
	Pwd    string  `json:"pwd"`
}

type loginResponse struct {
	User        authrepo.UserResponse `json:"user"`
	AccessToken string                `json:"access_token"`
	ExpiresAt   time.Time             `json:"expires_at"`
}

type refreshResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

const (
	refreshCookiePath     = "/api/v1.0"
	refreshCookieSameSite = http.SameSiteStrictMode
)

func (h *Handler) setRefreshTokenCookie(c *gin.Context, token string, ttl time.Duration) {
	secure := c.Request.TLS != nil
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.refreshCookieName,
		Value:    token,
		Path:     refreshCookiePath,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: refreshCookieSameSite,
	})
}

func (h *Handler) clearRefreshTokenCookie(c *gin.Context) {
	secure := c.Request.TLS != nil
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: refreshCookieSameSite,
	})
}

func (h *Handler) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	var req LoginReqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind login payload", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	if report := h.val.Struct(req); report.HasErrors() {
		response.Errors(c, http.StatusBadRequest, report.ToMap())
		return
	}

	input := authsvc.LoginInput{
		Email:  req.Email,
		Number: req.Number,
		Pwd:    req.Pwd,
	}

	user, tokens, err := h.authSvc.Login(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidCredentials):
			response.Unauthorized(c, "invalid email/number or password")
		case errors.Is(err, repository.ErrMissingContact):
			response.BadRequest(c, "email or number is required")
		default:
			logger.Error("login", zap.Error(err))
			response.InternalError(c, "something went wrong")
		}
		return
	}

	h.setRefreshTokenCookie(c, tokens.RefreshToken, h.jwtSvc.RefreshTTL())

	accessExpiresAt := time.Now().Add(h.jwtSvc.AccessTTL())

	response.OK(c, loginResponse{
		User:        user.ToResponse(),
		AccessToken: tokens.AccessToken,
		ExpiresAt:   accessExpiresAt,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	authUser, err := middleware.GetUserFromContext(c)
	if err != nil {
		response.Unauthorized(c, "authentication required")
		return
	}

	if err := h.authSvc.Logout(ctx, authUser.UserID); err != nil {
		logger.Error("logout", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	}

	response.OK(c, gin.H{"message": "logged out successfully"})
	h.clearRefreshTokenCookie(c)
}

func (h *Handler) Refresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	refreshToken, err := c.Cookie(h.refreshCookieName)
	if err != nil {
		response.Unauthorized(c, "refresh token is required")
		return
	}

	_, tokens, err := h.authSvc.RefreshTokens(ctx, refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			response.Unauthorized(c, "invalid refresh token")
		case errors.Is(err, repository.ErrTokenExpired):
			response.Unauthorized(c, "refresh token has expired")
		default:
			logger.Error("refresh token", zap.Error(err))
			response.InternalError(c, "something went wrong")
		}
		return
	}

	accessExpiresAt := time.Now().Add(h.jwtSvc.AccessTTL())

	response.OK(c, refreshResponse{
		AccessToken: tokens.AccessToken,
		ExpiresAt:   accessExpiresAt,
	})
	h.setRefreshTokenCookie(c, tokens.RefreshToken, h.jwtSvc.RefreshTTL())
}
