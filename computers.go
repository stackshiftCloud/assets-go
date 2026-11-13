package stackshift

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ComputersClient struct {
	client    *Client
	Keys      *ComputerKeysClient
	Snapshots *ComputerSnapshotsClient
	CI        *ComputerCIClient
}

type ComputerCIClient struct{ client *Client }
type ComputerSnapshotsClient struct{ client *Client }
type ComputerKeysClient struct{ client *Client }

type ComputerWaitError struct {
	Message  string
	Computer *Computer
}

func (e *ComputerWaitError) Error() string { return "stackshift: " + e.Message }

func ComputerIdempotencyKey() string {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	raw[6] = raw[6]&0x0f | 0x40
	raw[8] = raw[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
}

func computerSeg(value string) string { return url.PathEscape(value) }

func computerQuery(values map[string]string) string {
	query := url.Values{}
	for key, value := range values {
		if value != "" {
			query.Set(key, value)
		}
	}
	if len(query) == 0 {
		return ""
	}
	return "?" + query.Encode()
}

const computerConfirmDeleteHeader = "X-StackShift-Confirm-Delete"

func computerDeleteHeaders(idempotencyKey, confirm string) map[string]string {
	headers := computerIdempotency(idempotencyKey)
	headers[computerConfirmDeleteHeader] = confirm
	return headers
}

func computerIdempotency(key string) map[string]string {
	if key == "" {
		key = ComputerIdempotencyKey()
	}
	return map[string]string{"Idempotency-Key": key}
}

func (c *Client) computerRequest(ctx context.Context, method, path string, body any, headers map[string]string, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	response, err := c.computerSend(ctx, method, path, reader, headers, body != nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return computerAPIError(response.StatusCode, raw)
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func (c *Client) computerSend(ctx context.Context, method, path string, body io.Reader, headers map[string]string, hasJSON bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "StackShift-Go/0.1")
	if hasJSON {
		req.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	return c.httpClient.Do(req)
}

func (c *Client) computerBytes(ctx context.Context, method, path string, body io.Reader, headers map[string]string) ([]byte, error) {
	response, err := c.computerSend(ctx, method, path, body, headers, false)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, computerAPIError(response.StatusCode, raw)
	}
	return raw, nil
}

func computerAPIError(status int, raw []byte) error {
	var body struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &body) == nil && (body.Error.Code != "" || body.Error.Message != "") {
		return &APIError{StatusCode: status, Code: body.Error.Code, Message: firstNonEmpty(body.Error.Message, http.StatusText(status)), Details: body.Error.Details}
	}
	return &APIError{StatusCode: status, Message: firstNonEmpty(strings.TrimSpace(string(raw)), http.StatusText(status))}
}

type ComputerProject struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Created bool   `json:"created"`
}

func (c *ComputersClient) EnsureProject(ctx context.Context) (*ComputerProject, error) {
	var out ComputerProject
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/project", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Templates(ctx context.Context, projectID string) ([]ComputerTemplate, error) {
	var out []ComputerTemplate
	err := c.client.computerRequest(ctx, http.MethodGet, "/computers/templates"+computerQuery(map[string]string{"project_id": projectID}), nil, nil, &out)
	return out, err
}

type ComputerListOptions struct {
	State  string
	Cursor string
	Limit  int
}

func (c *ComputersClient) List(ctx context.Context, projectID string, options ComputerListOptions) (*ComputerPage, error) {
	var out ComputerPage
	query := map[string]string{"project_id": projectID, "state": options.State, "cursor": options.Cursor}
	if options.Limit > 0 {
		query["limit"] = strconv.Itoa(options.Limit)
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ListAll(ctx context.Context, projectID string, state string) ([]Computer, error) {
	var all []Computer
	cursor := ""
	for {
		page, err := c.List(ctx, projectID, ComputerListOptions{State: state, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if !page.HasMore || page.NextCursor == "" {
			return all, nil
		}
		cursor = page.NextCursor
	}
}

func (c *ComputersClient) Get(ctx context.Context, id string) (*Computer, error) {
	var out Computer
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Update(ctx context.Context, id string, input UpdateComputerInput) (*ComputerResult, error) {
	var out ComputerResult
	if err := c.client.computerRequest(ctx, http.MethodPatch, "/computers/"+computerSeg(id), input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Create(ctx context.Context, input CreateComputerInput, idempotencyKey string) (*ComputerResult, error) {
	var out ComputerResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) CreateAndWait(ctx context.Context, input CreateComputerInput, idempotencyKey string) (*Computer, error) {
	result, err := c.Create(ctx, input, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return c.Wait(ctx, result.Computer.ID, "running", time.Second)
}

func (c *ComputersClient) Delete(ctx context.Context, id, idempotencyKey string) (*ComputerResult, error) {
	return c.DeleteConfirmed(ctx, id, id, idempotencyKey)
}

func (c *ComputersClient) DeleteConfirmed(ctx context.Context, id, confirm, idempotencyKey string) (*ComputerResult, error) {
	var out ComputerResult
	if err := c.client.computerRequest(ctx, http.MethodDelete, "/computers/"+computerSeg(id), nil, computerDeleteHeaders(idempotencyKey, confirm), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Fork(ctx context.Context, id string, input ComputerForkInput, idempotencyKey string) (*ComputerForkResult, error) {
	var out ComputerForkResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/fork", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Pause(ctx context.Context, id, mode, idempotencyKey string) (*ComputerResult, error) {
	var body any
	if mode != "" {
		body = map[string]string{"mode": mode}
	}
	var out ComputerResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/pause", body, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Resume(ctx context.Context, id, idempotencyKey string) (*ComputerResult, error) {
	var out ComputerResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/resume", nil, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Restore(ctx context.Context, id, snapshotID, idempotencyKey string) (*ComputerResult, error) {
	var out ComputerResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/restore", map[string]string{"snapshot_id": snapshotID}, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Snapshot(ctx context.Context, id, name, mode, idempotencyKey string) (*ComputerSnapshot, error) {
	var out ComputerSnapshot
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/snapshots", map[string]string{"name": name, "mode": mode}, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Wait(ctx context.Context, id, state string, interval time.Duration) (*Computer, error) {
	if state == "" {
		state = "running"
	}
	if interval <= 0 {
		interval = time.Second
	}
	for {
		computer, err := c.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if computer.State == state {
			return computer, nil
		}
		if computer.State == "failed" || computer.State == "deleted" {
			return nil, &ComputerWaitError{Message: "computer is " + computer.State, Computer: computer}
		}
		select {
		case <-ctx.Done():
			return nil, &ComputerWaitError{Message: "computer is " + computer.State + ", not " + state, Computer: computer}
		case <-time.After(interval):
		}
	}
}

func (c *ComputersClient) Exec(ctx context.Context, id string, input ComputerExecInput) (*ComputerExecResult, error) {
	var out ComputerExecResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/exec", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ReadFile(ctx context.Context, id, path string) ([]byte, error) {
	return c.client.computerBytes(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/files"+computerQuery(map[string]string{"path": path}), nil, map[string]string{"Accept": "application/octet-stream"})
}

func (c *ComputersClient) WriteFile(ctx context.Context, id, path string, content []byte, mode int) error {
	query := map[string]string{"path": path}
	if mode > 0 {
		query["mode"] = strconv.Itoa(mode)
	}
	_, err := c.client.computerBytes(ctx, http.MethodPut, "/computers/"+computerSeg(id)+"/files"+computerQuery(query), bytes.NewReader(content), map[string]string{"Content-Type": "application/octet-stream"})
	return err
}

func (c *ComputersClient) ListFiles(ctx context.Context, id, path string) ([]ComputerFileEntry, error) {
	var out []ComputerFileEntry
	err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/files/list"+computerQuery(map[string]string{"path": path}), nil, nil, &out)
	return out, err
}

func (c *ComputersClient) PublishPort(ctx context.Context, id string, port int, visibility string) (*ComputerPort, error) {
	if visibility == "" {
		visibility = "private"
	}
	var out ComputerPort
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/ports", map[string]any{"port": port, "visibility": visibility}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) UnpublishPort(ctx context.Context, id string, port int) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/"+computerSeg(id)+"/ports/"+strconv.Itoa(port), nil, nil, nil)
}

func (c *ComputersClient) Act(ctx context.Context, id string, actions []ComputerAction, actor, stepID string) (*ComputerActionsResponse, error) {
	if actor == "" {
		actor = "agent"
	}
	if stepID == "" {
		stepID = ComputerIdempotencyKey()
	}
	var out ComputerActionsResponse
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/actions", ComputerActionsRequest{Actor: actor, StepID: stepID, Actions: actions}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ActWith(ctx context.Context, id string, actions []ComputerAction, options ComputerActOptions) (*ComputerActionsResponse, error) {
	if options.Actor == "" {
		options.Actor = "agent"
	}
	if options.StepID == "" {
		options.StepID = ComputerIdempotencyKey()
	}
	var out ComputerActionsResponse
	request := ComputerActionsRequest{Actor: options.Actor, StepID: options.StepID, Actions: actions, Task: options.Task, NextStep: options.NextStep}
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/actions", request, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ComputerScreenshotOptions struct {
	Format   string
	MaxWidth int
	WindowID string
}

func (c *ComputersClient) Screenshot(ctx context.Context, id string, options ComputerScreenshotOptions) (*ComputerScreenshot, error) {
	query := map[string]string{"format": options.Format, "window_id": options.WindowID}
	if options.MaxWidth > 0 {
		query["max_width"] = strconv.Itoa(options.MaxWidth)
	}
	var out ComputerScreenshot
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/screenshot"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) TakeOver(ctx context.Context, id string) (*ComputerInputLock, error) {
	var out ComputerInputLock
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/take-over", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) HandBack(ctx context.Context, id string) (*ComputerInputLock, error) {
	var out ComputerInputLock
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/hand-back", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Steps(ctx context.Context, id string, options ComputerStepOptions) (*ComputerStepPage, error) {
	query := map[string]string{"order": options.Order}
	if options.AfterSeq > 0 {
		query["after_seq"] = strconv.FormatInt(options.AfterSeq, 10)
	}
	if options.BeforeSeq > 0 {
		query["before_seq"] = strconv.FormatInt(options.BeforeSeq, 10)
	}
	if options.Limit > 0 {
		query["limit"] = strconv.Itoa(options.Limit)
	}
	var out ComputerStepPage
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/steps"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Recordings(ctx context.Context, id string) ([]ComputerRecording, error) {
	var out []ComputerRecording
	err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/recordings", nil, nil, &out)
	return out, err
}

func (c *ComputersClient) AskOwner(ctx context.Context, id string, input ComputerAskInput) (*ComputerApproval, error) {
	var out ComputerApproval
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/approvals", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Approvals(ctx context.Context, id, status string) ([]ComputerApproval, error) {
	var out []ComputerApproval
	err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/approvals"+computerQuery(map[string]string{"status": status}), nil, nil, &out)
	return out, err
}

func (c *ComputersClient) DecideApproval(ctx context.Context, approvalID string, approved bool) (*ComputerApproval, error) {
	var out ComputerApproval
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/approvals/"+computerSeg(approvalID)+"/decision", map[string]bool{"approved": approved}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) PauseAgent(ctx context.Context, id string) (*Computer, error) {
	var out Computer
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/pause-agent", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ResumeAgent(ctx context.Context, id string) (*Computer, error) {
	var out Computer
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/resume-agent", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Processes(ctx context.Context, id string) (*ComputerProcessList, error) {
	var out ComputerProcessList
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/processes", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Logs(ctx context.Context, id string, options ComputerLogOptions) (*ComputerLogPage, error) {
	query := map[string]string{"source": options.Source, "cursor": options.Cursor}
	if options.Limit > 0 {
		query["limit"] = strconv.Itoa(options.Limit)
	}
	var out ComputerLogPage
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/logs"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) DownloadRecording(ctx context.Context, id, recordingID string) ([]byte, error) {
	return c.client.computerBytes(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/recordings/"+computerSeg(recordingID), nil, map[string]string{"Accept": "video/webm"})
}

func (c *ComputersClient) Usage(ctx context.Context, projectID string, options ComputerUsageOptions) (*ComputerUsage, error) {
	var out ComputerUsage
	query := map[string]string{"project_id": projectID, "month": options.Month, "computer_id": options.ComputerID, "currency": options.Currency}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/usage"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Pricing(ctx context.Context, projectID, currency string) (*ComputerPricing, error) {
	var out ComputerPricing
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/pricing"+computerQuery(map[string]string{"project_id": projectID, "currency": currency}), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) PublicPricing(ctx context.Context) (*ComputerPublicPricing, error) {
	return FetchComputerPublicPricing(ctx, c.client.baseURL, c.client.httpClient)
}

func FetchComputerPublicPricing(ctx context.Context, baseURL string, httpClient HTTPDoer) (*ComputerPublicPricing, error) {
	var out ComputerPublicPricing
	if err := computerPublicCall(ctx, baseURL, httpClient, http.MethodGet, "/computers/public-pricing", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func computerPublicCall(ctx context.Context, baseURL string, httpClient HTTPDoer, method, path string, body any, out any) error {
	if baseURL == "" {
		baseURL = "https://api.stackshift.cloud/api/v1"
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(baseURL, "/")+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "StackShift-Go/0.1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return computerAPIError(response.StatusCode, raw)
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func (c *ComputersClient) Regions(ctx context.Context, projectID string) (*ComputerRegionList, error) {
	var out ComputerRegionList
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/regions"+computerQuery(map[string]string{"project_id": projectID}), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Settings(ctx context.Context, projectID string) (*ComputerAccountSettings, error) {
	var out ComputerAccountSettings
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/settings"+computerQuery(map[string]string{"project_id": projectID}), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) SaveSettings(ctx context.Context, projectID string, input SaveComputerSettingsInput) (*ComputerAccountSettings, error) {
	var out ComputerAccountSettings
	if err := c.client.computerRequest(ctx, http.MethodPut, "/computers/settings"+computerQuery(map[string]string{"project_id": projectID}), input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) System(ctx context.Context, id string) (*ComputerSystem, error) {
	var out ComputerSystem
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/system", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ListWindows(ctx context.Context, id string) (*ComputerWindowList, error) {
	var out ComputerWindowList
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/windows", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) SetTask(ctx context.Context, id string, input ComputerTaskInput) (*ComputerTask, error) {
	var out ComputerTask
	if err := c.client.computerRequest(ctx, http.MethodPut, "/computers/"+computerSeg(id)+"/task", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ClearTask(ctx context.Context, id string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/"+computerSeg(id)+"/task", nil, nil, nil)
}

func (c *ComputersClient) Setup(ctx context.Context, id string) (*ComputerSetup, error) {
	var out ComputerSetup
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/setup", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) StartSetup(ctx context.Context, id string, input ComputerSetupInput, idempotencyKey string) (*ComputerSetup, error) {
	var out ComputerSetup
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/setup", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) ImportImage(ctx context.Context, input ImportComputerImageInput, idempotencyKey string) (*ComputerImageImportResult, error) {
	var out ComputerImageImportResult
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/images/import", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) Approval(ctx context.Context, approvalID string) (*ComputerApproval, error) {
	var out ComputerApproval
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/approvals/"+computerSeg(approvalID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) WaitForApproval(ctx context.Context, approvalID string, interval time.Duration) (*ComputerApproval, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	for {
		approval, err := c.Approval(ctx, approvalID)
		if err != nil || approval.Status != "pending" {
			return approval, err
		}
		select {
		case <-ctx.Done():
			return approval, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (c *ComputersClient) OpenSession(ctx context.Context, id, kind string) (*ComputerSession, error) {
	var out ComputerSession
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/"+computerSeg(id)+"/sessions", map[string]string{"kind": kind}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputersClient) SocketURL(raw string, rows, columns int) (string, error) {
	base, err := url.Parse(c.client.baseURL)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	resolved := (&url.URL{Scheme: base.Scheme, Host: base.Host}).ResolveReference(parsed)
	switch resolved.Scheme {
	case "https":
		resolved.Scheme = "wss"
	case "http":
		resolved.Scheme = "ws"
	}
	if rows > 0 && columns > 0 {
		query := resolved.Query()
		query.Set("rows", strconv.Itoa(rows))
		query.Set("columns", strconv.Itoa(columns))
		resolved.RawQuery = query.Encode()
	}
	return resolved.String(), nil
}

func (c *ComputersClient) sessionURL(ctx context.Context, id, kind string, rows, columns int) (string, error) {
	session, err := c.OpenSession(ctx, id, kind)
	if err != nil {
		return "", err
	}
	return c.SocketURL(session.URL, rows, columns)
}

func (c *ComputersClient) CDPURL(ctx context.Context, id string) (string, error) {
	return c.sessionURL(ctx, id, "cdp", 0, 0)
}

func (c *ComputersClient) LiveURL(ctx context.Context, id string) (string, error) {
	return c.sessionURL(ctx, id, "live", 0, 0)
}

func (c *ComputersClient) ViewLink(ctx context.Context, id string) (string, error) {
	return c.sessionURL(ctx, id, "view_link", 0, 0)
}

func (c *ComputersClient) TerminalURL(ctx context.Context, id string, rows, columns int) (string, error) {
	return c.sessionURL(ctx, id, "terminal", rows, columns)
}

func (c *ComputersClient) Events(ctx context.Context, id string, afterSeq int64, reconnect bool, handle func(ComputerEvent) bool) error {
	last := afterSeq
	for {
		headers := map[string]string{"Accept": "text/event-stream"}
		query := map[string]string{}
		if last > 0 {
			headers["Last-Event-ID"] = strconv.FormatInt(last, 10)
			query["after_seq"] = strconv.FormatInt(last, 10)
		}
		response, err := c.client.computerSend(ctx, http.MethodGet, "/computers/"+computerSeg(id)+"/events"+computerQuery(query), nil, headers, false)
		if err == nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
			raw, _ := io.ReadAll(response.Body)
			response.Body.Close()
			return computerAPIError(response.StatusCode, raw)
		}
		if err == nil {
			stop, streamErr := readComputerEvents(response.Body, &last, handle)
			response.Body.Close()
			if stop {
				return nil
			}
			err = streamErr
		}
		if ctx.Err() != nil {
			return nil
		}
		if !reconnect {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
}

func readComputerEvents(body io.Reader, last *int64, handle func(ComputerEvent) bool) (bool, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 8<<20)
	var data []string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if len(data) == 0 {
				continue
			}
			var event ComputerEvent
			if err := json.Unmarshal([]byte(strings.Join(data, "\n")), &event); err != nil {
				return false, err
			}
			data = nil
			*last = event.Seq
			if !handle(event) {
				return true, nil
			}
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	return false, scanner.Err()
}

func (c *ComputerSnapshotsClient) List(ctx context.Context, projectID, computerID, cursor string, limit int) (*ComputerSnapshotPage, error) {
	query := map[string]string{"project_id": projectID, "computer_id": computerID, "cursor": cursor}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	var out ComputerSnapshotPage
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/snapshots"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputerSnapshotsClient) Delete(ctx context.Context, snapshotID string) error {
	return c.DeleteConfirmed(ctx, snapshotID, snapshotID)
}

func (c *ComputerSnapshotsClient) DeleteConfirmed(ctx context.Context, snapshotID, confirm string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/snapshots/"+computerSeg(snapshotID), nil, map[string]string{computerConfirmDeleteHeader: confirm}, nil)
}

func (c *ComputerCIClient) Overview(ctx context.Context, projectID string) (*ComputerCIOverview, error) {
	var out ComputerCIOverview
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/ci"+computerQuery(map[string]string{"project_id": projectID}), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputerCIClient) Labels(ctx context.Context, projectID string) ([]string, error) {
	overview, err := c.Overview(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return overview.Labels, nil
}

func (c *ComputerCIClient) Jobs(ctx context.Context, projectID string, limit int) ([]ComputerCIJob, error) {
	query := map[string]string{"project_id": projectID}
	if limit > 0 {
		query["limit"] = strconv.Itoa(limit)
	}
	var out struct {
		Items []ComputerCIJob `json:"items"`
	}
	if err := c.client.computerRequest(ctx, http.MethodGet, "/computers/ci/jobs"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *ComputerCIClient) Connect(ctx context.Context, projectID string, installationID int64) (*ComputerCIConnection, error) {
	var out ComputerCIConnection
	body := map[string]any{"project_id": projectID, "installation_id": installationID}
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/ci/connections", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputerCIClient) Disconnect(ctx context.Context, connectionID string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/ci/connections/"+computerSeg(connectionID), nil, nil, nil)
}

func (c *ComputerCIClient) ConnectCircleCI(ctx context.Context, input ComputerCircleCIConnectInput) (*ComputerCircleCIConnection, error) {
	var out ComputerCircleCIConnection
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/ci/circleci/connections", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputerCIClient) DisconnectCircleCI(ctx context.Context, connectionID string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/ci/circleci/connections/"+computerSeg(connectionID), nil, nil, nil)
}

func (c *ComputerKeysClient) List(ctx context.Context, projectID string) ([]ComputerKey, error) {
	var out []ComputerKey
	err := c.client.computerRequest(ctx, http.MethodGet, "/computers/keys"+computerQuery(map[string]string{"project_id": projectID}), nil, nil, &out)
	return out, err
}

func (c *ComputerKeysClient) Create(ctx context.Context, input CreateComputerKeyInput) (*ComputerCreatedKey, error) {
	var out ComputerCreatedKey
	if err := c.client.computerRequest(ctx, http.MethodPost, "/computers/keys", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *ComputerKeysClient) Revoke(ctx context.Context, keyID string) error {
	return c.client.computerRequest(ctx, http.MethodDelete, "/computers/keys/"+computerSeg(keyID), nil, nil, nil)
}

func ptrInt(value int) *int { return &value }

func ComputerClick(x, y int) ComputerAction {
	return ComputerAction{Type: "click", X: ptrInt(x), Y: ptrInt(y)}
}

func ComputerDoubleClick(x, y int) ComputerAction {
	return ComputerAction{Type: "double_click", X: ptrInt(x), Y: ptrInt(y)}
}

func ComputerMove(x, y int) ComputerAction {
	return ComputerAction{Type: "move", X: ptrInt(x), Y: ptrInt(y)}
}

func ComputerType(text string) ComputerAction { return ComputerAction{Type: "type", Text: text} }

func ComputerKeys(chords ...string) ComputerAction { return ComputerAction{Type: "key", Keys: chords} }

func ComputerScroll(x, y, dx, dy int) ComputerAction {
	return ComputerAction{Type: "scroll", X: ptrInt(x), Y: ptrInt(y), ScrollX: ptrInt(dx), ScrollY: ptrInt(dy)}
}

func ComputerDrag(path ...ComputerPoint) ComputerAction {
	return ComputerAction{Type: "drag", Path: path}
}

func ComputerOpenURL(address string) ComputerAction {
	return ComputerAction{Type: "open_url", URL: address}
}

func ComputerFillLogin(login string, submit bool) ComputerAction {
	return ComputerAction{Type: "fill_login", Login: login, Submit: submit}
}

func (a ComputerAction) WithScreenshot() ComputerAction {
	a.Screenshot = true
	return a
}

func (r ComputerActionResult) Err() error {
	if r.OK {
		return nil
	}
	if r.Error == nil {
		return errors.New("stackshift: computer action " + r.Type + " failed")
	}
	return &APIError{StatusCode: http.StatusOK, Code: r.Error.Code, Message: firstNonEmpty(r.Error.Message, "computer action "+r.Type+" failed")}
}
