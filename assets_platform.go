package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type AssetConnectionsClient struct{ client *Client }
type AssetStorageClient struct{ client *Client }
type AssetMigrationsClient struct{ client *Client }

// ForSpace selects an authorized Assets space for platform operations.
func (a *AssetsClient) ForSpace(spaceID string) *AssetsClient {
	scopedClient := *a.client
	scopedClient.assetSpaceID = spaceID
	return &AssetsClient{client: &scopedClient,
		Connections: &AssetConnectionsClient{client: &scopedClient},
		Storage:     &AssetStorageClient{client: &scopedClient},
		Migrations:  &AssetMigrationsClient{client: &scopedClient}}
}

func assetPlatformRequest(ctx context.Context, client *Client, method, path string, body any, headers map[string]string) (map[string]any, error) {
	var result map[string]any
	if err := client.requestWithHeaders(ctx, method, path, body, headers, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (a *AssetConnectionsClient) List(ctx context.Context) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodGet, "/assets/connections", nil, nil)
}

func (a *AssetConnectionsClient) Get(ctx context.Context, id string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodGet, "/assets/connections/"+url.PathEscape(id), nil, nil)
}

func (a *AssetConnectionsClient) Create(ctx context.Context, input map[string]any) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/connections", input, nil)
}

func (a *AssetConnectionsClient) Update(ctx context.Context, id string, input map[string]any) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPatch, "/assets/connections/"+url.PathEscape(id), input, nil)
}

func (a *AssetConnectionsClient) Verify(ctx context.Context, id string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/connections/"+url.PathEscape(id)+"/verify", nil, nil)
}

func (a *AssetConnectionsClient) Delete(ctx context.Context, id string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodDelete, "/assets/connections/"+url.PathEscape(id), nil, nil)
}

func (a *AssetStorageClient) Configure(ctx context.Context, bucketID string, revision int64, input map[string]any) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPut, "/assets/buckets/"+url.PathEscape(bucketID)+"/storage", input,
		map[string]string{"If-Match": "\"" + strconv.FormatInt(revision, 10) + "\""})
}

func (a *AssetStorageClient) ListS2Buckets(ctx context.Context) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodGet, "/assets/s2-buckets", nil, nil)
}

func (a *AssetStorageClient) CreateS2Bucket(ctx context.Context, input map[string]any) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/s2-buckets", input, nil)
}

func (a *AssetStorageClient) Relocate(ctx context.Context, bucketID string, input map[string]any) (map[string]any, error) {
	if input == nil {
		input = map[string]any{}
	}
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/buckets/"+url.PathEscape(bucketID)+"/relocations", input, nil)
}

func (a *AssetMigrationsClient) Start(ctx context.Context, input map[string]any) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/migrations", input, nil)
}

func (a *AssetMigrationsClient) List(ctx context.Context, limit int) (map[string]any, error) {
	path := "/assets/migrations"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	return assetPlatformRequest(ctx, a.client, http.MethodGet, path, nil, nil)
}

func (a *AssetMigrationsClient) Get(ctx context.Context, id string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodGet, "/assets/migrations/"+url.PathEscape(id), nil, nil)
}

func (a *AssetMigrationsClient) Cancel(ctx context.Context, id string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/migrations/"+url.PathEscape(id)+"/cancel", nil, nil)
}

func (a *AssetMigrationsClient) RetryItem(ctx context.Context, migrationID, itemID string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodPost, "/assets/migrations/"+url.PathEscape(migrationID)+"/items/"+url.PathEscape(itemID)+"/retry", nil, nil)
}

func (a *AssetMigrationsClient) Report(ctx context.Context, id string) (map[string]any, error) {
	return assetPlatformRequest(ctx, a.client, http.MethodGet, "/assets/migrations/"+url.PathEscape(id)+"/report?format=json", nil, nil)
}
