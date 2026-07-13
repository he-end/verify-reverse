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

type ChangeWANumberReqBody struct {
	Number string `json:"number"`
}

type ChangeWANumberResBody struct {
	Message string `json:"message"`
	Link    string `json:"link"`
}

func (h *Handler) ChangeWANumber(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	authUser, err := middleware.GetUserFromContext(c)
	if err != nil {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req ChangeWANumberReqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind change WA number payload", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	if report := h.val.Struct(req); report.HasErrors() {
		response.Errors(c, http.StatusBadRequest, report.ToMap())
		return
	}

	code, err := h.userSvc.InitiateWANumberChange(ctx, authUser.UserID, req.Number)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNumberTaken):
			response.Conflict(c, "phone number already in use")
		case errors.Is(err, repository.ErrDuplicateKey):
			response.BadRequest(c, "this is already your phone number")
		default:
			logger.Error("initiate WA number change", zap.Error(err))
			response.InternalError(c, "something went wrong")
		}
		return
	}

	link := h.wa.CreateLinkChange(*code)

	response.OK(c, ChangeWANumberResBody{
		Message: "link verifikasi WhatsApp telah disiapkan. silakan kirim pesan CHANGE: dari nomor baru Anda untuk konfirmasi.",
		Link:    link,
	})
}
