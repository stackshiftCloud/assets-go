package stackshift

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const teamLongPollSeconds = 25

type TeamsClient struct{ client *Client }

func (i UpdateTeamInput) MarshalJSON() ([]byte, error) {
	body := map[string]any{}
	if i.Job != nil {
		body["job"] = *i.Job
	}
	if i.StopRule != nil {
		body["stop_rule"] = *i.StopRule
	}
	if i.ClearSpendLimit {
		body["spend_limit"] = nil
	} else if i.SpendLimit != nil {
		body["spend_limit"] = i.SpendLimit
	}
	if i.AskBeforeActing != nil {
		body["ask_before_acting"] = *i.AskBeforeActing
	}
	return json.Marshal(body)
}

func teamPath(id string) string { return "/computers/teams/" + computerSeg(id) }

func (t *TeamsClient) Create(ctx context.Context, input CreateTeamInput, idempotencyKey string) (*Team, error) {
	var out Team
	if err := t.client.computerRequest(ctx, http.MethodPost, "/computers/teams", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type TeamListOptions struct {
	State  string
	Cursor string
	Limit  int
}

func (t *TeamsClient) List(ctx context.Context, projectID string, options TeamListOptions) (*TeamPage, error) {
	query := map[string]string{"project_id": projectID, "state": options.State, "cursor": options.Cursor}
	if options.Limit > 0 {
		query["limit"] = strconv.Itoa(options.Limit)
	}
	var out TeamPage
	if err := t.client.computerRequest(ctx, http.MethodGet, "/computers/teams"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) ListAll(ctx context.Context, projectID, state string) ([]Team, error) {
	var all []Team
	cursor := ""
	for {
		page, err := t.List(ctx, projectID, TeamListOptions{State: state, Cursor: cursor})
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

func (t *TeamsClient) Get(ctx context.Context, id string) (*Team, error) {
	var out Team
	if err := t.client.computerRequest(ctx, http.MethodGet, teamPath(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) Update(ctx context.Context, id string, input UpdateTeamInput) (*Team, error) {
	var out Team
	if err := t.client.computerRequest(ctx, http.MethodPatch, teamPath(id), input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) Delete(ctx context.Context, id string) (*Team, error) {
	var out Team
	if err := t.client.computerRequest(ctx, http.MethodDelete, teamPath(id), nil, map[string]string{"X-StackShift-Confirm-Delete": id}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) Pause(ctx context.Context, id string) (*Team, error) {
	var out Team
	if err := t.client.computerRequest(ctx, http.MethodPost, teamPath(id)+"/pause", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) Resume(ctx context.Context, id string) (*Team, error) {
	var out Team
	if err := t.client.computerRequest(ctx, http.MethodPost, teamPath(id)+"/resume", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) AddMember(ctx context.Context, teamID string, input TeamMemberInput, idempotencyKey string) (*TeamMember, error) {
	var out TeamMember
	if err := t.client.computerRequest(ctx, http.MethodPost, teamPath(teamID)+"/members", input, computerIdempotency(idempotencyKey), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) RemoveMember(ctx context.Context, teamID, memberID string, keepComputer bool) error {
	query := map[string]string{}
	if keepComputer {
		query["keep_computer"] = "true"
	}
	return t.client.computerRequest(ctx, http.MethodDelete, teamPath(teamID)+"/members/"+computerSeg(memberID)+computerQuery(query), nil, nil, nil)
}

func (t *TeamsClient) Send(ctx context.Context, teamID string, input SendTeamMessageInput) (*TeamMessage, error) {
	var out TeamMessage
	if err := t.client.computerRequest(ctx, http.MethodPost, teamPath(teamID)+"/messages", input, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type TeamMessageOptions struct {
	AfterSeq int64
	Limit    int
	Wait     int
}

func (t *TeamsClient) ListMessages(ctx context.Context, teamID string, options TeamMessageOptions) (*TeamMessagePage, error) {
	query := map[string]string{}
	if options.AfterSeq > 0 {
		query["after_seq"] = strconv.FormatInt(options.AfterSeq, 10)
	}
	if options.Limit > 0 {
		query["limit"] = strconv.Itoa(options.Limit)
	}
	if options.Wait > 0 {
		query["wait"] = strconv.Itoa(options.Wait)
	}
	var out TeamMessagePage
	if err := t.client.computerRequest(ctx, http.MethodGet, teamPath(teamID)+"/messages"+computerQuery(query), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) WaitForMessages(ctx context.Context, teamID string, afterSeq int64, timeout time.Duration) (*TeamMessagePage, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	for {
		wait := teamLongPollSeconds
		if !deadline.IsZero() {
			remaining := int(time.Until(deadline).Seconds())
			if remaining < 1 {
				return &TeamMessagePage{Items: []TeamMessage{}, NextSeq: afterSeq}, nil
			}
			if remaining < wait {
				wait = remaining
			}
		}
		page, err := t.ListMessages(ctx, teamID, TeamMessageOptions{AfterSeq: afterSeq, Wait: wait})
		if err != nil {
			return nil, err
		}
		if len(page.Items) > 0 {
			return page, nil
		}
		if page.NextSeq > afterSeq {
			afterSeq = page.NextSeq
		}
		if ctx.Err() != nil {
			return &TeamMessagePage{Items: []TeamMessage{}, NextSeq: afterSeq}, nil
		}
	}
}

func (t *TeamsClient) StreamMessages(ctx context.Context, teamID string, afterSeq int64, reconnect bool, handle func(TeamMessage) bool) error {
	last := afterSeq
	for {
		headers := map[string]string{"Accept": "text/event-stream"}
		query := map[string]string{}
		if last > 0 {
			headers["Last-Event-ID"] = strconv.FormatInt(last, 10)
			query["after_seq"] = strconv.FormatInt(last, 10)
		}
		response, err := t.client.computerSend(ctx, http.MethodGet, teamPath(teamID)+"/messages/stream"+computerQuery(query), nil, headers, false)
		if err == nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
			raw, _ := io.ReadAll(response.Body)
			response.Body.Close()
			return computerAPIError(response.StatusCode, raw)
		}
		if err == nil {
			stop, streamErr := readTeamMessages(response.Body, &last, handle)
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

func readTeamMessages(body io.Reader, last *int64, handle func(TeamMessage) bool) (bool, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 8<<20)
	var data []string
	name := ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if len(data) > 0 && (name == "" || name == "message") {
				var message TeamMessage
				if err := json.Unmarshal([]byte(strings.Join(data, "\n")), &message); err != nil {
					return false, err
				}
				*last = message.Seq
				if !handle(message) {
					return true, nil
				}
			}
			data, name = nil, ""
		case strings.HasPrefix(line, ":"):
		case strings.HasPrefix(line, "event:"):
			name = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	return false, scanner.Err()
}

func (t *TeamsClient) Files(ctx context.Context, teamID, path string) (*TeamFiles, error) {
	var out TeamFiles
	if err := t.client.computerRequest(ctx, http.MethodGet, teamPath(teamID)+"/files"+computerQuery(map[string]string{"path": path}), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) Network(ctx context.Context, teamID string) (*TeamNetworkMap, error) {
	var out TeamNetworkMap
	if err := t.client.computerRequest(ctx, http.MethodGet, teamPath(teamID)+"/network", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (t *TeamsClient) Activity(ctx context.Context, teamID string) (*TeamActivity, error) {
	var out TeamActivity
	if err := t.client.computerRequest(ctx, http.MethodGet, teamPath(teamID)+"/activity", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
