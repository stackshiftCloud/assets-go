package stackshift

import (
	"context"
	"net/http"
	"net/url"
)

type MailTemplateDiagnosticRequest struct {
	VersionID string         `json:"versionId,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Stream    string         `json:"stream,omitempty"`
}

type MailDiagnosticFinding struct {
	Code           string `json:"code"`
	Severity       string `json:"severity"`
	Field          string `json:"field"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

type MailTemplateDiagnosticResult struct {
	Status           string                  `json:"status"`
	Stream           string                  `json:"stream"`
	SubjectLength    int                     `json:"subjectLength"`
	HTMLBytes        int                     `json:"htmlBytes"`
	TextBytes        int                     `json:"textBytes"`
	LinkCount        int                     `json:"linkCount"`
	RemoteImageCount int                     `json:"remoteImageCount"`
	MissingAltCount  int                     `json:"missingAltCount"`
	Findings         []MailDiagnosticFinding `json:"findings"`
}

func (m *MailTemplatesClient) Diagnostics(ctx context.Context, templateID string, input MailTemplateDiagnosticRequest) (*MailTemplateDiagnosticResult, error) {
	var out MailTemplateDiagnosticResult
	err := m.client.mailRequest(ctx, http.MethodPost, "/mail/templates/"+url.PathEscape(templateID)+"/diagnostics", input, &out)
	return &out, err
}
