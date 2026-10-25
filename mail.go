package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type MailClient struct {
	Automations *MailAutomationsClient
	Forms       *MailFormsClient
	Segments    *MailSegmentsClient

	Billing        *MailBillingClient
	Migration      *MailMigrationClient
	Tracking       *MailTrackingClient
	client         *Client
	Domains        *MailDomainsClient
	Messages       *MailMessagesClient
	Suppressions   *MailSuppressionsClient
	Bounces        *MailBouncesClient
	OTP            *MailOTPClient
	Templates      *MailTemplatesClient
	Events         *MailEventsClient
	Webhooks       *MailWebhooksClient
	Streams        *MailStreamsClient
	Brands         *MailBrandsClient
	Audiences      *MailAudiencesClient
	Campaigns      *MailCampaignsClient
	Exports        *MailExportsClient
	MailboxDomains *MailboxDomainsClient
	Mailboxes      *MailboxesClient
	Aliases        *MailboxAliasesClient
	Groups         *MailboxGroupsClient
	Routes         *MailboxRoutesClient
	MailboxImports *MailboxImportsClient
	Limits         *MailLimitsClient
	Reputation     *MailReputationClient
	Scheduled      *MailScheduledClient
	Batch          *MailBatchClient
	Inbound        *MailInboundClient
	Analytics      *MailAnalyticsClient
}

type MailDomainsClient struct {
	client *Client
}

type MailMessagesClient struct {
	client *Client
}

type MailSuppressionsClient struct {
	client *Client
}

type MailBouncesClient struct {
	client *Client
}

type MailTemplatesClient struct {
	client *Client
}

type MailEventsClient struct {
	client *Client
}

type MailWebhooksClient struct {
	client *Client
}

type MailStreamsClient struct {
	client *Client
}

type MailLimitsClient struct {
	client *Client
}

type MailReputationClient struct {
	client *Client
	Events *MailReputationEventsClient
}

type MailReputationEventsClient struct {
	client *Client
}

type MailOTPClient struct {
	client     *Client
	Challenges *MailOTPChallengesClient
}

type MailOTPChallengesClient struct {
	client *Client
}

func (m *MailClient) Send(ctx context.Context, input MailSendRequest) (*MailSendResponse, error) {
	var response MailSendResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/send", input, &response)
	return &response, err
}

func (m *MailClient) SendTemplate(ctx context.Context, input MailSendTemplateRequest) (*MailSendTemplateResponse, error) {
	var response MailSendTemplateResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/send-template", input, &response)
	return &response, err
}

func (m *MailLimitsClient) Get(ctx context.Context) (*MailLimits, error) {
	var result MailLimits
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/limits", nil, &result)
	return &result, err
}

func (m *MailReputationClient) Get(ctx context.Context) (*MailReputation, error) {
	var result MailReputation
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/reputation", nil, &result)
	return &result, err
}

func (m *MailReputationEventsClient) List(ctx context.Context, options MailReputationEventListOptions) (*MailReputationEventListResult, error) {
	query := url.Values{}
	appendQuery(query, "severity", options.Severity)
	appendQuery(query, "type", options.Type)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	path := "/mail/reputation/events"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailReputationEventListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailDomainsClient) Create(ctx context.Context, domain string) (*MailDomain, error) {
	var result MailDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/domains", map[string]string{"domain": domain}, &result)
	return &result, err
}

func (m *MailDomainsClient) List(ctx context.Context) (*MailDomainListResult, error) {
	var result MailDomainListResult
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/domains", nil, &result)
	return &result, err
}

func (m *MailDomainsClient) Get(ctx context.Context, domainID string) (*MailDomain, error) {
	var result MailDomain
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/domains/"+url.PathEscape(domainID), nil, &result)
	return &result, err
}

func (m *MailDomainsClient) Verify(ctx context.Context, domainID string) (*MailDomain, error) {
	var result MailDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/domains/"+url.PathEscape(domainID)+"/verify", nil, &result)
	return &result, err
}

func (m *MailDomainsClient) Delete(ctx context.Context, domainID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/domains/"+url.PathEscape(domainID), nil, nil)
}

func (m *MailMessagesClient) List(ctx context.Context, options MailMessageListOptions) (*MailMessageListResult, error) {
	query := url.Values{}
	appendQuery(query, "status", options.Status)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	appendQuery(query, "to", options.To)
	appendQuery(query, "from", options.From)
	path := "/mail/messages"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailMessageListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailMessagesClient) Get(ctx context.Context, messageID string) (*MailMessage, error) {
	var message MailMessage
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/messages/"+url.PathEscape(messageID), nil, &message)
	return &message, err
}

func (m *MailMessagesClient) Attempts(ctx context.Context, messageID string) ([]MailSendAttempt, error) {
	var result struct {
		Data []MailSendAttempt `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/messages/"+url.PathEscape(messageID)+"/attempts", nil, &result)
	return result.Data, err
}

func (m *MailMessagesClient) Logs(ctx context.Context, messageID string) ([]MailLog, error) {
	var result struct {
		Data []MailLog `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/messages/"+url.PathEscape(messageID)+"/logs", nil, &result)
	return result.Data, err
}

func (m *MailMessagesClient) Bounces(ctx context.Context, messageID string) ([]MailBounce, error) {
	var result struct {
		Data []MailBounce `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/messages/"+url.PathEscape(messageID)+"/bounces", nil, &result)
	return result.Data, err
}

func (m *MailSuppressionsClient) List(ctx context.Context, options MailSuppressionListOptions) (*MailSuppressionListResult, error) {
	query := url.Values{}
	appendQuery(query, "email", options.Email)
	appendQuery(query, "reason", options.Reason)
	appendQuery(query, "brandId", options.BrandID)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	path := "/mail/suppressions"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailSuppressionListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailSuppressionsClient) Create(ctx context.Context, input MailSuppressionCreateRequest) (*MailSuppression, error) {
	var result MailSuppression
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/suppressions", input, &result)
	return &result, err
}

func (m *MailSuppressionsClient) Delete(ctx context.Context, suppressionID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/suppressions/"+url.PathEscape(suppressionID), nil, nil)
}

func (m *MailBouncesClient) List(ctx context.Context, options MailBounceListOptions) (*MailBounceListResult, error) {
	query := url.Values{}
	appendQuery(query, "email", options.Email)
	appendQuery(query, "type", options.Type)
	appendQuery(query, "messageId", options.MessageID)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	path := "/mail/bounces"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailBounceListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailTemplatesClient) Create(ctx context.Context, input MailTemplateCreateRequest) (*MailTemplate, error) {
	var result MailTemplate
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/templates", input, &result)
	return &result, err
}

func (m *MailTemplatesClient) List(ctx context.Context, options MailTemplateListOptions) (*MailTemplateListResult, error) {
	query := url.Values{}
	appendQuery(query, "status", options.Status)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	appendQuery(query, "search", options.Search)
	path := "/mail/templates"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailTemplateListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailTemplatesClient) Get(ctx context.Context, templateIDOrSlug string) (*MailTemplateDetail, error) {
	var result MailTemplateDetail
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/templates/"+url.PathEscape(templateIDOrSlug), nil, &result)
	return &result, err
}

func (m *MailTemplatesClient) Update(ctx context.Context, templateID string, input MailTemplateUpdateRequest) (*MailTemplate, error) {
	var result MailTemplate
	err := m.client.mailRequest(ctx, http.MethodPatch, "/mail/templates/"+url.PathEscape(templateID), input, &result)
	return &result, err
}

func (m *MailTemplatesClient) Delete(ctx context.Context, templateID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/templates/"+url.PathEscape(templateID), nil, nil)
}

func (m *MailTemplatesClient) Versions(ctx context.Context, templateID string) ([]MailTemplateVersion, error) {
	var result struct {
		Data []MailTemplateVersion `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/templates/"+url.PathEscape(templateID)+"/versions", nil, &result)
	return result.Data, err
}

func (m *MailTemplatesClient) GetVersion(ctx context.Context, templateID, versionID string) (*MailTemplateVersion, error) {
	var result MailTemplateVersion
	path := "/mail/templates/" + url.PathEscape(templateID) + "/versions/" + url.PathEscape(versionID)
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailTemplatesClient) ActivateVersion(ctx context.Context, templateID, versionID string) (*MailTemplate, error) {
	var result MailTemplate
	path := "/mail/templates/" + url.PathEscape(templateID) + "/versions/" + url.PathEscape(versionID) + "/activate"
	err := m.client.mailRequest(ctx, http.MethodPost, path, nil, &result)
	return &result, err
}

func (m *MailTemplatesClient) Preview(ctx context.Context, templateID string, input MailTemplatePreviewRequest) (*MailTemplatePreview, error) {
	var result MailTemplatePreview
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/templates/"+url.PathEscape(templateID)+"/preview", input, &result)
	return &result, err
}

func (m *MailTemplatesClient) TestSend(ctx context.Context, templateID string, input MailTemplateTestSendRequest) (*MailTemplateTestSendResponse, error) {
	var result MailTemplateTestSendResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/templates/"+url.PathEscape(templateID)+"/test-send", input, &result)
	return &result, err
}

func (m *MailOTPClient) Send(ctx context.Context, input MailOTPSendRequest) (*MailOTPSendResponse, error) {
	var response MailOTPSendResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/otp/send", input, &response)
	return &response, err
}

func (m *MailOTPClient) Verify(ctx context.Context, input MailOTPVerifyRequest) (*MailOTPVerifyResponse, error) {
	var response MailOTPVerifyResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/otp/verify", input, &response)
	return &response, err
}

func (m *MailOTPChallengesClient) List(ctx context.Context, options MailOTPChallengeListOptions) (*MailOTPChallengeListResult, error) {
	query := url.Values{}
	appendQuery(query, "to", options.To)
	appendQuery(query, "purpose", options.Purpose)
	appendQuery(query, "status", options.Status)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	path := "/mail/otp/challenges"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailOTPChallengeListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailOTPChallengesClient) Get(ctx context.Context, challengeID string) (*MailOTPChallenge, error) {
	var result MailOTPChallenge
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/otp/challenges/"+url.PathEscape(challengeID), nil, &result)
	return &result, err
}

func (m *MailOTPChallengesClient) Cancel(ctx context.Context, challengeID string) (*MailOTPChallenge, error) {
	var result MailOTPChallenge
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/otp/challenges/"+url.PathEscape(challengeID)+"/cancel", nil, &result)
	return &result, err
}

func appendQuery(values url.Values, key string, value any) {
	switch v := value.(type) {
	case string:
		if v != "" {
			values.Set(key, v)
		}
	case int:
		if v > 0 {
			values.Set(key, strconv.Itoa(v))
		}
	}
}
