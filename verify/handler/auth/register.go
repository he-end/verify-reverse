package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/verify/log"
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

type RegisterViaEmailResBody struct {
	Message string `json:"message"`
	Link    string `json:"link"`
	QRLink  string `json:"qr_link"`
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

	var link, qrLink *string

	code, _, err := h.authSvc.InitiateWAVerify(ctx, *req.Number, name, pwd)
	if err != nil {
		logger.Error("initiate WA verification", zap.Error(err))
	} else {
		link, qrLink, err = h.wa.CreateLinkRegister(ctx, *code, *req.Number)
		if err != nil {
			logger.Error("create QR link", zap.Error(err))
		}
	}

	c.Header("Conten-Type", "application/json")

	resBody := RegisterViaEmailResBody{
		Message: "jika nomor memenuhi syarat, link verifikasi WhatsApp telah disiapkan. silakan scan QR atau akses link ini.",
		// Link:    *link,
		// QRLink:  *qrLink,
	}

	if link == nil && qrLink == nil {
		logger.Warn("generate link error", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	} else {
		if link != nil {
			resBody.Link = *link
		}
		if qrLink != nil {
			resBody.QRLink = *qrLink
		}
	}

	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resBody); err != nil {
		logger.Warn("encode reseponse error", zap.Error(err))
		response.InternalError(c, "something went wrong")
		return
	}
	// response.OK(c, gin.H{
	// 	"message": "jika nomor memenuhi syarat, link verifikasi WhatsApp telah disiapkan",
	// 	"link":    link,
	// 	"qrlink":  qrLink,
	// })
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
