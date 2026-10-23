package supportactions

import (
	"math"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var actionName = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,79}$`)
var fieldName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func endpoint(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && len(raw) <= 2048 && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && u.Opaque == "" && !strings.Contains(raw, "#")
}
func reserved(name string) bool {
	for _, prefix := range []string{"customer_", "tenant_", "auth_"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	switch name {
	case "customer", "customer_id", "external_customer_id", "external_id", "contact_id", "workspace_id", "environment_id", "inbox_id", "conversation_id", "installation_id", "tenant_id", "user_id", "organization_id", "authorization", "scope", "identity", "credentials", "secret", "token":
		return true
	}
	return false
}
func validFields(fields map[string]Field, input bool) bool {
	if fields == nil || len(fields) > 32 {
		return false
	}
	for name, f := range fields {
		if !fieldName.MatchString(name) || (input && reserved(name)) {
			return false
		}
		switch f.Type {
		case "string":
			if f.MaxLength == nil || *f.MaxLength < 1 || *f.MaxLength > 16000 || f.Minimum != nil || f.Maximum != nil || len(f.Enum) > 64 {
				return false
			}
			seen := map[string]bool{}
			for _, item := range f.Enum {
				if !utf8.ValidString(item) || utf8.RuneCountInString(item) > *f.MaxLength || seen[item] {
					return false
				}
				seen[item] = true
			}
		case "number", "integer":
			if (f.MaxLength != nil && *f.MaxLength != 0) || len(f.Enum) > 0 {
				return false
			}
			for _, bound := range []*float64{f.Minimum, f.Maximum} {
				if bound != nil && (math.IsNaN(*bound) || math.IsInf(*bound, 0) || (f.Type == "integer" && (math.Trunc(*bound) != *bound || math.Abs(*bound) > float64(maxSafeInteger)))) {
					return false
				}
			}
			if f.Minimum != nil && f.Maximum != nil && *f.Minimum > *f.Maximum {
				return false
			}
		case "boolean":
			if (f.MaxLength != nil && *f.MaxLength != 0) || f.Minimum != nil || f.Maximum != nil || len(f.Enum) > 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// ValidateCatalog checks local declarations. Registration, installation consent,
// and public destination policy are still enforced by the server.
func ValidateCatalog(version string, actions []Action) error {
	if len(version) > 32 || !versionPattern.MatchString(version) || len(actions) < 1 || len(actions) > 32 {
		return ProtocolError("invalid_catalog")
	}
	byName := map[string]Action{}
	for _, a := range actions {
		_, exists := byName[a.Name]
		if exists || !actionName.MatchString(a.Name) || a.Version != version || !a.IdentityRequired || a.TimeoutSeconds < 1 || a.TimeoutSeconds > 30 || !endpoint(a.Endpoint) || !validFields(a.Input, true) || !validFields(a.Output, false) {
			return ProtocolError("invalid_catalog")
		}
		switch a.Risk {
		case "read", "bounded_write", "customer_change", "financial", "security":
		default:
			return ProtocolError("invalid_catalog")
		}
		if (a.Risk != "read" && (!a.IdempotencySupported || !endpoint(a.ReconciliationEndpoint) || a.VerificationOperation == "")) || (a.SafeRetry && !a.IdempotencySupported) || (a.ReconciliationEndpoint != "" && !endpoint(a.ReconciliationEndpoint)) {
			return ProtocolError("invalid_catalog")
		}
		byName[a.Name] = a
	}
	for _, a := range actions {
		if a.VerificationOperation != "" {
			verify, ok := byName[a.VerificationOperation]
			if !ok || verify.Risk != "read" {
				return ProtocolError("invalid_catalog")
			}
		}
	}
	return nil
}

type InstallationInput struct {
	VersionID        string   `json:"version_id"`
	EnvironmentID    string   `json:"environment_id"`
	InboxID          string   `json:"inbox_id"`
	Permissions      []string `json:"permissions"`
	ActionOperations []string `json:"action_operations"`
	ConfirmConsent   bool     `json:"confirm_consent"`
	Revision         int64    `json:"revision"`
}
type InstallationCredentials struct {
	ActionKeyID      string   `json:"action_key_id"`
	ActionSecret     string   `json:"action_secret"`
	ActionOperations []string `json:"action_operations"`
}
