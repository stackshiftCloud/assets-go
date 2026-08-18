package stackshift

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AssetTransformDefinition is the ordered, deterministic v2 media transform contract.
type AssetTransformDefinition struct {
	Version    int                       `json:"version"`
	MediaKind  string                    `json:"media_kind"`
	Animation  AssetAnimationPolicy      `json:"animation,omitempty"`
	Operations []AssetTransformOperation `json:"operations"`
	Output     AssetTransformOutput      `json:"output"`
}

type AssetAnimationPolicy string

const (
	AssetAnimationFirst    AssetAnimationPolicy = "first"
	AssetAnimationPreserve AssetAnimationPolicy = "preserve"
	AssetAnimationReject   AssetAnimationPolicy = "reject"
)

// AssetTransformOperation describes one ordered image or video operation.
type AssetTransformOperation struct {
	Type             string  `json:"type"`
	Mode             string  `json:"mode,omitempty"`
	Gravity          string  `json:"gravity,omitempty"`
	Width            int     `json:"width,omitempty"`
	Height           int     `json:"height,omitempty"`
	X                int     `json:"x,omitempty"`
	Y                int     `json:"y,omitempty"`
	Degrees          float64 `json:"degrees,omitempty"`
	Sigma            float64 `json:"sigma,omitempty"`
	Radius           int     `json:"radius,omitempty"`
	DPR              float64 `json:"dpr,omitempty"`
	Color            string  `json:"color,omitempty"`
	AssetID          string  `json:"asset_id,omitempty"`
	Opacity          int     `json:"opacity,omitempty"`
	StartSeconds     float64 `json:"start_seconds,omitempty"`
	EndSeconds       float64 `json:"end_seconds,omitempty"`
	FPS              float64 `json:"fps,omitempty"`
	TimestampSeconds float64 `json:"timestamp_seconds,omitempty"`
}

// AssetTransformOutput selects the derived media kind, encoding, and image quality.
type AssetTransformOutput struct {
	Kind    string `json:"kind"`
	Format  string `json:"format"`
	Quality int    `json:"quality,omitempty"`
}

// AssetTransformMaterializeInput selects exactly one v2 definition or named preset.
type AssetTransformMaterializeInput struct {
	Definition *AssetTransformDefinition `json:"definition,omitempty"`
	Preset     string                    `json:"preset,omitempty"`
}

// AssetDerivedAsset describes a persisted, replicated transform output.
type AssetDerivedAsset struct {
	ID                   string     `json:"id"`
	AssetID              string     `json:"asset_id"`
	SourceChecksumSHA256 string     `json:"source_checksum_sha256"`
	TransformHash        string     `json:"transform_hash"`
	NormalizedSpec       string     `json:"normalized_spec"`
	DefinitionVersion    int        `json:"definition_version"`
	OutputKind           string     `json:"output_kind"`
	DependencyChecksum   string     `json:"dependency_checksum_sha256,omitempty"`
	ProcessorRevision    string     `json:"processor_revision"`
	MimeType             string     `json:"mime_type"`
	Size                 int64      `json:"size"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	LastAccessedAt       *time.Time `json:"last_accessed_at,omitempty"`
}

// AssetImageGenerationInput creates a new private image through a durable job.
type AssetImageGenerationInput struct {
	Prompt string `json:"prompt"`
	Size   string `json:"size"`
	Format string `json:"format"`
	Bucket string `json:"bucket"`
	Key    string `json:"key,omitempty"`
}

// CreateTransformationV2 stores or revises a named deterministic v2 preset.
func (a *AssetsClient) CreateTransformationV2(
	ctx context.Context,
	name string,
	definition AssetTransformDefinition,
	eager bool,
) (*AssetTransformation, error) {
	var transformation AssetTransformation
	err := a.client.request(ctx, http.MethodPost, "/assets/transformations", map[string]any{
		"name": name, "definition": definition, "eager": eager,
	}, &transformation)
	return &transformation, err
}

// MaterializeTransform enqueues a persisted transform output. Input must select one definition or preset.
func (a *AssetsClient) MaterializeTransform(
	ctx context.Context,
	assetID string,
	input AssetTransformMaterializeInput,
	idempotencyKey string,
) (*AssetJobSubmission, error) {
	var job AssetJobSubmission
	headers := map[string]string{}
	if idempotencyKey != "" {
		headers["Idempotency-Key"] = idempotencyKey
	}
	err := a.client.requestWithHeaders(ctx, http.MethodPost,
		"/assets/"+url.PathEscape(assetID)+"/transforms", input, headers, &job)
	return &job, err
}

// ListDerivatives returns persisted outputs for an asset.
func (a *AssetsClient) ListDerivatives(ctx context.Context, assetID string) ([]AssetDerivedAsset, error) {
	var result struct {
		Derivatives []AssetDerivedAsset `json:"derivatives"`
	}
	err := a.client.request(ctx, http.MethodGet,
		"/assets/"+url.PathEscape(assetID)+"/derivatives", nil, &result)
	return result.Derivatives, err
}

// GenerateAsset enqueues private image generation. The idempotency key is required by the API.
func (a *AssetsClient) GenerateAsset(
	ctx context.Context,
	input AssetImageGenerationInput,
	idempotencyKey string,
) (*AssetJobSubmission, error) {
	var job AssetJobSubmission
	err := a.client.requestWithHeaders(ctx, http.MethodPost, "/assets/generations", input,
		map[string]string{"Idempotency-Key": idempotencyKey}, &job)
	return &job, err
}
