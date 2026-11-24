package stackshift

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type computerSSEMessage struct {
	Event string
	ID    string
	Data  string
}

func readComputerSSE(body io.Reader, handle func(computerSSEMessage) bool) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 8<<20)
	var message computerSSEMessage
	var data []string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			message.Data = strings.Join(data, "\n")
			ready := message.Data != "" || message.Event == "end"
			current := message
			message, data = computerSSEMessage{}, nil
			if ready && !handle(current) {
				return nil
			}
		case strings.HasPrefix(line, "event:"):
			message.Event = strings.TrimPrefix(strings.TrimPrefix(line, "event:"), " ")
		case strings.HasPrefix(line, "id:"):
			message.ID = strings.TrimPrefix(strings.TrimPrefix(line, "id:"), " ")
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	return scanner.Err()
}

func (c *ComputersClient) stream(ctx context.Context, method, path string, body any, headers map[string]string, handle func(computerSSEMessage) bool) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Accept"] = "text/event-stream"
	response, err := c.client.computerSend(ctx, method, path, reader, headers, body != nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		raw, _ := io.ReadAll(response.Body)
		return computerAPIError(response.StatusCode, raw)
	}
	err = readComputerSSE(response.Body, handle)
	if err != nil && ctx.Err() != nil {
		return nil
	}
	return err
}

func (c *ComputersClient) ExecStream(ctx context.Context, id string, input ComputerExecInput, handle func(ComputerExecEvent) bool) error {
	var failure error
	err := c.stream(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/exec", input, nil, func(message computerSSEMessage) bool {
		event := ComputerExecEvent{Event: message.Event}
		switch message.Event {
		case "start":
			var data struct {
				ID string `json:"id"`
			}
			failure = json.Unmarshal([]byte(message.Data), &data)
			event.ProcessID = data.ID
		case "stdout", "stderr":
			var data struct {
				Data string `json:"data"`
			}
			failure = json.Unmarshal([]byte(message.Data), &data)
			event.Data = data.Data
		case "exit":
			var exit ComputerExecExit
			failure = json.Unmarshal([]byte(message.Data), &exit)
			event.Exit = &exit
		case "error":
			var streamErr ComputerStreamError
			failure = json.Unmarshal([]byte(message.Data), &streamErr)
			event.Error = &streamErr
		}
		return failure == nil && handle(event)
	})
	if failure != nil {
		return failure
	}
	return err
}

func (c *ComputersClient) SendStdin(ctx context.Context, id, commandID string, data []byte, closeInput bool) error {
	query := map[string]string{}
	if closeInput {
		query["close"] = "true"
	}
	_, err := c.client.computerBytes(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/commands/"+computerSeg(commandID)+"/stdin"+computerQuery(query), bytes.NewReader(data), map[string]string{"Content-Type": "application/octet-stream"})
	return err
}

func (c *ComputersClient) History(ctx context.Context, id, cursor string, limit int) (*ComputerCommandPage, error) {
	query := map[string]string{"cursor": cursor}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	var out ComputerCommandPage
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/history"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Command(ctx context.Context, id, commandID string) (*ComputerCommandRecord, error) {
	if commandID == "" {
		commandID = "last"
	}
	var out ComputerCommandRecord
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/history/"+computerSeg(commandID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Jobs(ctx context.Context, id string) (*ComputerCommandList, error) {
	var out ComputerCommandList
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/jobs", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) JobLogs(ctx context.Context, id, jobID string, offset int64) (*ComputerJobLogs, error) {
	query := map[string]string{}
	if offset > 0 {
		query["offset"] = strconv.FormatInt(offset, 10)
	}
	var out ComputerJobLogs
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/jobs/"+computerSeg(jobID)+"/logs"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) FollowJobLogs(ctx context.Context, id, jobID string, offset int64, handle func(ComputerJobLogEvent) bool) error {
	query := map[string]string{"follow": "1"}
	headers := map[string]string{}
	if offset > 0 {
		query["offset"] = strconv.FormatInt(offset, 10)
		headers["Last-Event-ID"] = strconv.FormatInt(offset, 10)
	}
	var failure error
	err := c.stream(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/jobs/"+computerSeg(jobID)+"/logs"+computerQuery(query), nil, headers, func(message computerSSEMessage) bool {
		event := ComputerJobLogEvent{Event: message.Event}
		event.Offset, _ = strconv.ParseInt(message.ID, 10, 64)
		if message.Event == "exit" {
			var record ComputerCommandRecord
			failure = json.Unmarshal([]byte(message.Data), &record)
			event.Exit = &record
		} else {
			var data struct {
				Data string `json:"data"`
			}
			failure = json.Unmarshal([]byte(message.Data), &data)
			event.Data = data.Data
		}
		return failure == nil && handle(event)
	})
	if failure != nil {
		return failure
	}
	return err
}

func (c *ComputersClient) WaitJob(ctx context.Context, id, jobID string, timeoutSeconds int) (*ComputerCommandRecord, error) {
	query := map[string]string{}
	if timeoutSeconds > 0 {
		query["timeout_seconds"] = strconv.Itoa(timeoutSeconds)
	}
	var out ComputerCommandRecord
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/jobs/"+computerSeg(jobID)+"/wait"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) KillJob(ctx context.Context, id, jobID, signal string) (*ComputerCommandRecord, error) {
	body := map[string]string{}
	if signal != "" {
		body["signal"] = signal
	}
	var out ComputerCommandRecord
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/jobs/"+computerSeg(jobID)+"/kill", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ActivityLeases(ctx context.Context, id string) (*ComputerActivityLeaseList, error) {
	var out ComputerActivityLeaseList
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/activity-leases", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) PutActivityLease(ctx context.Context, id, task string, ttlSeconds int) (*ComputerActivityLeaseResult, error) {
	var out ComputerActivityLeaseResult
	if err := c.client.computerRequest(ctx, http.MethodPut, "/computers/"+computerSeg(id)+"/activity-leases/"+computerSeg(task), map[string]int{"ttl_seconds": ttlSeconds}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ReleaseActivityLease(ctx context.Context, id, task string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/"+computerSeg(id)+"/activity-leases/"+computerSeg(task), nil, nil, nil)
}

func (c *ComputersClient) Limits(ctx context.Context, id string) (*ComputerLifecycleLimits, error) {
	var out ComputerLifecycleLimits
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/limits", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Metrics(ctx context.Context, id, since string) (*ComputerMetrics, error) {
	var out ComputerMetrics
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/metrics"+computerQuery(map[string]string{"since": since}), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) CreateSSHToken(ctx context.Context, id string, ttlSeconds int) (*ComputerSSHToken, error) {
	body := map[string]int{}
	if ttlSeconds > 0 {
		body["ttl_seconds"] = ttlSeconds
	}
	var out ComputerSSHToken
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/ssh-tokens", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) RevokeSSHTokens(ctx context.Context, id string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/"+computerSeg(id)+"/ssh-tokens", nil, nil, nil)
}

func (c *ComputersClient) SSHCertificate(ctx context.Context, computerID, publicKey string) (*ComputerSSHCertificate, error) {
	var out ComputerSSHCertificate
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/ssh/certificates", map[string]string{"computer_id": computerID, "public_key": publicKey}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) PromptAgent(ctx context.Context, id string, input ComputerAgentPrompt) (*ComputerAgentConversation, error) {
	var out ComputerAgentConversation
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/agent/prompt", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) PromptAgentFanout(ctx context.Context, id string, input ComputerAgentFanoutPrompt) (*ComputerAgentFanout, error) {
	var out ComputerAgentFanout
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/agent/prompt", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) AttachAgentFile(ctx context.Context, id, localPath string) (string, error) {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return "", err
	}
	remote := "/workspace/.stackshift/attachments/" + strings.ReplaceAll(ComputerIdempotencyKey(), "-", "") + "/" + filepath.Base(localPath)
	if err := c.WriteFile(ctx, id, remote, content, 0); err != nil {
		return "", err
	}
	return remote, nil
}

func (c *ComputersClient) AgentProviders(ctx context.Context, id string, refresh bool) ([]ComputerAgentProviderInfo, error) {
	var out struct {
		Providers []ComputerAgentProviderInfo `json:"providers"`
	}
	query := map[string]string{}
	if refresh {
		query["refresh"] = "1"
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/providers"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Providers, nil
}

func (c *ComputersClient) AgentModels(ctx context.Context, id, provider, secretBindingID string) (*ComputerAgentModelList, error) {
	var out ComputerAgentModelList
	query := map[string]string{}
	if provider != "" {
		query["provider"] = provider
	}
	if secretBindingID != "" {
		query["secret_binding_id"] = secretBindingID
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/models"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) AgentResult(ctx context.Context, id, conversationID string, wait bool) (*ComputerAgentResult, error) {
	var out ComputerAgentResult
	query := map[string]string{}
	if wait {
		query["wait"] = "1"
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/conversations/"+computerSeg(conversationID)+"/result"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) AgentTurns(ctx context.Context, id string, limit int) ([]ComputerAgentTurn, error) {
	var out struct {
		Turns []ComputerAgentTurn `json:"turns"`
	}
	query := map[string]string{}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/turns"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Turns, nil
}

func (c *ComputersClient) AgentFanouts(ctx context.Context, id string) ([]ComputerAgentFanout, error) {
	var out struct {
		Fanouts []ComputerAgentFanout `json:"fanouts"`
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/fanouts", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Fanouts, nil
}

func (c *ComputersClient) AgentFanout(ctx context.Context, id, fanoutID string) (*ComputerAgentFanout, error) {
	var out ComputerAgentFanout
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/fanouts/"+computerSeg(fanoutID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) KeepAgentFork(ctx context.Context, id, fanoutID, forkComputerID string) (*ComputerAgentFanout, error) {
	var out ComputerAgentFanout
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/agent/fanouts/"+computerSeg(fanoutID)+"/keep", map[string]string{"computer_id": forkComputerID}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) AgentConversations(ctx context.Context, id string) ([]ComputerAgentConversation, error) {
	var out struct {
		Conversations []ComputerAgentConversation `json:"conversations"`
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/agent/conversations", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Conversations, nil
}

func agentEventsQuery(afterSeq int64, limit int, follow bool) map[string]string {
	query := map[string]string{}
	if afterSeq > 0 {
		query["after_seq"] = strconv.FormatInt(afterSeq, 10)
	}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	if follow {
		query["follow"] = "1"
	}
	return query
}

func (c *ComputersClient) AgentEvents(ctx context.Context, id, conversationID string, afterSeq int64, limit int) (*ComputerAgentEventPage, error) {
	var out ComputerAgentEventPage
	path := "/computers/" + computerSeg(id) + "/agent/conversations/" + computerSeg(conversationID) + "/events" + computerQuery(agentEventsQuery(afterSeq, limit, false))
	if err := c.client.computerRequest(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) FollowAgentEvents(ctx context.Context, id, conversationID string, afterSeq int64, limit int, handle func(ComputerAgentEvent) bool) error {
	headers := map[string]string{}
	if afterSeq > 0 {
		headers["Last-Event-ID"] = strconv.FormatInt(afterSeq, 10)
	}
	path := "/computers/" + computerSeg(id) + "/agent/conversations/" + computerSeg(conversationID) + "/events" + computerQuery(agentEventsQuery(afterSeq, limit, true))
	var failure error
	err := c.stream(ctx, http.MethodGet, path, nil, headers, func(message computerSSEMessage) bool {
		if message.Event == "end" {
			return false
		}
		var event ComputerAgentEvent
		if failure = json.Unmarshal([]byte(message.Data), &event); failure != nil {
			return false
		}
		return handle(event)
	})
	if failure != nil {
		return failure
	}
	return err
}

func (c *ComputersClient) SteerAgent(ctx context.Context, id, conversationID, text string) (*ComputerAgentConversation, error) {
	var out ComputerAgentConversation
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/agent/conversations/"+computerSeg(conversationID)+"/steer", map[string]string{"text": text}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) InterruptAgent(ctx context.Context, id, conversationID string) (*ComputerAgentConversation, error) {
	var out ComputerAgentConversation
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/agent/conversations/"+computerSeg(conversationID)+"/interrupt", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Remove(ctx context.Context, id, path string, recursive bool) error {
	query := map[string]string{"path": path}
	if recursive {
		query["recursive"] = "true"
	}
	_, err := c.client.computerBytes(ctx, http.MethodDelete, "/computers/"+computerSeg(id)+"/files"+computerQuery(query), nil, nil)
	return err
}

func (c *ComputersClient) Mkdir(ctx context.Context, id, path string, parents bool) error {
	body := map[string]any{"path": path}
	if parents {
		body["parents"] = true
	}
	return c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/files/mkdir", body, nil, nil)
}

func (c *ComputersClient) StartUpload(ctx context.Context, id string, input ComputerUploadInput, idempotencyKey string) (*ComputerUploadPlan, error) {
	var out ComputerUploadPlan
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/files/uploads", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) CompleteUpload(ctx context.Context, id, transferID string, input ComputerCompleteUploadInput) (*ComputerTransfer, error) {
	var out ComputerTransfer
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/files/uploads/"+computerSeg(transferID)+"/complete", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ResumeUpload(ctx context.Context, id, transferID string, input ComputerResumeUploadInput) (*ComputerUploadPlan, error) {
	var out ComputerUploadPlan
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/files/uploads/"+computerSeg(transferID)+"/resume", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) StartDownload(ctx context.Context, id, path, idempotencyKey string) (*ComputerTransfer, error) {
	var out ComputerTransfer
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/files/downloads", map[string]string{"path": path}, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Transfer(ctx context.Context, id, transferID string) (*ComputerTransfer, error) {
	var out ComputerTransfer
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/files/transfers/"+computerSeg(transferID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) AgentClaim(ctx context.Context, code string) (*ComputerAgentClaim, error) {
	var out ComputerAgentClaim
	if err := c.client.computerRequest(ctx, http.MethodGet, "/auth/agent/claims/"+computerSeg(code), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) DecideAgentClaim(ctx context.Context, code string, decision ComputerAgentClaimDecision) (*ComputerAgentClaim, error) {
	var out ComputerAgentClaim
	if err := c.client.computerRequest(ctx, http.MethodPost, "/auth/agent/claims/"+computerSeg(code)+"/decision", decision, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

var computerAgentWait = func(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func RegisterComputerAgent(ctx context.Context, baseURL string, httpClient HTTPDoer, input ComputerAgentRegistrationInput) (*ComputerAgentRegistration, error) {
	var out ComputerAgentRegistration
	if err := computerPublicCall(ctx, baseURL, httpClient, http.MethodPost, "/auth/agent", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func ExchangeComputerAgentClaim(ctx context.Context, baseURL string, httpClient HTTPDoer, claimToken string) (*ComputerAgentIssuedKey, error) {
	var out ComputerAgentIssuedKey
	if err := computerPublicCall(ctx, baseURL, httpClient, http.MethodPost, "/auth/agent/token", map[string]string{"claim_token": claimToken}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func WaitForComputerAgentKey(ctx context.Context, baseURL string, httpClient HTTPDoer, registration *ComputerAgentRegistration) (*ComputerAgentIssuedKey, error) {
	interval := time.Duration(max(registration.Interval, 1)) * time.Second
	deadline := time.Now().Add(time.Duration(registration.ExpiresIn) * time.Second)
	for {
		key, err := ExchangeComputerAgentClaim(ctx, baseURL, httpClient, registration.ClaimToken)
		if err == nil {
			return key, nil
		}
		var apiErr *APIError
		switch {
		case errors.As(err, &apiErr) && apiErr.Code == "slow_down":
			interval += 5 * time.Second
		case errors.As(err, &apiErr) && apiErr.Code == "authorization_pending":
		default:
			return nil, err
		}
		if time.Now().Add(interval).After(deadline) {
			return nil, &ComputerWaitError{Message: "the owner did not decide before the claim expired"}
		}
		if err := computerAgentWait(ctx, interval); err != nil {
			return nil, err
		}
	}
}
