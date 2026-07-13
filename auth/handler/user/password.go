package user

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
	"github.com/he-end/verify-reverse/auth/response"
)

type ChangePasswordReqBody struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (h *Handler) ChangePassword(c *gin.Context) {
	start := time.Now()
	defer func() { sleepRemaining(time.Since(start)) }()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	authUser, err := middleware.GetUserFromContext(c)
	if err != nil {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req ChangePasswordReqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind change password payload", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	if report := h.val.Struct(req); report.HasErrors() {
		response.Errors(c, http.StatusBadRequest, report.ToMap())
		return
	}

	err = h.userSvc.ChangePassword(ctx, authUser.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrWrongPassword):
			response.Unauthorized(c, "old password is incorrect")
		case errors.Is(err, repository.ErrNotFound):
			response.Unauthorized(c, "user not found")
		default:
			logger.Error("change password", zap.Error(err))
			response.InternalError(c, "something went wrong")
		}
		return
	}

	response.OK(c, gin.H{"message": "password changed successfully"})
}
