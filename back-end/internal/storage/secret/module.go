package secret

import (
	"context"
	"fmt"
	"time"

	"mm/config"

	"github.com/hashicorp/vault/api"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module("secret_db",
		fx.Provide(
			CreateSecretStorageConnection,
		),

		fx.Provide(
			NewAccountStorage,
			NewKeyStorage,
		),

		fx.Provide(
			fx.Annotate(
				NewStorage,
				fx.As(
					new(Storage),
				)),
		),

		fx.Invoke(
			func(lc fx.Lifecycle, apiClient *api.Client, cfg *config.Config, log *zap.Logger) {
				var renewCancel context.CancelFunc

				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						renewCtx, cancel := context.WithCancel(context.Background())
						renewCancel = cancel

						go maintainVaultToken(renewCtx, apiClient, cfg, log)

						return nil
					},
					OnStop: func(ctx context.Context) error {
						if renewCancel != nil {
							renewCancel()
						}

						token := apiClient.Token()

						if token == "" {
							return nil
						}

						err := apiClient.Auth().Token().RevokeSelfWithContext(ctx, token)

						if err != nil {
							return err
						}

						return nil
					},
				},
				)
			},
		),
	)
}

func maintainVaultToken(ctx context.Context, apiClient *api.Client, cfg *config.Config, log *zap.Logger) {
	const retryDelay = 5 * time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		secret, err := apiClient.Auth().Token().LookupSelfWithContext(ctx)
		if err != nil || secret == nil {
			if err != nil {
				log.Warn("failed to lookup vault token, trying to re-login", zap.Error(err))
			} else {
				log.Warn("vault token lookup returned nil, trying to re-login")
			}

			secret, err = LoginWithAppRole(ctx, apiClient, cfg)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				log.Error("failed to re-login to vault", zap.Error(err))

				if !sleepWithContext(ctx, retryDelay) {
					return
				}

				continue
			}

			log.Info("vault re-login succeeded")
		}

		renewable, ttl, err := tokenState(secret)
		if err != nil {
			log.Warn("failed to read vault token state, trying to re-login", zap.Error(err))
			if !reloginVault(ctx, apiClient, cfg, log, retryDelay) {
				return
			}
			continue
		}

		if !renewable {
			log.Warn("vault token is not renewable, trying to re-login")
			if !reloginVault(ctx, apiClient, cfg, log, retryDelay) {
				return
			}
			continue
		}

		waitBeforeRenew := nextRenewDelay(ttl)
		if !sleepWithContext(ctx, waitBeforeRenew) {
			return
		}

		renewedSecret, err := apiClient.Auth().Token().RenewSelfWithContext(ctx, 0)
		if err != nil {
			log.Warn("failed to renew vault token, trying to re-login", zap.Error(err))
			if !reloginVault(ctx, apiClient, cfg, log, retryDelay) {
				return
			}
			continue
		}

		_, renewedTTL, ttlErr := tokenState(renewedSecret)
		if ttlErr != nil {
			log.Info("vault token renewed")
			continue
		}

		log.Info("vault token renewed", zap.Duration("ttl", renewedTTL))
	}
}

func tokenState(secret *api.Secret) (bool, time.Duration, error) {
	if secret == nil {
		return false, 0, fmt.Errorf("secret is nil")
	}

	renewable, err := secret.TokenIsRenewable()
	if err != nil {
		return false, 0, fmt.Errorf("failed to read renewable flag: %w", err)
	}

	ttl, err := secret.TokenTTL()
	if err != nil {
		return false, 0, fmt.Errorf("failed to read token ttl: %w", err)
	}

	return renewable, ttl, nil
}

func reloginVault(ctx context.Context, apiClient *api.Client, cfg *config.Config, log *zap.Logger, retryDelay time.Duration) bool {
	if _, err := LoginWithAppRole(ctx, apiClient, cfg); err != nil {
		if ctx.Err() != nil {
			return false
		}

		log.Error("failed to re-login to vault", zap.Error(err))

		return sleepWithContext(ctx, retryDelay)
	}

	log.Info("vault re-login succeeded")

	return true
}

func nextRenewDelay(ttl time.Duration) time.Duration {
	switch {
	case ttl <= 0:
		return time.Second
	case ttl <= 2*time.Minute:
		return ttl / 2
	default:
		return ttl - time.Minute
	}
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
