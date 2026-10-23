package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type MailboxRouteTarget struct {
	Kind             string `json:"kind"`
	MailboxID        string `json:"mailboxId,omitempty"`
	GroupID          string `json:"groupId,omitempty"`
	WebhookURL       string `json:"webhookUrl,omitempty"`
	SupportChannelID string `json:"supportChannelId,omitempty"`
	ForwardTo        string `json:"forwardTo,omitempty"`
}

type MailboxDNSRecord struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Priority    *int   `json:"priority,omitempty"`
	Status      string `json:"status"`
	Required    bool   `json:"required"`
	Explanation string `json:"explanation,omitempty"`
}

type MailboxDNSApplied struct {
	At          time.Time `json:"at"`
	ByName      string    `json:"byName"`
	RecordCount int       `json:"recordCount"`
}

type MailboxDomainDNS struct {
	CloudflareConnected bool               `json:"cloudflareConnected"`
	LastApplied         *MailboxDNSApplied `json:"lastApplied"`
}

type MailboxDomain struct {
	ID                string              `json:"id"`
	WorkspaceID       string              `json:"workspaceId"`
	Domain            string              `json:"domain"`
	Status            string              `json:"status"`
	MXVerifiedAt      *time.Time          `json:"mxVerifiedAt,omitempty"`
	CatchAll          *MailboxRouteTarget `json:"catchAll,omitempty"`
	RetentionDays     *int                `json:"retentionDays,omitempty"`
	AddonStorageUnits int                 `json:"addonStorageUnits"`
	QuotaPoolBytes    int64               `json:"quotaPoolBytes"`
	UsedBytes         int64               `json:"usedBytes"`
	VerificationToken string              `json:"verificationToken"`
	Revision          int64               `json:"revision"`
	CreatedAt         time.Time           `json:"createdAt"`
	UpdatedAt         time.Time           `json:"updatedAt"`
	Records           []MailboxDNSRecord  `json:"records"`
	FoundMX           *string             `json:"foundMx"`
	DNS               MailboxDomainDNS    `json:"dns"`
}

type MailboxDomainUpdateRequest struct {
	CatchAll          *MailboxRouteTarget
	ClearCatchAll     bool
	RetentionDays     *int
	ClearRetention    bool
	AddonStorageUnits *int
}

func (r MailboxDomainUpdateRequest) body() (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	if r.ClearCatchAll {
		out["catchAll"] = json.RawMessage("null")
	} else if err := setRaw(out, "catchAll", r.CatchAll, r.CatchAll != nil); err != nil {
		return nil, err
	}
	if r.ClearRetention {
		out["retentionDays"] = json.RawMessage("null")
	} else if err := setRaw(out, "retentionDays", r.RetentionDays, r.RetentionDays != nil); err != nil {
		return nil, err
	}
	if err := setRaw(out, "addonStorageUnits", r.AddonStorageUnits, r.AddonStorageUnits != nil); err != nil {
		return nil, err
	}
	return out, nil
}

func setRaw(out map[string]json.RawMessage, key string, value any, present bool) error {
	if !present {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	out[key] = raw
	return nil
}

type MailboxInvitation struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expiresAt"`
	Expired   bool      `json:"expired"`
}

type Mailbox struct {
	ID                      string             `json:"id"`
	WorkspaceID             string             `json:"workspaceId"`
	DomainID                string             `json:"domainId"`
	LocalPart               string             `json:"localPart"`
	Address                 string             `json:"address"`
	DisplayName             string             `json:"displayName"`
	Kind                    string             `json:"kind"`
	OwnerUserID             string             `json:"ownerUserId,omitempty"`
	QuotaBytes              int64              `json:"quotaBytes"`
	UsedBytes               int64              `json:"usedBytes"`
	Plan                    string             `json:"plan"`
	Status                  string             `json:"status"`
	AIEnabled               bool               `json:"aiEnabled"`
	Revision                int64              `json:"revision"`
	CreatedAt               time.Time          `json:"createdAt"`
	UpdatedAt               time.Time          `json:"updatedAt"`
	Invitation              *MailboxInvitation `json:"invitation"`
	AIMicrocreditsThisMonth *int64             `json:"aiMicrocreditsThisMonth"`
}

type MailboxCreateRequest struct {
	DomainID    string `json:"domainId"`
	LocalPart   string `json:"localPart"`
	DisplayName string `json:"displayName,omitempty"`
	Kind        string `json:"kind,omitempty"`
	OwnerUserID string `json:"ownerUserId,omitempty"`
	OwnerEmail  string `json:"ownerEmail,omitempty"`
	Plan        string `json:"plan,omitempty"`
	QuotaBytes  int64  `json:"quotaBytes,omitempty"`
	AIEnabled   bool   `json:"aiEnabled,omitempty"`
}

type MailboxUpdateRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	QuotaBytes  *int64  `json:"quotaBytes,omitempty"`
	AIEnabled   *bool   `json:"aiEnabled,omitempty"`
	Plan        *string `json:"plan,omitempty"`
}

type MailboxMember struct {
	MailboxID string    `json:"mailboxId"`
	UserID    string    `json:"userId"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type MailboxAppPassword struct {
	ID         string     `json:"id"`
	MailboxID  string     `json:"mailboxId"`
	Label      string     `json:"label"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

type MailboxAppPasswordCreated struct {
	MailboxAppPassword
	Secret string `json:"secret"`
}

type MailboxDevice struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	Label      string     `json:"label"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type MailboxPasswordReset struct {
	Password   string `json:"password"`
	AppliedNow bool   `json:"appliedNow"`
}

type MailboxProfileLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type MailboxPlan struct {
	Slug             string `json:"slug"`
	QuotaBytes       int64  `json:"quotaBytes"`
	MonthlyPriceKobo int64  `json:"monthlyPriceKobo"`
}

type MailboxPricing struct {
	Plans                 []MailboxPlan `json:"plans"`
	AddonUnitBytes        int64         `json:"addonUnitBytes"`
	AddonPriceKobo        int64         `json:"addonPriceKobo"`
	PricesArePlaceholders bool          `json:"pricesArePlaceholders"`
}

type MailboxClientEndpoint struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Security string `json:"security"`
}

type MailboxClientSettings struct {
	Incoming *MailboxClientEndpoint `json:"incoming"`
	Outgoing *MailboxClientEndpoint `json:"outgoing"`
}

type MailboxAICredits struct {
	AvailableMicrocredits     int64     `json:"availableMicrocredits"`
	UsedThisMonthMicrocredits int64     `json:"usedThisMonthMicrocredits"`
	NextReset                 time.Time `json:"nextReset"`
}

type MailboxUsage struct {
	MailboxLimit  int               `json:"mailboxLimit"`
	MailboxesUsed int               `json:"mailboxesUsed"`
	AI            *MailboxAICredits `json:"ai"`
}

type MailboxAlias struct {
	ID        string    `json:"id"`
	MailboxID string    `json:"mailboxId"`
	Address   string    `json:"address"`
	SendAs    bool      `json:"sendAs"`
	CreatedAt time.Time `json:"createdAt"`
}

type MailboxAliasCreateRequest struct {
	MailboxID string `json:"mailboxId"`
	LocalPart string `json:"localPart"`
	DomainID  string `json:"domainId,omitempty"`
	SendAs    bool   `json:"sendAs,omitempty"`
}

type MailboxGroup struct {
	ID          string    `json:"id"`
	DomainID    string    `json:"domainId"`
	Address     string    `json:"address"`
	DisplayName string    `json:"displayName"`
	MemberIDs   []string  `json:"memberIds"`
	ModeratorID string    `json:"moderatorId,omitempty"`
	Revision    int64     `json:"revision"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type MailboxGroupCreateRequest struct {
	DomainID    string   `json:"domainId"`
	LocalPart   string   `json:"localPart"`
	DisplayName string   `json:"displayName,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
	ModeratorID string   `json:"moderatorId,omitempty"`
}

type MailboxGroupUpdateRequest struct {
	DisplayName    *string
	ModeratorID    *string
	ClearModerator bool
	MemberIDs      *[]string
}

func (r MailboxGroupUpdateRequest) body() (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	if err := setRaw(out, "displayName", r.DisplayName, r.DisplayName != nil); err != nil {
		return nil, err
	}
	if r.ClearModerator {
		out["moderatorId"] = json.RawMessage("null")
	} else if err := setRaw(out, "moderatorId", r.ModeratorID, r.ModeratorID != nil); err != nil {
		return nil, err
	}
	if r.MemberIDs != nil {
		members := *r.MemberIDs
		if members == nil {
			members = []string{}
		}
		if err := setRaw(out, "memberIds", members, true); err != nil {
			return nil, err
		}
	}
	return out, nil
}

type MailboxRoute struct {
	ID        string               `json:"id"`
	DomainID  string               `json:"domainId"`
	Pattern   string               `json:"pattern"`
	Priority  int                  `json:"priority"`
	Targets   []MailboxRouteTarget `json:"targets"`
	Revision  int64                `json:"revision"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type MailboxRouteCreateRequest struct {
	DomainID string               `json:"domainId"`
	Pattern  string               `json:"pattern"`
	Priority int                  `json:"priority,omitempty"`
	Targets  []MailboxRouteTarget `json:"targets"`
}

type MailboxRouteUpdateRequest struct {
	Priority *int                 `json:"priority,omitempty"`
	Targets  []MailboxRouteTarget `json:"targets,omitempty"`
}

type MailboxAppHealthIssue struct {
	RouteID          string    `json:"routeId"`
	Address          string    `json:"address"`
	Since            time.Time `json:"since"`
	UndeliveredCount int       `json:"undeliveredCount"`
	HolderName       *string   `json:"holderName"`
	URL              string    `json:"url"`
	LastError        string    `json:"lastError"`
	EndpointID       string    `json:"endpointId"`
}

type MailboxImportAccount struct {
	MailboxID string `json:"mailboxId"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type MailboxImportStartRequest struct {
	DomainID string                 `json:"domainId"`
	Source   string                 `json:"source"`
	Host     string                 `json:"host,omitempty"`
	Port     int                    `json:"port,omitempty"`
	Accounts []MailboxImportAccount `json:"accounts"`
}

type MailboxImportMailbox struct {
	MailboxID      string  `json:"mailboxId"`
	Address        string  `json:"address"`
	SourceUsername string  `json:"sourceUsername"`
	State          string  `json:"state"`
	TotalEmails    int64   `json:"totalEmails"`
	CopiedEmails   int64   `json:"copiedEmails"`
	TotalBytes     int64   `json:"totalBytes"`
	CopiedBytes    int64   `json:"copiedBytes"`
	MinutesLeft    *int    `json:"minutesLeft"`
	Message        *string `json:"message"`
}

type MailboxImportSwitch struct {
	RequestedFor *string    `json:"requestedFor"`
	ScheduledAt  *time.Time `json:"scheduledAt"`
	SwitchedAt   *time.Time `json:"switchedAt"`
}

type MailboxImport struct {
	ID          string                 `json:"id"`
	DomainID    string                 `json:"domainId"`
	Domain      string                 `json:"domain"`
	Source      string                 `json:"source"`
	SourceLabel string                 `json:"sourceLabel"`
	Host        string                 `json:"host"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Stage       string                 `json:"stage"`
	Switch      MailboxImportSwitch    `json:"switch"`
	Mailboxes   []MailboxImportMailbox `json:"mailboxes"`
}

type MailboxDomainsClient struct{ client *Client }
type MailboxesClient struct{ client *Client }
type MailboxAliasesClient struct{ client *Client }
type MailboxGroupsClient struct{ client *Client }
type MailboxRoutesClient struct{ client *Client }
type MailboxImportsClient struct{ client *Client }

func mailboxQuery(key, value string) string {
	if value == "" {
		return ""
	}
	return "?" + url.Values{key: {value}}.Encode()
}

func (m *MailboxDomainsClient) Create(ctx context.Context, domain string) (*MailboxDomain, error) {
	var out MailboxDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-domains", map[string]any{"domain": domain}, &out)
	return &out, err
}

func (m *MailboxDomainsClient) List(ctx context.Context) ([]MailboxDomain, error) {
	var out struct {
		Data []MailboxDomain `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-domains", nil, &out)
	return out.Data, err
}

func (m *MailboxDomainsClient) Get(ctx context.Context, id string) (*MailboxDomain, error) {
	var out MailboxDomain
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-domains/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailboxDomainsClient) Update(ctx context.Context, id string, input MailboxDomainUpdateRequest) (*MailboxDomain, error) {
	body, err := input.body()
	if err != nil {
		return nil, err
	}
	var out MailboxDomain
	err = m.client.mailRequest(ctx, http.MethodPatch, "/mail/mailbox-domains/"+url.PathEscape(id), body, &out)
	return &out, err
}

func (m *MailboxDomainsClient) Verify(ctx context.Context, id string) (*MailboxDomain, error) {
	var out MailboxDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-domains/"+url.PathEscape(id)+"/verify", nil, &out)
	return &out, err
}

func (m *MailboxDomainsClient) ApplyDNS(ctx context.Context, id string) (*MailboxDomain, error) {
	var out MailboxDomain
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-domains/"+url.PathEscape(id)+"/dns/apply", nil, &out)
	return &out, err
}

func (m *MailboxDomainsClient) DNSApplications(ctx context.Context, id string) ([]map[string]any, error) {
	var out struct {
		Data []map[string]any `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-domains/"+url.PathEscape(id)+"/dns/applications", nil, &out)
	return out.Data, err
}

func (m *MailboxDomainsClient) Delete(ctx context.Context, id string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/mailbox-domains/"+url.PathEscape(id), nil, nil)
}

func (m *MailboxesClient) Plans(ctx context.Context) (*MailboxPricing, error) {
	var out MailboxPricing
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-plans", nil, &out)
	return &out, err
}

func (m *MailboxesClient) ClientSettings(ctx context.Context) (*MailboxClientSettings, error) {
	var out MailboxClientSettings
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-client-settings", nil, &out)
	return &out, err
}

func (m *MailboxesClient) Usage(ctx context.Context) (*MailboxUsage, error) {
	var out MailboxUsage
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-usage", nil, &out)
	return &out, err
}

func (m *MailboxesClient) Create(ctx context.Context, input MailboxCreateRequest) (*Mailbox, error) {
	var out Mailbox
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes", input, &out)
	return &out, err
}

func (m *MailboxesClient) List(ctx context.Context, domainID string) ([]Mailbox, error) {
	var out struct {
		Data []Mailbox `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailboxes"+mailboxQuery("domainId", domainID), nil, &out)
	return out.Data, err
}

func (m *MailboxesClient) Get(ctx context.Context, id string) (*Mailbox, error) {
	var out Mailbox
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailboxes/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailboxesClient) Update(ctx context.Context, id string, input MailboxUpdateRequest) (*Mailbox, error) {
	var out Mailbox
	err := m.client.mailRequest(ctx, http.MethodPatch, "/mail/mailboxes/"+url.PathEscape(id), input, &out)
	return &out, err
}

func (m *MailboxesClient) Delete(ctx context.Context, id string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/mailboxes/"+url.PathEscape(id), nil, nil)
}

func (m *MailboxesClient) Suspend(ctx context.Context, id string) (*Mailbox, error) {
	var out Mailbox
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/suspend", nil, &out)
	return &out, err
}

func (m *MailboxesClient) Resume(ctx context.Context, id string) (*Mailbox, error) {
	var out Mailbox
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/resume", nil, &out)
	return &out, err
}

func (m *MailboxesClient) ResendInvitation(ctx context.Context, id string) (*Mailbox, error) {
	var out Mailbox
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/invitation/resend", nil, &out)
	return &out, err
}

func (m *MailboxesClient) ResetPassword(ctx context.Context, id string) (*MailboxPasswordReset, error) {
	var out MailboxPasswordReset
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/password-reset", nil, &out)
	return &out, err
}

func (m *MailboxesClient) CreateProfileLink(ctx context.Context, id string) (*MailboxProfileLink, error) {
	var out MailboxProfileLink
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/mobileconfig", nil, &out)
	return &out, err
}

func (m *MailboxesClient) ListDevices(ctx context.Context, id string) ([]MailboxDevice, error) {
	var out struct {
		Data []MailboxDevice `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailboxes/"+url.PathEscape(id)+"/devices", nil, &out)
	return out.Data, err
}

func (m *MailboxesClient) RevokeDevice(ctx context.Context, id, kind, deviceID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/mailboxes/"+url.PathEscape(id)+"/devices/"+url.PathEscape(kind)+"/"+url.PathEscape(deviceID), nil, nil)
}

func (m *MailboxesClient) ListMembers(ctx context.Context, id string) ([]MailboxMember, error) {
	var out struct {
		Data []MailboxMember `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailboxes/"+url.PathEscape(id)+"/members", nil, &out)
	return out.Data, err
}

func (m *MailboxesClient) AddMember(ctx context.Context, id, userID string) (*MailboxMember, error) {
	var out MailboxMember
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/members", map[string]any{"userId": userID}, &out)
	return &out, err
}

func (m *MailboxesClient) RemoveMember(ctx context.Context, id, userID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/mailboxes/"+url.PathEscape(id)+"/members/"+url.PathEscape(userID), nil, nil)
}

func (m *MailboxesClient) ListAppPasswords(ctx context.Context, id string) ([]MailboxAppPassword, error) {
	var out struct {
		Data []MailboxAppPassword `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailboxes/"+url.PathEscape(id)+"/app-passwords", nil, &out)
	return out.Data, err
}

func (m *MailboxesClient) CreateAppPassword(ctx context.Context, id, label string) (*MailboxAppPasswordCreated, error) {
	var out MailboxAppPasswordCreated
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailboxes/"+url.PathEscape(id)+"/app-passwords", map[string]any{"label": label}, &out)
	return &out, err
}

func (m *MailboxesClient) RevokeAppPassword(ctx context.Context, id, passwordID string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/mailboxes/"+url.PathEscape(id)+"/app-passwords/"+url.PathEscape(passwordID), nil, nil)
}

func (m *MailboxAliasesClient) Create(ctx context.Context, input MailboxAliasCreateRequest) (*MailboxAlias, error) {
	var out MailboxAlias
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/aliases", input, &out)
	return &out, err
}

func (m *MailboxAliasesClient) List(ctx context.Context, mailboxID string) ([]MailboxAlias, error) {
	var out struct {
		Data []MailboxAlias `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/aliases"+mailboxQuery("mailboxId", mailboxID), nil, &out)
	return out.Data, err
}

func (m *MailboxAliasesClient) Update(ctx context.Context, id string, sendAs bool) (*MailboxAlias, error) {
	var out MailboxAlias
	err := m.client.mailRequest(ctx, http.MethodPatch, "/mail/aliases/"+url.PathEscape(id), struct {
		SendAs bool `json:"sendAs"`
	}{SendAs: sendAs}, &out)
	return &out, err
}

func (m *MailboxAliasesClient) Delete(ctx context.Context, id string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/aliases/"+url.PathEscape(id), nil, nil)
}

func (m *MailboxGroupsClient) Create(ctx context.Context, input MailboxGroupCreateRequest) (*MailboxGroup, error) {
	var out MailboxGroup
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/groups", input, &out)
	return &out, err
}

func (m *MailboxGroupsClient) List(ctx context.Context, domainID string) ([]MailboxGroup, error) {
	var out struct {
		Data []MailboxGroup `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/groups"+mailboxQuery("domainId", domainID), nil, &out)
	return out.Data, err
}

func (m *MailboxGroupsClient) Get(ctx context.Context, id string) (*MailboxGroup, error) {
	var out MailboxGroup
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/groups/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailboxGroupsClient) Update(ctx context.Context, id string, input MailboxGroupUpdateRequest) (*MailboxGroup, error) {
	body, err := input.body()
	if err != nil {
		return nil, err
	}
	var out MailboxGroup
	err = m.client.mailRequest(ctx, http.MethodPatch, "/mail/groups/"+url.PathEscape(id), body, &out)
	return &out, err
}

func (m *MailboxGroupsClient) Delete(ctx context.Context, id string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/groups/"+url.PathEscape(id), nil, nil)
}

func (m *MailboxRoutesClient) Create(ctx context.Context, input MailboxRouteCreateRequest) (*MailboxRoute, error) {
	var out MailboxRoute
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/routes", input, &out)
	return &out, err
}

func (m *MailboxRoutesClient) List(ctx context.Context, domainID string) ([]MailboxRoute, error) {
	var out struct {
		Data []MailboxRoute `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/routes"+mailboxQuery("domainId", domainID), nil, &out)
	return out.Data, err
}

func (m *MailboxRoutesClient) Get(ctx context.Context, id string) (*MailboxRoute, error) {
	var out MailboxRoute
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/routes/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailboxRoutesClient) Update(ctx context.Context, id string, input MailboxRouteUpdateRequest) (*MailboxRoute, error) {
	var out MailboxRoute
	err := m.client.mailRequest(ctx, http.MethodPatch, "/mail/routes/"+url.PathEscape(id), input, &out)
	return &out, err
}

func (m *MailboxRoutesClient) Delete(ctx context.Context, id string) error {
	return m.client.mailRequest(ctx, http.MethodDelete, "/mail/routes/"+url.PathEscape(id), nil, nil)
}

func (m *MailboxRoutesClient) AppHealth(ctx context.Context, domainID string) (*MailboxAppHealthIssue, error) {
	var out struct {
		Issue *MailboxAppHealthIssue `json:"issue"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/routes/app-health"+mailboxQuery("domainId", domainID), nil, &out)
	return out.Issue, err
}

func (m *MailboxRoutesClient) RetryDeliveries(ctx context.Context, id string) (int, error) {
	var out struct {
		Retried int `json:"retried"`
	}
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/routes/"+url.PathEscape(id)+"/retry-deliveries", nil, &out)
	return out.Retried, err
}

func (m *MailboxImportsClient) Start(ctx context.Context, input MailboxImportStartRequest) (*MailboxImport, error) {
	var out MailboxImport
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-imports", input, &out)
	return &out, err
}

func (m *MailboxImportsClient) List(ctx context.Context) ([]MailboxImport, error) {
	var out struct {
		Data []MailboxImport `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-imports", nil, &out)
	return out.Data, err
}

func (m *MailboxImportsClient) Get(ctx context.Context, id string) (*MailboxImport, error) {
	var out MailboxImport
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/mailbox-imports/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailboxImportsClient) Cancel(ctx context.Context, id string) (*MailboxImport, error) {
	var out MailboxImport
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-imports/"+url.PathEscape(id)+"/cancel", nil, &out)
	return &out, err
}

func (m *MailboxImportsClient) Skip(ctx context.Context, id, mailboxID string) (*MailboxImport, error) {
	var out MailboxImport
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-imports/"+url.PathEscape(id)+"/mailboxes/"+url.PathEscape(mailboxID)+"/skip", nil, &out)
	return &out, err
}

func (m *MailboxImportsClient) Switch(ctx context.Context, id, when string) (*MailboxImport, error) {
	var out MailboxImport
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/mailbox-imports/"+url.PathEscape(id)+"/switch", map[string]any{"when": when}, &out)
	return &out, err
}
