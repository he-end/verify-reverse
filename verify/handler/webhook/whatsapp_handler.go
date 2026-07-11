package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/verify/log"
)

type webhookEntry struct {
	Changes []webhookChange `json:"changes"`
}

type webhookChange struct {
	Value webhookValue `json:"value"`
}

type webhookValue struct {
	Messages []webhookMessage `json:"messages"`
	Contacts []webhookContact `json:"contacts"`
}

type webhookMessage struct {
	From string       `json:"from"`
	Text *webhookText `json:"text"`
}

type webhookText struct {
	Body string `json:"body"`
}

type webhookContact struct {
	WaID   string `json:"wa_id"`
	WaName string `json:"wa_name"`
}

type webhookBody struct {
	Entry []webhookEntry `json:"entry"`
}

var verifyCodePattern = regexp.MustCompile(`^VERIFY:(VRFY-[A-Z0-9]{8})$`)

func (h *Handler) WhatsAppVerify(c *gin.Context) {
	challenge := c.Query("hub.challenge")
	if challenge != "" {
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusBadRequest)
}

func (h *Handler) WhatsAppHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.CtxLogger(c.Request.Context()).Error("failed to read webhook body", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	var payload webhookBody
	if err := json.Unmarshal(body, &payload); err != nil {
		log.CtxLogger(c.Request.Context()).Error("failed to parse webhook body", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if msg.From == "" || msg.Text == nil {
					continue
				}

				if !h.rateLimiter.Allow(msg.From) {
					c.Status(http.StatusOK)
					return
				}

				ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
				defer cancel()
				logger := log.CtxLogger(ctx)

				blocked, err := h.attemptRepo.IsBlocked(ctx, msg.From, "wa")
				if err != nil {
					logger.Warn("attempt check failed", zap.String("from", msg.From), zap.Error(err))
				} else if blocked {
					cancel()
					c.Status(http.StatusOK)
					return
				}

				verified, err := h.authSvc.IsAlreadyVerified(ctx, msg.From)
				if err != nil {
					logger.Warn("verified check failed", zap.String("from", msg.From), zap.Error(err))
				} else if verified {
					cancel()
					c.Status(http.StatusOK)
					return
				}

				matches := verifyCodePattern.FindStringSubmatch(msg.Text.Body)
				if len(matches) != 2 {
					if err := h.attemptRepo.RecordFailed(ctx, msg.From, "wa"); err != nil {
						logger.Error("record failed attempt", zap.String("from", msg.From), zap.Error(err))
					}
					cancel()
					continue
				}
				code := matches[1]

				user, err := h.authSvc.CompleteWAVerify(ctx, code)
				if err != nil {
					logger.Error("verification failed", zap.Error(err))
					if recErr := h.attemptRepo.RecordFailed(ctx, msg.From, "wa"); recErr != nil {
						logger.Error("record failed attempt", zap.String("from", msg.From), zap.Error(recErr))
					}
					cancel()
					c.Status(http.StatusOK)
					return
				}

				if err := h.attemptRepo.ResetAttempts(ctx, msg.From, "wa"); err != nil {
					logger.Warn("reset attempts failed", zap.String("from", msg.From), zap.Error(err))
				}

				h.wa.SendMessage(ctx, msg.From, "Verifikasi berhasil.")
				logger.Info("user verified via WhatsApp", zap.String("user_id", user.ID.String()))
				cancel()
				c.Status(http.StatusOK)
				return
			}
		}
	}

	c.Status(http.StatusOK)
}
