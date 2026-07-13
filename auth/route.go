package auth

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/he-end/verify-reverse/auth/middleware"
)

func (c *Container) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1.0")

	api.GET("/csrf-token", middleware.CSRFTokenHandler(c.cfg))
	api.POST("/whatsapp/", middleware.VerifyMetaWebhook(c.cfg.WebhookAppSecret), c.Webhook.WhatsAppHandler)

	csrf := api.Group("")
	csrf.Use(middleware.CSRFMiddleware(c.cfg))
	{
		csrf.POST("/wa-register", middleware.RateLimitRegister(5, time.Minute), c.Auth.RegisterViaWA)
		csrf.POST("/email-register", middleware.RateLimitRegister(5, time.Minute), c.Auth.RegisterViaEmail)
		csrf.POST("/login", c.Auth.Login)
		csrf.POST("/refresh", c.Auth.Refresh)

		protected := csrf.Group("")
		protected.Use(middleware.AuthMiddleware(c.JwtSvc))
		{
			protected.GET("/me", c.User.GetProfile)
			protected.POST("/logout", c.Auth.Logout)
			protected.PATCH("/me", c.User.UpdateProfile)
			protected.PUT("/me/password", c.User.ChangePassword)
			protected.PUT("/me/wa-number", c.User.ChangeWANumber)
		}
	}
}
