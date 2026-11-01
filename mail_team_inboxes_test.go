package stackshift

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type teamInboxCall struct {
	method string
	path   string
	query  url.Values
	body   string
}

func newTeamInboxClient(t *testing.T, calls *[]teamInboxCall, activity *[]string) *Client {
	t.Helper()
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		if request.Body != nil {
			raw, _ := io.ReadAll(request.Body)
			body = string(raw)
		}
		*calls = append(*calls, teamInboxCall{method: request.Method, path: request.URL.Path, query: request.URL.Query(), body: body})
		path := strings.TrimPrefix(request.URL.Path, "/api/v1/mail/mailboxes/mbx_1")
		switch {
		case strings.HasSuffix(path, "/discard"):
			return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
		case path == "/conversations":
			return jsonResponse(http.StatusOK, `{"conversations":[{"mailboxId":"mbx_1","threadId":"thr_1","status":"open","revision":4}],"counts":{"open":1,"waiting":0,"done":0},"nextCursor":null}`), nil
		case path == "/conversations/thr_1":
			return jsonResponse(http.StatusOK, `{"conversation":{"threadId":"thr_1","revision":4},"notes":[],"activity":[]}`), nil
		case path == "/conversations/thr_1/notes":
			return jsonResponse(http.StatusCreated, `{"id":"note_1","body":"checked logs"}`), nil
		case path == "/pending-replies":
			return jsonResponse(http.StatusOK, `{"pendingReplies":[{"id":"pr_1","state":"waiting","reasons":["New contact"]}]}`), nil
		case path == "/activity":
			if len(*activity) > 0 {
				next := (*activity)[0]
				*activity = (*activity)[1:]
				return jsonResponse(http.StatusOK, next), nil
			}
			return jsonResponse(http.StatusOK, `{"activity":[],"next":`+defaultString(request.URL.Query().Get("after"), "0")+`}`), nil
		case path == "/people":
			return jsonResponse(http.StatusOK, `{"people":[{"userId":"usr_1","name":"Ada Obi","email":"ada@example.com"}],"workers":[{"id":"wrk_1","mailboxId":"mbx_1","kind":"stackie","name":"Stackie","enabled":true,"sendLevel":"ask","wakeOn":"assigned","platformResources":[{"kind":"project","id":"prj_1"}],"revision":2}]}`), nil
		case path == "/team-inbox":
			return jsonResponse(http.StatusOK, `{"mailboxId":"mbx_1","name":"Support","forwardTo":[],"lockSeconds":120}`), nil
		}
		return jsonResponse(http.StatusOK, `{"threadId":"thr_1","revision":5}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestTeamInboxConversations(t *testing.T) {
	var calls []teamInboxCall
	client := newTeamInboxClient(t, &calls, new([]string))
	ctx := context.Background()
	base := "/api/v1/mail/mailboxes/mbx_1"

	list, err := client.Mail.TeamInboxes.Conversations.List(ctx, "mbx_1", TeamInboxConversationListOptions{Status: "open", Assignee: "worker:wrk_1", Pending: true, Cursor: "c1", Limit: 20})
	if err != nil || list.Counts.Open != 1 || list.Conversations[0].ThreadID != "thr_1" || list.NextCursor != nil {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	query := calls[0].query
	if calls[0].path != base+"/conversations" || query.Get("status") != "open" || query.Get("assignee") != "worker:wrk_1" || query.Get("pending") != "true" || query.Get("cursor") != "c1" || query.Get("limit") != "20" {
		t.Fatalf("call=%+v", calls[0])
	}

	detail, err := client.Mail.TeamInboxes.Conversations.Get(ctx, "mbx_1", "thr_1")
	if err != nil || detail.Conversation.Revision != 4 || calls[1].path != base+"/conversations/thr_1" {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}

	if _, err := client.Mail.TeamInboxes.Conversations.Assign(ctx, "mbx_1", "thr_1", TeamInboxAssignRequest{WorkerID: "wrk_1", ExpectedRevision: 4}); err != nil {
		t.Fatal(err)
	}
	if calls[2].method != http.MethodPut || calls[2].path != base+"/conversations/thr_1/assignee" || calls[2].body != `{"workerId":"wrk_1","expectedRevision":4}` {
		t.Fatalf("call=%+v", calls[2])
	}
	if _, err := client.Mail.TeamInboxes.Conversations.Assign(ctx, "mbx_1", "thr_1", TeamInboxAssignRequest{ExpectedRevision: 5}); err != nil || calls[3].body != `{"expectedRevision":5}` {
		t.Fatalf("call=%+v err=%v", calls[3], err)
	}

	if _, err := client.Mail.TeamInboxes.Conversations.SetStatus(ctx, "mbx_1", "thr_1", TeamInboxStatusRequest{Status: "done", ExpectedRevision: 5}); err != nil {
		t.Fatal(err)
	}
	if calls[4].method != http.MethodPut || calls[4].path != base+"/conversations/thr_1/status" || calls[4].body != `{"status":"done","expectedRevision":5}` {
		t.Fatalf("call=%+v", calls[4])
	}

	note, err := client.Mail.TeamInboxes.Conversations.AddNote(ctx, "mbx_1", "thr_1", TeamInboxNoteRequest{Body: "checked logs", Mentions: []string{"usr_1"}})
	if err != nil || note.ID != "note_1" || calls[5].method != http.MethodPost || calls[5].path != base+"/conversations/thr_1/notes" || calls[5].body != `{"body":"checked logs","mentions":["usr_1"]}` {
		t.Fatalf("note=%+v err=%v call=%+v", note, err, calls[5])
	}

	if _, err := client.Mail.TeamInboxes.Conversations.Reply(ctx, "mbx_1", "thr_1", TeamInboxReplyRequest{DraftEmailID: "draft_1", Then: "waiting"}); err != nil {
		t.Fatal(err)
	}
	if calls[6].method != http.MethodPost || calls[6].path != base+"/conversations/thr_1/send" || calls[6].body != `{"draftEmailId":"draft_1","then":"waiting"}` {
		t.Fatalf("call=%+v", calls[6])
	}
}

func TestTeamInboxPendingReplies(t *testing.T) {
	var calls []teamInboxCall
	client := newTeamInboxClient(t, &calls, new([]string))
	ctx := context.Background()
	base := "/api/v1/mail/mailboxes/mbx_1"

	pending, err := client.Mail.TeamInboxes.PendingReplies.List(ctx, "mbx_1")
	if err != nil || len(pending) != 1 || pending[0].State != "waiting" || pending[0].Reasons[0] != "New contact" || calls[0].path != base+"/pending-replies" {
		t.Fatalf("pending=%+v err=%v", pending, err)
	}
	edited := "Edited reply"
	if _, err := client.Mail.TeamInboxes.PendingReplies.Send(ctx, "mbx_1", "pr_1", TeamInboxPendingSendRequest{Body: &edited}); err != nil {
		t.Fatal(err)
	}
	if calls[1].method != http.MethodPost || calls[1].path != base+"/pending-replies/pr_1/send" || calls[1].body != `{"body":"Edited reply"}` {
		t.Fatalf("call=%+v", calls[1])
	}
	if _, err := client.Mail.TeamInboxes.PendingReplies.Send(ctx, "mbx_1", "pr_1", TeamInboxPendingSendRequest{}); err != nil || calls[2].body != `{}` {
		t.Fatalf("call=%+v err=%v", calls[2], err)
	}
	if err := client.Mail.TeamInboxes.PendingReplies.Discard(ctx, "mbx_1", "pr_1"); err != nil {
		t.Fatal(err)
	}
	if calls[3].method != http.MethodPost || calls[3].path != base+"/pending-replies/pr_1/discard" {
		t.Fatalf("call=%+v", calls[3])
	}
}

func TestTeamInboxActivityAndSettings(t *testing.T) {
	var calls []teamInboxCall
	pages := []string{`{"activity":[],"next":7}`, `{"activity":[{"id":9,"kind":"received"}],"next":9}`}
	client := newTeamInboxClient(t, &calls, &pages)
	ctx := context.Background()
	activity := client.Mail.TeamInboxes.Activity

	if _, err := activity.List(ctx, "mbx_1", TeamInboxActivityListOptions{After: 7, Limit: 50}); err != nil {
		t.Fatal(err)
	}
	if calls[0].query.Get("after") != "7" || calls[0].query.Get("limit") != "50" {
		t.Fatalf("call=%+v", calls[0])
	}

	calls = nil
	pages = []string{`{"activity":[],"next":7}`, `{"activity":[{"id":9,"kind":"received"}],"next":9}`}
	woke, err := activity.Wait(ctx, "mbx_1", TeamInboxActivityWaitOptions{After: 7, Timeout: 5 * time.Second, Interval: 10 * time.Millisecond})
	if err != nil || woke.Next != 9 || woke.Activity[0].Kind != "received" || len(calls) != 2 || calls[1].query.Get("after") != "7" {
		t.Fatalf("woke=%+v err=%v calls=%+v", woke, err, calls)
	}

	idle, err := activity.Wait(ctx, "mbx_1", TeamInboxActivityWaitOptions{After: 9, Timeout: 50 * time.Millisecond, Interval: 10 * time.Millisecond})
	if err != nil || len(idle.Activity) != 0 || idle.Next != 9 {
		t.Fatalf("idle=%+v err=%v", idle, err)
	}

	settings, err := client.Mail.TeamInboxes.Settings.Get(ctx, "mbx_1")
	if err != nil || settings.LockSeconds != 120 || calls[len(calls)-1].path != "/api/v1/mail/mailboxes/mbx_1/team-inbox" {
		t.Fatalf("settings=%+v err=%v", settings, err)
	}
}

func TestTeamInboxPeople(t *testing.T) {
	var calls []teamInboxCall
	client := newTeamInboxClient(t, &calls, &[]string{})

	result, err := client.Mail.TeamInboxes.People(context.Background(), "mbx_1")
	if err != nil || len(result.People) != 1 || result.People[0].Email != "ada@example.com" || result.People[0].UserID != "usr_1" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	worker := result.Workers[0]
	if worker.SendLevel != "ask" || !worker.Enabled || worker.PlatformResources[0].Kind != "project" || worker.Revision != 2 {
		t.Fatalf("worker=%+v", worker)
	}
	if calls[0].method != http.MethodGet || calls[0].path != "/api/v1/mail/mailboxes/mbx_1/people" {
		t.Fatalf("call=%+v", calls[0])
	}
}
