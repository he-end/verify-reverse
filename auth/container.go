package auth

import (
	"context"
	"time"

	"github.com/he-end/verify-reverse/auth/conf"
	"github.com/he-end/verify-reverse/auth/handler"
	"github.com/he-end/verify-reverse/auth/handler/auth"
	"github.com/he-end/verify-reverse/auth/handler/user"
	"github.com/he-end/verify-reverse/auth/handler/webhook"
	"github.com/he-end/verify-reverse/auth/log"
	"github.com/he-end/verify-reverse/auth/repository"
	authrepo "github.com/he-end/verify-reverse/auth/repository/auth"
	"github.com/he-end/verify-reverse/auth/service"
	authsvc "github.com/he-end/verify-reverse/auth/service/auth"
	usersvc "github.com/he-end/verify-reverse/auth/service/user"
	"go.uber.org/zap"
)

type Container struct {
	Auth    handler.AuthHandler
	User    handler.UserHandler
	Webhook handler.WhatsAppWebhookHandler
	WaSvc   *service.WaService
	JwtSvc  *authsvc.JWTService
	Val     *service.Validator
	cfg     *conf.Conf
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
	authService := authsvc.NewAuthService(authRepo, sessionRepo, verifyRepo, jwtSvc, cfg.AllowMultiSession, cfg.MaxSession)

	attemptRepo := authrepo.NewAttemptRepository(db)
	rateLimiter := service.NewRateLimiter(ctx)

	userSvc := usersvc.NewUserService(authRepo, sessionRepo, verifyRepo)

	wa.StartExpiredCleanup(ctx, verifyRepo, 5*time.Minute)

	return &Container{
		Auth:    auth.New(wa, val, authService, jwtSvc, cfg.RefreshCookieName),
		User:    user.New(wa, val, userSvc),
		Webhook: webhook.New(wa, val, authService, attemptRepo, rateLimiter, userSvc),
		WaSvc:   wa,
		JwtSvc:  jwtSvc,
		Val:     val,
		cfg:     cfg,
	}
}
