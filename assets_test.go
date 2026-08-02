package stackshift

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAssetTransformPresetsAndTransformationAPI(t *testing.T) {
	var calls []string
	client, err := New(Options{
		APIKey: "sk_test",
		HTTPClient: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls = append(calls, req.Method+" "+req.URL.String())
			var body []byte
			if req.Body != nil {
				body, _ = io.ReadAll(req.Body)
			}
			switch {
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/assets/transformations"):
				return jsonResponse(http.StatusOK, `{"success":true,"data":{"transformations":[{"id":"tr_1","name":"custom-card","spec":"w_640,h_360,c_fill,f_webp,q_82","normalized_spec":"w_640,h_360,c_fill,f_webp,q_82","created_at":"2026-05-11T00:00:00Z","updated_at":"2026-05-11T00:00:00Z"}]}}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/assets/transformations"):
				if !strings.Contains(string(body), `"name":"hero"`) {
					t.Fatalf("unexpected create body: %s", body)
				}
				return jsonResponse(http.StatusCreated, `{"success":true,"data":{"id":"tr_2","name":"hero","spec":"w_1600,h_900,c_fill,f_webp,q_84","normalized_spec":"w_1600,h_900,c_fill,f_webp,q_84","created_at":"2026-05-11T00:00:00Z","updated_at":"2026-05-11T00:00:00Z"}}`), nil
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/assets/asset_123/transform-url"):
				return jsonResponse(http.StatusCreated, `{"success":true,"data":{"url":"https://cdn.stackshift.test/assets/asset_123/tr/w_1600,h_900,c_fill,f_webp,q_84/hero.webp","expires_at":"2026-05-11T00:10:00Z"}}`), nil
			case req.Method == http.MethodDelete && strings.HasSuffix(req.URL.Path, "/assets/transformations/custom-card"):
				return jsonResponse(http.StatusOK, `{"success":true,"data":{"deleted":true}}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":{"code":"not_found","message":"Not found."}}`), nil
			}
		}),
		CDNBaseURL: "https://cdn.stackshift.test",
	})
	if err != nil {
		t.Fatal(err)
	}

	hero, ok := AssetTransformPresetByName("hero")
	if !ok || hero.Width != 1600 || hero.Category != "responsive-web" {
		t.Fatalf("unexpected hero preset: %+v", hero)
	}
	if got := AssetTransformSpec(hero.AssetTransformOptions); got != "w_1600,h_900,c_fill,f_webp,q_84,g_centre,a_first" {
		t.Fatalf("unexpected spec: %s", got)
	}
	if got := client.Assets.URL("asset_123", hero.AssetTransformOptions); got != "https://cdn.stackshift.test/assets/asset_123/tr/w_1600%2Ch_900%2Cc_fill%2Cf_webp%2Cq_84%2Cg_centre%2Ca_first" {
		t.Fatalf("unexpected URL: %s", got)
	}

	transformations, err := client.Assets.ListTransformations(context.Background())
	if err != nil || transformations[0].Name != "custom-card" {
		t.Fatalf("transformations=%+v err=%v", transformations, err)
	}

	created, err := client.Assets.CreateTransformation(context.Background(), "hero", hero.AssetTransformOptions)
	if err != nil || created.Name != "hero" {
		t.Fatalf("created=%+v err=%v", created, err)
	}

	signed, err := client.Assets.SignedTransformURL(context.Background(), "asset_123", hero.AssetTransformOptions)
	if err != nil || !strings.Contains(signed.URL, "/assets/asset_123/tr/") {
		t.Fatalf("signed=%+v err=%v", signed, err)
	}

	if err := client.Assets.DeleteTransformation(context.Background(), "custom-card"); err != nil {
		t.Fatal(err)
	}
	if calls[len(calls)-1] != "DELETE https://api.stackshift.cloud/api/v1/assets/transformations/custom-card" {
		t.Fatalf("unexpected delete call: %s", calls[len(calls)-1])
	}
}
