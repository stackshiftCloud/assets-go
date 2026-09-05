package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type MailTrackingOptions struct {
	Opens  *bool `json:"opens,omitempty"`
	Clicks *bool `json:"clicks,omitempty"`
}
type MailTrackingSettings struct {
	Opens  bool `json:"opens"`
	Clicks bool `json:"clicks"`
}
type MailEngagementCount struct {
	Kind               string     `json:"kind"`
	Variant            string     `json:"variant"`
	Link               string     `json:"link,omitempty"`
	Total              int64      `json:"total"`
	Unique             int64      `json:"unique"`
	SuspectedAutomated int64      `json:"suspectedAutomated"`
	FirstAt            *time.Time `json:"firstAt,omitempty"`
	LastAt             *time.Time `json:"lastAt,omitempty"`
}
type MailEngagementReport struct {
	OpenEngagedDelivered  int64                 `json:"openEngagedDelivered"`
	ClickEngagedDelivered int64                 `json:"clickEngagedDelivered"`
	Counts                []MailEngagementCount `json:"counts"`
	OpenEligible          int64                 `json:"openEligible"`
	ClickEligible         int64                 `json:"clickEligible"`
	OpenDelivered         int64                 `json:"openDelivered"`
	ClickDelivered        int64                 `json:"clickDelivered"`
	Scope                 string                `json:"scope"`
	RetentionDays         int                   `json:"retentionDays"`
}
type MailTrackingDomain struct {
	ID             string     `json:"id"`
	DomainID       string     `json:"domainId"`
	Hostname       string     `json:"hostname"`
	OwnershipToken string     `json:"ownershipToken"`
	Enabled        bool       `json:"enabled"`
	Status         string     `json:"status"`
	CheckedAt      *time.Time `json:"checkedAt"`
	DNSValidUntil  *time.Time `json:"dnsValidUntil"`
}
type MailTrackingSettingsResponse struct {
	Settings  MailTrackingSettings `json:"settings"`
	Available bool                 `json:"available"`
}
type MailTrackingDomainsResponse struct {
	Data      []MailTrackingDomain `json:"data"`
	Target    string               `json:"target"`
	Available bool                 `json:"available"`
}
type MailTrackingClient struct{ client *Client }

func (m *MailTrackingClient) Settings(ctx context.Context) (*MailTrackingSettingsResponse, error) {
	var out MailTrackingSettingsResponse
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/tracking/settings", nil, &out)
	return &out, err
}
func (m *MailTrackingClient) UpdateSettings(ctx context.Context, input MailTrackingSettings) (*MailTrackingSettings, error) {
	var out MailTrackingSettings
	err := m.client.mailRequest(ctx, http.MethodPut, "/mail/tracking/settings", input, &out)
	return &out, err
}
func (m *MailTrackingClient) Message(ctx context.Context, id string) (*MailEngagementReport, error) {
	var out MailEngagementReport
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/messages/"+url.PathEscape(id)+"/engagement", nil, &out)
	return &out, err
}
func (m *MailTrackingClient) Campaign(ctx context.Context, id string) (*MailEngagementReport, error) {
	var out MailEngagementReport
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/campaigns/"+url.PathEscape(id)+"/engagement", nil, &out)
	return &out, err
}
func (m *MailTrackingClient) Domains(ctx context.Context) (*MailTrackingDomainsResponse, error) {
	var out MailTrackingDomainsResponse
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/tracking/domains", nil, &out)
	return &out, err
}
func (m *MailTrackingClient) CreateDomain(ctx context.Context, domainID, hostname string) (*MailTrackingDomain, error) {
	var out MailTrackingDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/tracking/domains", map[string]string{"domainId": domainID, "hostname": hostname}, &out)
	return &out, err
}
func (m *MailTrackingClient) VerifyDomain(ctx context.Context, id string) (*MailTrackingDomain, error) {
	var out MailTrackingDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/tracking/domains/"+url.PathEscape(id)+"/verify", nil, &out)
	return &out, err
}
func (m *MailTrackingClient) UpdateDomain(ctx context.Context, id string, enabled, revoke bool) error {
	return m.client.mailRequest(ctx, http.MethodPut, "/mail/tracking/domains/"+url.PathEscape(id), map[string]bool{"enabled": enabled, "revoke": revoke}, nil)
}
