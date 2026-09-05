package stackshift

import (
	"context"
	"net/http"
	"net/url"
)

type MailBrandFields struct {
	Name            string   `json:"name"`
	LogoURL         string   `json:"logoUrl"`
	SendingDomainID *string  `json:"sendingDomainId"`
	DefaultFrom     string   `json:"defaultFrom"`
	ReplyTo         string   `json:"replyTo"`
	PostalAddress   string   `json:"postalAddress"`
	Website         string   `json:"website"`
	SupportEmail    string   `json:"supportEmail"`
	EditableFields  []string `json:"editableFields"`
}

type MailBrand struct {
	MailBrandFields
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspaceId"`
	Mode        MailMode `json:"mode"`
	Status      string   `json:"status"`
	Revision    int      `json:"revision"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type MailBrandUpdateRequest struct {
	MailBrandFields
	Revision int `json:"revision"`
}

type MailBrandsClient struct{ client *Client }

func (m *MailBrandsClient) Create(ctx context.Context, input MailBrandFields) (*MailBrand, error) {
	var out MailBrand
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/brands", input, &out)
	return &out, err
}

func (m *MailBrandsClient) List(ctx context.Context) ([]MailBrand, error) {
	var out struct {
		Data []MailBrand `json:"data"`
	}
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/brands", nil, &out)
	return out.Data, err
}

func (m *MailBrandsClient) Get(ctx context.Context, id string) (*MailBrand, error) {
	var out MailBrand
	err := m.client.mailRequest(ctx, http.MethodGet, "/mail/brands/"+url.PathEscape(id), nil, &out)
	return &out, err
}

func (m *MailBrandsClient) Update(ctx context.Context, id string, input MailBrandUpdateRequest) (*MailBrand, error) {
	var out MailBrand
	err := m.client.mailRequest(ctx, http.MethodPut, "/mail/brands/"+url.PathEscape(id), input, &out)
	return &out, err
}

func (m *MailBrandsClient) Archive(ctx context.Context, id string) (*MailBrand, error) {
	var out MailBrand
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/brands/"+url.PathEscape(id)+"/archive", nil, &out)
	return &out, err
}
