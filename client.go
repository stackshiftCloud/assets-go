package stackshift

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var errorCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func topLevelErrorFields(raw []byte) (string, map[string]any) {
	var body map[string]any
	if json.Unmarshal(raw, &body) != nil {
		return "", nil
	}
	code, _ := body["code"].(string)
	if text, ok := body["error"].(string); ok && code == "" && errorCodePattern.MatchString(text) {
		code = text
	}
	details := map[string]any{}
	for key, value := range body {
		switch key {
		case "success", "status", "status_code", "message", "error", "code":
		default:
			details[key] = value
		}
	}
	if len(details) == 0 {
		return code, nil
	}
	return code, details
}

const DefaultBaseURL = "https://api.stackshift.cloud/api/v1"
const DefaultMailBaseURL = "https://api.stackshift.cloud/api/v1"
const DefaultCDNBaseURL = "https://cdn.stackshift.cloud"

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	apiKey       string
	assetSpaceID string
	baseURL      string
	mailBaseURL  string
	cdnBaseURL   string
	httpClient   HTTPDoer

	Agents           *AgentsClient
	AICredits        *AICreditsClient
	Assets           *AssetsClient
	Databases        *DatabasesClient
	Mail             *MailClient
	Partner          *PartnerClient
	WorkloadSecurity *WorkloadSecurityClient
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Message envelopeMessage `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type envelopeMessage string

func (m *envelopeMessage) UnmarshalJSON(raw []byte) error {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		*m = envelopeMessage(text)
	}
	return nil
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]any
}

func (e *APIError) Error() string {
	return fmt.Sprintf("stackshift: API request failed with status %d: %s", e.StatusCode, e.Message)
}

func New(options Options) (*Client, error) {
	apiKey := strings.TrimSpace(options.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("STACKSHIFT_API_KEY"))
	}
	if apiKey == "" {
		return nil, errors.New("stackshift: API key is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(options.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	mailBaseURL := strings.TrimRight(strings.TrimSpace(options.MailBaseURL), "/")
	if mailBaseURL == "" {
		mailBaseURL = DefaultMailBaseURL
	}
	cdnBaseURL := strings.TrimRight(strings.TrimSpace(options.CDNBaseURL), "/")
	if cdnBaseURL == "" {
		cdnBaseURL = DefaultCDNBaseURL
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := &Client{
		apiKey:      apiKey,
		baseURL:     baseURL,
		mailBaseURL: mailBaseURL,
		cdnBaseURL:  cdnBaseURL,
		httpClient:  httpClient,
	}
	client.Assets = &AssetsClient{client: client}
	client.Assets.Connections = &AssetConnectionsClient{client: client}
	client.Assets.Storage = &AssetStorageClient{client: client}
	client.Assets.Migrations = &AssetMigrationsClient{client: client}
	client.Agents = &AgentsClient{}
	client.Agents.Connections = &AgentConnectionsClient{client: client, Operations: &AgentConnectionOperationsClient{client: client}}
	client.AICredits = &AICreditsClient{client: client}
	client.Databases = &DatabasesClient{client: client}
	client.Databases.Branches = &DatabaseBranchesClient{client: client}
	client.Partner = &PartnerClient{client: client}
	client.WorkloadSecurity = &WorkloadSecurityClient{client: client}
	client.Mail = &MailClient{
		client:       client,
		Domains:      &MailDomainsClient{client: client},
		Messages:     &MailMessagesClient{client: client},
		Suppressions: &MailSuppressionsClient{client: client},
		Bounces:      &MailBouncesClient{client: client},
		Templates:    &MailTemplatesClient{client: client},
		Events:       &MailEventsClient{client: client},
		Webhooks:     &MailWebhooksClient{client: client},
		Streams:      &MailStreamsClient{client: client},
		Billing:      &MailBillingClient{client: client},
		Automations:  &MailAutomationsClient{client: client},
		Forms:        &MailFormsClient{client: client},
		Segments:     &MailSegmentsClient{client: client},

		Migration:      &MailMigrationClient{client: client},
		Brands:         &MailBrandsClient{client: client},
		Audiences:      &MailAudiencesClient{client: client},
		Campaigns:      &MailCampaignsClient{client: client},
		Tracking:       &MailTrackingClient{client: client},
		Exports:        &MailExportsClient{client: client},
		MailboxDomains: &MailboxDomainsClient{client: client},
		Mailboxes:      &MailboxesClient{client: client},
		Agreements:     &MailAgreementsClient{client: client},
		Aliases:        &MailboxAliasesClient{client: client},
		Groups:         &MailboxGroupsClient{client: client},
		Routes:         &MailboxRoutesClient{client: client},
		MailboxImports: &MailboxImportsClient{client: client},
		Limits:         &MailLimitsClient{client: client},
		Scheduled:      &MailScheduledClient{client: client},
		Batch:          &MailBatchClient{client: client},
		Analytics:      &MailAnalyticsClient{client: client},
	}
	client.Mail.AgentInboxes = &AgentInboxesClient{client: client, now: time.Now, Messages: &AgentInboxMessagesClient{client: client}}
	client.Mail.TeamInboxes = &TeamInboxesClient{
		client:         client,
		Conversations:  &TeamInboxConversationsClient{client: client},
		PendingReplies: &TeamInboxPendingRepliesClient{client: client},
		Activity:       &TeamInboxActivityClient{client: client},
		Settings:       &TeamInboxSettingsClient{client: client},
	}
	client.Mail.Inbound = &MailInboundClient{
		client:   client,
		Domains:  &MailInboundDomainsClient{client: client},
		Messages: &MailInboundMessagesClient{client: client},
	}
	client.Mail.Reputation = &MailReputationClient{
		client: client,
		Events: &MailReputationEventsClient{client: client},
	}
	client.Mail.OTP = &MailOTPClient{
		client:     client,
		Challenges: &MailOTPChallengesClient{client: client},
	}
	return client, nil
}

func (c *Client) request(ctx context.Context, method, path string, body any, out any) error {
	return c.requestWithBaseHeaders(ctx, c.baseURL, method, path, body, nil, out)
}

func (c *Client) requestWithHeaders(ctx context.Context, method, path string, body any, headers map[string]string, out any) error {
	return c.requestWithBaseHeaders(ctx, c.baseURL, method, path, body, headers, out)
}

func (c *Client) mailRequest(ctx context.Context, method, path string, body any, out any) error {
	return c.requestWithBase(ctx, c.mailBaseURL, method, path, body, out)
}

func (c *Client) mailRequestWithHeaders(ctx context.Context, method, path string, body any, headers map[string]string, out any) error {
	return c.requestWithBaseHeaders(ctx, c.mailBaseURL, method, path, body, headers, out)
}

func (c *Client) requestWithBase(ctx context.Context, baseURL, method, path string, body any, out any) error {
	return c.requestWithBaseHeaders(ctx, baseURL, method, path, body, nil, out)
}

func (c *Client) requestWithBaseHeaders(ctx context.Context, baseURL, method, path string, body any, headers map[string]string, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(stripEmpty(body))
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "StackShift-Go/0.1")
	if c.assetSpaceID != "" && strings.HasPrefix(path, "/assets") {
		req.Header.Set("X-Asset-Space-ID", c.assetSpaceID)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	var envelope apiEnvelope
	var publicErr struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &envelope); err != nil {
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return &APIError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(raw))}
			}
			return err
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = json.Unmarshal(raw, &publicErr)
		if publicErr.Error.Message != "" || publicErr.Error.Code != "" {
			return &APIError{
				StatusCode: resp.StatusCode, Code: publicErr.Error.Code,
				Message: firstNonEmpty(publicErr.Error.Message, string(envelope.Message)), Details: publicErr.Error.Details,
			}
		}
		if code, details := topLevelErrorFields(raw); code != "" || details != nil {
			return &APIError{StatusCode: resp.StatusCode, Code: code, Message: firstNonEmpty(string(envelope.Message), http.StatusText(resp.StatusCode)), Details: details}
		}
		message := string(envelope.Message)
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return &APIError{StatusCode: resp.StatusCode, Message: message}
	}
	if out != nil && len(envelope.Data) > 0 && bytes.Contains(raw, []byte(`"success"`)) {
		return json.Unmarshal(envelope.Data, out)
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}
