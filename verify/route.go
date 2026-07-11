package verify

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/he-end/verify-reverse/verify/middleware"
)

func (c *Container) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1.0")
	{
		// api.GET("/whatsapp", c.Webhook.WhatsAppVerify) // this is Endpoint used for Verify Add URL Webhook in Whatsapp
		api.POST("/whatsapp/", c.Webhook.WhatsAppHandler)
		api.POST("/wa-register", middleware.RateLimitRegister(5, time.Minute), c.Auth.RegisterViaWA)
		api.POST("/email-register", middleware.RateLimitRegister(5, time.Minute), c.Auth.RegisterViaEmail)
		api.POST("/login", c.Auth.Login)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(c.JwtSvc))
		{
			protected.POST("/logout", c.Auth.Logout)
		}
	}
}
