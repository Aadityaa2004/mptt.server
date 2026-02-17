package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

type EmailService struct {
	smtpHost    string
	smtpPort    int
	encryption  string
	username    string
	password    string
	fromName    string
	fromAddress string
}

func NewEmailService(host string, port int, encryption, username, password, fromName, fromAddress string) *EmailService {
	return &EmailService{
		smtpHost:    host,
		smtpPort:    port,
		encryption:  encryption,
		username:    username,
		password:    password,
		fromName:    fromName,
		fromAddress: fromAddress,
	}
}

func (s *EmailService) SendHTML(to, subject, htmlBody string) error {
	addr := net.JoinHostPort(s.smtpHost, strconv.Itoa(s.smtpPort))

	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		s.fromName, s.fromAddress, to, subject, htmlBody)

	auth := smtp.PlainAuth("", s.username, s.password, s.smtpHost)

	if s.encryption == "starttls" {
		return s.sendWithSTARTTLS(addr, auth, to, []byte(msg))
	}

	// Plain SMTP
	return smtp.SendMail(addr, auth, s.fromAddress, []string{to}, []byte(msg))
}

func (s *EmailService) sendWithSTARTTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: s.smtpHost,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	if err := client.Mail(s.fromAddress); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close message failed: %w", err)
	}

	return client.Quit()
}
