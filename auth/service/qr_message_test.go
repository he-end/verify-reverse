package service

import "testing"

func TestCreateLinkRegister(t *testing.T) {
	svc := SetupWAService("token", "https://graph.facebook.com/v22.0", "12345", "628123456789")
	link := svc.CreateLinkRegister("VRFY-ABC12345")
	expected := "https://wa.me/628123456789?text=VERIFY%3AVRFY-ABC12345"
	if link != expected {
		t.Errorf("expected %q, got %q", expected, link)
	}
}
