package supportactions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
var stampPattern = regexp.MustCompile(`^[1-9][0-9]{0,15}$`)
var signaturePattern = regexp.MustCompile(`^v1=[0-9a-f]{64}$`)

const maxSafeInteger int64 = 9007199254740991

func validUUID(value string) bool {
	return uuidPattern.MatchString(value) && value != "00000000-0000-0000-0000-000000000000"
}
func positive(value int64) bool { return value > 0 && value <= maxSafeInteger }
func signature(raw []byte, context, keyID, secret, stamp, delivery string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.Join([]string{context, keyID, stamp, delivery, ""}, "\n")))
	_, _ = mac.Write(raw)
	return mac.Sum(nil)
}

// Sign signs exact raw bytes. Receipt callers must echo the request delivery ID.
func Sign(raw []byte, context, keyID, secret string, timestamp int64, delivery string) (map[string]string, error) {
	if (context != RequestContext && context != ReceiptContext) || !keyPattern.MatchString(keyID) || len(secret) < 32 || !positive(timestamp) || !validUUID(delivery) {
		return nil, ProtocolError("invalid_request")
	}
	stamp := strconv.FormatInt(timestamp, 10)
	return map[string]string{"x-stackshift-action-key-id": keyID, "x-stackshift-timestamp": stamp, "x-stackshift-delivery-id": delivery, "x-stackshift-action-signature": "v1=" + hex.EncodeToString(signature(raw, context, keyID, secret, stamp, delivery))}, nil
}
func normalize(headers map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(headers))
	for name, value := range headers {
		name = strings.ToLower(name)
		if _, exists := result[name]; exists {
			return nil, ProtocolError("invalid_headers")
		}
		result[name] = value
	}
	return result, nil
}
func verifySignature(raw []byte, headers map[string]string, context string, options Verification, limit int) (string, error) {
	if len(raw) == 0 || len(raw) > limit {
		return "", ProtocolError("body_limit")
	}
	h, err := normalize(headers)
	if err != nil {
		return "", err
	}
	now := options.Now.Unix()
	tolerance := int64(options.Tolerance.Seconds())
	if !positive(now) || tolerance <= 0 || tolerance > 300 {
		return "", ProtocolError("invalid_request")
	}
	stamp := h["x-stackshift-timestamp"]
	seconds, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil || !stampPattern.MatchString(stamp) || !positive(seconds) {
		return "", ProtocolError("invalid_headers")
	}
	delivery := h["x-stackshift-delivery-id"]
	if h["x-stackshift-action-key-id"] != options.KeyID || !validUUID(delivery) || !keyPattern.MatchString(options.KeyID) || len(options.Secret) < 32 {
		return "", ProtocolError("binding_mismatch")
	}
	if seconds < now-tolerance || seconds > now+tolerance {
		return "", ProtocolError("stale_signature")
	}
	signed := h["x-stackshift-action-signature"]
	if !signaturePattern.MatchString(signed) {
		return "", ProtocolError("invalid_signature")
	}
	actual, _ := hex.DecodeString(signed[3:])
	if !hmac.Equal(actual, signature(raw, context, options.KeyID, options.Secret, stamp, delivery)) {
		return "", ProtocolError("invalid_signature")
	}
	return delivery, nil
}
