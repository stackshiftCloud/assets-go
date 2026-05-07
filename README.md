# StackShift Assets Go SDK

Official Go SDK for StackShift Assets.

## Install

```bash
go get github.com/stackshiftCloud/assets-go
```

## Upload an asset

```go
package main

import (
	"context"
	"log"
	"os"

	stackshift "github.com/stackshiftCloud/assets-go"
)

func main() {
	client, err := stackshift.New(stackshift.Options{})
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open("avatar.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	asset, err := client.Assets.Upload(context.Background(), stackshift.AssetUpload{
		File:       file,
		FileName:   "avatar.png",
		Folder:     "avatars",
		Visibility: "public",
		Metadata:   map[string]any{"userId": "user_123"},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(asset.URL)
}
```

You do not pass a deployed StackShift project ID. The API key identifies the StackShift account, and StackShift resolves the default asset space internally.

StackShift chooses storage nodes internally. SDK users work with buckets, keys, asset metadata, public URLs, and signed URLs.
