package scheduler

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"agent/internal/repositories"
)

// DigestService sends daily automation digest emails
type DigestService struct {
	automationRepo  *repositories.AutomationRepository
	proposalRepo    *repositories.ProposalRepository
	smtpHost        string
	smtpPort        int
	smtpUsername    string
	smtpPassword    string
	fromEmail       string
	anthropicAPIKey string
}

// NewDigestService creates a new digest service
func NewDigestService(
	automationRepo *repositories.AutomationRepository,
	proposalRepo *repositories.ProposalRepository,
	smtpHost string,
	smtpPort int,
	smtpUsername string,
	smtpPassword string,
	fromEmail string,
	anthropicAPIKey string,
) *DigestService {
	if fromEmail == "" {
		fromEmail = smtpUsername
	}
	return &DigestService{
		automationRepo:  automationRepo,
		proposalRepo:    proposalRepo,
		smtpHost:        smtpHost,
		smtpPort:        smtpPort,
		smtpUsername:    smtpUsername,
		smtpPassword:    smtpPassword,
		fromEmail:       fromEmail,
		anthropicAPIKey: anthropicAPIKey,
	}
}

// SendDailyDigests sends digest emails for all enabled automations
func (s *DigestService) SendDailyDigests() {
	if s.smtpUsername == "" || s.smtpPassword == "" {
		log.Println("Digest service: SMTP not configured, skipping digests")
		return
	}

	now := time.Now()
	configs, err := s.automationRepo.GetDigestDueConfigs(now)
	if err != nil {
		log.Printf("Digest service: failed to get due configs: %v", err)
		return
	}

	if len(configs) == 0 {
		return
	}

	since := now.Add(-24 * time.Hour)
	sentCount := 0

	for i := range configs {
		cfg := &configs[i]
		if !s.isDigestTimeNow(cfg, now) {
			continue
		}
		if cfg.DigestEmails == nil || *cfg.DigestEmails == "" {
			continue
		}
		s.sendDigest(cfg, since)
		s.automationRepo.UpdateLastDigest(cfg.ID, now)
		sentCount++
	}

	if sentCount > 0 {
		log.Printf("Digest service: sent %d digests", sentCount)
	}
}

func (s *DigestService) isDigestTimeNow(cfg *repositories.AutomationConfigDTO, now time.Time) bool {
	digestTime := "09:00"
	if cfg.DigestTime != nil && *cfg.DigestTime != "" {
		digestTime = *cfg.DigestTime
	}

	currentTime := now.Format("15:04")

	// Check if current time is within 30 minute window of digest time
	targetHour, targetMin := parseTime(digestTime)
	currentHour, currentMin := parseTime(currentTime)

	targetMins := targetHour*60 + targetMin
	currentMins := currentHour*60 + currentMin

	diff := currentMins - targetMins
	if diff < 0 {
		diff = -diff
	}

	return diff <= 30
}

func parseTime(t string) (int, int) {
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return 9, 0
	}
	hour := 0
	min := 0
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &min)
	return hour, min
}

func (s *DigestService) sendDigest(cfg *repositories.AutomationConfigDTO, since time.Time) {
	proposals, err := s.proposalRepo.GetAutomationProposals(cfg.TenantID, cfg.UserID, since)
	if err != nil {
		log.Printf("Digest service: failed to get proposals for config %d: %v", cfg.ID, err)
		return
	}

	// Group proposals by status
	pending, approved, rejected := s.groupProposalsByStatus(proposals)

	// Build email body - use LLM if model configured
	body := s.buildDigestBody(cfg, pending, approved, rejected)
	subject := fmt.Sprintf("Job Lead Automation Digest - %s", time.Now().Format("Jan 2, 2006"))

	// Send to all configured emails
	s.sendToRecipients(cfg.DigestEmails, subject, body)
}

func (s *DigestService) groupProposalsByStatus(proposals []repositories.AutomationProposalDTO) (pending, approved, rejected []repositories.AutomationProposalDTO) {
	for _, p := range proposals {
		if p.Status == "pending" {
			pending = append(pending, p)
		}
		if p.Status == "approved" {
			approved = append(approved, p)
		}
		if p.Status == "rejected" {
			rejected = append(rejected, p)
		}
	}
	return
}

func (s *DigestService) sendToRecipients(digestEmails *string, subject, body string) {
	if digestEmails == nil {
		return
	}
	emails := strings.Split(*digestEmails, ",")
	for _, email := range emails {
		email = strings.TrimSpace(email)
		s.trySendEmail(email, subject, body)
	}
}

func (s *DigestService) trySendEmail(email, subject, body string) {
	if email == "" {
		return
	}
	if err := s.sendEmail(email, subject, body); err != nil {
		log.Printf("Digest service: failed to send to %s: %v", email, err)
		return
	}
	log.Printf("Digest service: sent digest to %s", email)
}

func (s *DigestService) buildDigestBody(cfg *repositories.AutomationConfigDTO, pending, approved, rejected []repositories.AutomationProposalDTO) string {
	rawBody := s.buildRawDigestBody(pending, approved, rejected)

	if cfg.DigestModel == nil || *cfg.DigestModel == "" || s.anthropicAPIKey == "" {
		return rawBody
	}

	llmBody := s.generateLLMDigest(cfg, rawBody, pending, approved, rejected)
	if llmBody == "" {
		return rawBody
	}
	return llmBody
}

func (s *DigestService) buildRawDigestBody(pending, approved, rejected []repositories.AutomationProposalDTO) string {
	var b strings.Builder

	b.WriteString("Job Lead Automation Daily Digest\n")
	b.WriteString("================================\n\n")

	b.WriteString("Summary:\n")
	b.WriteString(fmt.Sprintf("- Pending Review: %d\n", len(pending)))
	b.WriteString(fmt.Sprintf("- Approved: %d\n", len(approved)))
	b.WriteString(fmt.Sprintf("- Rejected: %d\n\n", len(rejected)))

	s.appendProposalSection(&b, "Pending Review", pending)
	s.appendProposalSection(&b, "Approved", approved)
	s.appendProposalSection(&b, "Rejected", rejected)

	b.WriteString("---\n")
	b.WriteString("This is an automated digest from your Job Lead Sourcing Bot.\n")

	return b.String()
}

func (s *DigestService) appendProposalSection(b *strings.Builder, title string, proposals []repositories.AutomationProposalDTO) {
	if len(proposals) == 0 {
		return
	}
	b.WriteString(title + ":\n")
	b.WriteString(strings.Repeat("-", len(title)) + "\n")
	for _, p := range proposals {
		b.WriteString(fmt.Sprintf("- %s at %s\n", p.JobTitle, p.Account))
	}
	b.WriteString("\n")
}

func (s *DigestService) generateLLMDigest(cfg *repositories.AutomationConfigDTO, rawBody string, pending, approved, rejected []repositories.AutomationProposalDTO) string {
	model := *cfg.DigestModel

	prompt := s.buildDigestPrompt(pending, approved, rejected)

	client := anthropic.NewClient(option.WithAPIKey(s.anthropicAPIKey))
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: "You are a helpful assistant that creates concise, insightful daily digest emails for job search automation. Write in a professional but friendly tone. Keep it brief but informative."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		log.Printf("Digest service: LLM generation failed: %v", err)
		return ""
	}

	return extractDigestText(resp)
}

func (s *DigestService) buildDigestPrompt(pending, approved, rejected []repositories.AutomationProposalDTO) string {
	var b strings.Builder
	b.WriteString("Create a daily digest email summarizing these job leads:\n\n")

	b.WriteString(fmt.Sprintf("PENDING REVIEW (%d):\n", len(pending)))
	for _, p := range pending {
		b.WriteString(fmt.Sprintf("- %s at %s\n", p.JobTitle, p.Account))
	}

	b.WriteString(fmt.Sprintf("\nAPPROVED (%d):\n", len(approved)))
	for _, p := range approved {
		b.WriteString(fmt.Sprintf("- %s at %s\n", p.JobTitle, p.Account))
	}

	b.WriteString(fmt.Sprintf("\nREJECTED (%d):\n", len(rejected)))
	for _, p := range rejected {
		b.WriteString(fmt.Sprintf("- %s at %s\n", p.JobTitle, p.Account))
	}

	b.WriteString("\nWrite a brief, professional digest email with:\n")
	b.WriteString("1. A quick summary of activity\n")
	b.WriteString("2. Any notable patterns or insights\n")
	b.WriteString("3. Action items if there are pending reviews\n")
	b.WriteString("\nKeep it under 300 words. Use plain text format.")

	return b.String()
}

func extractDigestText(resp *anthropic.Message) string {
	if resp == nil || len(resp.Content) == 0 {
		return ""
	}
	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return strings.TrimSpace(text)
}

func (s *DigestService) sendEmail(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.fromEmail, to, subject, body)

	return smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(msg))
}
