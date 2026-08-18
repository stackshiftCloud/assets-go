package stackshift

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAssetV2TransformAndWorkflowRoutes(t *testing.T) {
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
		case strings.HasSuffix(req.URL.Path, "/assets/workflows/catalog"):
			return jsonResponse(http.StatusOK, `{"success":true,"data":{"nodes":[{"type":"trigger.manual","version":"1","name":"Manual","category":"Triggers","description":"Manual trigger","input_ports":[],"output_ports":[],"config_schema":{},"risk":"read","requires_approval":false,"approval_mode":"none"}],"templates":[]}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/workflows/wf_1/publish"):
			if !strings.Contains(string(body), `"expected_revision":3`) {
				t.Fatalf("unexpected publish body: %s", body)
			}
			return jsonResponse(http.StatusCreated, `{"success":true,"data":{"id":"version_1","workflow_id":"wf_1","version":1,"semantic_digest":"sha256","definition":{"version":1,"nodes":[],"edges":[]},"layout":{"nodes":[],"viewport":{"x":0,"y":0,"zoom":1}},"created_by":"user_1","created_at":"2026-08-18T00:00:00Z"}}`), nil
		case strings.HasSuffix(req.URL.Path, "/assets/workflows/connections"):
			if strings.Contains(string(body), `"id"`) || !strings.Contains(string(body), `"host":"hooks.example.com"`) {
				t.Fatalf("unexpected connection request: %s", body)
			}
			return jsonResponse(http.StatusCreated, `{"success":true,"data":{"id":"connection_1","name":"webhook","host":"hooks.example.com","allowed_path_prefixes":["/v1/"],"default_headers":{},"revision":1,"created_at":"2026-08-18T00:00:00Z","updated_at":"2026-08-18T00:00:00Z"}}`), nil
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

	catalog, err := client.Assets.WorkflowCatalog(context.Background())
	if err != nil || len(catalog.Nodes) != 1 {
		t.Fatalf("catalog=%+v err=%v", catalog, err)
	}
	version, err := client.Assets.PublishWorkflow(context.Background(), "wf_1", 3)
	if err != nil || version.SemanticDigest != "sha256" {
		t.Fatalf("version=%+v err=%v", version, err)
	}
	connection, err := client.Assets.CreateWorkflowHTTPConnection(context.Background(), CreateAssetWorkflowHTTPConnectionInput{
		Name: "webhook", Host: "hooks.example.com", AllowedPathPrefixes: []string{"/v1/"},
	})
	if err != nil || connection.ID != "connection_1" {
		t.Fatalf("connection=%+v err=%v", connection, err)
	}
	if len(requests) != 4 {
		t.Fatalf("requests=%d", len(requests))
	}
}
