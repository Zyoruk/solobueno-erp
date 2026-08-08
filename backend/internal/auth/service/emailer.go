package service

import (
	"context"
	"log"

	"github.com/google/uuid"
)

// Emailer defines the interface for sending transactional auth emails
// (temporary passwords, tenant-link notifications, password resets).
// Real delivery (AWS SES per the project's stack) is a future integration;
// LogEmailer is the dev-safe default until that adapter is wired in.
type Emailer interface {
	// SendTemporaryPassword notifies a newly created user of their temporary password.
	SendTemporaryPassword(ctx context.Context, toEmail, tempPassword string) error

	// SendTenantLinked notifies an existing user that a new tenant role was added to their account.
	SendTenantLinked(ctx context.Context, toEmail string, tenantID uuid.UUID) error
}

// LogEmailer is a stub Emailer that logs instead of sending real email.
// Safe default for dev/test; swap for an SES-backed implementation in production.
type LogEmailer struct{}

// NewLogEmailer creates a new LogEmailer.
func NewLogEmailer() *LogEmailer {
	return &LogEmailer{}
}

func (e *LogEmailer) SendTemporaryPassword(ctx context.Context, toEmail, tempPassword string) error {
	log.Printf("[email stub] temporary password for %s: %s", toEmail, tempPassword)
	return nil
}

func (e *LogEmailer) SendTenantLinked(ctx context.Context, toEmail string, tenantID uuid.UUID) error {
	log.Printf("[email stub] %s was added to tenant %s", toEmail, tenantID)
	return nil
}

var _ Emailer = (*LogEmailer)(nil)
