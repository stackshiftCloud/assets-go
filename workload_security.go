package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type WorkloadSecurityClient struct {
	client *Client
}

type WorkloadAdmission struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id,omitempty"`
	DeploymentID   string    `json:"deployment_id,omitempty"`
	Environment    string    `json:"environment"`
	ExecutionKind  string    `json:"execution_kind"`
	ImageReference string    `json:"image_reference"`
	ImageDigest    string    `json:"image_digest"`
	PolicyVersion  string    `json:"policy_version"`
	Decision       string    `json:"decision"`
	Evidence       any       `json:"sanitized_evidence"`
	ReceiptDigest  string    `json:"receipt_digest"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type WorkloadExternalDependency struct {
	ID                   string    `json:"id"`
	ProjectID            string    `json:"project_id"`
	Environment          string    `json:"environment"`
	Hostname             string    `json:"hostname"`
	Port                 int       `json:"port"`
	Protocol             string    `json:"protocol"`
	Purpose              string    `json:"purpose"`
	TLSRequired          bool      `json:"tls_required"`
	PrivateRouteRequired bool      `json:"private_route_required"`
	Detected             bool      `json:"detected"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type WorkloadEgressGrant struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	DependencyID      string    `json:"dependency_id,omitempty"`
	Environment       string    `json:"environment"`
	Protocol          string    `json:"protocol"`
	Hostname          string    `json:"hostname,omitempty"`
	CIDR              string    `json:"cidr,omitempty"`
	PortStart         int       `json:"port_start"`
	PortEnd           int       `json:"port_end"`
	Purpose           string    `json:"purpose"`
	Status            string    `json:"status"`
	BroadPublicEgress bool      `json:"broad_public_egress"`
	ExpiresAt         time.Time `json:"expires_at"`
	ResolvedCIDRs     []string  `json:"resolved_cidrs,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WorkloadPrivateRoute struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	NodeID      string    `json:"node_id"`
	Environment string    `json:"environment"`
	RouteKind   string    `json:"route_kind"`
	CIDR        string    `json:"cidr"`
	Hostname    string    `json:"hostname,omitempty"`
	Purpose     string    `json:"purpose"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkloadNetworkDestination struct {
	Protocol  string    `json:"protocol"`
	Hostname  string    `json:"hostname,omitempty"`
	CIDRs     []string  `json:"cidrs,omitempty"`
	PortStart int       `json:"port_start"`
	PortEnd   int       `json:"port_end"`
	Purpose   string    `json:"purpose"`
	GrantID   string    `json:"grant_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type WorkloadNetworkPolicy struct {
	Version            string                       `json:"version"`
	Mode               string                       `json:"mode"`
	AllowStackShiftDNS bool                         `json:"allow_stackshift_dns"`
	AllowPublicWeb     bool                         `json:"allow_public_web"`
	Destinations       []WorkloadNetworkDestination `json:"destinations"`
	BlockPrivateRanges bool                         `json:"block_private_ranges"`
	BlockMetadata      bool                         `json:"block_metadata"`
	BlockAlternateDNS  bool                         `json:"block_alternate_dns"`
	BlockUnexpectedUDP bool                         `json:"block_unexpected_udp"`
}

type WorkloadPolicyInspection struct {
	Effective           bool                         `json:"effective"`
	Policy              *WorkloadNetworkPolicy       `json:"policy,omitempty"`
	PendingDependencies []WorkloadExternalDependency `json:"pending_dependencies,omitempty"`
}

type WorkloadSecurityIncident struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	DeploymentID      string    `json:"deployment_id,omitempty"`
	Environment       string    `json:"environment"`
	Category          string    `json:"category"`
	Confidence        string    `json:"confidence"`
	Status            string    `json:"status"`
	SanitizedSummary  string    `json:"sanitized_summary"`
	Evidence          any       `json:"sanitized_evidence"`
	QuarantinedAt     time.Time `json:"quarantined_at,omitempty"`
	EnforcementStatus string    `json:"enforcement_status"`
	ReleasedAt        time.Time `json:"released_at,omitempty"`
	ResolutionNote    string    `json:"resolution_note,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type DeclareWorkloadDependencyInput struct {
	Environment string `json:"environment"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Purpose     string `json:"purpose"`
	TLSRequired bool   `json:"tls_required"`
}

type RequestWorkloadGrantInput struct {
	DependencyID      string    `json:"dependency_id,omitempty"`
	Environment       string    `json:"environment"`
	Protocol          string    `json:"protocol"`
	Hostname          string    `json:"hostname,omitempty"`
	CIDR              string    `json:"cidr,omitempty"`
	PortStart         int       `json:"port_start"`
	PortEnd           int       `json:"port_end"`
	Purpose           string    `json:"purpose"`
	BroadPublicEgress bool      `json:"broad_public_egress"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type RequestWorkloadPrivateRouteInput struct {
	NodeID      string    `json:"node_id"`
	Environment string    `json:"environment"`
	RouteKind   string    `json:"route_kind"`
	CIDR        string    `json:"cidr"`
	Hostname    string    `json:"hostname,omitempty"`
	Purpose     string    `json:"purpose"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (w *WorkloadSecurityClient) Admissions(ctx context.Context, projectID string, limit int) ([]WorkloadAdmission, error) {
	var response struct {
		Admissions []WorkloadAdmission `json:"admissions"`
	}
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/admissions", "", "", limit), nil, &response)
	return response.Admissions, err
}

func (w *WorkloadSecurityClient) Policy(ctx context.Context, projectID, environment string) (*WorkloadPolicyInspection, error) {
	var response WorkloadPolicyInspection
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/policy", environment, "", 0), nil, &response)
	return &response, err
}

func (w *WorkloadSecurityClient) Dependencies(ctx context.Context, projectID, environment string) ([]WorkloadExternalDependency, error) {
	var response struct {
		Dependencies []WorkloadExternalDependency `json:"dependencies"`
	}
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/dependencies", environment, "", 0), nil, &response)
	return response.Dependencies, err
}

func (w *WorkloadSecurityClient) DeclareDependency(ctx context.Context, projectID string, input DeclareWorkloadDependencyInput) ([]WorkloadExternalDependency, error) {
	var response struct {
		Dependencies []WorkloadExternalDependency `json:"dependencies"`
	}
	err := w.client.request(ctx, http.MethodPost, securityPath(projectID, "/dependencies", "", "", 0), input, &response)
	return response.Dependencies, err
}

func (w *WorkloadSecurityClient) RefreshDependencies(ctx context.Context, projectID, environment string) ([]WorkloadExternalDependency, error) {
	var response struct {
		Dependencies []WorkloadExternalDependency `json:"dependencies"`
	}
	err := w.client.request(ctx, http.MethodPost, securityPath(projectID, "/dependencies/refresh", "", "", 0), map[string]string{"environment": environment}, &response)
	return response.Dependencies, err
}

func (w *WorkloadSecurityClient) Grants(ctx context.Context, projectID, environment string) ([]WorkloadEgressGrant, error) {
	var response struct {
		Grants []WorkloadEgressGrant `json:"grants"`
	}
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/grants", environment, "", 0), nil, &response)
	return response.Grants, err
}

func (w *WorkloadSecurityClient) RequestGrant(ctx context.Context, projectID string, input RequestWorkloadGrantInput) (*WorkloadEgressGrant, error) {
	var response WorkloadEgressGrant
	err := w.client.request(ctx, http.MethodPost, securityPath(projectID, "/grants", "", "", 0), input, &response)
	return &response, err
}

func (w *WorkloadSecurityClient) PrivateRoutes(ctx context.Context, projectID, environment string) ([]WorkloadPrivateRoute, error) {
	var response struct {
		PrivateRoutes []WorkloadPrivateRoute `json:"private_routes"`
	}
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/private-routes", environment, "", 0), nil, &response)
	return response.PrivateRoutes, err
}

func (w *WorkloadSecurityClient) RequestPrivateRoute(ctx context.Context, projectID string, input RequestWorkloadPrivateRouteInput) (*WorkloadPrivateRoute, error) {
	var response WorkloadPrivateRoute
	err := w.client.request(ctx, http.MethodPost, securityPath(projectID, "/private-routes", "", "", 0), input, &response)
	return &response, err
}

func (w *WorkloadSecurityClient) Incidents(ctx context.Context, projectID, status string, limit int) ([]WorkloadSecurityIncident, error) {
	var response struct {
		Incidents []WorkloadSecurityIncident `json:"incidents"`
	}
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/incidents", "", status, limit), nil, &response)
	return response.Incidents, err
}

func (w *WorkloadSecurityClient) Incident(ctx context.Context, projectID, incidentID string) (*WorkloadSecurityIncident, error) {
	var response struct {
		Incident WorkloadSecurityIncident `json:"incident"`
	}
	err := w.client.request(ctx, http.MethodGet, securityPath(projectID, "/incidents/"+url.PathEscape(incidentID), "", "", 0), nil, &response)
	return &response.Incident, err
}

func securityPath(projectID, suffix, environment, status string, limit int) string {
	path := "/projects/" + url.PathEscape(projectID) + "/security" + suffix
	query := url.Values{}
	if environment != "" {
		query.Set("environment", environment)
	}
	if status != "" {
		query.Set("status", status)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return path
}
