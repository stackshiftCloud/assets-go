package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type AICreditGrant struct {
	ID             string     `json:"id"`
	Kind           string     `json:"kind"`
	Source         string     `json:"source"`
	PolicyRevision int64      `json:"policy_revision"`
	Amount         int64      `json:"amount_microcredits"`
	Remaining      int64      `json:"remaining_microcredits"`
	AvailableAt    time.Time  `json:"available_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

type AICreditBalance struct {
	Available int64           `json:"available_microcredits"`
	Monthly   int64           `json:"monthly_microcredits"`
	Purchased int64           `json:"purchased_microcredits"`
	Reserved  int64           `json:"reserved_microcredits"`
	Shortfall int64           `json:"shortfall_microcredits"`
	NextReset time.Time       `json:"next_reset"`
	Grants    []AICreditGrant `json:"grants"`
}

type AICreditUsage struct {
	ID           string    `json:"id"`
	OperationKey string    `json:"operation_key"`
	UserID       string    `json:"user_id"`
	ProjectID    *string   `json:"project_id,omitempty"`
	SpaceID      *string   `json:"space_id,omitempty"`
	ParentTaskID *string   `json:"parent_task_id,omitempty"`
	Product      string    `json:"product"`
	Feature      string    `json:"feature"`
	FundingMode  string    `json:"funding_mode"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	Reserved     int64     `json:"reserved_microcredits"`
	Used         int64     `json:"used_microcredits"`
	Returned     int64     `json:"returned_microcredits"`
}

type AICreditUsagePage struct {
	Items              []AICreditUsage  `json:"items"`
	TotalsMicrocredits map[string]int64 `json:"totals_microcredits"`
}

type AICreditQuote struct {
	ID              string    `json:"id"`
	CatalogRevision int64     `json:"catalog_revision"`
	Microcredits    int64     `json:"credit_microcredits"`
	AmountMinor     int64     `json:"amount_minor"`
	Currency        string    `json:"currency"`
	ExpiresAt       time.Time `json:"expires_at"`
}

type AICreditCheckout struct {
	Reference        string `json:"reference"`
	Status           string `json:"status"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

type AIFundingPreferences map[string]string

type AICreditsClient struct{ client *Client }

func (c *AICreditsClient) Balance(ctx context.Context) (*AICreditBalance, error) {
	var result AICreditBalance
	err := c.client.request(ctx, http.MethodGet, "/billing/ai-credits", nil, &result)
	return &result, err
}
func (c *AICreditsClient) Usage(ctx context.Context, filters url.Values) (*AICreditUsagePage, error) {
	var result AICreditUsagePage
	err := c.client.request(ctx, http.MethodGet, "/billing/ai-credits/usage?"+filters.Encode(), nil, &result)
	return &result, err
}
func (c *AICreditsClient) Quote(ctx context.Context, credits int64, currency string) (*AICreditQuote, error) {
	var result AICreditQuote
	err := c.client.request(ctx, http.MethodPost, "/billing/ai-credits/top-ups/quote", map[string]any{"credits": credits, "currency": currency}, &result)
	return &result, err
}
func (c *AICreditsClient) Checkout(ctx context.Context, quoteID, idempotencyKey, provider string) (*AICreditCheckout, error) {
	var result AICreditCheckout
	err := c.client.request(ctx, http.MethodPost, "/billing/ai-credits/top-ups/checkout", map[string]string{"quote_id": quoteID, "idempotency_key": idempotencyKey, "provider": provider}, &result)
	return &result, err
}

func (c *AICreditsClient) Funding(ctx context.Context) (AIFundingPreferences, error) {
	var result AIFundingPreferences
	err := c.client.request(ctx, http.MethodGet, "/billing/ai-credits/funding", nil, &result)
	return result, err
}

func (c *AICreditsClient) SetFunding(ctx context.Context, product, mode string) (AIFundingPreferences, error) {
	var result AIFundingPreferences
	err := c.client.request(ctx, http.MethodPut, "/billing/ai-credits/funding", map[string]string{"product": product, "mode": mode}, &result)
	return result, err
}
