package raydium

import (
	"context"
	"fmt"
	"mm/config"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
	"resty.dev/v3"
)

const (
	DefaultRPCRetries = 2
)

type Client struct {
	RPC         solanarpc.SolanaRPC
	restyClient *resty.Client
	URL         string
}

func NewClient(rpc solanarpc.SolanaRPC, client *resty.Client, cfg *config.Config) *Client {

	return &Client{RPC: rpc, restyClient: client, URL: cfg.App.RaydiumRPCURL}
}

func (c *Client) RequireTokenAccount(ctx context.Context, acct solana.PublicKey) error {

	rpcWithRetries := solanarpc.WithRetries(c.RPC, DefaultRPCRetries)

	_, err := rpcWithRetries.GetTokenAccountBalance(ctx, acct)

	return err
}

func (c *Client) FindPoolByMints(ctx context.Context, mintA, mintB solana.PublicKey, typeProgramID ...solana.PublicKey) (*model.PoolResponse, error) {

	method := "/pools/info/mint"

	pageSize := 100

	response := FindPoolByMintsResponse{}

	data, err := c.restyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("mint1", mintA.String()).
		SetQueryParam("mint2", mintB.String()).
		SetQueryParam("poolType", "all").
		SetQueryParam("poolSortField", "liquidity").
		SetQueryParam("sortType", "desc").
		SetQueryParam("pageSize", fmt.Sprintf("%d", pageSize)).
		SetQueryParam("page", "1").
		SetResult(&response).
		Get(c.URL + method)

	if err != nil {
		return nil, apperrors.Internal("failed to get pools: ", err)
	}

	if data.IsError() {
		return nil, apperrors.Internal("failed to get data", data.Err)
	}

	result := make([]model.PoolResponse, 0, len(response.Data.Pools))

	for _, pool := range response.Data.Pools {
		if pool.IsPoolTypeOneOf(typeProgramID...) {
			result = append(result, pool)

		}
	}

	if len(result) == 0 {
		return nil, apperrors.Internal("pool not found")
	}

	if len(result) == 1 {
		return &result[0], nil
	}

	best := result[0]

	for i := 1; i < len(result); i++ {
		p := result[i]

		diff := (best.TVL - p.TVL) / best.TVL * 100

		if diff <= 15 {
			if p.FeeRate < best.FeeRate {
				best = p
			}
		} else {
			continue
		}
	}

	return &best, nil
}
