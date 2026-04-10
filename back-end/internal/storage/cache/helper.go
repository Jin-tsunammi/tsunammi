package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gagliardetto/solana-go"
)

func jupiterGetRateAndDecimals(ctx context.Context, url string, mints ...solana.PublicKey) (map[string]map[string]any, error) {
	strMints := make([]string, len(mints))

	for index, mint := range mints {
		strMints[index] = mint.String()
	}

	mintIds := strings.Join(strMints, ",")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?ids=%s", url, mintIds), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	result := make(map[string]map[string]any)
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func geckoGetRateAndDecimals(ctx context.Context, mints ...solana.PublicKey) (map[string]map[string]any, error) {
	if len(mints) == 0 {
		return make(map[string]map[string]any), nil
	}

	strMints := make([]string, len(mints))
	for index, mint := range mints {
		strMints[index] = mint.String()
	}

	mintIds := strings.Join(strMints, "%2C")

	// baseURL "https://api.geckoterminal.com/api/v2/networks/solana/tokens/multi"
	url := "https://api.geckoterminal.com/api/v2/networks/solana/tokens/multi"
	fullURL := fmt.Sprintf("%s/%s?include_inactive_source=true&include=top_pools&include_composition=false", strings.TrimRight(url, "/"), mintIds)

	fmt.Println("COINGECKO", fullURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", response.StatusCode)
	}

	var apiResponse struct {
		Data []struct {
			Attributes struct {
				Address  string `json:"address"`
				Decimals int    `json:"decimals"`
				PriceUSD string `json:"price_usd"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err = json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	resp := &http.Response{Body: response.Body}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	// 2. Конвертуємо байтовий зріз у string
	_ = string(bodyBytes)

	fmt.Println("BODY", apiResponse)

	result := make(map[string]map[string]any)
	for _, token := range apiResponse.Data {
		attr := token.Attributes
		fmt.Println("Atributes", attr.Address, attr.Decimals, attr.PriceUSD)
		price, _ := strconv.ParseFloat(attr.PriceUSD, 64)

		result[attr.Address] = map[string]any{
			"decimals": attr.Decimals,
			"usdPrice": price,
		}
	}

	fmt.Println("RESULT", result)

	return result, nil
}
