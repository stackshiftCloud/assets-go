package stackshift

import (
	"context"
	"net/url"
)

func (a *AssetVideoClient) Workspace(ctx context.Context, assetID string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/"+url.PathEscape(assetID)+"/video",
		nil,
		nil,
	)
}

func (a *AssetVideoClient) Process(ctx context.Context, assetID string, versionID string) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/versions/"+url.PathEscape(versionID)+"/video",
		nil,
		nil,
	)
}

func (a *AssetVideoClient) CreateSession(ctx context.Context, assetID string) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/playback-sessions",
		nil,
		nil,
	)
}

func (a *AssetVideoClient) RevokeSession(
	ctx context.Context,
	assetID string,
	sessionID string,
) (map[string]any, error) {
	return a.object(
		ctx,
		"DELETE",
		"/assets/"+url.PathEscape(assetID)+"/playback-sessions/"+url.PathEscape(sessionID),
		nil,
		nil,
	)
}

func (a *AssetVideoClient) Captions(ctx context.Context, assetID string, versionID string) ([]map[string]any, error) {
	result := []map[string]any{}
	err := a.call(ctx, "GET", "/assets/"+url.PathEscape(assetID)+"/versions/"+url.PathEscape(versionID)+"/video/captions", nil, nil, &result)
	return result, err
}

func (a *AssetVideoClient) AddCaption(
	ctx context.Context,
	assetID string,
	versionID string,
	input map[string]any,
) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/versions/"+url.PathEscape(versionID)+"/video/captions",
		input,
		nil,
	)
}

func (a *AssetVideoClient) RemoveCaption(
	ctx context.Context,
	assetID string,
	versionID string,
	captionID string,
) (map[string]any, error) {
	return a.object(
		ctx,
		"DELETE",
		"/assets/"+url.PathEscape(assetID)+"/versions/"+url.PathEscape(versionID)+"/video/captions/"+url.PathEscape(captionID),
		nil,
		nil,
	)
}

func (a *AssetVideoClient) Review(
	ctx context.Context,
	assetID string,
	versionID string,
	decision string,
	reason string,
) (map[string]any, error) {
	body := map[string]any{}
	body["decision"] = decision
	body["reason"] = reason
	return a.object(
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/versions/"+url.PathEscape(versionID)+"/video/review",
		body,
		nil,
	)
}

func (a *AssetVideoClient) Analytics(ctx context.Context, assetID string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/"+url.PathEscape(assetID)+"/video/analytics",
		nil,
		nil,
	)
}
