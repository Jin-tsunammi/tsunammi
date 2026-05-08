package model

import "fmt"

// UploadImageResponse is returned after a successful image upload to Lighthouse.
//
//	@Description	IPFS upload result for an image file
type UploadImageResponse struct {
	CID  string `json:"cid"  example:"bafkreigh2akiscaildcqv7bchbosqjg5qbapjklkpkhhwvafg5o4xtwmhq"`
	URL  string `json:"url"  example:"https://gateway.lighthouse.storage/ipfs/bafkreigh2akiscaildcqv7bchbosqjg5qbapjklkpkhhwvafg5o4xtwmhq"`
	Name string `json:"name" example:"logo.png"`
	Size string `json:"size" example:"204800"`
}

func NewUploadImageResponse(hash, name, size string) UploadImageResponse {
	return UploadImageResponse{
		CID:  hash,
		URL:  fmt.Sprintf("https://gateway.lighthouse.storage/ipfs/%s", hash),
		Name: name,
		Size: size,
	}
}

// UploadMetadataRequest is an arbitrary key-value JSON object representing NFT or token metadata.
//
//	@Description	Arbitrary metadata object to be uploaded as metadata.json (must not be empty)
type UploadMetadataRequest map[string]any

// UploadMetadataResponse is returned after a successful metadata upload to Lighthouse.
//
//	@Description	IPFS upload result for a metadata.json file
type UploadMetadataResponse struct {
	CID         string `json:"cid"          example:"bafkreigh2akiscaildcqv7bchbosqjg5qbapjklkpkhhwvafg5o4xtwmhq"`
	MetadataURL string `json:"metadata_url" example:"https://gateway.lighthouse.storage/ipfs/bafkreigh2akiscaildcqv7bchbosqjg5qbapjklkpkhhwvafg5o4xtwmhq"`
}

func NewUploadMetadataResponse(hash string) UploadMetadataResponse {
	return UploadMetadataResponse{
		CID:         hash,
		MetadataURL: fmt.Sprintf("https://gateway.lighthouse.storage/ipfs/%s", hash),
	}
}
