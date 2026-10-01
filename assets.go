package stackshift

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type AssetsClient struct {
	client      *Client
	Connections *AssetConnectionsClient
	Storage     *AssetStorageClient
	Migrations  *AssetMigrationsClient
}

var BuiltInAssetTransformPresets = []AssetTransformPreset{
	assetPreset("web-thumbnail", "responsive-web", "Compact grid thumbnail for libraries and dashboards.", 320, 180, "fill", "webp", "78"),
	assetPreset("web-card", "responsive-web", "Balanced card image for product and content listings.", 640, 360, "fill", "webp", "82"),
	assetPreset("feature-image", "responsive-web", "Wide content feature image with enough detail for retina screens.", 1200, 675, "fill", "webp", "84"),
	assetPreset("hero", "responsive-web", "Large hero image for landing pages and page headers.", 1600, 900, "fill", "webp", "84"),
	assetPreset("full-width-banner", "responsive-web", "Extra-wide banner for app shells and marketing sections.", 1920, 720, "fill", "webp", "82"),
	assetPreset("open-graph", "social", "Share preview image for Open Graph cards.", 1200, 630, "fill", "jpeg", "85"),
	assetPreset("square-post", "social", "Square image for social feeds.", 1080, 1080, "fill", "webp", "84"),
	assetPreset("story", "social", "Vertical story format for mobile social placements.", 1080, 1920, "fill", "webp", "82"),
	assetPreset("x-twitter-card", "social", "Landscape card optimized for X and Twitter previews.", 1200, 675, "fill", "jpeg", "84"),
	assetPreset("linkedin-share", "social", "LinkedIn share preview dimensions.", 1200, 627, "fill", "jpeg", "84"),
	assetPreset("product-square", "product", "Square product image with clean containment.", 1024, 1024, "fit", "webp", "86"),
	assetPreset("product-card", "product", "Product listing image for commerce grids.", 640, 640, "fit", "webp", "84"),
	assetPreset("product-zoom", "product", "High-detail product image for zoom interactions.", 2000, 2000, "fit", "jpeg", "90"),
	assetPreset("gallery-thumb", "product", "Compact gallery thumbnail for product detail pages.", 400, 400, "fill", "webp", "82"),
	assetPreset("avatar-small", "avatar-profile", "Small circular avatar source for dense UI.", 128, 128, "fill", "webp", "82"),
	assetPreset("avatar-retina", "avatar-profile", "Retina avatar source for account and profile UI.", 256, 256, "fill", "webp", "84"),
	assetPreset("profile-cover", "avatar-profile", "Wide profile cover image.", 1500, 500, "fill", "webp", "82"),
	assetPreset("webp-auto", "performance", "Auto-sized WebP delivery with browser-friendly compression.", 1280, 0, "fit", "webp", "78"),
	assetPreset("avif-compressed", "performance", "Highly compressed AVIF for modern browsers.", 1280, 0, "fit", "avif", "62"),
	assetPreset("lqip", "performance", "Tiny low-quality placeholder for progressive image loading.", 48, 0, "fit", "jpeg", "35"),
	assetPreset("blur-placeholder", "performance", "Small placeholder candidate for blurred loading states.", 32, 32, "fill", "jpeg", "30"),
	assetPreset("email-safe-600", "email-docs", "Email-safe content image with conservative width.", 600, 0, "fit", "jpeg", "82"),
	assetPreset("markdown-image", "email-docs", "Documentation image that fits most reading columns.", 800, 0, "fit", "webp", "82"),
	assetPreset("docs-screenshot", "email-docs", "Crisp screenshot size for guides and changelogs.", 1440, 900, "fit", "webp", "84"),
}

func AssetTransformPresetByName(name string) (AssetTransformPreset, bool) {
	for _, preset := range BuiltInAssetTransformPresets {
		if preset.Name == name {
			return preset, true
		}
	}
	return AssetTransformPreset{}, false
}

func AssetTransformSpec(options AssetTransformOptions) string {
	return transformSpec(options)
}

func (a *AssetsClient) Upload(ctx context.Context, upload AssetUpload) (*Asset, error) {
	body, contentType, err := assetMultipart(upload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.client.baseURL+"/assets/upload", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.client.apiKey)
	req.Header.Set("User-Agent", "StackShift-Go/0.1")
	req.Header.Set("Content-Type", contentType)
	var asset Asset
	if err := a.client.do(req, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (a *AssetsClient) List(ctx context.Context, options AssetListOptions) (*AssetListResult, error) {
	query := url.Values{}
	setQuery(query, "bucket", options.Bucket)
	setQuery(query, "prefix", options.Prefix)
	setQuery(query, "visibility", options.Visibility)
	setQuery(query, "query", options.Query)
	setQuery(query, "tag", options.Tag)
	setQuery(query, "mime_type", options.MimeType)
	setQuery(query, "type", options.Type)
	setQuery(query, "folder", options.Folder)
	setQuery(query, "checksum", options.Checksum)
	setQuery(query, "status", options.Status)
	setQuery(query, "scan_status", options.ScanStatus)
	setQuery(query, "replication_status", options.ReplicationStatus)
	setQuery(query, "moderation_status", options.ModerationStatus)
	setQuery(query, "quarantine_status", options.QuarantineStatus)
	setQuery(query, "sort", options.Sort)
	setQuery(query, "direction", options.Direction)
	setQuery(query, "cursor", options.Cursor)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	path := "/assets"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var result AssetListResult
	if err := a.client.request(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *AssetsClient) Get(ctx context.Context, assetID string) (*Asset, error) {
	var asset Asset
	if err := a.client.request(ctx, http.MethodGet, "/assets/"+url.PathEscape(assetID), nil, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (a *AssetsClient) Delete(ctx context.Context, assetID string, revision int64) (*AssetJobSubmission, error) {
	return a.DeleteWithIdempotency(ctx, assetID, revision, "")
}

func (a *AssetsClient) DeleteWithIdempotency(ctx context.Context, assetID string, revision int64, idempotencyKey string) (*AssetJobSubmission, error) {
	var result AssetJobSubmission
	headers := map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}
	if idempotencyKey != "" {
		headers["Idempotency-Key"] = idempotencyKey
	}
	err := a.client.requestWithHeaders(ctx, http.MethodDelete, "/assets/"+url.PathEscape(assetID), nil, headers, &result)
	return &result, err
}

func (a *AssetsClient) Replace(ctx context.Context, assetID string, revision int64, upload AssetUpload) (*Asset, error) {
	body, contentType, err := assetMultipart(upload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, a.client.baseURL+"/assets/"+url.PathEscape(assetID)+"/replace", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.client.apiKey)
	req.Header.Set("User-Agent", "StackShift-Go/0.1")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("If-Match", fmt.Sprintf("\"%d\"", revision))
	var asset Asset
	if err := a.client.do(req, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (a *AssetsClient) SignedURL(ctx context.Context, assetID string, options SignedURLOptions) (*SignedURL, error) {
	body := map[string]any{"expires_in": options.ExpiresIn}
	if options.MaxDownloads != nil {
		body["max_downloads"] = *options.MaxDownloads
	}
	var signed SignedURL
	if err := a.client.request(ctx, http.MethodPost, "/assets/"+url.PathEscape(assetID)+"/signed-url", body, &signed); err != nil {
		return nil, err
	}
	return &signed, nil
}

func (a *AssetsClient) SignedUploadURL(ctx context.Context, options SignedUploadURLOptions) (*SignedUploadURL, error) {
	return a.CreateUploadSession(ctx, options)
}

func (a *AssetsClient) CreateUploadSession(ctx context.Context, options SignedUploadURLOptions) (*SignedUploadURL, error) {
	var signed SignedUploadURL
	headers := map[string]string{}
	if options.IdempotencyKey != "" {
		headers["Idempotency-Key"] = options.IdempotencyKey
	}
	err := a.client.requestWithHeaders(ctx, http.MethodPost, "/assets/upload-sessions", map[string]any{
		"mode": options.Mode, "file_name": options.FileName, "file_size": options.FileSize, "mime_type": options.MimeType,
		"checksum_sha256": options.ChecksumSHA256, "part_size": options.PartSize,
		"bucket": options.Bucket, "key": options.Key, "folder": options.Folder, "visibility": options.Visibility,
		"cache_control": options.CacheControl, "metadata": options.Metadata, "expires_in": options.ExpiresIn, "max_bytes": options.MaxBytes,
	}, headers, &signed)
	return &signed, err
}

func (a *AssetsClient) CreateChunkedUploadSession(ctx context.Context, options SignedUploadURLOptions) (*SignedUploadURL, error) {
	options.Mode = "chunked"
	return a.CreateUploadSession(ctx, options)
}

func (a *AssetsClient) UploadChunk(ctx context.Context, uploadURLOrToken string, partNumber int, body io.Reader) error {
	raw, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	target := strings.TrimRight(uploadURLOrToken, "/")
	if !strings.HasPrefix(target, "http") {
		target = strings.TrimSuffix(a.client.baseURL, "/api/v1") + "/uploads/" + url.PathEscape(uploadURLOrToken)
	}
	target += "/parts/" + strconv.Itoa(partNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "StackShift-Go/0.1")
	req.Header.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(raw)))
	return a.client.do(req, nil)
}

func (a *AssetsClient) ResumeUploadSession(ctx context.Context, sessionID string) (*SignedUploadURL, error) {
	var signed SignedUploadURL
	err := a.client.request(ctx, http.MethodGet, "/assets/upload-sessions/"+url.PathEscape(sessionID), nil, &signed)
	return &signed, err
}

func (a *AssetsClient) CompleteUploadSession(ctx context.Context, sessionID string) (*Asset, error) {
	var asset Asset
	err := a.client.request(ctx, http.MethodPost, "/assets/upload-sessions/"+url.PathEscape(sessionID)+"/complete", nil, &asset)
	return &asset, err
}

func (a *AssetsClient) CancelUploadSession(ctx context.Context, sessionID string) error {
	return a.client.request(ctx, http.MethodDelete, "/assets/upload-sessions/"+url.PathEscape(sessionID), nil, nil)
}

func (a *AssetsClient) ListBuckets(ctx context.Context) ([]AssetBucket, error) {
	var result struct {
		Buckets []AssetBucket `json:"buckets"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/buckets", nil, &result)
	return result.Buckets, err
}

func (a *AssetsClient) CreateBucket(ctx context.Context, input AssetBucketInput) (*AssetBucket, error) {
	var bucket AssetBucket
	err := a.client.request(ctx, http.MethodPost, "/assets/buckets", input, &bucket)
	return &bucket, err
}

func (a *AssetsClient) UpdateBucket(ctx context.Context, bucketID string, revision int64, input AssetBucketInput) (*AssetBucket, error) {
	var bucket AssetBucket
	err := a.client.requestWithHeaders(ctx, http.MethodPut, "/assets/buckets/"+url.PathEscape(bucketID), input,
		map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}, &bucket)
	return &bucket, err
}

func (a *AssetsClient) DeleteBucket(ctx context.Context, bucketID string, revision int64) error {
	return a.client.requestWithHeaders(ctx, http.MethodDelete, "/assets/buckets/"+url.PathEscape(bucketID), nil,
		map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}, nil)
}

func (a *AssetsClient) SetTags(ctx context.Context, assetID string, revision int64, tags []string) error {
	return a.client.requestWithHeaders(ctx, http.MethodPut, "/assets/"+url.PathEscape(assetID)+"/tags", map[string]any{"tags": tags},
		map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}, nil)
}

func (a *AssetsClient) PatchMetadata(ctx context.Context, assetID string, revision int64, metadata map[string]any) error {
	return a.client.requestWithHeaders(ctx, http.MethodPatch, "/assets/"+url.PathEscape(assetID)+"/metadata", map[string]any{"metadata": metadata},
		map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}, nil)
}

func (a *AssetsClient) Patch(ctx context.Context, assetID string, revision int64, input AssetPatchInput) (*Asset, error) {
	var asset Asset
	err := a.client.requestWithHeaders(ctx, http.MethodPatch, "/assets/"+url.PathEscape(assetID), input,
		map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}, &asset)
	return &asset, err
}

func (a *AssetsClient) URL(assetID string, options AssetTransformOptions) string {
	return a.client.cdnBaseURL + "/assets/" + url.PathEscape(assetID) + "/tr/" + url.PathEscape(transformSpec(options))
}

func (a *AssetsClient) NamedURL(assetID, name string) string {
	return a.client.cdnBaseURL + "/t/" + url.PathEscape(name) + "/assets/" + url.PathEscape(assetID)
}

func (a *AssetsClient) VideoURL(assetID, kind, profile string) string {
	return a.client.cdnBaseURL + "/assets/" + url.PathEscape(assetID) + "/video/" + url.PathEscape(kind) + "/" + url.PathEscape(profile)
}

func (a *AssetsClient) PrivateVideoURL(assetID, kind, profile, token string) string {
	base := a.client.cdnBaseURL + "/private/assets/" + url.PathEscape(assetID)
	videoPath := "/video/" + url.PathEscape(kind) + "/" + url.PathEscape(profile)
	return base + videoPath + "?token=" + url.QueryEscape(token)
}

func (a *AssetsClient) VersionURL(assetID, versionID string) string {
	return a.client.cdnBaseURL + "/assets/" + url.PathEscape(assetID) + "/versions/" + url.PathEscape(versionID)
}

func (a *AssetsClient) BranchURL(assetID, branch string) string {
	return a.client.cdnBaseURL + "/assets/" + url.PathEscape(assetID) + "/branches/" + url.PathEscape(branch)
}

func (a *AssetsClient) AI(ctx context.Context, assetID string) (*AssetAIInfo, error) {
	var info AssetAIInfo
	err := a.client.request(ctx, http.MethodGet, "/assets/"+url.PathEscape(assetID)+"/ai", nil, &info)
	return &info, err
}

func (a *AssetsClient) RequestAIAnalyze(ctx context.Context, assetID string) (*AssetJobSubmission, error) {
	return a.requestAssetJob(ctx, "/assets/"+url.PathEscape(assetID)+"/ai/analyze")
}

func (a *AssetsClient) RequestAIAnalyzeWithIdempotency(ctx context.Context, assetID, key string) (*AssetJobSubmission, error) {
	return a.requestAssetJobWithIdempotency(ctx, "/assets/"+url.PathEscape(assetID)+"/ai/analyze", key)
}

func (a *AssetsClient) RequestModeration(ctx context.Context, assetID string) (*AssetJobSubmission, error) {
	return a.requestAssetJob(ctx, "/assets/"+url.PathEscape(assetID)+"/moderation")
}

func (a *AssetsClient) RequestModerationWithIdempotency(ctx context.Context, assetID, key string) (*AssetJobSubmission, error) {
	return a.requestAssetJobWithIdempotency(ctx, "/assets/"+url.PathEscape(assetID)+"/moderation", key)
}

func (a *AssetsClient) Transcript(ctx context.Context, assetID string) (*AssetTranscript, error) {
	var transcript AssetTranscript
	err := a.client.request(ctx, http.MethodGet, "/assets/"+url.PathEscape(assetID)+"/transcript", nil, &transcript)
	return &transcript, err
}

func (a *AssetsClient) RequestTranscript(ctx context.Context, assetID string) (*AssetJobSubmission, error) {
	return a.requestAssetJob(ctx, "/assets/"+url.PathEscape(assetID)+"/transcript")
}

func (a *AssetsClient) RequestTranscriptWithIdempotency(ctx context.Context, assetID, key string) (*AssetJobSubmission, error) {
	return a.requestAssetJobWithIdempotency(ctx, "/assets/"+url.PathEscape(assetID)+"/transcript", key)
}

func (a *AssetsClient) RequestSmartCrop(ctx context.Context, assetID string) (*AssetJobSubmission, error) {
	return a.requestAssetJob(ctx, "/assets/"+url.PathEscape(assetID)+"/smart-crop")
}

func (a *AssetsClient) RequestSmartCropWithIdempotency(ctx context.Context, assetID, key string) (*AssetJobSubmission, error) {
	return a.requestAssetJobWithIdempotency(ctx, "/assets/"+url.PathEscape(assetID)+"/smart-crop", key)
}

func (a *AssetsClient) RequestBackgroundRemoval(ctx context.Context, assetID string) (*AssetJobSubmission, error) {
	return a.requestAssetJob(ctx, "/assets/"+url.PathEscape(assetID)+"/background-remove")
}

func (a *AssetsClient) RequestBackgroundRemovalWithIdempotency(ctx context.Context, assetID, key string) (*AssetJobSubmission, error) {
	return a.requestAssetJobWithIdempotency(ctx, "/assets/"+url.PathEscape(assetID)+"/background-remove", key)
}

func (a *AssetsClient) Versions(ctx context.Context, assetID string) ([]AssetVersion, error) {
	var result struct {
		Versions []AssetVersion `json:"versions"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/"+url.PathEscape(assetID)+"/versions", nil, &result)
	return result.Versions, err
}

func (a *AssetsClient) CreateBranch(ctx context.Context, assetID, name, fromVersionID string) error {
	body := map[string]any{"name": name}
	if fromVersionID != "" {
		body["from_version_id"] = fromVersionID
	}
	return a.client.request(ctx, http.MethodPost, "/assets/"+url.PathEscape(assetID)+"/branches", body, nil)
}

func (a *AssetsClient) RestoreVersion(ctx context.Context, assetID, versionID string) error {
	return a.client.request(ctx, http.MethodPost, "/assets/"+url.PathEscape(assetID)+"/versions/"+url.PathEscape(versionID)+"/restore", nil, nil)
}

func (a *AssetsClient) PromoteBranch(ctx context.Context, assetID, branch string) error {
	return a.client.request(ctx, http.MethodPost, "/assets/"+url.PathEscape(assetID)+"/branches/"+url.PathEscape(branch)+"/promote", nil, nil)
}

func (a *AssetsClient) Collections(ctx context.Context) ([]AssetCollection, error) {
	var result struct {
		Collections []AssetCollection `json:"collections"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/collections", nil, &result)
	return result.Collections, err
}

func (a *AssetsClient) CreateCollection(ctx context.Context, input AssetCollection) (*AssetCollection, error) {
	var collection AssetCollection
	err := a.client.request(ctx, http.MethodPost, "/assets/collections", input, &collection)
	return &collection, err
}

func (a *AssetsClient) SavedSearches(ctx context.Context) ([]AssetSavedSearch, error) {
	var result struct {
		SavedSearches []AssetSavedSearch `json:"saved_searches"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/saved-searches", nil, &result)
	return result.SavedSearches, err
}

func (a *AssetsClient) SaveSearch(ctx context.Context, input AssetSavedSearch) (*AssetSavedSearch, error) {
	var search AssetSavedSearch
	err := a.client.request(ctx, http.MethodPost, "/assets/saved-searches", input, &search)
	return &search, err
}

func (a *AssetsClient) GetPolicy(ctx context.Context) (*AssetContentPolicy, error) {
	var policy AssetContentPolicy
	err := a.client.request(ctx, http.MethodGet, "/assets/policy", nil, &policy)
	return &policy, err
}

func (a *AssetsClient) UpdatePolicy(ctx context.Context, policy AssetContentPolicy) (*AssetContentPolicy, error) {
	var updated AssetContentPolicy
	err := a.client.request(ctx, http.MethodPut, "/assets/policy", policy, &updated)
	return &updated, err
}

func (a *AssetsClient) CreateTransformation(ctx context.Context, name string, options AssetTransformOptions) (*AssetTransformation, error) {
	var transformation AssetTransformation
	err := a.client.request(ctx, http.MethodPost, "/assets/transformations", map[string]any{
		"name": name, "width": options.Width, "height": options.Height, "crop": options.Crop,
		"format": options.Format, "quality": options.Quality,
	}, &transformation)
	return &transformation, err
}

func (a *AssetsClient) ListTransformations(ctx context.Context) ([]AssetTransformation, error) {
	var result struct {
		Transformations []AssetTransformation `json:"transformations"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/transformations", nil, &result)
	return result.Transformations, err
}

func (a *AssetsClient) DeleteTransformation(ctx context.Context, name string) error {
	return a.client.request(ctx, http.MethodDelete, "/assets/transformations/"+url.PathEscape(name), nil, nil)
}

func (a *AssetsClient) SignedTransformURL(ctx context.Context, assetID string, options AssetTransformOptions) (*SignedURL, error) {
	var signed SignedURL
	err := a.client.request(ctx, http.MethodPost, "/assets/"+url.PathEscape(assetID)+"/transform-url", map[string]any{
		"width": options.Width, "height": options.Height, "crop": options.Crop,
		"format": options.Format, "quality": options.Quality, "expires_in": options.ExpiresIn,
	}, &signed)
	return &signed, err
}

func (a *AssetsClient) requestAssetJob(ctx context.Context, path string) (*AssetJobSubmission, error) {
	return a.requestAssetJobWithIdempotency(ctx, path, "")
}

func (a *AssetsClient) requestAssetJobWithIdempotency(ctx context.Context, path, key string) (*AssetJobSubmission, error) {
	var job AssetJobSubmission
	headers := map[string]string{}
	if key != "" {
		headers["Idempotency-Key"] = key
	}
	err := a.client.requestWithHeaders(ctx, http.MethodPost, path, nil, headers, &job)
	return &job, err
}

func (a *AssetsClient) GetAIConfig(ctx context.Context) (*AssetAIAccountConfig, error) {
	var result AssetAIAccountConfig
	err := a.client.request(ctx, http.MethodGet, "/assets/ai/config", nil, &result)
	return &result, err
}

func (a *AssetsClient) UpdateAIConfig(ctx context.Context, revision int64, policy AssetAIPolicy) (*AssetAIPolicy, error) {
	var result AssetAIPolicy
	err := a.client.requestWithHeaders(ctx, http.MethodPut, "/assets/ai/config", policy,
		map[string]string{"If-Match": fmt.Sprintf("\"%d\"", revision)}, &result)
	return &result, err
}

func (a *AssetsClient) Job(ctx context.Context, jobID string) (*AssetJob, error) {
	var job AssetJob
	err := a.client.request(ctx, http.MethodGet, "/assets/jobs/"+url.PathEscape(jobID), nil, &job)
	return &job, err
}

func (a *AssetsClient) CancelJob(ctx context.Context, jobID string) (*AssetJob, error) {
	var job AssetJob
	err := a.client.request(ctx, http.MethodPost, "/assets/jobs/"+url.PathEscape(jobID)+"/cancel", nil, &job)
	return &job, err
}

func (a *AssetsClient) CreateLifecycleRule(ctx context.Context, input map[string]any) (*AssetLifecycleRule, error) {
	var rule AssetLifecycleRule
	err := a.client.request(ctx, http.MethodPost, "/assets/lifecycle-rules", input, &rule)
	return &rule, err
}

func (a *AssetsClient) ListLifecycleRules(ctx context.Context) ([]AssetLifecycleRule, error) {
	var result struct {
		LifecycleRules []AssetLifecycleRule `json:"lifecycle_rules"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/lifecycle-rules", nil, &result)
	return result.LifecycleRules, err
}

func (a *AssetsClient) DeleteLifecycleRule(ctx context.Context, ruleID string) error {
	return a.client.request(ctx, http.MethodDelete, "/assets/lifecycle-rules/"+url.PathEscape(ruleID), nil, nil)
}

func (a *AssetsClient) CreateDomain(ctx context.Context, domain string) (*AssetCustomDomain, error) {
	var result AssetCustomDomain
	err := a.client.request(ctx, http.MethodPost, "/assets/domains", map[string]string{"domain": domain}, &result)
	return &result, err
}

func (a *AssetsClient) ListDomains(ctx context.Context) ([]AssetCustomDomain, error) {
	var result struct {
		Domains []AssetCustomDomain `json:"domains"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/domains", nil, &result)
	return result.Domains, err
}

func (a *AssetsClient) VerifyDomain(ctx context.Context, domainID string) (*AssetCustomDomain, error) {
	var result AssetCustomDomain
	err := a.client.request(ctx, http.MethodPost, "/assets/domains/"+url.PathEscape(domainID)+"/verify", nil, &result)
	return &result, err
}

func (a *AssetsClient) DeleteDomain(ctx context.Context, domainID string) error {
	return a.client.request(ctx, http.MethodDelete, "/assets/domains/"+url.PathEscape(domainID), nil, nil)
}

func (a *AssetsClient) Analytics(ctx context.Context) (*AssetUsageSummary, error) {
	var result AssetUsageSummary
	err := a.client.request(ctx, http.MethodGet, "/assets/analytics", nil, &result)
	return &result, err
}

func (a *AssetsClient) UsageSummary(ctx context.Context) (*AssetUsageSummary, error) {
	var result AssetUsageSummary
	err := a.client.request(ctx, http.MethodGet, "/assets/summary", nil, &result)
	return &result, err
}

func (a *AssetsClient) Events(ctx context.Context, limit int) ([]AssetEvent, int, error) {
	var result struct {
		Events []AssetEvent `json:"events"`
		Total  int          `json:"total"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/events?limit="+strconv.Itoa(limit), nil, &result)
	return result.Events, result.Total, err
}

func (a *AssetsClient) CreateWebhook(ctx context.Context, endpoint string, eventTypes []string) (*AssetWebhook, string, error) {
	var result struct {
		Webhook AssetWebhook `json:"webhook"`
		Secret  string       `json:"secret"`
	}
	err := a.client.request(ctx, http.MethodPost, "/assets/webhooks", map[string]any{"url": endpoint, "event_types": eventTypes}, &result)
	return &result.Webhook, result.Secret, err
}

func (a *AssetsClient) ListWebhooks(ctx context.Context) ([]AssetWebhook, error) {
	var result struct {
		Webhooks []AssetWebhook `json:"webhooks"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/webhooks", nil, &result)
	return result.Webhooks, err
}

func (a *AssetsClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	return a.client.request(ctx, http.MethodDelete, "/assets/webhooks/"+url.PathEscape(webhookID), nil, nil)
}

func (a *AssetsClient) ListWebhookDeliveries(ctx context.Context, webhookID string, limit int) ([]AssetWebhookDelivery, error) {
	var result struct {
		Deliveries []AssetWebhookDelivery `json:"deliveries"`
	}
	err := a.client.request(ctx, http.MethodGet, "/assets/webhooks/"+url.PathEscape(webhookID)+"/deliveries?limit="+strconv.Itoa(limit), nil, &result)
	return result.Deliveries, err
}

func (a *AssetsClient) RetryWebhookDelivery(ctx context.Context, deliveryID string) (*AssetWebhookDelivery, error) {
	var result AssetWebhookDelivery
	err := a.client.request(ctx, http.MethodPost, "/assets/webhook-deliveries/"+url.PathEscape(deliveryID)+"/retry", nil, &result)
	return &result, err
}

func (a *AssetsClient) Purge(ctx context.Context, assetID string) error {
	return a.client.request(ctx, http.MethodPost, "/assets/"+url.PathEscape(assetID)+"/purge", nil, nil)
}

func (a *AssetsClient) Bulk(ctx context.Context, input map[string]any) (*AssetJobSubmission, error) {
	return a.BulkWithIdempotency(ctx, input, "")
}

func (a *AssetsClient) BulkWithIdempotency(ctx context.Context, input map[string]any, idempotencyKey string) (*AssetJobSubmission, error) {
	var result AssetJobSubmission
	headers := map[string]string{}
	if idempotencyKey != "" {
		headers["Idempotency-Key"] = idempotencyKey
	}
	err := a.client.requestWithHeaders(ctx, http.MethodPost, "/assets/bulk", input, headers, &result)
	return &result, err
}

func assetPreset(name, category, description string, width, height int, crop, format, quality string) AssetTransformPreset {
	return AssetTransformPreset{
		ID:          name,
		Name:        name,
		Category:    category,
		Description: description,
		AssetTransformOptions: AssetTransformOptions{
			Width: width, Height: height, Crop: crop, Format: format, Quality: quality,
		},
	}
}

func assetMultipart(upload AssetUpload) (io.ReadCloser, string, error) {
	reader, writerPipe := io.Pipe()
	writer := multipart.NewWriter(writerPipe)
	contentType := writer.FormDataContentType()

	go func() {
		err := writeAssetMultipart(writer, upload)
		if closeErr := writer.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = writerPipe.CloseWithError(err)
			return
		}
		_ = writerPipe.Close()
	}()

	return reader, contentType, nil
}

func writeAssetMultipart(writer *multipart.Writer, upload AssetUpload) error {
	part, err := writer.CreateFormFile("file", defaultString(upload.FileName, "asset"))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, upload.File); err != nil {
		return err
	}
	writeOptionalField(writer, "bucket", upload.Bucket)
	writeOptionalField(writer, "key", upload.Key)
	writeOptionalField(writer, "folder", upload.Folder)
	writeOptionalField(writer, "visibility", upload.Visibility)
	writeOptionalField(writer, "cache_control", upload.CacheControl)
	if upload.Metadata != nil {
		raw, err := json.Marshal(upload.Metadata)
		if err != nil {
			return err
		}
		_ = writer.WriteField("metadata", string(raw))
	}
	return nil
}

func transformSpec(options AssetTransformOptions) string {
	parts := make([]string, 0, 5)
	if options.Width > 0 {
		parts = append(parts, "w_"+strconv.Itoa(options.Width))
	}
	if options.Height > 0 {
		parts = append(parts, "h_"+strconv.Itoa(options.Height))
	}
	parts = append(parts, "c_"+defaultString(options.Crop, "fit"))
	parts = append(parts, "f_"+defaultString(options.Format, "auto"))
	parts = append(parts, "q_"+defaultString(options.Quality, "auto"))
	parts = append(parts, "g_"+defaultString(options.Gravity, "centre"))
	parts = append(parts, "a_"+defaultString(options.Animation, "first"))
	return strings.Join(parts, ",")
}

func writeOptionalField(writer *multipart.Writer, key, value string) {
	if value != "" {
		_ = writer.WriteField(key, value)
	}
}
