package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mm/config"
	"mm/internal/client/kucoinapi"
	"mm/internal/crypto"
	"mm/internal/model"
	"mm/internal/storage/cache"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/pkg/apperrors"
	repo "mm/pkg/repository"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type AccountService struct {
	AccountRepository     *repository.AccountRepository
	AccountEncryptor      crypto.Encryptor `name:"account_encryptor"`
	Kc                    kucoinapi.KuCoinApi
	Cfg                   *config.Config
	AccountStorage        *secret.AccountStorage
	RateCache             cache.RateStorage
	TransactionManager    *repo.TransactionManager
	UserHistoryRepository *repository.UserHistoryRepository
}

type accountServiceParams struct {
	fx.In

	AccountRepository    *repository.AccountRepository
	AccountEncryptor     crypto.Encryptor `name:"account_encryptor"`
	Kc                   kucoinapi.KuCoinApi
	Cfg                  *config.Config
	AccountStorage       *secret.AccountStorage
	RateCache            cache.RateStorage
	TransactionManager   *repo.TransactionManager
	UserActionRepository *repository.UserHistoryRepository
}

func NewAccountService(p accountServiceParams) *AccountService {
	return &AccountService{
		AccountRepository:     p.AccountRepository,
		AccountEncryptor:      p.AccountEncryptor,
		Kc:                    p.Kc,
		Cfg:                   p.Cfg,
		AccountStorage:        p.AccountStorage,
		RateCache:             p.RateCache,
		TransactionManager:    p.TransactionManager,
		UserHistoryRepository: p.UserActionRepository,
	}
}

func (s *AccountService) AddExchangeAccount(ctx context.Context, req *model.AddExchangeAccountReq, userID uint64) (*model.Account, error) {

	acc, err := s.Kc.GetKeyInfo(ctx, &model.Key{
		ApiKey:     req.ApiKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
	})

	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidKeys) {
			return nil, apperrors.BadRequest("invalid api keys")
		}
		return nil, err
	}

	if !strings.Contains(acc.Perm, "Transfer") {
		return nil, apperrors.Unauthorized("no withdraw permission", err)
	}

	account := new(model.Account)

	message := fmt.Sprintf("created account from %s with name %s", account.ExchangeName, account.Name)

	userEvent := model.NewUserHistory(userID, model.ActionAddAPI, message)

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			account, err = s.AccountRepository.WithTx(tx).AddExchangeAccount(ctx, &model.Account{
				Name:              req.Name,
				ExchangeID:        req.ExchangeID,
				UserID:            userID,
				ExchangeAccountId: acc.UID,
				ExchangeApiName:   acc.ApiName,
				Status:            model.AccountStatusActive,
			})
			if err != nil {
				return err
			}

			userEvent, err = s.UserHistoryRepository.WithTx(tx).Save(ctx, userEvent)

			if err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return nil, apperrors.Internal("failed to add exchange account", err)
	}

	err = s.AccountStorage.Save(ctx, secret.AccountSecret{
		AccountID:         account.ID,
		UserID:            userID,
		ExchangeAccountID: acc.UID,
		ApiKey:            req.ApiKey,
		SecretKey:         req.SecretKey,
		Passphrase:        req.Passphrase,
	})

	fmt.Println("SECRETERR", err)

	if err != nil {
		_ = s.TransactionManager.WithinTransaction(ctx,
			func(ctx context.Context, tx bun.Tx) error {

				if err = s.AccountRepository.WithTx(tx).Delete(ctx, account.ID); err != nil {
					return err
				}

				if err = s.UserHistoryRepository.WithTx(tx).Delete(ctx, userEvent.ID); err != nil {
					return err
				}

				return nil

			},
		)

		return nil, apperrors.Internal("failed to save account", err)
	}

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			account.Status = model.AccountStatusActive
			err = s.AccountRepository.WithTx(tx).Update(ctx, account)

			if err != nil {
				return err
			}

			err = s.UserHistoryRepository.WithTx(tx).Update(ctx, userEvent)

			if err != nil {
				return err
			}

			return nil

		},
	)

	account.ApiKey = req.ApiKey
	account.SecretKey = req.SecretKey
	account.Passphrase = req.Passphrase

	return account, nil
}

func (s *AccountService) GetAccountByIDAndUserID(ctx context.Context, id, userID uint64) (*model.AccountResponse, error) {
	account, err := s.AccountRepository.GetByIDAndUserID(ctx, id, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, apperrors.Internal("failed to get account", err)
	}

	accountResponse, err := s.accountToAccountResponse(ctx, *account)

	if err != nil {
		return nil, apperrors.Internal("failed to get rate", err)
	}

	return accountResponse, nil
}

func (s *AccountService) GetAccountsByUserID(ctx context.Context, parsedPage, parsedPageSize int, userID uint64) (*model.AccountsWithPaginationResponse, error) {
	accounts, total, err := s.AccountRepository.FindAllByUserID(ctx, parsedPage, parsedPageSize, userID)

	if err != nil {
		return nil, apperrors.Internal("failed to get accounts", err)
	}

	accountRes := make([]model.AccountResponse, 0, len(accounts))

	for _, account := range accounts {

		accountResponse, err := s.accountToAccountResponse(ctx, account)
		if err != nil {
			return nil, apperrors.Internal("failed to get rate", err)
		}

		accountRes = append(accountRes, *accountResponse)

	}

	return &model.AccountsWithPaginationResponse{
		Accounts: accountRes,
		Total:    total,
		Page:     parsedPage,
		PageSize: parsedPageSize,
	}, nil
}

func (s *AccountService) DeleteAccountByIDAndUserID(ctx context.Context, id, userID uint64) error {

	acc, err := s.AccountRepository.GetByIDAndUserID(ctx, id, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return apperrors.Internal("failed to get account", err)
	}

	acc.Status = model.AccountStatusDeleted

	err = s.AccountRepository.Update(ctx, acc)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("account not found")
		}
		return apperrors.Internal("failed to delete account", err)
	}

	err = s.AccountStorage.Delete(ctx, userID, acc.ExchangeAccountId, acc.ID)

	if err != nil {
		acc.Status = model.AccountStatusActive
		_ = s.AccountRepository.Update(ctx, acc)
		return apperrors.Internal("failed to delete secret", err)
	}

	message := fmt.Sprintf("deleted account from %s with name %s", acc.ExchangeName, acc.Name)

	userEvent := model.NewUserHistory(userID, model.ActionDeleteAPI, message)

	_ = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {

			err = s.AccountRepository.WithTx(tx).DeleteByIDAndUserID(ctx, acc.ID, userID)
			if err != nil {
				return err
			}

			userEvent, err = s.UserHistoryRepository.WithTx(tx).Save(ctx, userEvent)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return nil
}

func (s *AccountService) accountToAccountResponse(ctx context.Context, account model.Account) (*model.AccountResponse, error) {

	rate, err := s.RateCache.GetUSDRate(ctx, solana.SolMint)
	if err != nil {
		rate, err = s.RateCache.GetUSDRate(ctx, solana.SolMint)
		if err != nil {
			return nil, err
		}
	}

	return &model.AccountResponse{
		ID:               account.ID,
		Name:             account.Name,
		Exchange:         account.ExchangeName,
		Status:           account.Status,
		AccountId:        account.ExchangeAccountId,
		ApiName:          account.ExchangeApiName,
		WithdrawLimit:    account.WithdrawLimit,
		TotalDepositsSOL: account.DepositBalance,
		TotalDepositsUSD: account.DepositBalance * rate,
		CreatedAt:        account.CreatedAt,
	}, nil
}
