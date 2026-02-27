package service

// EmailSender defines the interface for sending HTML emails (mockable for tests)
type EmailSender interface {
	SendHTML(to, subject, htmlBody string) error
}
