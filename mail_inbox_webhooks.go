package stackshift

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	InboxEventConversationReceived      = "mailbox.conversation.received"
	InboxEventConversationAssigned      = "mailbox.conversation.assigned"
	InboxEventConversationStatusChanged = "mailbox.conversation.status_changed"
	InboxEventConversationNoteAdded     = "mailbox.conversation.note_added"
	InboxEventConversationReplyWaiting  = "mailbox.conversation.reply_waiting"
	InboxEventConversationSent          = "mailbox.conversation.sent"
	InboxEventAgentInboxCreated         = "mailbox.agent_inbox.created"
	InboxEventAgentInboxDeleted         = "mailbox.agent_inbox.deleted"
	InboxEventSendingPaused             = "mailbox.sending.paused"
	InboxEventSendingResumed            = "mailbox.sending.resumed"
)

type InboxEventActor struct {
	Kind     string `json:"kind"`
	UserID   string `json:"userId,omitempty"`
	WorkerID string `json:"workerId,omitempty"`
}

type InboxEventActivity struct {
	ID   int64           `json:"id"`
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type InboxConversationEventData struct {
	MailboxID    string                `json:"mailboxId"`
	ThreadID     string                `json:"threadId"`
	Conversation TeamInboxConversation `json:"conversation"`
	Actor        InboxEventActor       `json:"actor"`
	Activity     InboxEventActivity    `json:"activity"`
}

type InboxMailboxEventData struct {
	MailboxID string            `json:"mailboxId,omitempty"`
	Address   string            `json:"address,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	JobID     string            `json:"jobId,omitempty"`
	Reason    string            `json:"reason,omitempty"`
}

type InboxWebhookEvent struct {
	Mode        string          `json:"mode,omitempty"`
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	CreatedAt   time.Time       `json:"createdAt"`
	OccurredAt  time.Time       `json:"occurredAt"`
	WorkspaceID string          `json:"workspaceId"`
	Data        json.RawMessage `json:"data"`
}

func ParseInboxWebhook(payload []byte) (*InboxWebhookEvent, error) {
	var event InboxWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (e *InboxWebhookEvent) ConversationData() (*InboxConversationEventData, error) {
	var data InboxConversationEventData
	err := json.Unmarshal(e.Data, &data)
	return &data, err
}

func (e *InboxWebhookEvent) MailboxData() (*InboxMailboxEventData, error) {
	var data InboxMailboxEventData
	err := json.Unmarshal(e.Data, &data)
	return &data, err
}

func VerifyInboxWebhook(payload []byte, signatureHeader, secret string) bool {
	timestamp := ""
	for _, part := range strings.Split(signatureHeader, ",") {
		if key, value, ok := strings.Cut(strings.TrimSpace(part), "="); ok && key == "t" {
			timestamp = value
		}
	}
	return VerifyWebhookSignature(MailWebhookSignatureInput{
		RawBody: payload, SignatureHeader: signatureHeader, TimestampHeader: timestamp, Secret: secret,
	})
}
