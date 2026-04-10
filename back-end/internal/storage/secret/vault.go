package secret

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

var _ Storage = (*vault)(nil)

var NotFoundError = fmt.Errorf("secret not found")

type vault struct {
	ApiClient *api.Client
}

func NewStorage(apiClient *api.Client) Storage {
	return &vault{ApiClient: apiClient}
}

func (v *vault) GetSecret(ctx context.Context, path string) (map[string]any, error) {
	secret, err := v.ApiClient.Logical().ReadWithContext(ctx, path)

	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, NotFoundError
	}

	data, ok := secret.Data["data"]
	if !ok {
		return nil, NotFoundError
	}

	switch typedData := data.(type) {
	case map[string]any:
		if typedData == nil {
			return nil, NotFoundError
		}

		return typedData, nil
	case value:
		if typedData.Data == nil {
			return nil, NotFoundError
		}

		return typedData.Data, nil
	case *value:
		if typedData == nil || typedData.Data == nil {
			return nil, NotFoundError
		}

		return typedData.Data, nil
	default:
		return nil, fmt.Errorf("unexpected vault kv2 secret format for path %s", path)
	}
}

func (v *vault) SaveSecret(ctx context.Context, path string, secret map[string]any) error {
	_, err := v.ApiClient.Logical().WriteWithContext(ctx, path, map[string]any{"data": secret})
	return err
}

func (v *vault) DeleteSecret(ctx context.Context, path string) error {
	_, err := v.ApiClient.Logical().DeleteWithContext(ctx, path)
	return err
}
