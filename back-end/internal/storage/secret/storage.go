package secret

import (
	"context"
	"fmt"
	"mm/config"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
)

type Storage interface {
	GetSecret(ctx context.Context, path string) (map[string]interface{}, error)
	SaveSecret(ctx context.Context, path string, secret map[string]interface{}) error
	DeleteSecret(ctx context.Context, path string) error
}

func CreateSecretStorageConnection(cfg *config.Config) (*api.Client, error) {

	defaultConfig := api.DefaultConfig()
	defaultConfig.Address = cfg.Vault.Address

	client, err := api.NewClient(defaultConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	if _, err = LoginWithAppRole(context.Background(), client, cfg); err != nil {
		return nil, err
	}

	return client, nil
}

func LoginWithAppRole(ctx context.Context, client *api.Client, cfg *config.Config) (*api.Secret, error) {
	appRoleAuth, err := approle.NewAppRoleAuth(
		cfg.Vault.RoleID,
		&approle.SecretID{FromString: cfg.Vault.SecretID},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create AppRole: %w", err)
	}

	authInfo, err := client.Auth().Login(ctx, appRoleAuth)

	if err != nil {
		return nil, fmt.Errorf("failed to login to vault: %w", err)
	}

	if authInfo == nil {
		return nil, fmt.Errorf("failed to login to vault: authInfo is nil")
	}

	return authInfo, nil
}
