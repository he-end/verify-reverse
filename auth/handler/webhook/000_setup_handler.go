package webhook

import (
	"github.com/he-end/verify-reverse/auth/repository/auth"
	"github.com/he-end/verify-reverse/auth/service"
	authsvc "github.com/he-end/verify-reverse/auth/service/auth"
	usersvc "github.com/he-end/verify-reverse/auth/service/user"
)

type Handler struct {
	wa          *service.WaService
	val         *service.Validator
	authSvc     *authsvc.AuthService
	attemptRepo *auth.AttemptRepository
	rateLimiter *service.RateLimiter
	userSvc     *usersvc.UserService
}

func New(
	wa *service.WaService,
	val *service.Validator,
	authSvc *authsvc.AuthService,
	attemptRepo *auth.AttemptRepository,
	rateLimiter *service.RateLimiter,
	userSvc *usersvc.UserService,
) *Handler {
	return &Handler{
		wa:          wa,
		val:         val,
		authSvc:     authSvc,
		attemptRepo: attemptRepo,
		rateLimiter: rateLimiter,
		userSvc:     userSvc,
	}
}
