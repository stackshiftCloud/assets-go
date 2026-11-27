package stackshift

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const EventsQueueName = "_events"

type EventSubject struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type EventActor struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Event struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Version     int             `json:"version"`
	Time        time.Time       `json:"time"`
	Product     string          `json:"product"`
	Subject     EventSubject    `json:"subject"`
	ProjectID   string          `json:"project_id,omitempty"`
	Environment string          `json:"environment,omitempty"`
	Actor       EventActor      `json:"actor"`
	Data        json.RawMessage `json:"data"`
}

type EventFilter struct {
	Path  string `json:"path"`
	Op    string `json:"op"`
	Value any    `json:"value,omitempty"`
}

type EventFilters struct {
	SubjectKind string            `json:"subject_kind,omitempty"`
	SubjectIDs  []string          `json:"subject_ids,omitempty"`
	ProjectIDs  []string          `json:"project_ids,omitempty"`
	Data        []EventFilter     `json:"data,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type EventCatalog struct {
	Version  any              `json:"version"`
	Products []map[string]any `json:"products"`
	Events   []map[string]any `json:"events"`
}

type EventListOptions struct {
	Type      string
	Product   string
	ProjectID string
	Since     string
	Cursor    string
	Limit     int
	TeamID    string
}

type EventPage struct {
	Events     []Event `json:"events"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

type EventSubscriptionInput struct {
	TeamID      string        `json:"team_id,omitempty"`
	ProjectID   string        `json:"project_id,omitempty"`
	Description string        `json:"description,omitempty"`
	URL         string        `json:"url"`
	EventTypes  []string      `json:"event_types,omitempty"`
	Filters     *EventFilters `json:"filters,omitempty"`
}

type EventSubscriptionCreated struct {
	Subscription map[string]any `json:"subscription"`
	Secret       string         `json:"secret"`
}

type EventRuleInput struct {
	TeamID       string         `json:"team_id,omitempty"`
	ProjectID    string         `json:"project_id,omitempty"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	EventTypes   []string       `json:"event_types"`
	Filters      *EventFilters  `json:"filters,omitempty"`
	TargetType   string         `json:"target_type"`
	TargetConfig map[string]any `json:"target_config"`
	Status       string         `json:"status,omitempty"`
}

type EventRulePreview struct {
	Matched24h int      `json:"matched_24h"`
	Types      []string `json:"types"`
}

type EventDeliveryListOptions struct {
	Status         string
	DeadLetter     bool
	SubscriptionID string
	RuleID         string
	EventID        string
	TeamID         string
	Cursor         string
	Limit          int
}

type EventDeliveryPage struct {
	Deliveries []map[string]any `json:"deliveries"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

type EventDeliveryDetail struct {
	Delivery map[string]any   `json:"delivery"`
	Attempts []map[string]any `json:"attempts"`
}

type EventsClient struct {
	client *Client
}

type EventSignatureError struct {
	Code string
}

func (e *EventSignatureError) Error() string {
	switch e.Code {
	case "missing_signature":
		return "stackshift: the StackShift-Signature header is missing"
	case "expired":
		return "stackshift: the StackShift-Signature timestamp is outside the allowed tolerance"
	case "invalid_body":
		return "stackshift: the event body is not valid JSON"
	}
	return "stackshift: the StackShift-Signature header does not match the body"
}

type VerifyEventInput struct {
	RawBody         []byte
	SignatureHeader string
	Secret          string
	Tolerance       time.Duration
	Timestamp       int64
	Now             time.Time
}

func VerifyEvent(input VerifyEventInput) (*Event, error) {
	if strings.TrimSpace(input.SignatureHeader) == "" {
		return nil, &EventSignatureError{Code: "missing_signature"}
	}
	timestamp := input.Timestamp
	if timestamp == 0 {
		timestamp = signedTimestamp(input.SignatureHeader)
	}
	if timestamp == 0 || input.Secret == "" {
		return nil, &EventSignatureError{Code: "invalid_signature"}
	}
	mac := hmac.New(sha256.New, []byte(input.Secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10) + "."))
	mac.Write(input.RawBody)
	provided, err := hex.DecodeString(signatureV1(input.SignatureHeader))
	if err != nil || len(provided) == 0 || !hmac.Equal(mac.Sum(nil), provided) {
		return nil, &EventSignatureError{Code: "invalid_signature"}
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	tolerance := input.Tolerance
	if tolerance == 0 {
		tolerance = 5 * time.Minute
	}
	if tolerance > 0 {
		signedAt := time.Unix(timestamp, 0)
		if now.Sub(signedAt) > tolerance || signedAt.Sub(now) > tolerance {
			return nil, &EventSignatureError{Code: "expired"}
		}
	}
	var event Event
	if err := json.Unmarshal(input.RawBody, &event); err != nil {
		return nil, &EventSignatureError{Code: "invalid_body"}
	}
	return &event, nil
}

func (e *EventsClient) Verify(input VerifyEventInput) (*Event, error) {
	return VerifyEvent(input)
}

func signedTimestamp(header string) int64 {
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && key == "t" {
			timestamp, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0
			}
			return timestamp
		}
	}
	return 0
}

func MatchEventType(pattern, eventType string) bool {
	if pattern == "*" || pattern == eventType {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return false
	}
	parts := strings.Split(pattern, "*")
	if !strings.HasPrefix(eventType, parts[0]) {
		return false
	}
	rest := eventType[len(parts[0]):]
	for index, part := range parts[1:] {
		if index == len(parts)-2 {
			return strings.HasSuffix(rest, part)
		}
		position := strings.Index(rest, part)
		if position < 0 {
			return false
		}
		rest = rest[position+len(part):]
	}
	return true
}

type EventInvocation struct {
	TriggerType string          `json:"trigger_type"`
	QueueName   string          `json:"queue_name"`
	Payload     json.RawMessage `json:"payload"`
}

type EventHandler func(ctx context.Context, event Event) error

func OnEvent(pattern string, handler EventHandler) func(ctx context.Context, invocation EventInvocation) error {
	return func(ctx context.Context, invocation EventInvocation) error {
		if invocation.QueueName != "" && invocation.QueueName != EventsQueueName {
			return nil
		}
		var event Event
		if err := json.Unmarshal(invocation.Payload, &event); err != nil || event.ID == "" || event.Type == "" {
			return nil
		}
		if !MatchEventType(pattern, event.Type) {
			return nil
		}
		return handler(ctx, event)
	}
}

func eventScopePath(path, teamID, projectID string) string {
	query := url.Values{}
	setQuery(query, "team_id", teamID)
	setQuery(query, "project_id", projectID)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return path
}

func (e *EventsClient) Catalog(ctx context.Context) (*EventCatalog, error) {
	var result EventCatalog
	err := e.client.request(ctx, http.MethodGet, "/events/catalog", nil, &result)
	return &result, err
}

func (e *EventsClient) List(ctx context.Context, options EventListOptions) (*EventPage, error) {
	query := url.Values{}
	setQuery(query, "type", options.Type)
	setQuery(query, "product", options.Product)
	setQuery(query, "project_id", options.ProjectID)
	setQuery(query, "since", options.Since)
	setQuery(query, "cursor", options.Cursor)
	setQuery(query, "team_id", options.TeamID)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	path := "/events"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result EventPage
	err := e.client.request(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (e *EventsClient) Get(ctx context.Context, eventID string) (*Event, error) {
	var result Event
	err := e.client.request(ctx, http.MethodGet, "/events/"+url.PathEscape(eventID), nil, &result)
	return &result, err
}

func (e *EventsClient) CreateSubscription(ctx context.Context, input EventSubscriptionInput) (*EventSubscriptionCreated, error) {
	var result EventSubscriptionCreated
	err := e.client.request(ctx, http.MethodPost, "/events/subscriptions", input, &result)
	return &result, err
}

func (e *EventsClient) ListSubscriptions(ctx context.Context, teamID, projectID string) ([]map[string]any, error) {
	var result struct {
		Subscriptions []map[string]any `json:"subscriptions"`
	}
	err := e.client.request(ctx, http.MethodGet, eventScopePath("/events/subscriptions", teamID, projectID), nil, &result)
	return result.Subscriptions, err
}

func (e *EventsClient) GetSubscription(ctx context.Context, subscriptionID string) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodGet, "/events/subscriptions/"+url.PathEscape(subscriptionID), nil, &result)
	return result, err
}

func (e *EventsClient) UpdateSubscription(ctx context.Context, subscriptionID string, input map[string]any) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodPatch, "/events/subscriptions/"+url.PathEscape(subscriptionID), input, &result)
	return result, err
}

func (e *EventsClient) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	return e.client.request(ctx, http.MethodDelete, "/events/subscriptions/"+url.PathEscape(subscriptionID), nil, nil)
}

func (e *EventsClient) RotateSubscriptionSecret(ctx context.Context, subscriptionID string) (string, error) {
	var result struct {
		Secret string `json:"secret"`
	}
	err := e.client.request(ctx, http.MethodPost, "/events/subscriptions/"+url.PathEscape(subscriptionID)+"/rotate-secret", nil, &result)
	return result.Secret, err
}

func (e *EventsClient) TestSubscription(ctx context.Context, subscriptionID string) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodPost, "/events/subscriptions/"+url.PathEscape(subscriptionID)+"/test", nil, &result)
	return result, err
}

func (e *EventsClient) CreateRule(ctx context.Context, input EventRuleInput) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodPost, "/events/rules", input, &result)
	return result, err
}

func (e *EventsClient) ListRules(ctx context.Context, teamID, projectID string) ([]map[string]any, error) {
	var result struct {
		Rules []map[string]any `json:"rules"`
	}
	err := e.client.request(ctx, http.MethodGet, eventScopePath("/events/rules", teamID, projectID), nil, &result)
	return result.Rules, err
}

func (e *EventsClient) GetRule(ctx context.Context, ruleID string) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodGet, "/events/rules/"+url.PathEscape(ruleID), nil, &result)
	return result, err
}

func (e *EventsClient) UpdateRule(ctx context.Context, ruleID string, input map[string]any) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodPatch, "/events/rules/"+url.PathEscape(ruleID), input, &result)
	return result, err
}

func (e *EventsClient) DeleteRule(ctx context.Context, ruleID string) error {
	return e.client.request(ctx, http.MethodDelete, "/events/rules/"+url.PathEscape(ruleID), nil, nil)
}

func (e *EventsClient) PreviewRule(ctx context.Context, input EventRuleInput) (*EventRulePreview, error) {
	var result EventRulePreview
	err := e.client.request(ctx, http.MethodPost, "/events/rules/preview", input, &result)
	return &result, err
}

func (e *EventsClient) TestRule(ctx context.Context, ruleID string) (string, error) {
	var result struct {
		DeliveryID string `json:"delivery_id"`
	}
	err := e.client.request(ctx, http.MethodPost, "/events/rules/"+url.PathEscape(ruleID)+"/test", nil, &result)
	return result.DeliveryID, err
}

func (e *EventsClient) ListDeliveries(ctx context.Context, options EventDeliveryListOptions) (*EventDeliveryPage, error) {
	query := url.Values{}
	setQuery(query, "status", options.Status)
	setQuery(query, "subscription_id", options.SubscriptionID)
	setQuery(query, "rule_id", options.RuleID)
	setQuery(query, "event_id", options.EventID)
	setQuery(query, "team_id", options.TeamID)
	setQuery(query, "cursor", options.Cursor)
	if options.DeadLetter {
		query.Set("dead_letter", "true")
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	path := "/events/deliveries"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result EventDeliveryPage
	err := e.client.request(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (e *EventsClient) GetDelivery(ctx context.Context, deliveryID string) (*EventDeliveryDetail, error) {
	var result EventDeliveryDetail
	err := e.client.request(ctx, http.MethodGet, "/events/deliveries/"+url.PathEscape(deliveryID), nil, &result)
	return &result, err
}

func (e *EventsClient) ReplayDelivery(ctx context.Context, deliveryID string) (map[string]any, error) {
	var result map[string]any
	err := e.client.request(ctx, http.MethodPost, "/events/deliveries/"+url.PathEscape(deliveryID)+"/replay", nil, &result)
	return result, err
}
