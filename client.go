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
	"strings"
)

const DefaultBaseURL = "https://api.stackshift.cloud/api/v1"
const DefaultMailBaseURL = "https://api.stackshift.cloud/v1"
const DefaultCDNBaseURL = "https://cdn.stackshift.cloud"

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	apiKey      string
	baseURL     string
	mailBaseURL string
	cdnBaseURL  string
	httpClient  HTTPDoer

	Assets *AssetsClient
	Mail   *MailClient
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
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
		Audiences:    &MailAudiencesClient{client: client},
		Campaigns:    &MailCampaignsClient{client: client},
		Exports:      &MailExportsClient{client: client},
		Limits:       &MailLimitsClient{client: client},
		Scheduled:    &MailScheduledClient{client: client},
		Batch:        &MailBatchClient{client: client},
		Analytics:    &MailAnalyticsClient{client: client},
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
				Message: publicErr.Error.Message, Details: publicErr.Error.Details,
			}
		}
		message := envelope.Message
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
