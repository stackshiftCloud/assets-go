package stackshift

import "time"

type MailLimits struct {
	Tier               string `json:"tier"`
	Status             string `json:"status"`
	DailyLimit         int    `json:"dailyLimit"`
	DailyUsed          int    `json:"dailyUsed"`
	HourlyLimit        int    `json:"hourlyLimit"`
	HourlyUsed         int    `json:"hourlyUsed"`
	PerMinuteLimit     int    `json:"perMinuteLimit"`
	CurrentWarmupStage int    `json:"currentWarmupStage"`
	ManualReviewStatus string `json:"manualReviewStatus"`
}

type MailReputationWorkspace struct {
	Status          string   `json:"status"`
	Score           int      `json:"score"`
	Tier            string   `json:"tier"`
	BounceRate      float64  `json:"bounceRate"`
	DeferralRate    float64  `json:"deferralRate"`
	SuppressionRate float64  `json:"suppressionRate"`
	ComplaintRate   *float64 `json:"complaintRate"`
}

type MailDomainReputation struct {
	Domain             string   `json:"domain"`
	Status             string   `json:"status"`
	Score              int      `json:"score"`
	BounceRate         float64  `json:"bounceRate"`
	DeferralRate       float64  `json:"deferralRate"`
	ComplaintRate      *float64 `json:"complaintRate"`
	CurrentWarmupStage int      `json:"currentWarmupStage"`
}

type MailReputation struct {
	Workspace MailReputationWorkspace `json:"workspace"`
	Domains   []MailDomainReputation  `json:"domains"`
}

type MailReputationEvent struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Severity  string         `json:"severity"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type MailReputationEventListOptions struct {
	Severity string
	Type     string
	Limit    int
	Cursor   string
}

type MailReputationEventListResult struct {
	Data       []MailReputationEvent `json:"data"`
	NextCursor string                `json:"nextCursor,omitempty"`
}
