package webhook

import (
	"github.com/he-end/verify-reverse/verify/repository/auth"
	"github.com/he-end/verify-reverse/verify/service"
	authsvc "github.com/he-end/verify-reverse/verify/service/auth"
)

type Handler struct {
	wa          *service.WaService
	val         *service.Validator
	authSvc     *authsvc.AuthService
	attemptRepo *auth.AttemptRepository
	rateLimiter *service.RateLimiter
}

func New(
	wa *service.WaService,
	val *service.Validator,
	authSvc *authsvc.AuthService,
	attemptRepo *auth.AttemptRepository,
	rateLimiter *service.RateLimiter,
) *Handler {
	return &Handler{
		wa:          wa,
		val:         val,
		authSvc:     authSvc,
		attemptRepo: attemptRepo,
		rateLimiter: rateLimiter,
	}
}
