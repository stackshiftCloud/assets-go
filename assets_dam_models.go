package stackshift

import (
	"context"
	"net/url"
)

func (a *AssetDAMClient) Models(ctx context.Context, assetID string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/"+url.PathEscape(assetID)+"/models",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) EstimateRender(
	ctx context.Context,
	assetID string,
	revision int64,
	input map[string]any,
) (*AssetRenderEstimate, error) {
	headers := mediaRevision(revision)
	body := map[string]any{}
	for key, value := range input {
		body[key] = value
	}
	body["operation"] = "estimate"
	return assetMediaObject[AssetRenderEstimate](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/models",
		body,
		headers,
	)
}

func (a *AssetDAMClient) RenderProduct(
	ctx context.Context,
	assetID string,
	revision int64,
	idempotencyKey string,
	input map[string]any,
	maximumUnits int64,
) (*AssetRenderSubmission, error) {
	headers := mediaRevision(revision)
	headers["Idempotency-Key"] = idempotencyKey
	body := map[string]any{}
	for key, value := range input {
		body[key] = value
	}
	body["operation"] = "render"
	body["maximum_units"] = maximumUnits
	return assetMediaObject[AssetRenderSubmission](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/models",
		body,
		headers,
	)
}

func (a *AssetDAMClient) ProcessModel(
	ctx context.Context,
	assetID string,
	revision int64,
	idempotencyKey string,
) (map[string]any, error) {
	headers := mediaRevision(revision)
	headers["Idempotency-Key"] = idempotencyKey
	return a.object(
		ctx,
		"POST",
		"/assets/"+url.PathEscape(assetID)+"/models",
		nil,
		headers,
	)
}

func (a *AssetDAMClient) Galleries(ctx context.Context, query url.Values) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		mediaQuery("/assets/galleries", query),
		nil,
		nil,
	)
}

func (a *AssetDAMClient) CreateGallery(ctx context.Context, input map[string]any) (*AssetGallery, error) {
	return assetMediaObject[AssetGallery](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/galleries",
		input,
		nil,
	)
}

func (a *AssetDAMClient) UpdateGallery(
	ctx context.Context,
	id string,
	revision int64,
	input map[string]any,
) (*AssetGallery, error) {
	headers := mediaRevision(revision)
	return assetMediaObject[AssetGallery](
		&a.assetMediaClient,
		ctx,
		"PUT",
		"/assets/galleries/"+url.PathEscape(id),
		input,
		headers,
	)
}

func (a *AssetDAMClient) TransitionGallery(
	ctx context.Context,
	id string,
	revision int64,
	action string,
) (*AssetGallery, error) {
	headers := mediaRevision(revision)
	return assetMediaObject[AssetGallery](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/galleries/"+url.PathEscape(id)+"/"+url.PathEscape(action),
		nil,
		headers,
	)
}

func (a *AssetDAMClient) ResolveGallery(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/galleries/"+url.PathEscape(id)+"/resolve",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) PreviewGallery(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/galleries/"+url.PathEscape(id)+"/resolve?preview=true",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) ResolveGalleryAR(
	ctx context.Context,
	id string,
	publicationID string,
) (map[string]any, error) {
	body := map[string]any{}
	body["publication_id"] = publicationID
	return a.object(
		ctx,
		"POST",
		"/assets/galleries/"+url.PathEscape(id)+"/ar",
		body,
		nil,
	)
}
