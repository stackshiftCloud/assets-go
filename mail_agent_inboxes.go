package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

const (
	agentInboxMaxServerWait = 50 * time.Second
	agentInboxDefaultWait   = 30 * time.Second
)

type AgentInbox struct {
	ID             string            `json:"id"`
	Address        string            `json:"address"`
	DisplayName    string            `json:"displayName"`
	DomainID       string            `json:"domainId"`
	OnSharedDomain bool              `json:"onSharedDomain"`
	Labels         map[string]string `json:"labels"`
	Metadata       json.RawMessage   `json:"metadata"`
	JobID          string            `json:"jobId,omitempty"`
	ClientRef      string            `json:"clientRef,omitempty"`
	SendMode       string            `json:"sendMode"`
	ExpiresAt      *time.Time        `json:"expiresAt"`
	CreatedVia     string            `json:"createdVia"`
	ConnectionID   string            `json:"connectionId,omitempty"`
	Status         string            `json:"status"`
	SendingPaused  bool              `json:"sendingPaused"`
	LastMessageAt  *time.Time        `json:"lastMessageAt"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	DeletedAt      *time.Time        `json:"deletedAt,omitempty"`
	DeletedReason  string            `json:"deletedReason,omitempty"`
}

type AgentInboxCreateRequest struct {
	Prefix         string            `json:"prefix,omitempty"`
	Domain         string            `json:"domain,omitempty"`
	LocalPart      string            `json:"localPart,omitempty"`
	DisplayName    string            `json:"displayName,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Metadata       json.RawMessage   `json:"metadata,omitempty"`
	JobID          string            `json:"jobId,omitempty"`
	ClientRef      string            `json:"clientRef,omitempty"`
	ExpiresAt      *time.Time        `json:"expiresAt,omitempty"`
	ExpiresIn      int64             `json:"expiresIn,omitempty"`
	SendMode       string            `json:"sendMode,omitempty"`
	IdempotencyKey string            `json:"-"`
}

type AgentInboxUpdateRequest struct {
	DisplayName *string            `json:"displayName,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
	Metadata    json.RawMessage    `json:"metadata,omitempty"`
	ExpiresAt   *time.Time         `json:"expiresAt,omitempty"`
	ClearExpiry bool               `json:"clearExpiry,omitempty"`
	SendMode    *string            `json:"sendMode,omitempty"`
}

type AgentInboxListOptions struct {
	Labels map[string]string
	JobID  string
	Query  string
	Domain string
	Limit  int
	Cursor string
}

type AgentInboxPage struct {
	Data       []AgentInbox `json:"data"`
	NextCursor string       `json:"nextCursor,omitempty"`
}

type AgentInboxDeleteRequest struct {
	IDs    []string          `json:"ids,omitempty"`
	JobID  string            `json:"jobId,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type AgentInboxAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type AgentInboxAttachment struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

type AgentInboxMessage struct {
	ID          string                 `json:"id"`
	ThreadID    string                 `json:"threadId"`
	From        AgentInboxAddress      `json:"from"`
	To          []AgentInboxAddress    `json:"to"`
	Cc          []AgentInboxAddress    `json:"cc,omitempty"`
	Subject     string                 `json:"subject"`
	ReceivedAt  time.Time              `json:"receivedAt"`
	Preview     string                 `json:"preview"`
	Text        string                 `json:"text,omitempty"`
	HTML        string                 `json:"html,omitempty"`
	Codes       []string               `json:"codes"`
	Links       []string               `json:"links"`
	Attachments []AgentInboxAttachment `json:"attachments,omitempty"`
	Junk        bool                   `json:"junk,omitempty"`
}

type AgentInboxMessageListOptions struct {
	After string
	Limit int
}

type AgentInboxMessagePage struct {
	Data       []AgentInboxMessage `json:"data"`
	NextCursor string              `json:"nextCursor,omitempty"`
}

type AgentInboxMessageFilter struct {
	From    string
	Subject string
}

type AgentInboxWaitOptions struct {
	From    string
	Subject string
	Timeout time.Duration
	Since   time.Time
}

type agentInboxWaitResult struct {
	Message  *AgentInboxMessage `json:"message"`
	TimedOut bool               `json:"timedOut"`
}

type AgentInboxSendRequest struct {
	To               []string `json:"to"`
	Cc               []string `json:"cc,omitempty"`
	Subject          string   `json:"subject"`
	Text             string   `json:"text,omitempty"`
	HTML             string   `json:"html,omitempty"`
	ReplyToMessageID string   `json:"replyToMessageId,omitempty"`
}

type AgentInboxSendResult struct {
	Status         string   `json:"status"`
	MessageID      string   `json:"messageId,omitempty"`
	ThreadID       string   `json:"threadId,omitempty"`
	PendingReplyID string   `json:"pendingReplyId,omitempty"`
	Reasons        []string `json:"reasons,omitempty"`
}

type AgentInboxLimits struct {
	TrustTier              string    `json:"trustTier"`
	MailPlan               string    `json:"mailPlan"`
	Included               int       `json:"included"`
	Live                   int       `json:"live"`
	LiveShared             int       `json:"liveShared"`
	LiveSharedLimit        int       `json:"liveSharedLimit"`
	LiveOwnDomainLimit     int       `json:"liveOwnDomainLimit"`
	CreatedThisHour        int       `json:"createdThisHour"`
	CreatedPerHourLimit    int       `json:"createdPerHourLimit"`
	SendModes              []string  `json:"sendModes"`
	SharedRecipientsToday  int       `json:"sharedRecipientsToday"`
	SharedRecipientsPerDay int       `json:"sharedRecipientsPerDay"`
	InboxRecipientsPerDay  int       `json:"inboxRecipientsPerDay"`
	SharedSendingStatus    string    `json:"sharedSendingStatus"`
	SharedSendingReason    string    `json:"sharedSendingReason,omitempty"`
	ExtrasAllowed          bool      `json:"extrasAllowed"`
	NextExtraMonthlyKobo   int64     `json:"nextExtraMonthlyKobo"`
	SharedDomain           string    `json:"sharedDomain"`
	ResetsAt               time.Time `json:"resetsAt"`
}

type AgentInboxPriceBand struct {
	From        int   `json:"from"`
	To          int   `json:"to"`
	MonthlyKobo int64 `json:"monthlyKobo"`
	DailyKobo   int64 `json:"dailyKobo"`
}

type AgentInboxPricing struct {
	IncludedByPlan map[string]int        `json:"includedByPlan"`
	Bands          []AgentInboxPriceBand `json:"bands"`
	QuotaBytes     int64                 `json:"quotaBytes"`
	SharedDomain   string                `json:"sharedDomain"`
}

type AgentInboxesClient struct {
	client   *Client
	now      func() time.Time
	Messages *AgentInboxMessagesClient
}

type AgentInboxMessagesClient struct{ client *Client }

func agentInboxHeaders(key string) map[string]string {
	if key == "" {
		return nil
	}
	return map[string]string{"Idempotency-Key": key}
}

func (a *AgentInboxesClient) Create(ctx context.Context, input AgentInboxCreateRequest) (*AgentInbox, error) {
	var out AgentInbox
	err := a.client.mailRequestWithHeaders(ctx, http.MethodPost, "/mail/agent-inboxes", input, agentInboxHeaders(input.IdempotencyKey), &out)
	return &out, err
}

func (a *AgentInboxesClient) CreateBatch(ctx context.Context, inboxes []AgentInboxCreateRequest, idempotencyKey string) ([]AgentInbox, error) {
	var out struct {
		Data []AgentInbox `json:"data"`
	}
	err := a.client.mailRequestWithHeaders(ctx, http.MethodPost, "/mail/agent-inboxes/batch", map[string]any{"inboxes": inboxes}, agentInboxHeaders(idempotencyKey), &out)
	return out.Data, err
}

func (a *AgentInboxesClient) List(ctx context.Context, options AgentInboxListOptions) (*AgentInboxPage, error) {
	query := url.Values{}
	keys := make([]string, 0, len(options.Labels))
	for key := range options.Labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		query.Add("label", key+":"+options.Labels[key])
	}
	setQuery(query, "jobId", options.JobID)
	setQuery(query, "q", options.Query)
	setQuery(query, "domain", options.Domain)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	setQuery(query, "cursor", options.Cursor)
	path := "/mail/agent-inboxes"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out AgentInboxPage
	err := a.client.mailRequest(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (a *AgentInboxesClient) Get(ctx context.Context, inboxID string) (*AgentInbox, error) {
	var out AgentInbox
	err := a.client.mailRequest(ctx, http.MethodGet, "/mail/agent-inboxes/"+url.PathEscape(inboxID), nil, &out)
	return &out, err
}

func (a *AgentInboxesClient) Update(ctx context.Context, inboxID string, input AgentInboxUpdateRequest) (*AgentInbox, error) {
	var out AgentInbox
	err := a.client.mailRequest(ctx, http.MethodPatch, "/mail/agent-inboxes/"+url.PathEscape(inboxID), input, &out)
	return &out, err
}

func (a *AgentInboxesClient) Delete(ctx context.Context, inboxID string) error {
	return a.client.mailRequest(ctx, http.MethodDelete, "/mail/agent-inboxes/"+url.PathEscape(inboxID), nil, nil)
}

func (a *AgentInboxesClient) DeleteMany(ctx context.Context, input AgentInboxDeleteRequest) (int, error) {
	chosen := 0
	for _, present := range []bool{len(input.IDs) > 0, input.JobID != "", len(input.Labels) > 0} {
		if present {
			chosen++
		}
	}
	if chosen != 1 {
		return 0, errors.New("stackshift: DeleteMany needs exactly one of IDs, JobID or Labels")
	}
	var out struct {
		Deleted int `json:"deleted"`
	}
	err := a.client.mailRequest(ctx, http.MethodPost, "/mail/agent-inboxes/delete", input, &out)
	return out.Deleted, err
}

func (a *AgentInboxesClient) FinishJob(ctx context.Context, jobID string) (int, error) {
	return a.DeleteMany(ctx, AgentInboxDeleteRequest{JobID: jobID})
}

func (a *AgentInboxesClient) WaitForMessage(ctx context.Context, inboxID string, options AgentInboxWaitOptions) (*AgentInboxMessage, error) {
	started := a.now()
	since := options.Since
	if since.IsZero() {
		since = started
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = agentInboxDefaultWait
	}
	deadline := started.Add(timeout)
	for {
		wait := min(agentInboxMaxServerWait, max(time.Second, (deadline.Sub(a.now())+time.Second-1)/time.Second*time.Second))
		query := url.Values{}
		query.Set("since", since.UTC().Format(time.RFC3339Nano))
		setQuery(query, "from", options.From)
		setQuery(query, "subject", options.Subject)
		query.Set("timeout", strconv.Itoa(int(wait/time.Second)))
		var result agentInboxWaitResult
		path := "/mail/agent-inboxes/" + url.PathEscape(inboxID) + "/messages/wait?" + query.Encode()
		if err := a.client.mailRequest(ctx, http.MethodGet, path, nil, &result); err != nil {
			return nil, err
		}
		if result.Message != nil {
			return result.Message, nil
		}
		if !a.now().Before(deadline) {
			return nil, nil
		}
	}
}

func (a *AgentInboxesClient) Send(ctx context.Context, inboxID string, input AgentInboxSendRequest) (*AgentInboxSendResult, error) {
	var out AgentInboxSendResult
	err := a.client.mailRequest(ctx, http.MethodPost, "/mail/agent-inboxes/"+url.PathEscape(inboxID)+"/send", input, &out)
	return &out, err
}

func (a *AgentInboxesClient) Limits(ctx context.Context) (*AgentInboxLimits, error) {
	var out AgentInboxLimits
	err := a.client.mailRequest(ctx, http.MethodGet, "/mail/agent-inboxes/limits", nil, &out)
	return &out, err
}

func (a *AgentInboxesClient) Pricing(ctx context.Context) (*AgentInboxPricing, error) {
	var out AgentInboxPricing
	err := a.client.mailRequest(ctx, http.MethodGet, "/mail/agent-inboxes/pricing", nil, &out)
	return &out, err
}

func (m *AgentInboxMessagesClient) List(ctx context.Context, inboxID string, options AgentInboxMessageListOptions) (*AgentInboxMessagePage, error) {
	query := url.Values{}
	setQuery(query, "after", options.After)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	path := "/mail/agent-inboxes/" + url.PathEscape(inboxID) + "/messages"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out AgentInboxMessagePage
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (m *AgentInboxMessagesClient) Latest(ctx context.Context, inboxID string, filter AgentInboxMessageFilter) (*AgentInboxMessage, error) {
	query := url.Values{}
	setQuery(query, "from", filter.From)
	setQuery(query, "subject", filter.Subject)
	path := "/mail/agent-inboxes/" + url.PathEscape(inboxID) + "/messages/latest"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out AgentInboxMessage
	err := m.client.mailRequest(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (m *AgentInboxMessagesClient) Get(ctx context.Context, inboxID, messageID string) (*AgentInboxMessage, error) {
	var out AgentInboxMessage
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/agent-inboxes/"+url.PathEscape(inboxID)+"/messages/"+url.PathEscape(messageID), nil, &out)
	return &out, err
}
