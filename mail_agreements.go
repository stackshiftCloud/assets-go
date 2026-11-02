package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type MailAgreementsClient struct {
	client *Client
}

func (a *MailAgreementsClient) List(ctx context.Context, params ListAgreementsParams) (*AgreementListResult, error) {
	query := url.Values{}
	appendQuery(query, "inboxId", params.InboxID)
	appendQuery(query, "state", params.State)
	appendQuery(query, "counterparty", params.Counterparty)
	appendQuery(query, "cursor", params.Cursor)
	appendQuery(query, "limit", params.Limit)
	path := "/mail/agreements"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result AgreementListResult
	err := a.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (a *MailAgreementsClient) Get(ctx context.Context, threadID string) (*AgreementInspect, error) {
	var result AgreementInspect
	err := a.client.mailRequest(ctx, http.MethodGet, agreementPath(threadID, ""), nil, &result)
	return &result, err
}

func (a *MailAgreementsClient) Record(ctx context.Context, threadID string) (json.RawMessage, error) {
	var record json.RawMessage
	err := a.client.mailRequest(ctx, http.MethodGet, agreementPath(threadID, "/record"), nil, &record)
	return record, err
}

func (a *MailAgreementsClient) Wait(ctx context.Context, threadID, after string, timeoutSeconds int) (*AgreementWaitResult, error) {
	query := url.Values{}
	appendQuery(query, "after", after)
	appendQuery(query, "timeout", timeoutSeconds)
	path := agreementPath(threadID, "/wait")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result AgreementWaitResult
	err := a.client.mailRequest(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (a *MailAgreementsClient) WaitForChange(ctx context.Context, threadID, after string) (*AgreementInspect, error) {
	cursor := after
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := a.Wait(ctx, threadID, cursor, 0)
		if err != nil {
			return nil, err
		}
		if !result.TimedOut {
			return &result.AgreementInspect, nil
		}
		cursor = defaultString(result.ChangeCursor, cursor)
	}
}

func (a *MailAgreementsClient) Propose(ctx context.Context, inboxID string, input ProposeAgreementInput, options AgreementSendOptions) (*AgreementSendResult, error) {
	return a.send(ctx, "/mail/agent-inboxes/"+url.PathEscape(inboxID)+"/agreements", input, options)
}

func (a *MailAgreementsClient) Respond(ctx context.Context, threadID string, input RespondAgreementInput, options AgreementSendOptions) (*AgreementSendResult, error) {
	return a.send(ctx, agreementPath(threadID, "/respond"), input, options)
}

func (a *MailAgreementsClient) Accept(ctx context.Context, threadID, offerID string, slot *int, options AgreementSendOptions) (*AgreementSendResult, error) {
	return a.Respond(ctx, threadID, RespondAgreementInput{OfferID: offerID, Decision: "accept", Slot: slot}, options)
}

func (a *MailAgreementsClient) Decline(ctx context.Context, threadID, offerID, reason, note string, options AgreementSendOptions) (*AgreementSendResult, error) {
	return a.Respond(ctx, threadID, RespondAgreementInput{OfferID: offerID, Decision: "decline", Reason: reason, Note: note}, options)
}

func (a *MailAgreementsClient) Cancel(ctx context.Context, threadID string, input CancelAgreementInput, options AgreementSendOptions) (*AgreementSendResult, error) {
	return a.send(ctx, agreementPath(threadID, "/cancel"), input, options)
}

func (a *MailAgreementsClient) CountReply(ctx context.Context, threadID, offerID, replyMessageID string, options AgreementSendOptions) (*AgreementSendResult, error) {
	body := map[string]string{"offerId": offerID, "replyMessageId": replyMessageID}
	return a.send(ctx, agreementPath(threadID, "/count-reply"), body, options)
}

func (a *MailAgreementsClient) GetLimits(ctx context.Context, inboxID string) (*AgreementLimits, error) {
	var result AgreementLimits
	err := a.client.mailRequest(ctx, http.MethodGet, "/mail/agent-inboxes/"+url.PathEscape(inboxID)+"/agreement-limits", nil, &result)
	return &result, err
}

func (a *MailAgreementsClient) Check(ctx context.Context, record json.RawMessage) (*AgreementCheckResult, error) {
	return a.check(ctx, record)
}

func (a *MailAgreementsClient) CheckJWS(ctx context.Context, jws string) (*AgreementCheckResult, error) {
	return a.check(ctx, map[string]string{"jws": jws})
}

func (a *MailAgreementsClient) check(ctx context.Context, body any) (*AgreementCheckResult, error) {
	var result AgreementCheckResult
	err := a.client.mailRequest(ctx, http.MethodPost, "/public/agreements/check", body, &result)
	return &result, err
}

func (a *MailAgreementsClient) send(ctx context.Context, path string, body any, options AgreementSendOptions) (*AgreementSendResult, error) {
	if options.IdempotencyKey == "" {
		return nil, errors.New("stackshift: idempotency key is required")
	}
	var result AgreementSendResult
	headers := map[string]string{"Idempotency-Key": options.IdempotencyKey}
	err := a.client.mailRequestWithHeaders(ctx, http.MethodPost, path, body, headers, &result)
	return &result, err
}

func agreementPath(threadID, suffix string) string {
	return "/mail/agreements/" + url.PathEscape(threadID) + suffix
}
