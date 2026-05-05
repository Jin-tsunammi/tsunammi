package helius

import (
	"fmt"
	"mm/config"
	"net/http"

	"resty.dev/v3"
)

type Client struct {
	c        *resty.Client
	endpoint string
}

func NewClient(c *resty.Client, cfg *config.Config) *Client {
	return &Client{
		c:        c,
		endpoint: cfg.App.SolanaRPCURL,
	}
}

// GetPriorityFeeEstimate fetches Helius priority fee recommendation by level based on relevant accounts.
// Recommendations are returned in Microlamports/CU.
func (c *Client) GetPriorityFeeEstimate(accounts []string) (*PriorityFeeLevels, error) {
	req := &priorityFeeRequest{
		Jsonrpc: "2.0",
		ID:      "1",
		Method:  "getPriorityFeeEstimate",
		Params: []Params{
			{
				AccountKeys: accounts,
				Options: Options{
					IncludeAllPriorityFeeLevels: true,
				},
			},
		},
	}

	var res Result[PriorityFeeEstimate]
	data, err := c.c.R().EnableGenerateCurlCmd().
		SetContentType("application/json").
		SetBody(req).
		SetResult(&res).
		Post(c.endpoint)

	if err != nil {
		return nil, err
	}

	if data.StatusCode() != http.StatusOK {
		if res.Error != nil {
			return nil, fmt.Errorf("helius http %d, rpc error code %d: %s", data.StatusCode(), res.Error.Code, res.Error.Message)
		}
		return nil, fmt.Errorf("helius http %d: %s", data.StatusCode(), data.String())
	}

	if res.Error != nil {
		return nil, fmt.Errorf("error code %d, message: %s", res.Error.Code, res.Error.Message)
	}

	return &res.Result.Levels, nil
}
