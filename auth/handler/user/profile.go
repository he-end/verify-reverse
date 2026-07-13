package user

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth/log"
	"github.com/he-end/verify-reverse/auth/middleware"
	"github.com/he-end/verify-reverse/auth/response"
	"github.com/he-end/verify-reverse/auth/service"
	usersvc "github.com/he-end/verify-reverse/auth/service/user"
)

type Handler struct {
	wa      *service.WaService
	val     *service.Validator
	userSvc *usersvc.UserService
}

func New(wa *service.WaService, val *service.Validator, userSvc *usersvc.UserService) *Handler {
	h := &Handler{wa: wa, val: val, userSvc: userSvc}
	h.registerValidator(val)
	return h
}

type UpdateProfileReqBody struct {
	Name string `json:"name"`
}

func (h *Handler) GetProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	authUser, err := middleware.GetUserFromContext(c)
	if err != nil {
		response.Unauthorized(c, "authentication required")
		return
	}

	user, err := h.userSvc.GetProfile(ctx, authUser.UserID)
	if err != nil {
		logger.Error("get profile", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	}

	response.OK(c, user.ToResponse())
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	authUser, err := middleware.GetUserFromContext(c)
	if err != nil {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req UpdateProfileReqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind update profile payload", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	if report := h.val.Struct(req); report.HasErrors() {
		response.Errors(c, http.StatusBadRequest, report.ToMap())
		return
	}

	user, err := h.userSvc.UpdateProfile(ctx, authUser.UserID, usersvc.UpdateProfileInput{
		Name: req.Name,
	})
	if err != nil {
		logger.Error("update profile", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	}

	response.OK(c, user.ToResponse())
}
