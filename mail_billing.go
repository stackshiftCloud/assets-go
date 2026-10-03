package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type MailBillingSettings struct {
	PrepaidEnabled bool  `json:"prepaidEnabled"`
	CapKobo        int64 `json:"capKobo"`
	Revision       int   `json:"revision"`
}
type MailBillingSummary struct {
	CatalogID              string              `json:"catalogId,omitempty"`
	PlanSlug               string              `json:"planSlug,omitempty"`
	MonthlyPriceKobo       int64               `json:"monthlyPriceKobo"`
	OverageAllowed         *bool               `json:"overageAllowed,omitempty"`
	ReconciliationRequired bool                `json:"reconciliationRequired"`
	Active                 bool                `json:"active"`
	Mode                   MailMode            `json:"mode"`
	Currency               string              `json:"currency"`
	PeriodStart            string              `json:"periodStart"`
	PeriodEnd              string              `json:"periodEnd"`
	RateVersion            int64               `json:"rateVersion"`
	RecipientPriceKobo     int64               `json:"recipientPriceKobo"`
	Included               int64               `json:"included"`
	Used                   int64               `json:"used"`
	Reserved               int64               `json:"reserved"`
	Remaining              int64               `json:"remaining"`
	PrepaidBalanceKobo     int64               `json:"prepaidBalanceKobo"`
	SpentKobo              int64               `json:"spentKobo"`
	ReservedKobo           int64               `json:"reservedKobo"`
	Settings               MailBillingSettings `json:"settings"`
}
type MailCampaignEstimate struct {
	MailBillingSummary
	RecipientCount     int64    `json:"recipientCount"`
	RequiredCreditKobo int64    `json:"requiredCreditKobo"`
	BlockingReasons    []string `json:"blockingReasons"`
}
type MailCampaignBudget struct {
	Revision        int   `json:"revision"`
	ApprovedMaxKobo int64 `json:"approvedMaxKobo"`
}
type MailBillingClient struct{ client *Client }

func (m *MailBillingClient) Summary(ctx context.Context) (*MailBillingSummary, error) {
	var out MailBillingSummary
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/billing", nil, &out)
	return &out, err
}
func (m *MailCampaignsClient) Estimate(ctx context.Context, id string) (*MailCampaignEstimate, error) {
	var out MailCampaignEstimate
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/campaigns/"+url.PathEscape(id)+"/estimate", nil, &out)
	return &out, err
}
func (m *MailCampaignsClient) SendWithBudget(ctx context.Context, id string, budget MailCampaignBudget) (*MailCampaign, error) {
	var out MailCampaign
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns/"+url.PathEscape(id)+"/send", budget, &out)
	return &out, err
}
func (m *MailCampaignsClient) ScheduleWithBudget(ctx context.Context, id string, sendAt time.Time, budget MailCampaignBudget) (*MailCampaign, error) {
	var out MailCampaign
	input := struct {
		MailCampaignBudget
		SendAt time.Time `json:"sendAt"`
	}{budget, sendAt}
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/campaigns/"+url.PathEscape(id)+"/schedule", input, &out)
	return &out, err
}
