package kucoinapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mm/internal/model"
	"net/http"
	"strconv"
	"time"

	"resty.dev/v3"
)

func doSigned[T any](ctx context.Context, c *kuCoinApiClient, key *model.Key, method, path, query string, body any) (*Response[T], error) {
	fullUrl := fmt.Sprintf("%s%s", c.baseUrl, path)

	bytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err := c.doSignedRequest(ctx, key, method, path, query, fullUrl, bytes)
	if err != nil {
		return nil, err
	}

	rateLimitRemaining := resp.Header().Get("gw-ratelimit-remaining")
	rateLimitReset := resp.Header().Get("gw-ratelimit-reset")

	var response Response[T]
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	parsedLimitReset, err := strconv.ParseInt(rateLimitReset, 10, 64)
	if err != nil {
		return nil, err
	}

	parsedLimitRemaining, err := strconv.Atoi(rateLimitRemaining)
	if err != nil {
		return nil, err
	}

	response.RateLimitRemaining = parsedLimitRemaining
	response.RateLimitReset = time.Duration(parsedLimitReset) * time.Millisecond

	if response.Code != OK {
		if response.Code == FORBIDDEN {
			return nil, InvalidApiKeyError
		}
		return nil, fmt.Errorf("%d_%s", response.Code, response.Msg)
	}

	return &response, nil
}

// hashPassphrase hashes Kucoin passphrase using apiSecret (HMAC-SHA256 + Base64).
func hashPassphrase(passphrase, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(passphrase))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// signRequest creates a Kucoin API signature for authentication.
func (c *kuCoinApiClient) signRequest(timestamp, method, path, query, body, secret string) string {
	// Create the full path with query
	fullPath := path
	if query != "" {
		fullPath = path + "?" + query
	}

	// Prehash: timestamp + method + fullPath + body
	preHash := timestamp + method + fullPath + body

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(preHash))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// doSignedRequest performs a signed request (GET or POST) with the necessary auth headers.
func (c *kuCoinApiClient) doSignedRequest(ctx context.Context, key *model.Key, method, path, query, fullUrl string, body []byte) (*resty.Response, error) {
	requestFunc := func(cli *resty.Client) (*resty.Response, error) {
		// reqwindow
		timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

		var bodyStr string
		if body != nil && method != "GET" {
			bodyStr = string(body)
		}

		signature := c.signRequest(timestamp, method, path, query, bodyStr, key.SecretKey)

		// Hash passphrase for v2
		hashedPassphrase := hashPassphrase(key.Passphrase, key.SecretKey)

		req := cli.R().
			SetContext(ctx).
			SetHeader("KC-API-KEY", key.ApiKey).
			SetHeader("KC-API-PASSPHRASE", hashedPassphrase).
			SetHeader("KC-API-TIMESTAMP", timestamp).
			SetHeader("KC-API-SIGN", signature).
			SetHeader("KC-API-KEY-VERSION", "2").
			SetHeader("Content-Type", "application/json").
			SetQueryString(query)

		if body != nil && method != "GET" {
			req.SetBody(body)
		}

		switch method {
		case http.MethodGet:
			return req.Get(fullUrl)
		case http.MethodPost:
			return req.Post(fullUrl)
		default:
			return nil, fmt.Errorf("unsupported HTTP method: %s", method)
		}
	}

	resp, err := requestFunc(c.resty)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return nil, err
	}

	return resp, nil
}
