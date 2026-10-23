package supportactions

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func defaults(options Verification) Verification {
	if options.Now.IsZero() {
		options.Now = time.Now()
	}
	if options.Tolerance == 0 {
		options.Tolerance = 300 * time.Second
	}
	return options
}
func uniqueValue(decoder *json.Decoder, depth int) error {
	if depth > 32 {
		return ProtocolError("invalid_request")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		seen := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return ProtocolError("invalid_request")
			}
			seen[name] = true
			if err := uniqueValue(decoder, depth+1); err != nil {
				return err
			}
		}
	} else if delimiter == '[' {
		for decoder.More() {
			if err := uniqueValue(decoder, depth+1); err != nil {
				return err
			}
		}
	} else {
		return ProtocolError("invalid_request")
	}
	_, err = decoder.Token()
	return err
}
func decode(raw []byte, value any) error {
	if !utf8.Valid(raw) {
		return ProtocolError("invalid_request")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := uniqueValue(decoder, 0); err != nil {
		return ProtocolError("invalid_request")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ProtocolError("invalid_request")
	}
	decoder = json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return ProtocolError("invalid_request")
	}
	return nil
}
func shape(value any, required, optional string) bool {
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return false
	}
	allowed := map[string]bool{}
	for _, key := range strings.Fields(required) {
		if _, ok := object[key]; !ok {
			return false
		}
		allowed[key] = true
	}
	for _, key := range strings.Fields(optional) {
		allowed[key] = true
	}
	for key := range object {
		if !allowed[key] {
			return false
		}
	}
	return true
}

// ValidateFields checks the registered flat primitive schema without coercion.
func ValidateFields(value map[string]any, fields map[string]Field) error {
	if value == nil {
		return ProtocolError("invalid_request")
	}
	for name := range value {
		if _, ok := fields[name]; !ok {
			return ProtocolError("invalid_request")
		}
	}
	for name, field := range fields {
		item, exists := value[name]
		if !exists {
			if field.Required {
				return ProtocolError("invalid_request")
			}
			continue
		}
		switch field.Type {
		case "string":
			text, ok := item.(string)
			if !ok {
				return ProtocolError("invalid_request")
			}
			limit := 16000
			if field.MaxLength != nil {
				limit = *field.MaxLength
			}
			if utf8.RuneCountInString(text) > limit {
				return ProtocolError("invalid_request")
			}
			if len(field.Enum) > 0 {
				found := false
				for _, entry := range field.Enum {
					found = found || entry == text
				}
				if !found {
					return ProtocolError("invalid_request")
				}
			}
		case "boolean":
			if _, ok := item.(bool); !ok {
				return ProtocolError("invalid_request")
			}
		case "number", "integer":
			var number float64
			switch n := item.(type) {
			case json.Number:
				var err error
				number, err = n.Float64()
				if err != nil {
					return ProtocolError("invalid_request")
				}
			case float64:
				number = n
			case int:
				number = float64(n)
			case int64:
				number = float64(n)
			default:
				return ProtocolError("invalid_request")
			}
			if math.IsNaN(number) || math.IsInf(number, 0) || (field.Type == "integer" && (math.Trunc(number) != number || math.Abs(number) > float64(maxSafeInteger))) || (field.Minimum != nil && number < *field.Minimum) || (field.Maximum != nil && number > *field.Maximum) {
				return ProtocolError("invalid_request")
			}
		default:
			return ProtocolError("invalid_request")
		}
	}
	return nil
}

// Verify authenticates exact bytes before decoding or selecting an action.
func Verify(raw []byte, headers map[string]string, options Verification) (Request, error) {
	options = defaults(options)
	var request Request
	if _, err := verifySignature(raw, headers, RequestContext, options, RequestLimit); err != nil {
		return request, err
	}
	var body map[string]any
	if err := decode(raw, &body); err != nil {
		return request, err
	}
	if !shape(body, "schema request_id installation_id installation_revision app_version_id action scope customer arguments authorization issued_at expires_at mode", "target_request_id") || !shape(body["action"], "name version", "") || !shape(body["scope"], "workspace_id environment_id inbox_id conversation_id contact_id", "") || !shape(body["customer"], "external_id", "") || !shape(body["authorization"], "revision mode", "grant_id") {
		return request, ProtocolError("invalid_request")
	}
	if err := decode(raw, &request); err != nil {
		return request, err
	}
	if request.Schema != RequestContext || request.InstallationID != options.InstallationID || options.ExpectedCustomerExternalID == "" || utf8.RuneCountInString(options.ExpectedCustomerExternalID) > 256 || request.Customer.ExternalID != options.ExpectedCustomerExternalID {
		return request, ProtocolError("binding_mismatch")
	}
	ids := []string{request.RequestID, request.InstallationID, request.AppVersionID, request.Scope.WorkspaceID, request.Scope.EnvironmentID, request.Scope.InboxID, request.Scope.ConversationID, request.Scope.ContactID}
	for _, id := range ids {
		if !validUUID(id) {
			return request, ProtocolError("invalid_request")
		}
	}
	if !positive(request.InstallationRevision) || !positive(request.Authorization.Revision) {
		return request, ProtocolError("invalid_request")
	}
	auth := request.Authorization
	if auth.Mode != "automatic" && auth.Mode != "customer" && auth.Mode != "staff" {
		return request, ProtocolError("invalid_request")
	}
	if _, present := body["authorization"].(map[string]any)["grant_id"]; present && !validUUID(auth.GrantID) {
		return request, ProtocolError("invalid_request")
	}
	now := options.Now.Unix()
	if request.IssuedAt < 0 || request.IssuedAt > maxSafeInteger || !positive(request.ExpiresAt) || request.IssuedAt > now+int64(options.Tolerance.Seconds()) || request.ExpiresAt <= now || request.ExpiresAt <= request.IssuedAt {
		return request, ProtocolError("request_expired")
	}
	action, ok := findAction(options.Actions, request.Action)
	if !ok {
		return request, ProtocolError("action_not_registered")
	}
	if request.Mode == "reconcile" {
		if !validUUID(request.TargetRequestID) || request.TargetRequestID == request.RequestID || request.Arguments == nil || len(request.Arguments) != 0 {
			return request, ProtocolError("invalid_request")
		}
	} else {
		_, targetPresent := body["target_request_id"]
		if request.Mode != "execute" || targetPresent {
			return request, ProtocolError("invalid_request")
		}
		if err := ValidateFields(request.Arguments, action.Input); err != nil {
			return request, err
		}
	}
	return request, nil
}
func findAction(actions []Action, ref ActionRef) (Action, bool) {
	for _, action := range actions {
		if action.Name == ref.Name && action.Version == ref.Version {
			return action, true
		}
	}
	return Action{}, false
}

var errorPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,79}$`)

func validateReceipt(receipt Receipt) error {
	if receipt.Schema != ReceiptContext || !validUUID(receipt.RequestID) || receipt.Result == nil || (receipt.ErrorCode != "" && !errorPattern.MatchString(receipt.ErrorCode)) {
		return ProtocolError("invalid_request")
	}
	switch receipt.State {
	case "succeeded":
		return nil
	case "accepted", "failed", "unknown":
		if len(receipt.Result) == 0 {
			return nil
		}
	}
	return ProtocolError("invalid_request")
}
func EncodeReceipt(receipt Receipt) ([]byte, error) {
	if err := validateReceipt(receipt); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(receipt)
	if err != nil {
		return nil, ProtocolError("invalid_request")
	}
	if len(raw) > ReceiptLimit {
		return nil, ProtocolError("body_limit")
	}
	var bounded map[string]any
	if err := decode(raw, &bounded); err != nil {
		return nil, err
	}
	return raw, nil
}
func VerifyReceipt(raw []byte, headers map[string]string, options Verification, requestID, deliveryID string) (Receipt, error) {
	options = defaults(options)
	var receipt Receipt
	delivery, err := verifySignature(raw, headers, ReceiptContext, options, ReceiptLimit)
	if err != nil {
		return receipt, err
	}
	var body map[string]any
	if err := decode(raw, &body); err != nil {
		return receipt, err
	}
	if !shape(body, "schema request_id state result", "error_code") {
		return receipt, ProtocolError("invalid_request")
	}
	if err := decode(raw, &receipt); err != nil {
		return receipt, err
	}
	if err := validateReceipt(receipt); err != nil {
		return receipt, err
	}
	if _, exists := body["error_code"]; exists && receipt.ErrorCode == "" {
		return receipt, ProtocolError("invalid_request")
	}
	if receipt.RequestID != requestID || delivery != deliveryID {
		return receipt, ProtocolError("binding_mismatch")
	}
	return receipt, nil
}
