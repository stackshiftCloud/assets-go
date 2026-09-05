package stackshift

import (
	"context"
	"net/url"
)

func (a *AssetDAMClient) Publications(ctx context.Context, query url.Values) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		mediaQuery("/assets/publications", query),
		nil,
		nil,
	)
}

func (a *AssetDAMClient) PublicationRenditions(ctx context.Context, assetID string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/"+url.PathEscape(assetID)+"/publication-renditions",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) CreatePublication(ctx context.Context, input map[string]any) (*AssetPublication, error) {
	return assetMediaObject[AssetPublication](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/publications",
		input,
		nil,
	)
}

func (a *AssetDAMClient) UpdatePublication(
	ctx context.Context,
	id string,
	revision int64,
	input map[string]any,
) (*AssetPublication, error) {
	headers := mediaRevision(revision)
	return assetMediaObject[AssetPublication](
		&a.assetMediaClient,
		ctx,
		"PUT",
		"/assets/publications/"+url.PathEscape(id),
		input,
		headers,
	)
}

func (a *AssetDAMClient) TransitionPublication(
	ctx context.Context,
	id string,
	revision int64,
	action string,
	note string,
) (*AssetPublication, error) {
	headers := mediaRevision(revision)
	body := map[string]any{}
	body["note"] = note
	return assetMediaObject[AssetPublication](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/publications/"+url.PathEscape(id)+"/"+url.PathEscape(action),
		body,
		headers,
	)
}

func (a *AssetDAMClient) ResolvePublication(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/publications/"+url.PathEscape(id)+"/resolve",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) PreviewPublication(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/publications/"+url.PathEscape(id)+"/preview",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) ResolvePublicationAR(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/publications/"+url.PathEscape(id)+"/ar",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) CreatePickerCapability(ctx context.Context, input map[string]any) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/picker-capabilities",
		input,
		nil,
	)
}

func (a *AssetDAMClient) RevokePickerCapability(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"DELETE",
		"/assets/picker-capabilities/"+url.PathEscape(id),
		nil,
		nil,
	)
}
