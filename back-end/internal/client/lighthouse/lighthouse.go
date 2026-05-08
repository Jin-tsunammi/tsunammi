package lighthouse

import (
	"context"
	"io"
	"mm/config"

	lhsdk "github.com/lighthouse-web3/lighthouse-go-sdk/lighthouse"
)

type UploadResult struct {
	Hash string
	Name string
	Size string
}

type Client struct {
	lh *lhsdk.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		lh: lhsdk.NewClient(nil,
			lhsdk.WithAPIKey(cfg.Lighthouse.ApiKey),
		),
	}
}

func (c *Client) UploadReader(ctx context.Context, name string, size int64, r io.Reader) (*UploadResult, error) {
	upload, err := c.lh.Storage().UploadReader(ctx, name, size, r)
	if err != nil {
		return nil, err
	}
	return &UploadResult{
		Hash: upload.Hash,
		Name: upload.Name,
		Size: upload.Size,
	}, nil
}
