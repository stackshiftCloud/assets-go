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

	StorefrontDomainLimit int     `json:"storefrontDomainLimit"`
	MailboxLimit          int     `json:"mailboxLimit"`
	DisplayName           *string `json:"displayName"`
	SupportEmail          *string `json:"supportEmail"`
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
	StorefrontDomainLimit int     `json:"storefrontDomainLimit"`
	StorefrontDomainsUsed int     `json:"storefrontDomainsUsed"`
	MailboxLimit          int     `json:"mailboxLimit"`
	MailboxesUsed         int     `json:"mailboxesUsed"`
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
	Mode                 string `json:"mode"`
	ForwardTo            string `json:"forwardTo,omitempty"`
	MailboxRecoveryEmail string `json:"mailboxRecoveryEmail,omitempty"`
	MailboxPlan          string `json:"mailboxPlan,omitempty"`
	SendPasswordLink     *bool  `json:"sendPasswordLink,omitempty"`
}

type PartnerWebsiteInput struct {
	ProjectID string `json:"projectId"`
}

type PartnerDNSInput struct {
	Method          string `json:"method"`
	CloudflareToken string `json:"cloudflareToken,omitempty"`
}

type PartnerDomainPurchaseInput struct {
	Domain      string               `json:"domain"`
	Years       int                  `json:"years"`
	Currency    string               `json:"currency,omitempty"`
	MerchantRef string               `json:"merchantRef"`
	Registrant  PartnerRegistrant    `json:"registrant"`
	Brand       PartnerBrandInput    `json:"brand"`
	Replies     PartnerRepliesInput  `json:"replies"`
	Website     *PartnerWebsiteInput `json:"website,omitempty"`
}

type PartnerDomainConnectInput struct {
	Domain      string               `json:"domain"`
	MerchantRef string               `json:"merchantRef"`
	Brand       PartnerBrandInput    `json:"brand"`
	Replies     PartnerRepliesInput  `json:"replies"`
	Website     *PartnerWebsiteInput `json:"website,omitempty"`
	DNS         PartnerDNSInput      `json:"dns"`
	Registrant  *PartnerRegistrant   `json:"registrant,omitempty"`
}

type PartnerDomainTransferInInput struct {
	PartnerDomainConnectInput
	AuthCode   string   `json:"authCode"`
	Years      int      `json:"years,omitempty"`
	Currency   string   `json:"currency,omitempty"`
	ExtraHosts []string `json:"extraHosts,omitempty"`
}

type PartnerConnectedTransferInInput struct {
	AuthCode   string   `json:"authCode"`
	Years      int      `json:"years,omitempty"`
	Currency   string   `json:"currency,omitempty"`
	ExtraHosts []string `json:"extraHosts,omitempty"`
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

type PartnerWebsiteHost struct {
	Hostname      string   `json:"hostname"`
	DNS           string   `json:"dns"`
	Found         []string `json:"found"`
	HTTPS         string   `json:"https"`
	HTTPSRenewsAt *string  `json:"httpsRenewsAt"`
}

type PartnerWebsiteRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	State string `json:"state"`
}

type PartnerWebsiteQueue struct {
	Ahead           int    `json:"ahead"`
	EstimatedLiveAt string `json:"estimatedLiveAt"`
}

type PartnerWebsite struct {
	ProjectID   string                 `json:"projectId"`
	ProjectName string                 `json:"projectName"`
	State       string                 `json:"state"`
	StartedAt   *string                `json:"startedAt"`
	Hosts       []PartnerWebsiteHost   `json:"hosts"`
	Records     []PartnerWebsiteRecord `json:"records"`
	Queue       *PartnerWebsiteQueue   `json:"queue"`
	LiveAt      *string                `json:"liveAt"`
	Failure     *PartnerFailure        `json:"failure"`
}

type PartnerWebsiteProject struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Hostname         string `json:"hostname"`
	MerchantWebsites int    `json:"merchantWebsites"`
}

type PartnerDNSRecordCopy struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PartnerTransfer struct {
	State        string                 `json:"state"`
	StartedAt    *string                `json:"startedAt"`
	CompletedAt  *string                `json:"completedAt"`
	AmountMinor  int64                  `json:"amountMinor"`
	Currency     string                 `json:"currency"`
	Refunded     bool                   `json:"refunded"`
	DNSCopied    []PartnerDNSRecordCopy `json:"dnsCopied"`
	DNSNotCopied []PartnerDNSRecordCopy `json:"dnsNotCopied"`
	Failure      *PartnerFailure        `json:"failure"`
}

type PartnerConnect struct {
	Method      string                 `json:"method"`
	Connected   bool                   `json:"connected"`
	Nameservers []string               `json:"nameservers"`
	Records     []PartnerWebsiteRecord `json:"records"`
	CheckedAt   *string                `json:"checkedAt"`
	DeadlineAt  *string                `json:"deadlineAt"`
}

type PartnerMailStatus struct {
	Migration *string `json:"migration"`
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
	Source             string              `json:"source"`
	ConnectMethod      *string             `json:"connectMethod"`
	Connect            *PartnerConnect     `json:"connect"`
	Website            *PartnerWebsite     `json:"website"`
	Transfer           *PartnerTransfer    `json:"transfer"`
	Mail               PartnerMailStatus   `json:"mail"`
}

type PartnerDomainListOptions struct {
	MerchantRef string
	Query       string
	Attention   bool
	Website     string
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
	PartnerDomainID         string         `json:"partnerDomainId"`
	MerchantRef             string         `json:"merchantRef"`
	Domain                  string         `json:"domain"`
	BrandID                 *string        `json:"brandId"`
	BrandName               string         `json:"brandName"`
	MarketingRecipients     int            `json:"marketingRecipients"`
	TransactionalRecipients int            `json:"transactionalRecipients"`
	DomainChargesMinor      int64          `json:"domainChargesMinor"`
	MailboxCharges          []PartnerMoney `json:"mailboxCharges"`
	RefundedMinor           int64          `json:"refundedMinor"`
	Currency                string         `json:"currency"`
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

type PartnerMoney struct {
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}

type PartnerMailbox struct {
	ID                 string        `json:"id"`
	Address            string        `json:"address"`
	Kind               string        `json:"kind"`
	DisplayName        string        `json:"displayName"`
	DeliversTo         *string       `json:"deliversTo"`
	Status             string        `json:"status"`
	RecoveryEmail      *string       `json:"recoveryEmail"`
	PasswordSetAt      *string       `json:"passwordSetAt"`
	PasswordLinkSentAt *string       `json:"passwordLinkSentAt"`
	LastSignInAt       *string       `json:"lastSignInAt"`
	LastSignInApp      *string       `json:"lastSignInApp"`
	StorageUsedBytes   *int64        `json:"storageUsedBytes"`
	QuotaBytes         *int64        `json:"quotaBytes"`
	Plan               *string       `json:"plan"`
	Price              *PartnerMoney `json:"price"`
	RenewsAt           *string       `json:"renewsAt"`
}

type PartnerMailboxRemoved struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

type PartnerMailboxInput struct {
	LocalPart        string `json:"localPart"`
	DisplayName      string `json:"displayName,omitempty"`
	RecoveryEmail    string `json:"recoveryEmail,omitempty"`
	Plan             string `json:"plan,omitempty"`
	SendPasswordLink *bool  `json:"sendPasswordLink,omitempty"`
}

type PartnerMailboxUpdate struct {
	DisplayName   *string `json:"displayName,omitempty"`
	RecoveryEmail *string `json:"recoveryEmail,omitempty"`
}

type PartnerPasswordLinkInput struct {
	Reset bool `json:"reset"`
	Send  bool `json:"send"`
}

type PartnerPasswordLink struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt"`
}

type PartnerMailboxCreated struct {
	Mailbox      PartnerMailbox       `json:"mailbox"`
	PasswordLink *PartnerPasswordLink `json:"passwordLink"`
}

type PartnerAliasInput struct {
	LocalPart string `json:"localPart"`
	DeliverTo string `json:"deliverTo"`
}

type PartnerAliasUpdate struct {
	DeliverTo string `json:"deliverTo"`
}

type PartnerMailMigrationAccountInput struct {
	MailboxID string `json:"mailboxId"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type PartnerMailMigrationInput struct {
	Source   string                             `json:"source"`
	Host     string                             `json:"host,omitempty"`
	Port     int                                `json:"port,omitempty"`
	Accounts []PartnerMailMigrationAccountInput `json:"accounts"`
	Switch   string                             `json:"switch"`
}

type PartnerMailMigrationAccount struct {
	MailboxID      string  `json:"mailboxId"`
	Address        string  `json:"address"`
	Username       string  `json:"username"`
	State          string  `json:"state"`
	CopiedEmails   int64   `json:"copiedEmails"`
	TotalEmails    int64   `json:"totalEmails"`
	CopiedBytes    int64   `json:"copiedBytes"`
	TotalBytes     int64   `json:"totalBytes"`
	CaughtUpEmails *int64  `json:"caughtUpEmails"`
	MinutesLeft    *int    `json:"minutesLeft"`
	Message        *string `json:"message"`
}

type PartnerMailMigrationRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority"`
	Found    bool   `json:"found"`
}

type PartnerMailMigrationSwitchState struct {
	RequestedFor *string `json:"requestedFor"`
	ScheduledAt  *string `json:"scheduledAt"`
	SwitchedAt   *string `json:"switchedAt"`
	AnsweredAt   *string `json:"answeredAt"`
}

type PartnerMailMigrationEmails struct {
	Copied   int64 `json:"copied"`
	Total    int64 `json:"total"`
	CaughtUp int64 `json:"caughtUp"`
}

type PartnerMailMigration struct {
	ID          string                          `json:"id"`
	State       string                          `json:"state"`
	StartedBy   string                          `json:"startedBy"`
	Source      string                          `json:"source"`
	Host        string                          `json:"host"`
	Switch      PartnerMailMigrationSwitchState `json:"switch"`
	CatchUpAt   *string                         `json:"catchUpAt"`
	CompletedAt *string                         `json:"completedAt"`
	CreatedAt   string                          `json:"createdAt"`
	UpdatedAt   string                          `json:"updatedAt"`
	Emails      PartnerMailMigrationEmails      `json:"emails"`
	Mailboxes   []PartnerMailMigrationAccount   `json:"mailboxes"`
	Records     []PartnerMailMigrationRecord    `json:"records"`
	Failure     *PartnerFailure                 `json:"failure"`
}

type PartnerMailMigrationCredentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type PartnerMailSwitchResult struct {
	Switch    PartnerMailMigrationSwitchState `json:"switch"`
	Migration *PartnerMailMigration           `json:"migration"`
}

type PartnerMailMigrationSwitchInput struct {
	When string `json:"when"`
}

type PartnerWebsiteLiveEvent struct {
	PartnerEventBase
	ProjectID string   `json:"projectId"`
	Hosts     []string `json:"hosts"`
}

type PartnerWebsiteFailedEvent struct {
	PartnerEventBase
	Failure PartnerFailure `json:"failure"`
}

type PartnerWebsiteDNSChangedEvent struct {
	PartnerEventBase
	Host  string   `json:"host"`
	Found []string `json:"found"`
}

type PartnerWebsiteMovedEvent struct {
	PartnerEventBase
	FromProjectID string `json:"fromProjectId"`
	ProjectID     string `json:"projectId"`
}

type PartnerWebsiteRemovedEvent struct {
	PartnerEventBase
	Reason string `json:"reason"`
}

type PartnerWebsiteCertificateExpiringEvent struct {
	PartnerEventBase
	Host      string `json:"host"`
	ExpiresAt string `json:"expiresAt"`
}

type PartnerDomainConnectEvent struct {
	PartnerEventBase
	Method string `json:"method"`
	Reason string `json:"reason,omitempty"`
}

type PartnerDomainTransferStartedEvent struct {
	PartnerEventBase
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}

type PartnerDomainTransferWaitingEvent struct {
	PartnerEventBase
	Reason       string                 `json:"reason"`
	AmountMinor  int64                  `json:"amountMinor,omitempty"`
	Currency     string                 `json:"currency,omitempty"`
	DNSNotCopied []PartnerDNSRecordCopy `json:"dnsNotCopied,omitempty"`
}

type PartnerDomainTransferredInEvent struct {
	PartnerEventBase
	DNSCopied          []PartnerDNSRecordCopy `json:"dnsCopied"`
	DNSNotCopied       []PartnerDNSRecordCopy `json:"dnsNotCopied"`
	RegisteredDomainID string                 `json:"registeredDomainId"`
}

type PartnerDomainTransferFailedEvent struct {
	PartnerEventBase
	Failure  PartnerFailure `json:"failure"`
	Refunded bool           `json:"refunded"`
}

type PartnerMailboxCreatedEvent struct {
	PartnerEventBase
	MailboxID        string `json:"mailboxId"`
	Address          string `json:"address"`
	PasswordLinkSent bool   `json:"passwordLinkSent"`
}

type PartnerMailboxEvent struct {
	PartnerEventBase
	MailboxID string `json:"mailboxId"`
	Address   string `json:"address"`
	By        string `json:"by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type PartnerMailMigrationEvent struct {
	PartnerEventBase
	MigrationID  string          `json:"migrationId"`
	Source       string          `json:"source,omitempty"`
	MailboxCount int             `json:"mailboxCount,omitempty"`
	StartedBy    string          `json:"startedBy,omitempty"`
	Emails       int64           `json:"emails,omitempty"`
	CaughtUp     int64           `json:"caughtUp,omitempty"`
	SwitchedAt   string          `json:"switchedAt,omitempty"`
	CatchUpAt    *string         `json:"catchUpAt,omitempty"`
	Failure      *PartnerFailure `json:"failure,omitempty"`
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
	setQuery(query, "website", options.Website)
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

func (p *PartnerClient) ConnectDomain(ctx context.Context, input PartnerDomainConnectInput, idempotencyKey string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.requestWithHeaders(ctx, http.MethodPost, "/partner/domains/connect", input, map[string]string{"Idempotency-Key": idempotencyKey}, &out)
	return &out, err
}

func (p *PartnerClient) TransferInDomain(ctx context.Context, input PartnerDomainTransferInInput, idempotencyKey string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.requestWithHeaders(ctx, http.MethodPost, "/partner/domains/transfer-in", input, map[string]string{"Idempotency-Key": idempotencyKey}, &out)
	return &out, err
}

func (p *PartnerClient) TransferInConnectedDomain(ctx context.Context, id string, input PartnerConnectedTransferInInput, idempotencyKey string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.requestWithHeaders(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/transfer-in", input, map[string]string{"Idempotency-Key": idempotencyKey}, &out)
	return &out, err
}

func (p *PartnerClient) SwitchTransferNameservers(ctx context.Context, id string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/transfer-in/switch-nameservers", nil, &out)
	return &out, err
}

func (p *PartnerClient) SetWebsite(ctx context.Context, id string, input PartnerWebsiteInput) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.request(ctx, http.MethodPut, "/partner/domains/"+url.PathEscape(id)+"/website", input, &out)
	return &out, err
}

func (p *PartnerClient) RemoveWebsite(ctx context.Context, id string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.request(ctx, http.MethodDelete, "/partner/domains/"+url.PathEscape(id)+"/website", nil, &out)
	return &out, err
}

func (p *PartnerClient) CheckWebsite(ctx context.Context, id string) (*PartnerMerchantDomain, error) {
	var out PartnerMerchantDomain
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/website/check", nil, &out)
	return &out, err
}

func (p *PartnerClient) ListWebsiteProjects(ctx context.Context) ([]PartnerWebsiteProject, error) {
	var out []PartnerWebsiteProject
	err := p.client.request(ctx, http.MethodGet, "/partner/website-projects", nil, &out)
	return out, err
}

func (p *PartnerClient) ListMailboxes(ctx context.Context, id string) ([]PartnerMailbox, error) {
	var out []PartnerMailbox
	err := p.client.request(ctx, http.MethodGet, "/partner/domains/"+url.PathEscape(id)+"/mailboxes", nil, &out)
	return out, err
}

func (p *PartnerClient) CreateMailbox(ctx context.Context, id string, input PartnerMailboxInput, idempotencyKey string) (*PartnerMailboxCreated, error) {
	var out PartnerMailboxCreated
	err := p.client.requestWithHeaders(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mailboxes", input, map[string]string{"Idempotency-Key": idempotencyKey}, &out)
	return &out, err
}

func (p *PartnerClient) UpdateMailbox(ctx context.Context, id, mailboxID string, input PartnerMailboxUpdate) (*PartnerMailbox, error) {
	var out PartnerMailbox
	err := p.client.request(ctx, http.MethodPatch, "/partner/domains/"+url.PathEscape(id)+"/mailboxes/"+url.PathEscape(mailboxID), input, &out)
	return &out, err
}

func (p *PartnerClient) CreateMailboxPasswordLink(ctx context.Context, id, mailboxID string, input PartnerPasswordLinkInput) (*PartnerPasswordLink, error) {
	var out PartnerPasswordLink
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mailboxes/"+url.PathEscape(mailboxID)+"/password-link", input, &out)
	return &out, err
}

func (p *PartnerClient) PauseMailbox(ctx context.Context, id, mailboxID string) (*PartnerMailbox, error) {
	var out PartnerMailbox
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mailboxes/"+url.PathEscape(mailboxID)+"/pause", nil, &out)
	return &out, err
}

func (p *PartnerClient) ResumeMailbox(ctx context.Context, id, mailboxID string) (*PartnerMailbox, error) {
	var out PartnerMailbox
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mailboxes/"+url.PathEscape(mailboxID)+"/resume", nil, &out)
	return &out, err
}

func (p *PartnerClient) RemoveMailbox(ctx context.Context, id, mailboxID string) (*PartnerMailboxRemoved, error) {
	var out PartnerMailboxRemoved
	err := p.client.request(ctx, http.MethodDelete, "/partner/domains/"+url.PathEscape(id)+"/mailboxes/"+url.PathEscape(mailboxID), nil, &out)
	return &out, err
}

func (p *PartnerClient) CreateAlias(ctx context.Context, id string, input PartnerAliasInput) (*PartnerMailbox, error) {
	var out PartnerMailbox
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/aliases", input, &out)
	return &out, err
}

func (p *PartnerClient) UpdateAlias(ctx context.Context, id, aliasID string, input PartnerAliasUpdate) (*PartnerMailbox, error) {
	var out PartnerMailbox
	err := p.client.request(ctx, http.MethodPatch, "/partner/domains/"+url.PathEscape(id)+"/aliases/"+url.PathEscape(aliasID), input, &out)
	return &out, err
}

func (p *PartnerClient) RemoveAlias(ctx context.Context, id, aliasID string) error {
	return p.client.request(ctx, http.MethodDelete, "/partner/domains/"+url.PathEscape(id)+"/aliases/"+url.PathEscape(aliasID), nil, nil)
}

func (p *PartnerClient) StartMailMigration(ctx context.Context, id string, input PartnerMailMigrationInput, idempotencyKey string) (*PartnerMailMigration, error) {
	var out PartnerMailMigration
	err := p.client.requestWithHeaders(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mail-migration", input, map[string]string{"Idempotency-Key": idempotencyKey}, &out)
	return &out, err
}

func (p *PartnerClient) GetMailMigration(ctx context.Context, id string) (*PartnerMailMigration, error) {
	var out PartnerMailMigration
	err := p.client.request(ctx, http.MethodGet, "/partner/domains/"+url.PathEscape(id)+"/mail-migration", nil, &out)
	return &out, err
}

func (p *PartnerClient) SwitchMailMigration(ctx context.Context, id string, input PartnerMailMigrationSwitchInput) (*PartnerMailSwitchResult, error) {
	var out PartnerMailSwitchResult
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mail-migration/switch", input, &out)
	return &out, err
}

func (p *PartnerClient) CancelMailMigration(ctx context.Context, id string) (*PartnerMailMigration, error) {
	var out PartnerMailMigration
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mail-migration/cancel", nil, &out)
	return &out, err
}

func (p *PartnerClient) RetryMailMigration(ctx context.Context, id string) (*PartnerMailMigration, error) {
	var out PartnerMailMigration
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mail-migration/retry", nil, &out)
	return &out, err
}

func (p *PartnerClient) RetryMailMigrationAccount(ctx context.Context, id, mailboxID string, credentials *PartnerMailMigrationCredentials) (*PartnerMailMigration, error) {
	var out PartnerMailMigration
	var body any
	if credentials != nil {
		body = *credentials
	}
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mail-migration/accounts/"+url.PathEscape(mailboxID)+"/retry", body, &out)
	return &out, err
}

func (p *PartnerClient) SkipMailMigrationAccount(ctx context.Context, id, mailboxID string) (*PartnerMailMigration, error) {
	var out PartnerMailMigration
	err := p.client.request(ctx, http.MethodPost, "/partner/domains/"+url.PathEscape(id)+"/mail-migration/accounts/"+url.PathEscape(mailboxID)+"/skip", nil, &out)
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
