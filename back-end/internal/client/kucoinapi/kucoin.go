package kucoinapi

import (
	"context"
	"errors"
	"mm/config"
	"mm/internal/model"
	"mm/pkg/apperrors"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"resty.dev/v3"
)

var InvalidApiKeyError = errors.New("api key is not valid")

//go:generate mockgen -source=kucoin.go -destination=mocks/kucoin_mock.go -package=kcMocks
type KuCoinApi interface {
	GetKeyInfo(ctx context.Context, key *model.Key) (*KeyInfo, error)
	GetSolBalance(ctx context.Context, acc *model.Account) (float64, error)
	GetSolWithdrawDetails(ctx context.Context, key *model.Key) (*WithdrawQuotas, error)
	Withdraw(ctx context.Context, key *model.Key, req *WithdrawReq) (withdrawalID string, rateLimitRemaining int, rateLimitReset time.Duration, err error)
	GetWithdrawalStatus(ctx context.Context, key *model.Key, withdrawalID string) (*WithdrawalStatusResp, error)
}

type kuCoinApiClient struct {
	resty   *resty.Client
	baseUrl string
	log     *zap.Logger
}

func NewKuCoinApiClient(cfg *config.Config, l *zap.Logger, client *resty.Client) KuCoinApi {
	return &kuCoinApiClient{
		resty:   client,
		baseUrl: cfg.App.KucoinBaseUrl,
		log:     l,
	}
}

func (c *kuCoinApiClient) GetKeyInfo(ctx context.Context, key *model.Key) (*KeyInfo, error) {
	resp, err := doSigned[KeyInfo](ctx, c, key, http.MethodGet, "/api/v1/user/api-key", "", nil)
	if err != nil {
		c.log.Error("Failed to get key info", zap.Error(err))
		return nil, apperrors.ErrInvalidKeys
	}

	return &resp.Data, nil
}

func (c *kuCoinApiClient) GetSolBalance(ctx context.Context, acc *model.Account) (float64, error) {
	values := url.Values{}
	values.Set("currency", "SOL")
	res, err := doSigned[[]Balance](ctx, c, &acc.Key, http.MethodGet, "/api/v1/accounts", values.Encode(), nil)
	if err != nil {
		c.log.Error("Failed to get SOL balance", zap.Error(err))
		return 0, err
	}

	if len(res.Data) == 0 {
		c.log.Error("No SOL balance found")
		return 0, errors.New("no SOL balance found")
	}

	return res.Data[0].Available, nil
}

func (c *kuCoinApiClient) GetSolWithdrawDetails(ctx context.Context, key *model.Key) (*WithdrawQuotas, error) {
	values := url.Values{}
	values.Set("currency", "SOL")
	res, err := doSigned[WithdrawQuotas](ctx, c, key, http.MethodGet, "/api/v1/withdrawals/quotas", values.Encode(), nil)
	if err != nil {
		c.log.Error("Failed to get SOL withdraw fee", zap.Error(err))
		return nil, err
	}

	return &res.Data, nil
}

func (c *kuCoinApiClient) Withdraw(ctx context.Context, key *model.Key, req *WithdrawReq) (withdrawalID string, rateLimitRemaining int, rateLimitReset time.Duration, err error) {
	resp, err := doSigned[WithdrawResp](ctx, c, key, http.MethodPost, "/api/v3/withdrawals", "", req)
	if err != nil {
		c.log.Error("Failed to withdraw", zap.Error(err))

		return "", 0, time.Duration(0), err
	}

	return resp.Data.WithdrawalID, resp.RateLimitRemaining, resp.RateLimitReset, nil
}

func (c *kuCoinApiClient) GetWithdrawalStatus(ctx context.Context, key *model.Key, withdrawalID string) (*WithdrawalStatusResp, error) {
	resp, err := doSigned[WithdrawalStatusResp](ctx, c, key, http.MethodGet, "/api/v1/withdrawals/"+withdrawalID, "", nil)
	if err != nil {
		c.log.Error("Failed to get withdrawal status", zap.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
