package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type MailMigrationChecklist struct {
	Domain    string   `json:"domain"`
	Provider  string   `json:"provider"`
	Completed []string `json:"completed"`
	Revision  int      `json:"revision"`
	UpdatedAt string   `json:"updatedAt"`
}
type MailMigrationClient struct{ client *Client }

// Audit reads public DNS only. DKIM selectors must be supplied explicitly.
func (m *MailMigrationClient) Audit(ctx context.Context, domain string, selectors []string) (map[string]any, error) {
	var out map[string]any
	q := url.Values{"domain": {domain}, "selectors": {strings.Join(selectors, ",")}}
	err := m.client.request(ctx, http.MethodGet, "/mail/audit?"+q.Encode(), nil, &out)
	return out, err
}
func (m *MailMigrationClient) Plan(ctx context.Context, domain, provider string, selectors []string) (map[string]any, error) {
	var out map[string]any
	q := url.Values{"domain": {domain}, "provider": {provider}, "selectors": {strings.Join(selectors, ",")}}
	err := m.client.request(ctx, http.MethodGet, "/mail/migration-plan?"+q.Encode(), nil, &out)
	return out, err
}
func (m *MailMigrationClient) Checklist(ctx context.Context, domain, provider string) (*MailMigrationChecklist, error) {
	var out MailMigrationChecklist
	q := url.Values{"domain": {domain}, "provider": {provider}}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/migration/checklist?"+q.Encode(), nil, &out)
	return &out, err
}
func (m *MailMigrationClient) SaveChecklist(ctx context.Context, input MailMigrationChecklist) (*MailMigrationChecklist, error) {
	var out MailMigrationChecklist
	err := m.client.mailRequest(ctx, http.MethodPut, "/mail/migration/checklist", map[string]any{"domain": input.Domain, "provider": input.Provider, "completed": input.Completed, "revision": input.Revision}, &out)
	return &out, err
}
