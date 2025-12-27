package tools

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// EmailTools holds SMTP configuration for sending emails
type EmailTools struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	permChecker  PermissionChecker
}

// NewEmailTools creates email tools with SMTP configuration
func NewEmailTools(host string, port int, username, password, fromEmail string, permChecker PermissionChecker) *EmailTools {
	if fromEmail == "" {
		fromEmail = username
	}
	return &EmailTools{
		smtpHost:     host,
		smtpPort:     port,
		smtpUsername: username,
		smtpPassword: password,
		fromEmail:    fromEmail,
		permChecker:  permChecker,
	}
}

// --- Tool Parameter Structs ---

// SendEmailParams defines parameters for sending email
type SendEmailParams struct {
	ToEmail string `json:"to_email" jsonschema:"Recipient email address"`
	Subject string `json:"subject" jsonschema:"Email subject line"`
	Body    string `json:"body" jsonschema:"Email body text (include all content here)"`
}

// SendEmailResult is the result of sending an email
type SendEmailResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Tool Functions ---

// SendEmailCtx sends an email via SMTP
func (t *EmailTools) SendEmailCtx(ctx context.Context, params SendEmailParams) (*SendEmailResult, error) {
	if t.permChecker != nil && !t.permChecker.CanAccess(ctx, "email:send") {
		log.Printf("🚫 Permission denied: email:send")
		return &SendEmailResult{Success: false, Message: "Permission denied: email:send"}, nil
	}

	if t.smtpUsername == "" || t.smtpPassword == "" {
		log.Printf("SMTP credentials not configured")
		return &SendEmailResult{
			Success: false,
			Message: "Email not configured. SMTP_USERNAME and SMTP_PASSWORD environment variables are required.",
		}, nil
	}

	if params.ToEmail == "" {
		return &SendEmailResult{
			Success: false,
			Message: "Recipient email address is required.",
		}, nil
	}

	// Create email message
	msg := buildEmailMessage(t.fromEmail, params.ToEmail, params.Subject, params.Body)

	// Send via SMTP with STARTTLS
	addr := fmt.Sprintf("%s:%d", t.smtpHost, t.smtpPort)
	auth := smtp.PlainAuth("", t.smtpUsername, t.smtpPassword, t.smtpHost)

	err := smtp.SendMail(addr, auth, t.fromEmail, []string{params.ToEmail}, msg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return &SendEmailResult{
			Success: false,
			Message: fmt.Sprintf("Failed to send email: %v", err),
		}, nil
	}

	log.Printf("Email sent successfully to %s", params.ToEmail)
	return &SendEmailResult{
		Success: true,
		Message: fmt.Sprintf("Email sent successfully to %s", params.ToEmail),
	}, nil
}

// buildEmailMessage creates an RFC 2822 formatted email
func buildEmailMessage(from, to, subject, body string) []byte {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return []byte(msg.String())
}

