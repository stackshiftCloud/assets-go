package stackshift

import "time"

type MailSendingDomainHealth struct {
	State         string         `json:"state"`
	LastResult    string         `json:"lastResult"`
	LastCheckedAt *time.Time     `json:"lastCheckedAt"`
	LastSuccessAt *time.Time     `json:"lastSuccessAt"`
	NextCheckAt   time.Time      `json:"nextCheckAt"`
	DegradedAt    *time.Time     `json:"degradedAt"`
	Issues        []MailDNSIssue `json:"issues"`
}

type MailDNSIssue struct {
	Type     string   `json:"type,omitempty"`
	Record   string   `json:"record"`
	Name     string   `json:"name"`
	Expected string   `json:"expected"`
	Observed []string `json:"observed"`
	Status   string   `json:"status"`
	Scope    string   `json:"scope"`
}
