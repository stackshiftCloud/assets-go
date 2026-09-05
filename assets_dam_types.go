package stackshift

type AssetMetadataSchema struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Revision          int64            `json:"revision"`
	PublishedRevision int64            `json:"published_revision"`
	Fields            []map[string]any `json:"fields"`
}

type AssetPublication struct {
	ID               string           `json:"id"`
	AssetID          string           `json:"asset_id"`
	VersionID        string           `json:"version_id"`
	Channel          string           `json:"channel"`
	Status           string           `json:"status"`
	Revision         int64            `json:"revision"`
	MetadataRevision int64            `json:"metadata_revision"`
	Metadata         map[string]any   `json:"metadata"`
	Renditions       []map[string]any `json:"renditions"`
	Rights           map[string]any   `json:"rights"`
	ReviewedBy       string           `json:"reviewed_by"`
	ReviewNote       string           `json:"review_note"`
}

type AssetGallery struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Revision          int64            `json:"revision"`
	PublishedRevision int64            `json:"published_revision"`
	Entries           []map[string]any `json:"entries"`
}

type AssetRenderEstimate struct {
	Units     int64  `json:"units"`
	Remaining int64  `json:"remaining"`
	Reused    bool   `json:"reused"`
	SetID     string `json:"set_id,omitempty"`
}

type AssetRenderSubmission struct {
	SetID  string `json:"set_id"`
	Status string `json:"status"`
	JobID  string `json:"job_id,omitempty"`
	Units  int64  `json:"units,omitempty"`
	Reused bool   `json:"reused,omitempty"`
}
