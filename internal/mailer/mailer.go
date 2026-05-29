package mailer

import "log"

type Mailer interface {
	SendMagicLink(email, token string) error
}

type LogMailer struct{}

func (m *LogMailer) SendMagicLink(email, token string) error {
	log.Printf("magic link for %s: /auth/verify?token=%s", email, token)
	return nil
}
