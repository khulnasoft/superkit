package auth

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to string, token string) error
}

var defaultEmailSender EmailSender = &noopEmailSender{}

type noopEmailSender struct{}

func (n *noopEmailSender) SendVerificationEmail(ctx context.Context, to string, token string) error {
	return nil
}

func SetEmailSender(sender EmailSender) {
	defaultEmailSender = sender
}

func SendVerificationEmail(ctx context.Context, to string, token string) error {
	return defaultEmailSender.SendVerificationEmail(ctx, to, token)
}
