package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/verify/log"
	"github.com/he-end/verify-reverse/verify/middleware"
	"github.com/he-end/verify-reverse/verify/repository"
	authrepo "github.com/he-end/verify-reverse/verify/repository/auth"
	"github.com/he-end/verify-reverse/verify/response"
	authsvc "github.com/he-end/verify-reverse/verify/service/auth"
)

type LoginReqBody struct {
	Email  *string `json:"email"`
	Number *string `json:"number"`
	Pwd    string  `json:"pwd"`
}

type authResponse struct {
	User  authrepo.UserResponse `json:"user"`
	Token authsvc.TokenPair     `json:"token"`
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

	response.OK(c, authResponse{
		User:  user.ToResponse(),
		Token: *tokens,
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
}
