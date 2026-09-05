package stackshift

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestAssetMediaRequests(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, method, path, revision, key string
		body                              map[string]any
		call                              func(*Client) error
	}{
		{name: "scoped metadata read", method: "GET", path: "/api/v1/assets/a%2Fb/metadata-editor", call: func(c *Client) error {
			_, err := c.Assets.DAM("space-1").MetadataEditor(context.Background(), "a/b")
			return err
		}},
		{name: "published schema revision", method: "POST", path: "/api/v1/assets/metadata-schemas/schema/publish", revision: `"7"`, call: func(c *Client) error {
			_, err := c.Assets.DAM("space-1").PublishMetadataSchema(context.Background(), "schema", 7)
			return err
		}},
		{name: "review note and revision", method: "POST", path: "/api/v1/assets/publications/pub/request-changes", revision: `"3"`, body: map[string]any{"note": "Correct attribution"}, call: func(c *Client) error {
			_, err := c.Assets.DAM("space-1").TransitionPublication(context.Background(), "pub", 3, "request-changes", "Correct attribution")
			return err
		}},
		{name: "empty actions preserved", method: "PUT", path: "/api/v1/assets/collaborators/member", revision: `"1"`, body: map[string]any{"actions": []any{}}, call: func(c *Client) error {
			_, err := c.Assets.DAM("space-1").UpdateCollaborator(context.Background(), "member", 1, []string{})
			return err
		}},
		{name: "render admission and zero budget", method: "POST", path: "/api/v1/assets/model/models", revision: `"2"`, key: "render-1", body: map[string]any{"operation": "render", "maximum_units": float64(0), "outputs": []any{"proof"}}, call: func(c *Client) error {
			input := map[string]any{"operation": "estimate", "outputs": []string{"proof"}}
			_, err := c.Assets.DAM("space-1").RenderProduct(context.Background(), "model", 2, "render-1", input, 0)
			if input["operation"] != "estimate" {
				t.Error("mutated caller input")
			}
			return err
		}},
		{name: "gallery revision", method: "POST", path: "/api/v1/assets/galleries/gallery/publish", revision: `"8"`, call: func(c *Client) error {
			_, err := c.Assets.DAM("space-1").TransitionGallery(context.Background(), "gallery", 8, "publish")
			return err
		}},
		{name: "publication cursor escaping", method: "GET", path: "/api/v1/assets/publications?cursor=a%2B%2Fb&status=published", call: func(c *Client) error {
			_, err := c.Assets.DAM("space-1").Publications(context.Background(), url.Values{"cursor": {"a+/b"}, "status": {"published"}})
			return err
		}},
		{name: "version pinned caption", method: "POST", path: "/api/v1/assets/film/versions/v1/video/captions", body: map[string]any{"is_default": false, "language": "en"}, call: func(c *Client) error {
			_, err := c.Assets.Video("space-1").AddCaption(context.Background(), "film", "v1", map[string]any{"is_default": false, "language": "en"})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(Options{APIKey: "account-key", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != tt.method || req.URL.RequestURI() != tt.path {
					t.Fatalf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
				}
				if req.Header.Get("Authorization") != "Bearer account-key" || req.Header.Get("X-Asset-Space-ID") != "space-1" {
					t.Fatal("incorrect authentication or space")
				}
				if req.Header.Get("If-Match") != tt.revision || req.Header.Get("Idempotency-Key") != tt.key {
					t.Fatal("missing mutation preconditions")
				}
				if tt.body != nil {
					var body map[string]any
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(body, tt.body) {
						t.Fatalf("body=%#v want=%#v", body, tt.body)
					}
				} else if req.Body != nil {
					raw, err := io.ReadAll(req.Body)
					if err != nil || len(raw) > 0 {
						t.Fatalf("unexpected body %q: %v", raw, err)
					}
				}
				return jsonResponse(200, `{"success":true,"data":{"id":"result"}}`), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAssetPlaybackCredentialIsolation(t *testing.T) {
	t.Parallel()
	calls := 0
	client, err := New(Options{APIKey: "account-key", BaseURL: "https://example.test/custom/api/v1", HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.URL.Path != "/playback/sessions/renew" || req.Header.Get("Authorization") != "Bearer session-key" || req.Header.Get("X-Asset-Space-ID") != "" {
			t.Fatal("playback crossed credential boundary")
		}
		return jsonResponse(403, `{"error":{"code":"revoked","message":"Access revoked"}}`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Assets.Video("space-1").RenewSession(context.Background(), "session-key")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 403 || apiErr.Code != "revoked" {
		t.Fatalf("lost access error: %v", err)
	}
	if _, err := client.Assets.Video("").RenewSession(context.Background(), " "); err == nil {
		t.Fatal("accepted empty credential")
	}
	if calls != 1 {
		t.Fatalf("unexpected retries: %d", calls)
	}
}
