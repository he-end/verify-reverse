package model

type SendMessageModels struct {
	MessagingProduct string        `json:"messaging_product"`
	To               string        `json:"to"`
	RecipientType    string        `json:"recipient_type"`
	Type             string        `json:"type"`
	Text             *TextBody     `json:"text,omitempty"`
	Template         *TemplateBody `json:"template,omitempty"`
}

type TextBody struct {
	Body string `json:"body"`
}

type LanguageBody struct {
	Code string `json:"code"`
}

type TemplateBody struct {
	Name     string        `json:"name"`
	Language *LanguageBody `json:"language"`
}
