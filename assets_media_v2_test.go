package stackshift

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAssetV2TransformRoutes(t *testing.T) {
	t.Parallel()
	var requests []*http.Request
	client, err := New(Options{APIKey: "sk_test", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req)
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}
		switch {
		case strings.HasSuffix(req.URL.Path, "/assets/asset_1/transforms"):
			if req.Header.Get("Idempotency-Key") != "render-1" || !strings.Contains(string(body), `"timestamp_seconds":12.5`) {
				t.Fatalf("unexpected transform request: headers=%v body=%s", req.Header, body)
			}
			return jsonResponse(http.StatusAccepted, `{"success":true,"data":{"job_id":"job_1","status":"queued","status_url":"/api/v1/assets/jobs/job_1"}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/asset_1/transforms/estimate"):
			if !strings.Contains(string(body), `"version_id":"version_1"`) || !strings.Contains(string(body), `"type":"remove_background"`) {
				t.Fatalf("unexpected estimate request: %s", body)
			}
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"cost_micros":30000,"currency":"USD","steps":[{"index":1,"type":"remove_background","cost_micros":30000}]}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/asset_1/usage"):
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"window_days":30,"referrers":[{"host":"shop.example.com","requests":120}]}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/asset_1/hold"):
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"kind":"scan","finding":"f","explanation":"e","explained_by":"scan_result","version_id":"version_1","upload":{"window_minutes":10,"files":1,"held":[{"id":"asset_1","name":"a.zip"}]},"same_content":0}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/asset_1/reupload-requests") && req.Method == http.MethodPost:
			if !strings.Contains(string(body), `"email":"owner@example.com"`) {
				t.Fatalf("unexpected reupload request: %s", body)
			}
			return jsonResponse(http.StatusCreated, `{"success":true,"data":{"id":"req_1","email":"owner@example.com","status":"sent","expires_at":"2026-08-25T00:00:00Z","created_at":"2026-08-18T00:00:00Z"}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/asset_1/reupload-requests"):
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"requests":[{"id":"req_1","email":"owner@example.com","status":"sent","expires_at":"2026-08-25T00:00:00Z","created_at":"2026-08-18T00:00:00Z"}]}}`), nil
		default:
			return jsonResponse(http.StatusNotFound, `{"error":{"code":"not_found","message":"Not found."}}`), nil
		}
	})})
	if err != nil {
		t.Fatal(err)
	}

	definition := AssetTransformDefinition{Version: 2, MediaKind: "video",
		Operations: []AssetTransformOperation{{Type: "timestamp", TimestampSeconds: 12.5}},
		Output:     AssetTransformOutput{Kind: "thumbnail", Format: "webp", Quality: 82}}
	job, err := client.Assets.MaterializeTransform(context.Background(), "asset_1",
		AssetTransformMaterializeInput{Definition: &definition}, "render-1")
	if err != nil || job.JobID != "job_1" {
		t.Fatalf("job=%+v err=%v", job, err)
	}

	estimate, err := client.Assets.EstimateTransform(context.Background(), "asset_1", "version_1", AssetTransformDefinition{
		Version: 2, MediaKind: "image", Operations: []AssetTransformOperation{{Type: "remove_background"}},
		Output: AssetTransformOutput{Kind: "image", Format: "webp"},
	})
	if err != nil || estimate.CostMicros != 30000 || len(estimate.Steps) != 1 {
		t.Fatalf("estimate=%+v err=%v", estimate, err)
	}

	usage, err := client.Assets.Usage(context.Background(), "asset_1")
	if err != nil || usage.Referrers[0].Requests != 120 {
		t.Fatalf("usage=%+v err=%v", usage, err)
	}
	hold, err := client.Assets.Hold(context.Background(), "asset_1")
	if err != nil || hold.Upload.Held[0].Name != "a.zip" {
		t.Fatalf("hold=%+v err=%v", hold, err)
	}
	reupload, err := client.Assets.CreateReuploadRequest(context.Background(), "asset_1", AssetReuploadRequestInput{Email: "owner@example.com"})
	if err != nil || reupload.Status != "sent" {
		t.Fatalf("reupload=%+v err=%v", reupload, err)
	}
	reuploads, err := client.Assets.ReuploadRequests(context.Background(), "asset_1")
	if err != nil || len(reuploads) != 1 {
		t.Fatalf("reuploads=%+v err=%v", reuploads, err)
	}

	if len(requests) != 6 {
		t.Fatalf("requests=%d", len(requests))
	}
}
