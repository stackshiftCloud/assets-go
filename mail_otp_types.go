package stackshift

import "time"

type MailOTPSendRequest struct {
	Simulation     *MailSimulationOptions `json:"simulation,omitempty"`
	To             string                 `json:"to"`
	Purpose        string                 `json:"purpose"`
	From           string                 `json:"from,omitempty"`
	ExpiresIn      string                 `json:"expiresIn,omitempty"`
	CodeLength     int                    `json:"codeLength,omitempty"`
	Metadata       map[string]any         `json:"metadata,omitempty"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
	Subject        string                 `json:"subject,omitempty"`
	BrandName      string                 `json:"brandName,omitempty"`
}

type MailOTPVerifyRequest struct {
	To          string `json:"to,omitempty"`
	Purpose     string `json:"purpose,omitempty"`
	ChallengeID string `json:"challengeId,omitempty"`
	Code        string `json:"code"`
}

type MailOTPChallengeListOptions struct {
	To      string
	Purpose string
	Status  string
	Limit   int
	Cursor  string
}

type MailOTPSendResponse struct {
	Mode              MailMode  `json:"mode"`
	ID                string    `json:"id"`
	To                string    `json:"to"`
	Purpose           string    `json:"purpose"`
	Status            string    `json:"status"`
	ExpiresAt         time.Time `json:"expiresAt"`
	AttemptsRemaining int       `json:"attemptsRemaining"`
	ResendAvailableAt time.Time `json:"resendAvailableAt"`
	MessageID         *string   `json:"messageId,omitempty"`
	IdempotencyStatus string    `json:"idempotencyStatus,omitempty"`
}

type MailOTPVerifyResponse struct {
	Verified    bool   `json:"verified"`
	ChallengeID string `json:"challengeId"`
	Status      string `json:"status"`
}

type MailOTPChallenge struct {
	Mode              MailMode       `json:"mode"`
	ID                string         `json:"id"`
	To                string         `json:"to"`
	Purpose           string         `json:"purpose"`
	Status            string         `json:"status"`
	Attempts          int            `json:"attempts"`
	MaxAttempts       int            `json:"maxAttempts"`
	AttemptsRemaining int            `json:"attemptsRemaining,omitempty"`
	ExpiresAt         time.Time      `json:"expiresAt"`
	ResendAvailableAt *time.Time     `json:"resendAvailableAt,omitempty"`
	MessageID         *string        `json:"messageId,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
	VerifiedAt        *time.Time     `json:"verifiedAt,omitempty"`
	FailedAt          *time.Time     `json:"failedAt,omitempty"`
	CanceledAt        *time.Time     `json:"canceledAt,omitempty"`
}

type MailOTPChallengeListResult struct {
	Data       []MailOTPChallenge `json:"data"`
	NextCursor string             `json:"nextCursor,omitempty"`
}
