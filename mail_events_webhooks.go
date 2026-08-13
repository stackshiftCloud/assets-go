package stackshift

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (m *MailEventsClient) List(ctx context.Context, options MailEventListOptions) (*MailEventListResult, error) {
	query := url.Values{}
	appendQuery(query, "type", options.Type)
	appendQuery(query, "resourceType", options.ResourceType)
	appendQuery(query, "resourceId", options.ResourceID)
	appendQuery(query, "messageId", options.MessageID)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	appendQuery(query, "from", options.From)
	appendQuery(query, "to", options.To)
	path := "/mail/events"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailEventListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailEventsClient) Get(ctx context.Context, eventID string) (*MailEvent, error) {
	var event MailEvent
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/events/"+url.PathEscape(eventID), nil, &event)
	return &event, err
}

func (m *MailMessagesClient) Timeline(ctx context.Context, messageID string) ([]MailTimelineItem, error) {
	return m.TimelineForRecipient(ctx, messageID, "")
}

func (m *MailMessagesClient) TimelineForRecipient(ctx context.Context, messageID, recipient string) ([]MailTimelineItem, error) {
	var result struct {
		Data []MailTimelineItem `json:"data"`
	}
	path := "/mail/messages/" + url.PathEscape(messageID) + "/timeline"
	if recipient != "" {
		path += "?recipient=" + url.QueryEscape(recipient)
	}
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return result.Data, err
}

func (m *MailWebhooksClient) Replay(ctx context.Context, webhookID string, input MailWebhookReplayRequest) (*MailWebhookReplay, error) {
	var result MailWebhookReplay
	path := "/mail/webhooks/" + url.PathEscape(webhookID) + "/replay"
	headers := map[string]string{"Idempotency-Key": input.IdempotencyKey}
	body := map[string]any{"eventIds": input.EventIDs}
	err := m.client.mailRequestWithHeaders(ctx, http.MethodPost, path, body, headers, &result)
	return &result, err
}

func (m *MailWebhooksClient) Create(ctx context.Context, input MailWebhookCreateRequest) (*MailWebhookCreateResponse, error) {
	var result MailWebhookCreateResponse
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/webhooks", input, &result)
	return &result, err
}

func (m *MailWebhooksClient) List(ctx context.Context, options MailWebhookListOptions) (*MailWebhookListResult, error) {
	query := url.Values{}
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	path := "/mail/webhooks"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailWebhookListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailWebhooksClient) Get(ctx context.Context, webhookID string) (*MailWebhook, error) {
	var webhook MailWebhook
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/webhooks/"+url.PathEscape(webhookID), nil, &webhook)
	return &webhook, err
}

func (m *MailWebhooksClient) Update(ctx context.Context, webhookID string, input MailWebhookUpdateRequest) (*MailWebhook, error) {
	var webhook MailWebhook
	path := "/mail/webhooks/" + url.PathEscape(webhookID)
	err := m.client.mailRequest(ctx, http.MethodPatch, path, input, &webhook)
	return &webhook, err
}

func (m *MailWebhooksClient) RotateSecret(ctx context.Context, webhookID string) (string, error) {
	var result struct {
		Secret string `json:"secret"`
	}
	path := "/mail/webhooks/" + url.PathEscape(webhookID) + "/rotate-secret"
	err := m.client.mailRequest(ctx, http.MethodPost, path, nil, &result)
	return result.Secret, err
}

func (m *MailWebhooksClient) Delete(ctx context.Context, webhookID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/webhooks/"+url.PathEscape(webhookID), nil, nil)
}

func (m *MailWebhooksClient) Deliveries(ctx context.Context, webhookID string, options MailWebhookDeliveryListOptions) (*MailWebhookDeliveryListResult, error) {
	query := url.Values{}
	appendQuery(query, "status", options.Status)
	appendQuery(query, "limit", options.Limit)
	appendQuery(query, "cursor", options.Cursor)
	path := "/mail/webhooks/" + url.PathEscape(webhookID) + "/deliveries"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result MailWebhookDeliveryListResult
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailWebhooksClient) GetDelivery(ctx context.Context, deliveryID string) (*MailWebhookDeliveryDetail, error) {
	var result MailWebhookDeliveryDetail
	path := "/mail/webhook-deliveries/" + url.PathEscape(deliveryID)
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (m *MailWebhooksClient) RetryDelivery(ctx context.Context, deliveryID string) (*MailWebhookDelivery, error) {
	var result MailWebhookDelivery
	path := "/mail/webhook-deliveries/" + url.PathEscape(deliveryID) + "/retry"
	err := m.client.mailRequest(ctx, http.MethodPost, path, nil, &result)
	return &result, err
}

func (m *MailWebhooksClient) VerifySignature(input MailWebhookSignatureInput) bool {
	return VerifyWebhookSignature(input)
}

func VerifyWebhookSignature(input MailWebhookSignatureInput) bool {
	if input.Secret == "" || len(input.RawBody) == 0 || input.SignatureHeader == "" || input.TimestampHeader == "" {
		return false
	}
	timestamp, err := strconv.ParseInt(strings.TrimSpace(input.TimestampHeader), 10, 64)
	if err != nil {
		return false
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	tolerance := input.Tolerance
	if tolerance == 0 {
		tolerance = 5 * time.Minute
	}
	eventTime := time.Unix(timestamp, 0)
	if now.Sub(eventTime) > tolerance || eventTime.Sub(now) > tolerance {
		return false
	}
	expected := webhookSignature(input.Secret, input.TimestampHeader, input.RawBody)
	provided := signatureV1(input.SignatureHeader)
	if provided == "" {
		return false
	}
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	providedBytes, err := hex.DecodeString(provided)
	if err != nil {
		return false
	}
	return hmac.Equal(expectedBytes, providedBytes)
}

func webhookSignature(secret, timestamp string, rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}

func signatureV1(header string) string {
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && key == "v1" {
			return value
		}
	}
	return ""
}
