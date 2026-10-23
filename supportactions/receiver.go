package supportactions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Receive commits a durable execution claim before calling either handler.
// A handler error is uncertain: it is stored as unknown, never re-dispatched.
func Receive(ctx context.Context, raw []byte, headers map[string]string, options Verification, store DurableStore, execute, reconcile Handler) (SignedResponse, error) {
	options = defaults(options)
	request, err := Verify(raw, headers, options)
	if err != nil {
		return SignedResponse{}, err
	}
	h, _ := normalize(headers)
	delivery := h["x-stackshift-delivery-id"]
	digestBytes := sha256.Sum256(raw)
	digest := hex.EncodeToString(digestBytes[:])
	deliveryKey := request.InstallationID + ":" + delivery
	deliveryClaim, err := store.Claim(ctx, "delivery", deliveryKey, digest)
	if err != nil {
		return SignedResponse{}, err
	}
	if deliveryClaim.State != "acquired" || deliveryClaim.Fence == "" {
		return SignedResponse{}, ProtocolError("delivery_replay")
	}
	key := strings.Join([]string{request.InstallationID, request.Action.Name, request.Action.Version, request.RequestID, request.Mode}, ":")
	claim, err := store.Claim(ctx, "execution", key, digest)
	if err != nil {
		return SignedResponse{}, err
	}
	var response []byte
	switch claim.State {
	case "conflict":
		return SignedResponse{}, ProtocolError("idempotency_conflict")
	case "completed", "unknown":
		if len(claim.Response) > 0 {
			response = claim.Response
			break
		}
		if claim.State == "completed" {
			return SignedResponse{}, ProtocolError("invalid_store_claim")
		}
		response, err = EncodeReceipt(Receipt{Schema: ReceiptContext, RequestID: request.RequestID, State: "unknown", Result: map[string]any{}, ErrorCode: "execution_unknown"})
	case "running":
		response, err = EncodeReceipt(Receipt{Schema: ReceiptContext, RequestID: request.RequestID, State: "accepted", Result: map[string]any{}, ErrorCode: "execution_in_progress"})
	case "acquired":
		if claim.Fence == "" {
			return SignedResponse{}, ProtocolError("invalid_store_claim")
		}
		handler := execute
		if request.Mode == "reconcile" {
			handler = reconcile
		}
		receipt, handlerErr := handler(ctx, request)
		receipt.Schema = ReceiptContext
		receipt.RequestID = request.RequestID
		action, _ := findAction(options.Actions, request.Action)
		if handlerErr == nil && receipt.State == "succeeded" {
			handlerErr = ValidateFields(receipt.Result, action.Output)
		}
		if handlerErr == nil {
			response, handlerErr = EncodeReceipt(receipt)
		}
		if handlerErr != nil {
			receipt.State = "unknown"
			response, err = EncodeReceipt(Receipt{Schema: ReceiptContext, RequestID: request.RequestID, State: "unknown", Result: map[string]any{}, ErrorCode: "handler_outcome_unknown"})
		}
		if err == nil {
			err = store.Complete(ctx, "execution", key, claim.Fence, response, receipt.State == "unknown")
		}
	default:
		return SignedResponse{}, ProtocolError("invalid_store_claim")
	}
	if err != nil {
		return SignedResponse{}, err
	}
	if err = store.Complete(ctx, "delivery", deliveryKey, deliveryClaim.Fence, response, false); err != nil {
		return SignedResponse{}, err
	}
	signed, err := Sign(response, ReceiptContext, options.KeyID, options.Secret, options.Now.Unix(), delivery)
	return SignedResponse{Body: response, Headers: signed}, err
}
