package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/verify/log"
	"github.com/he-end/verify-reverse/verify/repository"
	authrepo "github.com/he-end/verify-reverse/verify/repository/auth"
	"github.com/he-end/verify-reverse/verify/response"
	"github.com/he-end/verify-reverse/verify/service"
	authsvc "github.com/he-end/verify-reverse/verify/service/auth"
)

type Handler struct {
	wa      *service.WaService
	val     *service.Validator
	authSvc *authsvc.AuthService
	jwtSvc  *authsvc.JWTService
}

func New(
	wa *service.WaService,
	val *service.Validator,
	authSvc *authsvc.AuthService,
	jwtSvc *authsvc.JWTService,
) *Handler {
	h := &Handler{wa: wa, val: val, authSvc: authSvc, jwtSvc: jwtSvc}
	h.registerValidator(val)
	return h
}

type RegisterViaWAReqBody struct {
	Number     *string `json:"number,require"`
	Name       *string `json:"name"`
	Pwd        *string `json:"pwd"`
	ConfirmPwd *string `json:"confirm_pwd"`
}

type RegisterViaEmailReqBody struct {
	Email      *string `json:"email"`
	Name       *string `json:"name"`
	Pwd        *string `json:"pwd"`
	ConfirmPwd *string `json:"confirm_pwd"`
}

type waRegisterResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Link      string    `json:"link"`
	QrLink    string    `json:"qr_link"`
	Message   string    `json:"message"`
}

type emailRegisterResponse struct {
	User  authrepo.UserResponse `json:"user"`
	Token authsvc.TokenPair     `json:"token"`
}

func (h *Handler) RegisterViaWA(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	var req RegisterViaWAReqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind register WA payload", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	if report := h.val.Struct(req); report.HasErrors() {
		response.Errors(c, http.StatusBadRequest, report.ToMap())
		return
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	pwd := ""
	if req.Pwd != nil {
		pwd = *req.Pwd
	}

	code, expiresAt, err := h.authSvc.InitiateWAVerify(ctx, *req.Number, name, pwd)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			response.Conflict(c, "phone number already registered")
		case errors.Is(err, repository.ErrVerificationPending):
			response.TooManyRequests(c, "verification already pending, wait or retry after expiry")
		default:
			logger.Error("initiate WA verification", zap.Error(err))
			response.InternalError(c, "something went wrong")
		}
		return
	}

	link, qrLink, err := h.wa.CreateLinkRegister(ctx, *code, *req.Number)
	if err != nil {
		logger.Error("create QR link", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	}
	resBody := waRegisterResponse{
		Code:      *code,
		ExpiresAt: expiresAt,
		Link:      *link,
		QrLink:    *qrLink,
		Message:   "Scan QR code and send the verification message to complete registration",
	}
	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resBody); err != nil {
		response.InternalError(c, "something went wrong")
		return
	}
	// response.OK(c, waRegisterResponse{
	// 	Code:      *code,
	// 	ExpiresAt: expiresAt,
	// 	Link:      *link,
	// 	QrLink:    *qrLink,
	// 	Message:   "Scan QR code and send the verification message to complete registration",
	// })
}

func (h *Handler) RegisterViaEmail(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	logger := log.CtxLogger(ctx)

	var req RegisterViaEmailReqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("bind register email payload", zap.Error(err))
		response.BadRequest(c, "invalid request body")
		return
	}

	if report := h.val.Struct(req); report.HasErrors() {
		response.Errors(c, http.StatusBadRequest, report.ToMap())
		return
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	pwd := ""
	if req.Pwd != nil {
		pwd = *req.Pwd
	}

	user, tokens, err := h.authSvc.Register(ctx, authsvc.RegisterInput{
		Email: req.Email,
		Name:  name,
		Pwd:   pwd,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			response.Conflict(c, "email already registered")
		case errors.Is(err, repository.ErrMissingContact):
			response.BadRequest(c, "email is required")
		default:
			logger.Error("register via email", zap.Error(err))
			response.InternalError(c, "something went wrong")
		}
		return
	}

	response.OK(c, emailRegisterResponse{
		User:  user.ToResponse(),
		Token: *tokens,
	})
}
