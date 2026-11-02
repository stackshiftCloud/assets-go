package stackshift

import (
	"encoding/json"
	"time"
)

type AgreementMoney struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type AgreementItem struct {
	Description string          `json:"description"`
	Quantity    string          `json:"quantity"`
	Unit        string          `json:"unit,omitempty"`
	UnitPrice   *AgreementMoney `json:"unit_price,omitempty"`
}

type AgreementSlot struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	TimeZone string    `json:"time_zone"`
}

type AgreementCounterparty struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	OrgName string `json:"orgName"`
	Party   string `json:"party"`
	Check   string `json:"check"`
}

type AgreementOpenOffer struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	ExpiresAt time.Time `json:"expiresAt"`
	FromUs    bool      `json:"fromUs"`
}

type AgreementWaiting struct {
	Kind       string `json:"kind"`
	ApprovalID string `json:"approvalId"`
}

type AgreementThread struct {
	ID           string                `json:"id"`
	InboxID      string                `json:"inboxId"`
	InboxAddress string                `json:"inboxAddress"`
	InboxName    string                `json:"inboxName"`
	Channel      string                `json:"channel"`
	State        string                `json:"state"`
	EndedReason  *string               `json:"endedReason"`
	Counterparty AgreementCounterparty `json:"counterparty"`
	What         string                `json:"what"`
	Amount       *AgreementMoney       `json:"amount"`
	Direction    string                `json:"direction"`
	OpenOffer    *AgreementOpenOffer   `json:"openOffer"`
	RecordKind   *string               `json:"recordKind"`
	Waiting      *AgreementWaiting     `json:"waiting"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

type AgreementCounts struct {
	Open          int `json:"open"`
	WaitingForYou int `json:"waiting_for_you"`
	Agreed        int `json:"agreed"`
	Ended         int `json:"ended"`
}

type AgreementListResult struct {
	Threads    []AgreementThread `json:"threads"`
	NextCursor string            `json:"nextCursor"`
	Counts     AgreementCounts   `json:"counts"`
}

type ListAgreementsParams struct {
	InboxID      string
	State        string
	Counterparty string
	Cursor       string
	Limit        int
}

type AgreementParty struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
	Party   string `json:"party,omitempty"`
}

type AgreementAuthority struct {
	DecidedBy     string `json:"decidedBy"`
	Level         string `json:"level"`
	LimitsVersion int    `json:"limitsVersion"`
}

type AgreementInfluence struct {
	Quote           string `json:"quote"`
	SourceMessageID string `json:"sourceMessageId"`
}

type AgreementMessage struct {
	ID            string              `json:"id"`
	Direction     string              `json:"direction"`
	Type          string              `json:"type"`
	Check         string              `json:"check"`
	OfferState    string              `json:"offerState"`
	CreatedAt     time.Time           `json:"createdAt"`
	ExpiresAt     *time.Time          `json:"expiresAt"`
	From          AgreementParty      `json:"from"`
	To            AgreementParty      `json:"to"`
	Summary       string              `json:"summary"`
	Subject       string              `json:"subject"`
	Text          string              `json:"text"`
	Items         []AgreementItem     `json:"items"`
	Total         *AgreementMoney     `json:"total"`
	DeliveryBy    string              `json:"deliveryBy"`
	WantedBy      string              `json:"wantedBy"`
	Slots         []AgreementSlot     `json:"slots"`
	Purpose       string              `json:"purpose"`
	Location      string              `json:"location"`
	Note          string              `json:"note"`
	Reason        string              `json:"reason"`
	Authority     *AgreementAuthority `json:"authority"`
	MailMessageID string              `json:"mailMessageId"`
	Influence     *AgreementInfluence `json:"influence"`
}

type AgreementSigner struct {
	OrgName string `json:"orgName"`
	Party   string `json:"party"`
	Address string `json:"address"`
	Kind    string `json:"kind"`
	Kid     string `json:"kid"`
	Revoked bool   `json:"revoked"`
}

type AgreementCheckResult struct {
	Result   string            `json:"result"`
	Reason   string            `json:"reason"`
	Summary  string            `json:"summary"`
	Signers  []AgreementSigner `json:"signers"`
	Terms    json.RawMessage   `json:"terms"`
	AgreedAt *time.Time        `json:"agreedAt"`
	Warnings []string          `json:"warnings"`
}

type AgreementRecord struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind"`
	AgreedAt    time.Time            `json:"agreedAt"`
	Hash        string               `json:"hash"`
	CancelledAt *time.Time           `json:"cancelledAt"`
	Check       AgreementCheckResult `json:"check"`
	CheckURL    string               `json:"checkUrl"`
	DownloadURL string               `json:"downloadUrl"`
}

type AgreementReply struct {
	MailMessageID string    `json:"mailMessageId"`
	From          string    `json:"from"`
	Text          string    `json:"text"`
	ReceivedAt    time.Time `json:"receivedAt"`
	OfferID       string    `json:"offerId"`
}

type AgreementInspect struct {
	Thread       AgreementThread    `json:"thread"`
	Messages     []AgreementMessage `json:"messages"`
	Record       *AgreementRecord   `json:"record"`
	Replies      []AgreementReply   `json:"replies"`
	ChangeCursor string             `json:"changeCursor"`
}

type AgreementWaitResult struct {
	AgreementInspect
	TimedOut bool `json:"timedOut"`
}

type AgreementSendOptions struct {
	IdempotencyKey string
}

type ProposeAgreementInput struct {
	Type           string          `json:"type"`
	ThreadID       string          `json:"threadId,omitempty"`
	To             string          `json:"to,omitempty"`
	InReplyTo      string          `json:"inReplyTo,omitempty"`
	Replaces       string          `json:"replaces,omitempty"`
	Items          []AgreementItem `json:"items,omitempty"`
	Total          *AgreementMoney `json:"total,omitempty"`
	WantedBy       string          `json:"wantedBy,omitempty"`
	DeliveryBy     string          `json:"deliveryBy,omitempty"`
	Slots          []AgreementSlot `json:"slots,omitempty"`
	Purpose        string          `json:"purpose,omitempty"`
	Location       string          `json:"location,omitempty"`
	OpenForSeconds int             `json:"openForSeconds,omitempty"`
	Note           string          `json:"note,omitempty"`
}

type RespondAgreementInput struct {
	OfferID  string `json:"offerId"`
	Decision string `json:"decision"`
	Slot     *int   `json:"slot,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Note     string `json:"note,omitempty"`
}

type CancelAgreementInput struct {
	OfferID string `json:"offerId,omitempty"`
	Note    string `json:"note,omitempty"`
}

type AgreementSendResult struct {
	ThreadID   string `json:"threadId"`
	MessageID  string `json:"messageId"`
	State      string `json:"state"`
	ApprovalID string `json:"approvalId,omitempty"`
}

type AgreementLimitMoney struct {
	BuyPerDeal  string `json:"buyPerDeal"`
	BuyPerDay   string `json:"buyPerDay"`
	SellPerDeal string `json:"sellPerDeal"`
	SellPerDay  string `json:"sellPerDay"`
}

type AgreementTimeWindow struct {
	Weekdays          []int  `json:"weekdays"`
	From              string `json:"from"`
	Until             string `json:"until"`
	TimeZone          string `json:"timeZone"`
	EarliestDaysAhead int    `json:"earliestDaysAhead"`
	LatestDaysAhead   int    `json:"latestDaysAhead"`
	MaxSlotMinutes    int    `json:"maxSlotMinutes"`
}

type AgreementUsedToday struct {
	Buy  string `json:"buy"`
	Sell string `json:"sell"`
}

type AgreementLimits struct {
	InboxID          string                         `json:"inboxId"`
	InboxAddress     string                         `json:"inboxAddress"`
	Enabled          bool                           `json:"enabled"`
	AllowedTypes     []string                       `json:"allowedTypes"`
	Money            map[string]AgreementLimitMoney `json:"money"`
	Counterparties   string                         `json:"counterparties"`
	CounterpartyList []string                       `json:"counterpartyList"`
	MaxOpenSeconds   int                            `json:"maxOpenSeconds"`
	TimeWindow       AgreementTimeWindow            `json:"timeWindow"`
	Version          int                            `json:"version"`
	UpdatedAt        time.Time                      `json:"updatedAt"`
	UsedToday        map[string]AgreementUsedToday  `json:"usedToday"`
	SentLastHour     int                            `json:"sentLastHour"`
	SentToday        int                            `json:"sentToday"`
	NewAccount       bool                           `json:"newAccount"`
}
