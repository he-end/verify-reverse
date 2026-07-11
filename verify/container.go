package verify

import (
	"context"
	"time"

	"github.com/he-end/verify-reverse/verify/conf"
	"github.com/he-end/verify-reverse/verify/handler"
	"github.com/he-end/verify-reverse/verify/handler/auth"
	"github.com/he-end/verify-reverse/verify/handler/webhook"
	"github.com/he-end/verify-reverse/verify/log"
	"github.com/he-end/verify-reverse/verify/repository"
	authrepo "github.com/he-end/verify-reverse/verify/repository/auth"
	"github.com/he-end/verify-reverse/verify/service"
	authsvc "github.com/he-end/verify-reverse/verify/service/auth"
	"go.uber.org/zap"
)

type Container struct {
	Auth    handler.AuthHandler
	Webhook handler.WhatsAppWebhookHandler
	WaSvc   *service.WaService
	JwtSvc  *authsvc.JWTService
	Val     *service.Validator
}

func NewContainer(ctx context.Context, cfg *conf.Conf) *Container {
	wa := service.SetupWAService(cfg.TokenWhatsApp, cfg.BaseURLGraphAPI, cfg.PhoneNumberID, cfg.WhatsAppPhone)
	val := service.Default()

	db, err := repository.NewPostgresDB(ctx, cfg.DBConf.DSN())
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	// if err := repository.RunMigrations(ctx, db); err != nil {
	// 	log.Fatal("failed to run migrations", zap.Error(err))
	// }

	jwtSvc := authsvc.NewJWTService(cfg.JWTConf)
	authRepo := authrepo.NewAuthRepository(db)
	sessionRepo := authrepo.NewSessionRepository(db)
	verifyRepo := authrepo.NewVerificationRepository(db)
	authService := authsvc.NewAuthService(authRepo, sessionRepo, verifyRepo, jwtSvc)

	attemptRepo := authrepo.NewAttemptRepository(db)
	rateLimiter := service.NewRateLimiter(ctx)

	wa.StartExpiredCleanup(ctx, verifyRepo, 5*time.Minute)

	return &Container{
		Auth:    auth.New(wa, val, authService, jwtSvc),
		Webhook: webhook.New(wa, val, authService, attemptRepo, rateLimiter),
		WaSvc:   wa,
		JwtSvc:  jwtSvc,
		Val:     val,
	}
}
