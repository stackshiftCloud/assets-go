package stackshift

import (
	"io"
	"time"
)

type Options struct {
	APIKey     string
	BaseURL    string
	HTTPClient HTTPDoer
}

type Asset struct {
	ID                string         `json:"id"`
	Bucket            string         `json:"bucket"`
	Key               string         `json:"key"`
	OriginalName      string         `json:"original_name"`
	MimeType          string         `json:"mime_type"`
	Size              int64          `json:"size"`
	ChecksumSHA256    string         `json:"checksum_sha256"`
	Visibility        string         `json:"visibility"`
	Status            string         `json:"status"`
	ReplicationStatus string         `json:"replication_status"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CacheControl      string         `json:"cache_control,omitempty"`
	URL               string         `json:"url,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type AssetUpload struct {
	File         io.Reader
	FileName     string
	Bucket       string
	Key          string
	Folder       string
	Visibility   string
	CacheControl string
	Metadata     map[string]any
}

type AssetListOptions struct {
	Bucket     string
	Prefix     string
	Visibility string
	Limit      int
	Cursor     string
}

type AssetListResult struct {
	Assets     []Asset `json:"assets"`
	Total      int     `json:"total"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

type SignedURLOptions struct {
	ExpiresIn    string
	MaxDownloads *int
}

type SignedURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SignedUploadURLOptions struct {
	Bucket       string
	Key          string
	Folder       string
	Visibility   string
	CacheControl string
	Metadata     map[string]any
	ExpiresIn    string
	MaxBytes     int64
}

type SignedUploadURL struct {
	UploadSessionID string    `json:"upload_session_id"`
	URL             string    `json:"url"`
	Method          string    `json:"method"`
	ExpiresAt       time.Time `json:"expires_at"`
	MaxBytes        int64     `json:"max_bytes"`
}
