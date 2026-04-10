package jito

import (
	"context"
	"fmt"
	"mm/config"
	"mm/internal/client/solanarpc"
	"mm/internal/client/solanaws"
	"mm/pkg/pool"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	jitogo "github.com/weeaa/jito-go"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"resty.dev/v3"
)

var Module = fx.Module("jito",
	fx.Provide(func(restyClient *resty.Client, jitoClients *pool.CloseableRoundRobin[*searcher_client.Client], cfg *config.Config, solanaWS *solanaws.Client, solanaRPC solanarpc.SolanaRPC, logger *zap.Logger) *Client {
		var tipAccounts []solana.PublicKey
		switch cfg.App.Environment {
		case config.EnvironmentProduction:
			tipAccounts = jitogo.MainnetTipAccounts
		case config.EnvironmentStage:
			tipAccounts = jitogo.MainnetTipAccounts
		case config.EnvironmentDev:
			tipAccounts = jitogo.TestnetTipAccounts

		}
		return NewClient(restyClient, solanaWS, solanaRPC, jitoClients, tipAccounts, cfg, logger)
	}),

	fx.Provide(
		func(cfg *config.Config) ([]*searcher_client.Client, error) {
			ctx := context.Background()

			rpcURLs := cfg.Jito.RpcURLs
			proxyURLs := append([]string{""}, cfg.Jito.ProxyURLs...)

			clients := make([]*searcher_client.Client, 0, len(rpcURLs))

			for _, proxyURL := range proxyURLs {
				for _, rpcURL := range rpcURLs {

					client, err := searcher_client.NewNoAuth(
						ctx,
						rpcURL,
						rpc.New(fmt.Sprintf("https://%s/api/v1/bundles", rpcURL)),
						rpc.New(cfg.App.SolanaRPCURL),
						proxyURL,
						nil,
					)

					if err != nil {
						return nil, err
					}

					clients = append(clients, client)
				}
			}

			return clients, nil
		},
	),
)
