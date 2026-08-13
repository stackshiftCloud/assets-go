package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type MailScheduledClient struct{ client *Client }
type MailBatchClient struct{ client *Client }
type MailAnalyticsClient struct{ client *Client }
type MailInboundClient struct {
	client   *Client
	Domains  *MailInboundDomainsClient
	Messages *MailInboundMessagesClient
}
type MailInboundDomainsClient struct{ client *Client }
type MailInboundMessagesClient struct{ client *Client }

func (m *MailScheduledClient) Create(ctx context.Context, input MailScheduledCreateRequest) (*MailScheduledMessage, error) {
	var result MailScheduledMessage
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/scheduled", input, &result)
	return &result, err
}

func (m *MailScheduledClient) List(ctx context.Context) (*MailScheduledListResult, error) {
	var result MailScheduledListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/scheduled", nil, &result)
	return &result, err
}

func (m *MailScheduledClient) Get(ctx context.Context, id string) (*MailScheduledMessage, error) {
	var result MailScheduledMessage
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/scheduled/"+url.PathEscape(id), nil, &result)
	return &result, err
}

func (m *MailScheduledClient) Cancel(ctx context.Context, id string) (*MailScheduledMessage, error) {
	var result MailScheduledMessage
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/scheduled/"+url.PathEscape(id)+"/cancel", nil, &result)
	return &result, err
}

func (m *MailBatchClient) Send(ctx context.Context, input MailBatchSendRequest) (*MailBatch, error) {
	var result MailBatch
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/batch", input, &result)
	return &result, err
}

func (m *MailBatchClient) SendTemplate(ctx context.Context, input MailBatchTemplateRequest) (*MailBatch, error) {
	var result MailBatch
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/batch-template", input, &result)
	return &result, err
}

func (m *MailBatchClient) List(ctx context.Context) (*MailBatchListResult, error) {
	var result MailBatchListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/batches", nil, &result)
	return &result, err
}

func (m *MailBatchClient) Get(ctx context.Context, id string) (*MailBatch, error) {
	var result MailBatch
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/batches/"+url.PathEscape(id), nil, &result)
	return &result, err
}

func (m *MailBatchClient) Items(ctx context.Context, id string) (*MailBatchItemListResult, error) {
	var result MailBatchItemListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/batches/"+url.PathEscape(id)+"/items", nil, &result)
	return &result, err
}

func (m *MailInboundDomainsClient) Create(ctx context.Context, domain string) (*MailInboundDomain, error) {
	var result MailInboundDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/inbound/domains", map[string]string{"domain": domain}, &result)
	return &result, err
}

func (m *MailInboundDomainsClient) List(ctx context.Context) (*MailInboundDomainListResult, error) {
	var result MailInboundDomainListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/inbound/domains", nil, &result)
	return &result, err
}

func (m *MailInboundDomainsClient) Verify(ctx context.Context, id string) (*MailInboundDomain, error) {
	var result MailInboundDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/inbound/domains/"+url.PathEscape(id)+"/verify", nil, &result)
	return &result, err
}

func (m *MailInboundMessagesClient) List(ctx context.Context) (*MailInboundMessageListResult, error) {
	var result MailInboundMessageListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/inbound/messages", nil, &result)
	return &result, err
}

func (m *MailInboundMessagesClient) Get(ctx context.Context, id string) (*MailInboundMessage, error) {
	var result MailInboundMessage
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/inbound/messages/"+url.PathEscape(id), nil, &result)
	return &result, err
}

func (m *MailAnalyticsClient) Get(ctx context.Context, options MailAnalyticsOptions) (*MailAnalytics, error) {
	query := url.Values{}
	appendQuery(query, "range", options.Range)
	if !options.From.IsZero() {
		query.Set("from", options.From.Format(time.RFC3339))
	}
	if !options.To.IsZero() {
		query.Set("to", options.To.Format(time.RFC3339))
	}
	appendQuery(query, "interval", options.Interval)
	appendQuery(query, "domain", options.Domain)
	appendQuery(query, "template", options.Template)
	appendQuery(query, "event", options.Event)
	path := "/mail/analytics"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailAnalytics
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}
