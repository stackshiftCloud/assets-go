package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type PartnerAccount struct {
	UserID             string  `json:"userId"`
	Name               string  `json:"name"`
	Status             string  `json:"status"`
	ComplaintPauseRate float64 `json:"complaintPauseRate"`
	BouncePauseRate    float64 `json:"bouncePauseRate"`
	DedicatedPoolName  *string `json:"dedicatedPoolName"`
	DailyPurchaseLimit int     `json:"dailyPurchaseLimit"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
}

type PartnerWalletBalance struct {
	Currency       string `json:"currency"`
	BalanceMinor   int64  `json:"balance_minor"`
	HeldMinor      int64  `json:"held_minor"`
	DisputedMinor  int64  `json:"disputed_minor"`
	AvailableMinor int64  `json:"available_minor"`
}

type PartnerLimits struct {
	MaxComplaintPauseRate float64 `json:"maxComplaintPauseRate"`
	MaxBouncePauseRate    float64 `json:"maxBouncePauseRate"`
	MinimumRecipients     int     `json:"minimumRecipients"`
	WindowDays            int     `json:"windowDays"`
	ShortWindowHours      int     `json:"shortWindowHours"`
	DailyPurchaseLimit    int     `json:"dailyPurchaseLimit"`
	WarningFraction       float64 `json:"warningFraction"`
	MerchantSharePercent  int     `json:"merchantSharePercent"`
	MerchantCount         int     `json:"merchantCount"`
	DailySendingLimit     int     `json:"dailySendingLimit"`
	HourlySendingLimit    int     `json:"hourlySendingLimit"`
}

type PartnerOverview struct {
	Account       PartnerAccount         `json:"account"`
	Wallet        []PartnerWalletBalance `json:"wallet"`
	WebhookEvents []string               `json:"webhookEvents"`
	Limits        PartnerLimits          `json:"limits"`
}

type PartnerSettingsInput struct {
	ComplaintPauseRate float64 `json:"complaintPauseRate"`
	BouncePauseRate    float64 `json:"bouncePauseRate"`
}

type PartnerRegistrant struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
}

type PartnerBrandInput struct {
	Name          string `json:"name"`
	FromLocalPart string `json:"fromLocalPart,omitempty"`
	PostalAddress string `json:"postalAddress"`
	Website       string `json:"website,omitempty"`
	SupportEmail  string `json:"supportEmail,omitempty"`
}

type PartnerRepliesInput struct {
	Mode              string `json:"mode"`
	ForwardTo         string `json:"forwardTo,omitempty"`
	MailboxOwnerEmail string `json:"mailboxOwnerEmail,omitempty"`
	MailboxPlan       string `json:"mailboxPlan,omitempty"`
}

type PartnerDomainPurchaseInput struct {
	Domain      string              `json:"domain"`
	Years       int                 `json:"years"`
	Currency    string              `json:"currency,omitempty"`
	MerchantRef string              `json:"merchantRef"`
	Registrant  PartnerRegistrant   `json:"registrant"`
	Brand       PartnerBrandInput   `json:"brand"`
	Replies     PartnerRepliesInput `json:"replies"`
}

type PartnerFailureFact struct {
	Label string  `json:"label"`
	Value string  `json:"value"`
	Tone  string  `json:"tone"`
	At    *string `json:"at"`
}

type PartnerFailureFix struct {
	Label  string `json:"label"`
	Action string `json:"action,omitempty"`
	Href   string `json:"href,omitempty"`
}

type PartnerFailure struct {
	Code        string               `json:"code"`
	Title       string               `json:"title"`
	Headline    string               `json:"headline"`
	Explanation string               `json:"explanation"`
	Since       *string              `json:"since"`
	Facts       []PartnerFailureFact `json:"facts"`
	Data        map[string]any       `json:"data"`
	Fix         *PartnerFailureFix   `json:"fix"`
	CanRetry    bool                 `json:"canRetry"`
}

type PartnerDiagnosis struct {
	CheckedAt string `json:"checkedAt"`
	Summary   string `json:"summary"`
}

type PartnerPause struct {
	Reason         string   `json:"reason"`
	Rate           float64  `json:"rate"`
	Limit          float64  `json:"limit"`
	Window         string   `json:"window"`
	CampaignID     *string  `json:"campaignId"`
	CampaignName   *string  `json:"campaignName"`
	ListID         *string  `json:"listId"`
	ListName       *string  `json:"listName"`
	ListImportedAt *string  `json:"listImportedAt"`
	ListShare      *float64 `json:"listShare"`
	RecoversAt     *string  `json:"recoversAt"`
}

type PartnerWaitingFor struct {
	Kind           string `json:"kind"`
	Currency       string `json:"currency"`
	RequiredMinor  int64  `json:"requiredMinor"`
	AvailableMinor int64  `json:"availableMinor"`
}

type PartnerRefund struct {
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
	At          string `json:"at"`
}

type PartnerMerchantDomain struct {
	ID                 string              `json:"id"`
	PartnerUserID      string              `json:"partnerUserId"`
	MerchantRef        string              `json:"merchantRef"`
	Domain             string              `json:"domain"`
	Years              int                 `json:"years"`
	DomainOrderID      *string             `json:"domainOrderId"`
	RegisteredDomainID *string             `json:"registeredDomainId"`
	MailDomainID       *string             `json:"mailDomainId"`
	BrandID            *string             `json:"brandId"`
	MailboxDomainID    *string             `json:"mailboxDomainId"`
	MailboxID          *string             `json:"mailboxId"`
	Brand              PartnerBrandInput   `json:"brand"`
	Registrant         PartnerRegistrant   `json:"registrant"`
	Replies            PartnerRepliesInput `json:"replies"`
	State              string              `json:"state"`
	FailedState        *string             `json:"failedState"`
	LastError          string              `json:"lastError"`
	Attempts           int                 `json:"attempts"`
	NextAttemptAt      string              `json:"nextAttemptAt"`
	LastAttemptAt      *string             `json:"lastAttemptAt"`
	WaitingFor         *PartnerWaitingFor  `json:"waitingFor"`
	WarmupDay          int                 `json:"warmupDay"`
	HealthWarnedAt     *string             `json:"healthWarnedAt"`
	TransferredOutAt   *string             `json:"transferredOutAt"`
	AmountMinor        int64               `json:"amountMinor"`
	Currency           string              `json:"currency"`
	WalletReference    string              `json:"walletReference"`
	CreatedAt          string              `json:"createdAt"`
	UpdatedAt          string              `json:"updatedAt"`
	ExpiryDate         *string             `json:"expiryDate"`
	BrandPausedAt      *string             `json:"brandPausedAt"`
	BrandPauseReason   string              `json:"brandPauseReason"`
	Pause              *PartnerPause       `json:"pause"`
	Failure            *PartnerFailure     `json:"failure"`
	Diagnosis          *PartnerDiagnosis   `json:"diagnosis"`
}

type PartnerDomainListOptions struct {
	MerchantRef string
	Query       string
	Attention   bool
	Limit       int
	Offset      int
}

type PartnerWarmupSummary struct {
	Status               string `json:"status"`
	Day                  int    `json:"day"`
	TotalDays            int    `json:"totalDays"`
	MarketingUnlocked    bool   `json:"marketingUnlocked"`
	MarketingOpensInDays int    `json:"marketingOpensInDays"`
}

type PartnerDomainListItem struct {
	PartnerMerchantDomain
	Health              *PartnerBrandHealth   `json:"health"`
	TransactionalHealth *PartnerBrandHealth   `json:"transactionalHealth"`
	Refunded            bool                  `json:"refunded"`
	SentByDay           []int                 `json:"sentByDay"`
	Warmup              *PartnerWarmupSummary `json:"warmup"`
}

type PartnerDomainList struct {
	Domains        []PartnerDomainListItem `json:"domains"`
	Total          int                     `json:"total"`
	AttentionTotal int                     `json:"attentionTotal"`
}

type PartnerSenders struct {
	Marketing     string `json:"marketing"`
	Transactional string `json:"transactional"`
	ReplyTo       string `json:"replyTo"`
}

type PartnerHealthWindow struct {
	Attempted     int     `json:"attempted"`
	Accepted      int     `json:"accepted"`
	Bounced       int     `json:"bounced"`
	Complained    int     `json:"complained"`
	BounceRate    float64 `json:"bounceRate"`
	ComplaintRate float64 `json:"complaintRate"`
}

type PartnerBrandHealth struct {
	PartnerHealthWindow
	Last24h PartnerHealthWindow `json:"last24h"`
}

type PartnerWarmup struct {
	Status                 string `json:"status"`
	Day                    int    `json:"day"`
	TotalDays              int    `json:"totalDays"`
	DailyLimit             int    `json:"dailyLimit"`
	MarketingDailyLimit    int    `json:"marketingDailyLimit"`
	MarketingUnlocked      bool   `json:"marketingUnlocked"`
	CleanTransactionalDays int    `json:"cleanTransactionalDays"`
	RequiredCleanDays      int    `json:"requiredCleanDays"`
	FirstCampaignCap       int    `json:"firstCampaignCap"`
	Paused                 bool   `json:"paused"`
	MarketingSentToday     int    `json:"marketingSentToday"`
	MarketingOpensInDays   int    `json:"marketingOpensInDays"`
}

type PartnerRecord struct {
	Domain   string `json:"domain"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority"`
	Verified bool   `json:"verified"`
	Purpose  string `json:"purpose"`
}

type PartnerReplies struct {
	PartnerRepliesInput
	MailboxOwnerName *string `json:"mailboxOwnerName"`
}

type PartnerDomainView struct {
	PartnerMerchantDomain
	Senders             PartnerSenders      `json:"senders"`
	BrandStatus         *MailBrand          `json:"brandStatus"`
	Health              *PartnerBrandHealth `json:"health"`
	TransactionalHealth *PartnerBrandHealth `json:"transactionalHealth"`
	Warmup              *PartnerWarmup      `json:"warmup"`
	Replies             PartnerReplies      `json:"replies"`
	Records             []PartnerRecord     `json:"records"`
	Refunded            bool                `json:"refunded"`
	Refund              *PartnerRefund      `json:"refund"`
	WebhookDeliveredAt  *string             `json:"webhookDeliveredAt"`
}

type PartnerTransferOut struct {
	Domain   string `json:"domain"`
	AuthCode string `json:"authCode"`
}

type PartnerUsageOptions struct {
	From string
	To   string
}

type PartnerBrandUsage struct {
	PartnerDomainID         string  `json:"partnerDomainId"`
	MerchantRef             string  `json:"merchantRef"`
	Domain                  string  `json:"domain"`
	BrandID                 *string `json:"brandId"`
	BrandName               string  `json:"brandName"`
	MarketingRecipients     int     `json:"marketingRecipients"`
	TransactionalRecipients int     `json:"transactionalRecipients"`
	DomainChargesMinor      int64   `json:"domainChargesMinor"`
	RefundedMinor           int64   `json:"refundedMinor"`
	Currency                string  `json:"currency"`
}

type PartnerUsage struct {
	From   string              `json:"from"`
	To     string              `json:"to"`
	Brands []PartnerBrandUsage `json:"brands"`
}

type PartnerComplaintSummary struct {
	Complained int `json:"complained"`
	Suppressed int `json:"suppressed"`
}

type PartnerComplaintSource struct {
	AudienceID   string  `json:"audienceId"`
	AudienceName string  `json:"audienceName"`
	ImportedAt   *string `json:"importedAt"`
	Contacts     int     `json:"contacts"`
	NeverOpened  int     `json:"neverOpened"`
	Complained   int     `json:"complained"`
}

type PartnerBulkSuppressionInput struct {
	AudienceID string `json:"audienceId"`
	Reason     string `json:"reason"`
}

type PartnerBulkSuppression struct {
	ID         string `json:"id"`
	AudienceID string `json:"audienceId"`
	Reason     string `json:"reason"`
	Suppressed int    `json:"suppressed"`
}

type PartnerRequestInput struct {
	Kind    string `json:"kind"`
	Message string `json:"message,omitempty"`
}

type PartnerRequest struct {
	ID         string  `json:"id"`
	Kind       string  `json:"kind"`
	Message    string  `json:"message"`
	Status     string  `json:"status"`
	ResolvedAt *string `json:"resolvedAt"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

type PartnerEventBase struct {
	PartnerDomainID string  `json:"partnerDomainId"`
	MerchantRef     string  `json:"merchantRef"`
	Domain          string  `json:"domain"`
	State           string  `json:"state"`
	BrandID         *string `json:"brandId,omitempty"`
}

type PartnerBrandHealthEvent struct {
	PartnerEventBase
	BounceRate     float64             `json:"bounceRate"`
	ComplaintRate  float64             `json:"complaintRate"`
	Recipients     int                 `json:"recipients"`
	Last24h        PartnerHealthWindow `json:"last24h"`
	BounceLimit    float64             `json:"bounceLimit"`
	ComplaintLimit float64             `json:"complaintLimit"`
	Reason         string              `json:"reason,omitempty"`
}

type PartnerDomainRenewedEvent struct {
	PartnerEventBase
	ExpiryDate         string `json:"expiryDate"`
	PreviousExpiryDate string `json:"previousExpiryDate"`
}

type PartnerDomainRenewalDueEvent struct {
	PartnerEventBase
	ExpiryDate string `json:"expiryDate"`
	AutoRenew  bool   `json:"autoRenew"`
}

type PartnerDomainRenewalFailedEvent struct {
	PartnerEventBase
	ExpiryDate  string `json:"expiryDate"`
	Reason      string `json:"reason"`
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}

type PartnerClient struct{ client *Client }

func (p *PartnerClient) Overview(ctx context.Context) (*PartnerOverview, error) {
	var out PartnerOverview
	err := p.client.request(ctx, http.MethodGet, "/partner", nil, &out)
	return &out, err
}

func (p *PartnerClient) UpdateSettings(ctx context.Context, input PartnerSettingsInput) (*PartnerAccount, error) {
	var out PartnerAccount
	err := p.client.request(ctx, http.MethodPut, "/partner/settings", input, &out)
	return &out, err
}

func (p *PartnerClient) PurchaseDomain(ctx context.Context, input PartnerDomainPurchaseInput, idempotencyKey string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.requestWithHeaders(ctx, http.MethodPost, "/partner/domains", input, map[string]string{"Idempotency-Key": idempotencyKey}, &out)
	return &out, err
}

func (p *PartnerClient) ListDomains(ctx context.Context, options PartnerDomainListOptions) (*PartnerDomainList, error) {
	query := url.Values{}
	setQuery(query, "merchantRef", options.MerchantRef)
	setQuery(query, "q", options.Query)
	if options.Attention {
		query.Set("attention", "true")
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Offset > 0 {
		query.Set("offset", strconv.Itoa(options.Offset))
	}
	var out PartnerDomainList
	err := p.client.request(ctx, http.MethodGet, "/partner/domains"+queryString(query), nil, &out)
	return &out, err
}

func (p *PartnerClient) GetDomain(ctx context.Context, id string) (*PartnerDomainView, error) {
	var out PartnerDomainView
	err := p.client.request(ctx, http.MethodGet, "/partner/domains/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (p *PartnerClient) RetryDomain(ctx context.Context, id string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/retry", nil, &out)
	return &out, err
}

func (p *PartnerClient) TransferOut(ctx context.Context, id string) (*PartnerTransferOut, error) {
	var out PartnerTransferOut
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/transfer-out", nil, &out)
	return &out, err
}

func (p *PartnerClient) GetDiagnosis(ctx context.Context, id string) (*PartnerDiagnosis, error) {
	var out PartnerDiagnosis
	err := p.client.request(ctx, http.MethodGet, "/partner/domains/"+url.PathEscape(id)+"/diagnosis", nil, &out)
	return &out, err
}

func (p *PartnerClient) RestoreDNS(ctx context.Context, id string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/restore-dns", nil, &out)
	return &out, err
}

func (p *PartnerClient) ResumeBrand(ctx context.Context, brandID string) (*MailBrand, error) {
	var out MailBrand
	err := p.client.request(ctx, http.MethodPost, "/partner/brands/"+url.PathEscape(brandID)+"/resume", nil, &out)
	return &out, err
}

func (p *PartnerClient) GetComplaints(ctx context.Context, brandID string) (*PartnerComplaintSummary, error) {
	var out PartnerComplaintSummary
	err := p.client.request(ctx, http.MethodGet, "/partner/brands/"+url.PathEscape(brandID)+"/complaints", nil, &out)
	return &out, err
}

func (p *PartnerClient) ListComplaintSources(ctx context.Context, brandID string) ([]PartnerComplaintSource, error) {
	var out []PartnerComplaintSource
	err := p.client.request(ctx, http.MethodGet, "/partner/brands/"+url.PathEscape(brandID)+"/complaints/sources", nil, &out)
	return out, err
}

func (p *PartnerClient) SuppressNeverOpened(ctx context.Context, brandID string, input PartnerBulkSuppressionInput) (*PartnerBulkSuppression, error) {
	var out PartnerBulkSuppression
	err := p.client.request(ctx, http.MethodPost, "/partner/brands/"+url.PathEscape(brandID)+"/suppressions/bulk", input, &out)
	return &out, err
}

func (p *PartnerClient) UndoSuppression(ctx context.Context, brandID, bulkID string) (int, error) {
	var out struct {
		Restored int `json:"restored"`
	}
	err := p.client.request(ctx, http.MethodPost, "/partner/brands/"+url.PathEscape(brandID)+"/suppressions/bulk/"+url.PathEscape(bulkID)+"/undo", nil, &out)
	return out.Restored, err
}

func (p *PartnerClient) ListRequests(ctx context.Context) ([]PartnerRequest, error) {
	var out struct {
		Requests []PartnerRequest `json:"requests"`
	}
	err := p.client.request(ctx, http.MethodGet, "/partner/requests", nil, &out)
	return out.Requests, err
}

func (p *PartnerClient) CreateRequest(ctx context.Context, input PartnerRequestInput) (*PartnerRequest, error) {
	var out PartnerRequest
	err := p.client.request(ctx, http.MethodPost, "/partner/requests", input, &out)
	return &out, err
}

func (p *PartnerClient) Usage(ctx context.Context, options PartnerUsageOptions) (*PartnerUsage, error) {
	query := url.Values{}
	setQuery(query, "from", options.From)
	setQuery(query, "to", options.To)
	var out PartnerUsage
	err := p.client.request(ctx, http.MethodGet, "/partner/usage"+queryString(query), nil, &out)
	return &out, err
}

func queryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}
	return "?" + query.Encode()
}
