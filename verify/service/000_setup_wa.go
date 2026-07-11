package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type whatsappConf struct {
	TokenWhatsApp   *string
	BaseURLGraphAPI *string
	PhoneNumberID   *string
	PhoneNumber     *string
}

type WaService struct {
	conf *whatsappConf
	buildMessage
}

func SetupWAService(token string, baseUrGraphAPI string, phoneNumberID string, phoneNumber string) *WaService {
	return &WaService{
		conf: &whatsappConf{
			TokenWhatsApp:   &token,
			BaseURLGraphAPI: &baseUrGraphAPI,
			PhoneNumberID:   &phoneNumberID,
			PhoneNumber:     &phoneNumber,
		},
		buildMessage: &msg{},
	}
}

func (w *whatsappConf) buildURL(path *string) (string, error) {
	if path == nil {
		return *w.BaseURLGraphAPI, nil
	}
	result, err := url.JoinPath(*w.BaseURLGraphAPI, *path)
	if err != nil {
		return "", fmt.Errorf("build URL path: %w", err)
	}
	return result, nil
}

func (w *whatsappConf) buildReq(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	newReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build HTTP request: %w", err)
	}
	newReq.Header.Set("Authorization", fmt.Sprintf("Bearer %v", *w.TokenWhatsApp))
	if body != nil {
		newReq.Header.Set("Content-Type", "application/json")
	}
	return newReq, nil
}
