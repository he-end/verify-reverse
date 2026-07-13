package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/he-end/verify-reverse/auth/log"
	"github.com/he-end/verify-reverse/auth/model"
	"github.com/he-end/verify-reverse/auth/repository/auth"
)

type CreateQRRes struct {
	Code         string `json:"code"`
	PrefilledMsg string `json:"prefilled_message"`
	DeepLinkURL  string `json:"deep_link_url"`
	QRImgURL     string `json:"qr_image_url"`
}

type CreateQRReq struct {
	PrefilledMsg string `json:"prefilled_message"`
	TypeMedia    string `json:"generate_qr_image"`
}

func (s *WaService) createQR(ctx context.Context, msg string) (*CreateQRRes, error) {
	path := fmt.Sprintf("/%v/%v", *s.conf.PhoneNumberID, "message_qrdls")
	url, err := s.conf.buildURL(&path)
	if err != nil {
		return nil, fmt.Errorf("build create QR URL: %w", err)
	}
	createBody := CreateQRReq{
		PrefilledMsg: msg,
		TypeMedia:    "PNG",
	}

	readBody, err := json.Marshal(createBody)
	if err != nil {
		return nil, fmt.Errorf("marshal create QR body: %w", err)
	}
	buff := bytes.NewBufferString(string(readBody))
	req, err := s.conf.buildReq(ctx, http.MethodPost, url, buff)
	if err != nil {
		return nil, fmt.Errorf("build create QR request: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send create QR request: %w", err)
	}
	defer res.Body.Close()

	var result CreateQRRes
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode create QR response: %w", err)
	}
	return &result, nil
}

type GetQRRes struct {
	Data []QR `json:"data"`
}

type QR struct {
	Code         string `json:"code"`
	PrefilledMsg string `json:"prefilled_message"`
	DeepLinkURL  string `json:"deep_link_url"`
}

func (s *WaService) getQr(ctx context.Context, code *string) (*GetQRRes, error) {
	path := fmt.Sprintf("/%v/message_qrdls", *s.conf.PhoneNumberID)
	if code != nil {
		path += fmt.Sprintf("/%v", *code)
	}

	url, err := s.conf.buildURL(&path)
	if err != nil {
		return nil, fmt.Errorf("build get QR URL: %w", err)
	}
	req, err := s.conf.buildReq(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build get QR request: %w", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send get QR request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		return nil, fmt.Errorf("cannot get QR")
	}

	var dataQr GetQRRes
	if code == nil {
		if err := json.NewDecoder(res.Body).Decode(&dataQr); err != nil {
			return nil, fmt.Errorf("decode get QR list response: %w", err)
		}
		return &dataQr, nil
	}

	var qr QR
	if err := json.NewDecoder(res.Body).Decode(&qr); err != nil {
		return nil, fmt.Errorf("decode get QR response: %w", err)
	}
	dataQr.Data = []QR{qr}
	return &dataQr, nil
}

type errorResponse struct {
	Error struct {
		Msg         string `json:"message"`
		Type        string `json:"type"`
		IsTransient string `json:"is_transient"`
		Code        int    `json:"code"`
		FbTraceID   string `json:"fbtrace_id"`
	} `json:"error"`
}

func (s *WaService) deleteQR(ctx context.Context, code string) error {
	type ResOK struct {
		Success bool `json:"success"`
	}

	path := fmt.Sprintf("/%v/message_qrdls/%v", *s.conf.PhoneNumberID, code)
	url, err := s.conf.buildURL(&path)
	if err != nil {
		return fmt.Errorf("build delete QR URL: %w", err)
	}
	req, err := s.conf.buildReq(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("build delete QR request: %w", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send delete QR request: %w", err)
	}
	defer res.Body.Close()

	var resOK ResOK
	if err := json.NewDecoder(res.Body).Decode(&resOK); err != nil {
		return fmt.Errorf("decode delete QR response: %w", err)
	}

	if !resOK.Success {
		return fmt.Errorf("delete QR: success field false")
	}
	return nil
}

func (s *WaService) CreateLinkRegister(codeRegister string) string {
	msg := url.QueryEscape("VERIFY:" + codeRegister)
	return fmt.Sprintf("https://wa.me/%s?text=%s", *s.conf.PhoneNumber, msg)
}

func (s *WaService) SendMessage(ctx context.Context, to, text string) error {
	path := fmt.Sprintf("/%v/messages", *s.conf.PhoneNumberID)
	url, err := s.conf.buildURL(&path)
	if err != nil {
		return fmt.Errorf("build send message URL: %w", err)
	}

	body := model.SendMessageModels{
		MessagingProduct: "whatsapp",
		To:               to,
		RecipientType:    "individual",
		Type:             "text",
		Text: &model.TextBody{
			Body: text,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal send message body: %w", err)
	}

	req, err := s.conf.buildReq(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("build send message request: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send message request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		var errRes errorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("send message failed with status %d", res.StatusCode)
		}
		return fmt.Errorf("send message failed: %s", errRes.Error.Msg)
	}

	return nil
}

func (s *WaService) StartExpiredCleanup(ctx context.Context, repo *auth.VerificationRepository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				n, err := repo.DeleteExpired(ctx)
				if err != nil {
					log.Error("cleanup expired codes", zap.Error(err))
				} else if n > 0 {
					log.Info("cleaned expired verification codes", zap.Int64("count", n))
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *WaService) TruncateQR(ctx context.Context) {
	log.Info("start to truncate QR")
	allQR, err := s.getQr(ctx, nil)
	if err != nil {
		log.Error("truncate QR: get QR failed", zap.Error(err))
	}
	log.Info("getting QR codes", zap.Int("count", len(allQR.Data)))
	deleted := 0
	for _, qr := range allQR.Data {
		if err := s.deleteQR(ctx, qr.Code); err != nil {
			log.Error("truncate QR: delete QR failed", zap.Error(err))
		}
		deleted++
		randomInt := rand.Intn(15-4+1) + 4
		time.Sleep(time.Duration(randomInt))
	}
	log.Info("truncate QR done", zap.Int("deleted", deleted))
}
