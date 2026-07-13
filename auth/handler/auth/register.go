package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth/log"
	"github.com/he-end/verify-reverse/auth/response"
	"github.com/he-end/verify-reverse/auth/service"
	authsvc "github.com/he-end/verify-reverse/auth/service/auth"
)

type Handler struct {
	wa                *service.WaService
	val               *service.Validator
	authSvc           *authsvc.AuthService
	jwtSvc            *authsvc.JWTService
	refreshCookieName string
}

func New(
	wa *service.WaService,
	val *service.Validator,
	authSvc *authsvc.AuthService,
	jwtSvc *authsvc.JWTService,
	refreshCookieName string,
) *Handler {
	h := &Handler{wa: wa, val: val, authSvc: authSvc, jwtSvc: jwtSvc, refreshCookieName: refreshCookieName}
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

// RegisterViaWAResBody is the response for WhatsApp registration.
// The client MUST:
//  1. Parse the "link" field — it is a wa.me deep link URL.
//  2. Render a QR code from the link using a client-side library
//     (e.g., qrcode.js: QRCode.toDataURL(link) → <img src="...">).
//  3. Display both the QR image AND the link as a clickable fallback.
//  4. When the user scans the QR or clicks the link, WhatsApp opens
//     with the verification message pre-filled. The user taps "Send".
//  5. The server receives the message via webhook, verifies the code,
//     and sends a confirmation reply via WhatsApp.
type RegisterViaWAResBody struct {
	Message string `json:"message"`
	Link    string `json:"link"`
}

func (h *Handler) RegisterViaWA(c *gin.Context) {
	start := time.Now()
	defer func() { sleepRemaining(time.Since(start)) }()

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

	code, _, err := h.authSvc.InitiateWAVerify(ctx, *req.Number, name, pwd)
	if err != nil {
		logger.Error("initiate WA verification", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	}

	link := h.wa.CreateLinkRegister(*code)

	response.OK(c, RegisterViaWAResBody{
		Message: "jika nomor memenuhi syarat, link verifikasi WhatsApp telah disiapkan. silakan scan QR atau akses link ini.",
		Link:    link,
	})
}

func (h *Handler) RegisterViaEmail(c *gin.Context) {
	start := time.Now()
	defer func() { sleepRemaining(time.Since(start)) }()

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

	_, _, err := h.authSvc.Register(ctx, authsvc.RegisterInput{
		Email: req.Email,
		Name:  name,
		Pwd:   pwd,
	})
	if err != nil {
		logger.Error("register via email", zap.Error(err))
	}

	response.OK(c, gin.H{"message": "jika email memenuhi syarat, link verifikasi telah disiapkan"})
}
