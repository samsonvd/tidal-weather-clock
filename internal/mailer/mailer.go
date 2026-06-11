package mailer

import (
	"log"

	"github.com/resend/resend-go/v3"
)

func getMagicLink(baseUrl, token string) string {
	return baseUrl + "/auth/verify?token=" + token
}

type Mailer interface {
	SendMagicLink(email, token string) error
}

type LogMailer struct {
	baseUrl string
}

func NewLogMailer(baseUrl string) *LogMailer {
	return &LogMailer{baseUrl: baseUrl}
}

func (m *LogMailer) SendMagicLink(email, token string) error {
	log.Printf("magic link for %s: %s", email, getMagicLink(m.baseUrl, token))
	return nil
}

type ResendMailer struct {
	baseUrl      string
	resendClient *resend.Client
}

func NewResendMailer(resendApiKey, baseUrl string) *ResendMailer {
	return &ResendMailer{resendClient: resend.NewClient(resendApiKey), baseUrl: baseUrl}
}

func (m *ResendMailer) SendMagicLink(email, token string) error {
	magicLink := getMagicLink(m.baseUrl, token)
	params := &resend.SendEmailRequest{
		From: "Kath's Weather <no-reply@kaths-weather.app>",
		To:   []string{email},
		Template: &resend.EmailTemplate{
			Id: "welcome-email",
			Variables: map[string]any{
				"magic_link_url": magicLink,
				"expiry_text":    "15 minutes",
			},
		},
	}
	sent, err := m.resendClient.Emails.Send(params)
	if err != nil {
		return err
	}
	log.Printf("sent magic link: %v", sent)
	return nil
}
