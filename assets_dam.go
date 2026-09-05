package stackshift

import (
	"context"
	"net/url"
)

func (a *AssetDAMClient) Spaces(ctx context.Context) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/spaces",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) Collaborators(ctx context.Context) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/collaborators",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) Invite(ctx context.Context, input map[string]any) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/collaborators",
		input,
		nil,
	)
}

func (a *AssetDAMClient) AcceptInvitation(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"POST",
		"/assets/invitations/"+url.PathEscape(id)+"/accept",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) UpdateCollaborator(
	ctx context.Context,
	id string,
	revision int64,
	actions []string,
) (map[string]any, error) {
	headers := mediaRevision(revision)
	body := map[string]any{}
	body["actions"] = actions
	return a.object(
		ctx,
		"PUT",
		"/assets/collaborators/"+url.PathEscape(id),
		body,
		headers,
	)
}

func (a *AssetDAMClient) RevokeCollaborator(ctx context.Context, id string, revision int64) (map[string]any, error) {
	headers := mediaRevision(revision)
	return a.object(
		ctx,
		"DELETE",
		"/assets/collaborators/"+url.PathEscape(id),
		nil,
		headers,
	)
}

func (a *AssetDAMClient) MetadataSchemas(ctx context.Context) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/metadata-schemas",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) MetadataEditor(ctx context.Context, assetID string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/"+url.PathEscape(assetID)+"/metadata-editor",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) CreateMetadataSchema(ctx context.Context, input map[string]any) (*AssetMetadataSchema, error) {
	return assetMediaObject[AssetMetadataSchema](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/metadata-schemas",
		input,
		nil,
	)
}

func (a *AssetDAMClient) UpdateMetadataSchema(
	ctx context.Context,
	id string,
	revision int64,
	input map[string]any,
) (*AssetMetadataSchema, error) {
	headers := mediaRevision(revision)
	return assetMediaObject[AssetMetadataSchema](
		&a.assetMediaClient,
		ctx,
		"PUT",
		"/assets/metadata-schemas/"+url.PathEscape(id),
		input,
		headers,
	)
}

func (a *AssetDAMClient) PublishMetadataSchema(
	ctx context.Context,
	id string,
	revision int64,
) (*AssetMetadataSchema, error) {
	headers := mediaRevision(revision)
	return assetMediaObject[AssetMetadataSchema](
		&a.assetMediaClient,
		ctx,
		"POST",
		"/assets/metadata-schemas/"+url.PathEscape(id)+"/publish",
		nil,
		headers,
	)
}

func (a *AssetDAMClient) BucketGovernance(ctx context.Context, id string) (map[string]any, error) {
	return a.object(
		ctx,
		"GET",
		"/assets/buckets/"+url.PathEscape(id)+"/governance",
		nil,
		nil,
	)
}

func (a *AssetDAMClient) SaveBucketGovernance(
	ctx context.Context,
	id string,
	input map[string]any,
) (map[string]any, error) {
	return a.object(
		ctx,
		"PUT",
		"/assets/buckets/"+url.PathEscape(id)+"/governance",
		input,
		nil,
	)
}

func (a *AssetDAMClient) MigrateBucketGovernance(
	ctx context.Context,
	id string,
	revision int64,
) (map[string]any, error) {
	headers := mediaRevision(revision)
	return a.object(
		ctx,
		"POST",
		"/assets/buckets/"+url.PathEscape(id)+"/governance/migrate",
		nil,
		headers,
	)
}
