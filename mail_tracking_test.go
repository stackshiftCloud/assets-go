package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMailTrackingAndDraftContracts(t *testing.T) {
	var paths []string
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		paths = append(paths, req.Method+" "+req.URL.Path)
		if req.Body != nil {
			var body map[string]any
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if strings.HasSuffix(req.URL.Path, "/schedule") && body["revision"] != float64(3) {
				t.Fatalf("revision lost: %v", body)
			}
			if strings.HasSuffix(req.URL.Path, "/settings") && body["opens"] != false {
				t.Fatalf("false override lost: %v", body)
			}
		}
		return jsonResponse(200, `{"id":"campaign","revision":3,"openEngagedDelivered":2}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err = client.Mail.Tracking.UpdateSettings(ctx, MailTrackingSettings{Clicks: true}); err != nil {
		t.Fatal(err)
	}
	report, err := client.Mail.Tracking.Message(ctx, "msg_test")
	if err != nil {
		t.Fatal(err)
	}
	if report.OpenEngagedDelivered != 2 {
		t.Fatalf("report not decoded: %+v", report)
	}
	if _, err = client.Mail.Campaigns.Schedule(ctx, "campaign", time.Now(), 3); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 || !strings.HasSuffix(paths[1], "/mail/messages/msg_test/engagement") {
		t.Fatalf("unexpected routes: %v", paths)
	}
}
