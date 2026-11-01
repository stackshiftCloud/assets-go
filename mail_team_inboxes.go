package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	teamInboxDefaultWait = 30 * time.Second
	teamInboxDefaultPoll = 2 * time.Second
)

type TeamInboxActor struct {
	UserID   string `json:"userId,omitempty"`
	WorkerID string `json:"workerId,omitempty"`
}

type TeamInboxLock struct {
	Holder     TeamInboxActor `json:"holder"`
	AcquiredAt time.Time      `json:"acquiredAt"`
	ExpiresAt  time.Time      `json:"expiresAt"`
}

type TeamInboxConversation struct {
	MailboxID        string          `json:"mailboxId"`
	ThreadID         string          `json:"threadId"`
	Status           string          `json:"status"`
	Assignee         TeamInboxActor  `json:"assignee"`
	Subject          string          `json:"subject"`
	Contact          string          `json:"contact"`
	LastInboundAt    *time.Time      `json:"lastInboundAt,omitempty"`
	LastOutboundAt   *time.Time      `json:"lastOutboundAt,omitempty"`
	StatusChangedAt  time.Time       `json:"statusChangedAt"`
	Revision         int64           `json:"revision"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	Lock             *TeamInboxLock  `json:"lock,omitempty"`
	PendingReplyID   string          `json:"pendingReplyId,omitempty"`
	NoteCount        int             `json:"noteCount"`
	LatestNoteAuthor *TeamInboxActor `json:"latestNoteAuthor,omitempty"`
}

type TeamInboxNote struct {
	ID        string         `json:"id"`
	MailboxID string         `json:"mailboxId"`
	ThreadID  string         `json:"threadId"`
	Author    TeamInboxActor `json:"author"`
	Body      string         `json:"body"`
	Mentions  []string       `json:"mentions"`
	CreatedAt time.Time      `json:"createdAt"`
	EditedAt  *time.Time     `json:"editedAt,omitempty"`
}

type TeamInboxActivity struct {
	ID        int64           `json:"id"`
	MailboxID string          `json:"mailboxId"`
	ThreadID  string          `json:"threadId"`
	Kind      string          `json:"kind"`
	Actor     TeamInboxActor  `json:"actor"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
}

type TeamInboxPendingReply struct {
	ID            string          `json:"id"`
	MailboxID     string          `json:"mailboxId"`
	ThreadID      string          `json:"threadId"`
	WorkerID      string          `json:"workerId"`
	DraftEmailID  string          `json:"draftEmailId"`
	Recipients    []string        `json:"recipients"`
	Subject       string          `json:"subject"`
	Body          string          `json:"body"`
	BodySHA256    string          `json:"bodySha256"`
	ThreadVersion string          `json:"threadVersion"`
	Reasons       []string        `json:"reasons"`
	LookedAt      json.RawMessage `json:"lookedAt"`
	State         string          `json:"state"`
	DecidedBy     string          `json:"decidedBy,omitempty"`
	DecidedAt     *time.Time      `json:"decidedAt,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type TeamInboxCounts struct {
	Open    int `json:"open"`
	Waiting int `json:"waiting"`
	Done    int `json:"done"`
}

type TeamInboxSettings struct {
	MailboxID   string    `json:"mailboxId"`
	Name        string    `json:"name"`
	ForwardTo   []string  `json:"forwardTo"`
	LockSeconds int       `json:"lockSeconds"`
	Revision    int64     `json:"revision"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TeamInboxPlatformResource struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type TeamInboxWorker struct {
	ID                string                      `json:"id"`
	MailboxID         string                      `json:"mailboxId"`
	Kind              string                      `json:"kind"`
	ConnectionID      string                      `json:"connectionId,omitempty"`
	Name              string                      `json:"name"`
	Enabled           bool                        `json:"enabled"`
	SendLevel         string                      `json:"sendLevel"`
	WakeOn            string                      `json:"wakeOn"`
	Instructions      string                      `json:"instructions"`
	PlatformResources []TeamInboxPlatformResource `json:"platformResources"`
	CreatedBy         string                      `json:"createdBy,omitempty"`
	Revision          int64                       `json:"revision"`
	CreatedAt         time.Time                   `json:"createdAt"`
	UpdatedAt         time.Time                   `json:"updatedAt"`
}

type TeamInboxPerson struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type TeamInboxPeopleResult struct {
	People  []TeamInboxPerson `json:"people"`
	Workers []TeamInboxWorker `json:"workers"`
}

type TeamInboxConversationListOptions struct {
	Status   string
	Assignee string
	Pending  bool
	Cursor   string
	Limit    int
}

type TeamInboxConversationList struct {
	Conversations []TeamInboxConversation `json:"conversations"`
	Counts        TeamInboxCounts         `json:"counts"`
	NextCursor    *string                 `json:"nextCursor"`
}

type TeamInboxConversationDetail struct {
	Conversation TeamInboxConversation  `json:"conversation"`
	Notes        []TeamInboxNote        `json:"notes"`
	Activity     []TeamInboxActivity    `json:"activity"`
	Pending      *TeamInboxPendingReply `json:"pending,omitempty"`
}

type TeamInboxAssignRequest struct {
	UserID           string `json:"userId,omitempty"`
	WorkerID         string `json:"workerId,omitempty"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

type TeamInboxStatusRequest struct {
	Status           string `json:"status"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

type TeamInboxNoteRequest struct {
	Body     string   `json:"body"`
	Mentions []string `json:"mentions,omitempty"`
}

type TeamInboxReplyRequest struct {
	DraftEmailID string `json:"draftEmailId"`
	Then         string `json:"then,omitempty"`
}

type TeamInboxPendingSendRequest struct {
	Body *string `json:"body,omitempty"`
}

type TeamInboxActivityListOptions struct {
	After int64
	Limit int
}

type TeamInboxActivityWaitOptions struct {
	After    int64
	Limit    int
	Timeout  time.Duration
	Interval time.Duration
}

type TeamInboxActivityResult struct {
	Activity []TeamInboxActivity `json:"activity"`
	Next     int64               `json:"next"`
}

type TeamInboxesClient struct {
	client         *Client
	Conversations  *TeamInboxConversationsClient
	PendingReplies *TeamInboxPendingRepliesClient
	Activity       *TeamInboxActivityClient
	Settings       *TeamInboxSettingsClient
}

type TeamInboxConversationsClient struct{ client *Client }
type TeamInboxPendingRepliesClient struct{ client *Client }
type TeamInboxActivityClient struct{ client *Client }
type TeamInboxSettingsClient struct{ client *Client }

func teamInboxPath(mailboxID string, rest ...string) string {
	path := "/mail/mailboxes/" + url.PathEscape(mailboxID)
	for _, part := range rest {
		path += "/" + url.PathEscape(part)
	}
	return path
}

func (c *TeamInboxConversationsClient) List(ctx context.Context, mailboxID string, options TeamInboxConversationListOptions) (*TeamInboxConversationList, error) {
	query := url.Values{}
	setQuery(query, "status", options.Status)
	setQuery(query, "assignee", options.Assignee)
	if options.Pending {
		query.Set("pending", "true")
	}
	setQuery(query, "cursor", options.Cursor)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	path := teamInboxPath(mailboxID, "conversations")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out TeamInboxConversationList
	err := c.client.mailRequest(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (c *TeamInboxConversationsClient) Get(ctx context.Context, mailboxID, threadID string) (*TeamInboxConversationDetail, error) {
	var out TeamInboxConversationDetail
	err := c.client.mailRequest(ctx, http.MethodGet, teamInboxPath(mailboxID, "conversations", threadID), nil, &out)
	return &out, err
}

func (c *TeamInboxConversationsClient) Assign(ctx context.Context, mailboxID, threadID string, input TeamInboxAssignRequest) (*TeamInboxConversation, error) {
	var out TeamInboxConversation
	err := c.client.mailRequest(ctx, http.MethodPut, teamInboxPath(mailboxID, "conversations", threadID, "assignee"), input, &out)
	return &out, err
}

func (c *TeamInboxConversationsClient) SetStatus(ctx context.Context, mailboxID, threadID string, input TeamInboxStatusRequest) (*TeamInboxConversation, error) {
	var out TeamInboxConversation
	err := c.client.mailRequest(ctx, http.MethodPut, teamInboxPath(mailboxID, "conversations", threadID, "status"), input, &out)
	return &out, err
}

func (c *TeamInboxConversationsClient) AddNote(ctx context.Context, mailboxID, threadID string, input TeamInboxNoteRequest) (*TeamInboxNote, error) {
	var out TeamInboxNote
	err := c.client.mailRequest(ctx, http.MethodPost, teamInboxPath(mailboxID, "conversations", threadID, "notes"), input, &out)
	return &out, err
}

func (c *TeamInboxConversationsClient) Reply(ctx context.Context, mailboxID, threadID string, input TeamInboxReplyRequest) (*TeamInboxConversation, error) {
	var out TeamInboxConversation
	err := c.client.mailRequest(ctx, http.MethodPost, teamInboxPath(mailboxID, "conversations", threadID, "send"), input, &out)
	return &out, err
}

func (c *TeamInboxPendingRepliesClient) List(ctx context.Context, mailboxID string) ([]TeamInboxPendingReply, error) {
	var out struct {
		PendingReplies []TeamInboxPendingReply `json:"pendingReplies"`
	}
	err := c.client.mailRequest(ctx, http.MethodGet, teamInboxPath(mailboxID, "pending-replies"), nil, &out)
	return out.PendingReplies, err
}

func (c *TeamInboxPendingRepliesClient) Send(ctx context.Context, mailboxID, pendingReplyID string, input TeamInboxPendingSendRequest) (*TeamInboxConversation, error) {
	var out TeamInboxConversation
	err := c.client.mailRequest(ctx, http.MethodPost, teamInboxPath(mailboxID, "pending-replies", pendingReplyID, "send"), input, &out)
	return &out, err
}

func (c *TeamInboxPendingRepliesClient) Discard(ctx context.Context, mailboxID, pendingReplyID string) error {
	return c.client.mailRequest(ctx, http.MethodPost, teamInboxPath(mailboxID, "pending-replies", pendingReplyID, "discard"), nil, nil)
}

func (c *TeamInboxActivityClient) List(ctx context.Context, mailboxID string, options TeamInboxActivityListOptions) (*TeamInboxActivityResult, error) {
	query := url.Values{}
	if options.After > 0 {
		query.Set("after", strconv.FormatInt(options.After, 10))
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	path := teamInboxPath(mailboxID, "activity")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out TeamInboxActivityResult
	err := c.client.mailRequest(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (c *TeamInboxActivityClient) Wait(ctx context.Context, mailboxID string, options TeamInboxActivityWaitOptions) (*TeamInboxActivityResult, error) {
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = teamInboxDefaultWait
	}
	interval := options.Interval
	if interval <= 0 {
		interval = teamInboxDefaultPoll
	}
	deadline := time.Now().Add(timeout)
	cursor := options.After
	for {
		result, err := c.List(ctx, mailboxID, TeamInboxActivityListOptions{After: cursor, Limit: options.Limit})
		if err != nil {
			return nil, err
		}
		if len(result.Activity) > 0 {
			return result, nil
		}
		cursor = result.Next
		if !time.Now().Add(interval).Before(deadline) {
			return &TeamInboxActivityResult{Activity: []TeamInboxActivity{}, Next: cursor}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (c *TeamInboxSettingsClient) Get(ctx context.Context, mailboxID string) (*TeamInboxSettings, error) {
	var out TeamInboxSettings
	err := c.client.mailRequest(ctx, http.MethodGet, teamInboxPath(mailboxID, "team-inbox"), nil, &out)
	return &out, err
}

func (c *TeamInboxesClient) People(ctx context.Context, mailboxID string) (*TeamInboxPeopleResult, error) {
	var out TeamInboxPeopleResult
	err := c.client.mailRequest(ctx, http.MethodGet, teamInboxPath(mailboxID, "people"), nil, &out)
	return &out, err
}
