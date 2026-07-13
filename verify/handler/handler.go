package handler

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	RegisterViaWA(c *gin.Context)
	RegisterViaEmail(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	Refresh(c *gin.Context)
}

type WhatsAppWebhookHandler interface {
	WhatsAppVerify(c *gin.Context)
	WhatsAppHandler(c *gin.Context)
}
