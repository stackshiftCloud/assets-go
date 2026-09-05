package stackshift

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type assetMediaClient struct {
	client  *Client
	spaceID string
}
type AssetDAMClient struct{ assetMediaClient }
type AssetVideoClient struct{ assetMediaClient }

// DAM selects the asset space for governed media operations; an empty ID uses the owner's space.
func (a *AssetsClient) DAM(spaceID string) *AssetDAMClient {
	return &AssetDAMClient{assetMediaClient: assetMediaClient{client: a.client, spaceID: spaceID}}
}

// Video selects the asset space for native video management.
func (a *AssetsClient) Video(spaceID string) *AssetVideoClient {
	return &AssetVideoClient{assetMediaClient: assetMediaClient{client: a.client, spaceID: spaceID}}
}

func (a *assetMediaClient) call(
	ctx context.Context,
	method,
	path string,
	body any,
	headers map[string]string,
	out any,
) error {
	if headers == nil {
		headers = map[string]string{}
	}
	if a.spaceID != "" {
		headers["X-Asset-Space-ID"] = a.spaceID
	}
	// Preserve explicit zero, false and empty metadata values through the legacy transport.
	var payload any
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode assets request: %w", err)
		}
		payload = json.RawMessage(raw)
	}
	return a.client.requestWithHeaders(
		ctx,
		method,
		path,
		payload,
		headers,
		out,
	)
}

func (a *assetMediaClient) object(
	ctx context.Context,
	method,
	path string,
	body any,
	headers map[string]string,
) (map[string]any, error) {
	result := map[string]any{}
	if err := a.call(ctx, method, path, body, headers, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func mediaRevision(revision int64) map[string]string {
	return map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}
}

func mediaQuery(path string, query url.Values) string {
	if len(query) == 0 {
		return path
	}
	return path + "?" + query.Encode()
}

func (a *AssetVideoClient) playback(ctx context.Context, path, credential string, body any) (map[string]any, error) {
	base, err := url.Parse(a.client.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse assets base url: %w", err)
	}
	if strings.TrimSpace(credential) == "" {
		return nil, fmt.Errorf("playback credential is required")
	}
	result := map[string]any{}
	err = a.client.requestWithBaseHeaders(ctx, base.Scheme+"://"+base.Host, http.MethodPost,
		"/playback"+path, body, map[string]string{"Authorization": "Bearer " + credential}, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *AssetVideoClient) RenewSession(ctx context.Context, credential string) (map[string]any, error) {
	return a.playback(
		ctx,
		"/sessions/renew",
		credential,
		map[string]any{},
	)
}

func (a *AssetVideoClient) Events(
	ctx context.Context,
	credential string,
	events []AssetPlaybackEvent,
) (map[string]any, error) {
	return a.playback(
		ctx,
		"/events",
		credential,
		map[string]any{"events": events},
	)
}

type AssetPlaybackEvent struct {
	Sequence int64   `json:"sequence"`
	Kind     string  `json:"kind"`
	Value    float64 `json:"value"`
}

func assetMediaObject[T any](
	a *assetMediaClient,
	ctx context.Context,
	method,
	path string,
	body any,
	headers map[string]string,
) (*T, error) {
	var result T
	if err := a.call(ctx, method, path, body, headers, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
