package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAssetPlatformRequests(t *testing.T) {
	tests := []struct {
		name, method, path, revision string
		body                         map[string]any
		call                         func(*Client) error
	}{
		{"Cloudinary connection", "POST", "/api/v1/assets/connections", "",
			map[string]any{"provider": "cloudinary"}, func(c *Client) error {
				_, err := c.Assets.Connections.Create(context.Background(), map[string]any{"provider": "cloudinary"})
				return err
			}},
		{"storage revision", "PUT", "/api/v1/assets/buckets/a%2Fb/storage", `"7"`,
			map[string]any{"storage_mode": "stackshift_s2"}, func(c *Client) error {
				_, err := c.Assets.Storage.Configure(context.Background(), "a/b", 7, map[string]any{"storage_mode": "stackshift_s2"})
				return err
			}},
		{"migration scope", "POST", "/api/v1/assets/migrations", "",
			map[string]any{"source_prefix": "campaign/"}, func(c *Client) error {
				_, err := c.Assets.Migrations.Start(context.Background(), map[string]any{"source_prefix": "campaign/"})
				return err
			}},
		{"migration report", "GET", "/api/v1/assets/migrations/m%2F1/report?format=json", "", nil, func(c *Client) error {
			_, err := c.Assets.Migrations.Report(context.Background(), "m/1")
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := New(Options{APIKey: "account", HTTPClient: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Method != test.method || request.URL.RequestURI() != test.path ||
					request.Header.Get("Authorization") != "Bearer account" ||
					request.Header.Get("X-Asset-Space-ID") != "space-1" ||
					request.Header.Get("If-Match") != test.revision {
					t.Fatalf("unexpected request: %s %s %+v", request.Method, request.URL, request.Header)
				}
				if test.body != nil {
					var body map[string]any
					if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
						t.Fatal(err)
					}
					for key, expected := range test.body {
						if body[key] != expected {
							t.Fatalf("%s=%v, want %v", key, body[key], expected)
						}
					}
				}
				return jsonResponse(200, `{"success":true,"data":{"id":"result"}}`), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			client.Assets = client.Assets.ForSpace("space-1")
			if err := test.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}
