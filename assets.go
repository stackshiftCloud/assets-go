package stackshift

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

type AssetsClient struct {
	client *Client
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

func (a *AssetsClient) Delete(ctx context.Context, assetID string) error {
	return a.client.request(ctx, http.MethodDelete, "/assets/"+url.PathEscape(assetID), nil, nil)
}

func (a *AssetsClient) Replace(ctx context.Context, assetID string, upload AssetUpload) (*Asset, error) {
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
	var asset Asset
	if err := a.client.do(req, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (a *AssetsClient) SignedURL(ctx context.Context, assetID string, options SignedURLOptions) (*SignedURL, error) {
	body := map[string]any{"expiresIn": options.ExpiresIn}
	if options.MaxDownloads != nil {
		body["maxDownloads"] = *options.MaxDownloads
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
	err := a.client.request(ctx, http.MethodPost, "/assets/upload-sessions", map[string]any{
		"bucket": options.Bucket, "key": options.Key, "folder": options.Folder, "visibility": options.Visibility,
		"cacheControl": options.CacheControl, "metadata": options.Metadata, "expiresIn": options.ExpiresIn, "maxBytes": options.MaxBytes,
	}, &signed)
	return &signed, err
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
	writeOptionalField(writer, "cacheControl", upload.CacheControl)
	if upload.Metadata != nil {
		raw, err := json.Marshal(upload.Metadata)
		if err != nil {
			return err
		}
		_ = writer.WriteField("metadata", string(raw))
	}
	return nil
}

func writeOptionalField(writer *multipart.Writer, key, value string) {
	if value != "" {
		_ = writer.WriteField(key, value)
	}
}
