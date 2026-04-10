package service

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mm/internal/client/kucoinapi"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/pkg/apperrors"
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type DepositService struct {
	AccountRepository      *repository.AccountRepository
	Kc                     kucoinapi.KuCoinApi
	DepositRepository      *repository.DepositRepository
	DepositOrderRepository *repository.DepositOrderRepository
	WalletRepository       *repository.WalletRepository
	ProjectRepository      *repository.ProjectRepository
	UserActionRepository   *repository.UserHistoryRepository
	AccountStorage         *secret.AccountStorage
	SolanaRPC              solanarpc.SolanaRPC
}

func NewDepositService(
	accountRepository *repository.AccountRepository,
	kc kucoinapi.KuCoinApi,
	depositRepository *repository.DepositRepository,
	depositOrderRepository *repository.DepositOrderRepository,
	walletRepository *repository.WalletRepository,
	projectRepository *repository.ProjectRepository,
	userActionRepository *repository.UserHistoryRepository,
	accountStorage *secret.AccountStorage,
	solanaRPC solanarpc.SolanaRPC,
) *DepositService {
	return &DepositService{
		AccountRepository:      accountRepository,
		Kc:                     kc,
		DepositRepository:      depositRepository,
		DepositOrderRepository: depositOrderRepository,
		WalletRepository:       walletRepository,
		ProjectRepository:      projectRepository,
		UserActionRepository:   userActionRepository,
		AccountStorage:         accountStorage,
		SolanaRPC:              solanaRPC,
	}
}

func (s *DepositService) DepositSolana(ctx context.Context, userID uint64, req *model.DepositSolanaReq) (*model.DepositResponse, error) {
	acc, err := s.fetchAccount(ctx, req.AccountID, userID)

	if err != nil {
		return nil, err
	}

	var (
		quotas  kucoinapi.WithdrawQuotas
		balance float64
		wallets []model.Wallet
	)

	g, errctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		ccy, err := s.Kc.GetSolWithdrawDetails(errctx, &acc.Key)
		if err != nil {
			return err
		}
		quotas = *ccy
		return nil
	})

	g.Go(func() error {

		balanceSOL, err := s.Kc.GetSolBalance(errctx, acc)
		if err != nil {
			return err
		}

		balance = balanceSOL

		return nil
	})

	g.Go(func() error {

		n := req.Quantity

		project, err := s.ProjectRepository.FetchProjectWithWalletsByID(errctx, req.ProjectID, userID)

		if err != nil {
			return err
		}

		if len(project.Wallets) < n {
			return errors.New("not enough wallets")
		}

		w := make([]model.Wallet, 0, n)

		slices.SortFunc(project.Wallets, func(i, j model.Wallet) int {
			return cmp.Compare(i.ID, j.ID)
		})

		wallets = append(w, project.Wallets[:n]...)

		return nil
	})

	if err = g.Wait(); err != nil {
		if errors.Is(err, kucoinapi.InvalidApiKeyError) {
			acc.Status = model.AccountInactive

			err = s.AccountRepository.Update(ctx, acc)
			if err != nil {
				return nil, apperrors.Internal("failed to update account", err)
			}

			return nil, apperrors.Unauthorized("invalid api key")
		}
		return nil, apperrors.Internal(err.Error())

	}

	if acc.Status == model.AccountInactive {
		acc.Status = model.AccountActive

		err = s.AccountRepository.Update(ctx, acc)
		if err != nil {
			return nil, apperrors.Internal("failed to update account", err)
		}

	}

	if len(wallets) == 0 {
		log.Info().Msg("No wallets to deposit")
		return nil, apperrors.NotFound("no wallets to deposit")
	}

	if req.MinAmount < quotas.MinWithdraw && req.MaxAmount < quotas.MinWithdraw {
		return nil, apperrors.BadRequest(fmt.Sprintf("min amount %.6f or max amount %.6f is less than minimum withdraw amount %.6f", req.MinAmount, req.MaxAmount, quotas.MinWithdraw))
	}

	if balance < float64(len(wallets))*(req.MinAmount+quotas.MinFee) && balance < float64(len(wallets))*(req.MaxAmount+quotas.MinFee) {
		return nil, apperrors.BadRequest(fmt.Sprintf("insufficient balance %.6f to cover %d wallets with amount %.6f-%.6f + fee %.6f each transfer", balance, len(wallets), req.MinAmount, req.MaxAmount, quotas.MinFee))
	}

	order := &model.DepositOrder{
		AccountID: acc.ID,
		MinAmount: req.MinAmount,
		MaxAmount: req.MaxAmount,
		Status:    model.DepositAwaitingApproval,
		ProjectID: req.ProjectID,
	}
	for _, wallet := range wallets {
		order.WalletIDs = append(order.WalletIDs, wallet.ID)
	}

	order, err = s.DepositOrderRepository.Save(ctx, order)
	if err != nil {
		return nil, apperrors.Internal("failed to create deposit order", err)
	}

	amountSOL := req.MaxAmount * float64(len(wallets))
	fee := quotas.MinFee * float64(len(wallets))

	return &model.DepositResponse{
		Status:  model.DepositAwaitingApproval,
		OrderID: order.ID,
		Amount:  amountSOL,
		Fee:     fee,
	}, nil
}

func (s *DepositService) ProcessDepositOrder(ctx context.Context, userID uint64, id uint64) (*model.DepositProcessResponse, error) {
	order, err := s.DepositOrderRepository.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("deposit not found")
		}
		return nil, apperrors.Internal("failed to get deposit", err)
	}

	if order.Status == model.DepositCompleted {
		return nil, apperrors.BadRequest("deposit already completed")
	}

	if order.Status == model.DepositFailed {
		return nil, apperrors.BadRequest("deposit already failed")
	}

	if order.Status == model.DepositPending {
		return nil, apperrors.BadRequest("deposit already processed")
	}

	wallets, err := s.WalletRepository.FetchWalletsByIdsAndUserID(ctx, order.WalletIDs, userID)

	if err != nil {
		return nil, apperrors.Internal("failed to get wallets", err)
	}

	acc, err := s.fetchAccount(ctx, order.AccountID, userID)

	if err != nil {
		return nil, err
	}

	order.Status = model.DepositPending

	err = s.DepositOrderRepository.Update(ctx, order)
	if err != nil {
		return nil, apperrors.Internal("failed to update deposit order", err)
	}

	go s.processDeposits(order, wallets, acc)

	return &model.DepositProcessResponse{
		Status:  model.DepositPending,
		OrderID: order.ID,
	}, nil
}

func (s *DepositService) GetDepositStatus(ctx context.Context, id uint64) (*model.DepositOrder, error) {
	order, err := s.DepositOrderRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("deposit not found")
		}
		return nil, apperrors.Internal("failed to get deposit", err)
	}

	return order, nil
}

func (s *DepositService) GetDepositHistory(ctx context.Context, userID uint64, page, pageSize int) (*model.PaginationDepositHistoryResponse, error) {
	depositHistory, total, err := s.DepositRepository.GetDepositHistory(ctx, userID, page, pageSize)

	if err != nil {
		return nil, apperrors.Internal("failed to get deposit history", err)
	}

	for i := 0; i < len(depositHistory); i++ {

		deposit := &depositHistory[i]

		transactions, err := s.fetchTransactionBalance(ctx, deposit.Transactions)
		if err != nil {
			return nil, apperrors.Internal("failed to get transaction balance", err)
		}

		deposit.TotalSumSOL = transactions[0].SumSOL * float64(len(transactions))
		deposit.Transactions = transactions
	}

	return &model.PaginationDepositHistoryResponse{
		Deposits: depositHistory,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *DepositService) GetDepositHistoryByProjectID(ctx context.Context, projectID, userID uint64) (*model.DepositHistoryResponse, error) {
	depositHistoryResponse, err := s.DepositRepository.GetDepositHistoryByProjectID(ctx, projectID, userID)

	if err != nil {
		return nil, apperrors.Internal("failed to get deposit history", err)
	}

	if depositHistoryResponse == nil {
		return nil, nil
	}

	transactions, err := s.fetchTransactionBalance(ctx, depositHistoryResponse.Transactions)
	if err != nil {
		return nil, apperrors.Internal("failed to get transaction balance", err)
	}

	depositHistoryResponse.TotalSumSOL = transactions[0].SumSOL * float64(len(transactions))
	depositHistoryResponse.Transactions = transactions

	return depositHistoryResponse, nil
}

func (s *DepositService) fetchAccount(ctx context.Context, accID, userID uint64) (*model.Account, error) {

	acc, err := s.AccountRepository.GetByIDAndUserID(ctx, accID, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("account not found")
		}
		return nil, apperrors.Internal("failed to get account", err)
	}

	key, err := s.AccountStorage.Get(ctx, userID, acc.ExchangeAccountId, acc.ID)

	if err != nil {
		return nil, apperrors.Internal("failed to get account", err)
	}

	acc.Key = *key

	return acc, nil
}

func (s *DepositService) fetchTransactionBalance(ctx context.Context, transactions []model.Transaction) ([]model.Transaction, error) {
	wallets := make([]model.Wallet, 0, len(transactions))

	for _, trans := range transactions {
		wallets = append(wallets, model.Wallet{
			PublicKey: trans.PublicKey,
		})
	}
	_, _, err := fetchWalletsBalanceWithTotal(ctx, wallets, s.SolanaRPC, 1.0)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(wallets); i++ {
		(&transactions[i]).BalanceSOL = wallets[i].BalanceSOL
	}

	return transactions, nil
}

func (s *DepositService) processDeposits(order *model.DepositOrder, wallets []model.Wallet, acc *model.Account) {
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(wallets))
	failedCh := make(chan string, len(wallets))

	signal := make(chan struct{})

	deposits := make([]model.Deposit, len(wallets))

	for index, wallet := range wallets {
		deposit := model.Deposit{
			DepositOrderID: order.ID,
			WalletID:       wallet.ID,
			Status:         model.DepositPending,
		}
		deposits[index] = deposit
	}

	err := s.DepositRepository.CreateAll(context.Background(), deposits)

	if err != nil {
		order.Status = model.DepositFailed
		_ = s.DepositOrderRepository.Update(context.Background(), order)
		log.Err(err).Msg("Failed to create deposit")
		return
	}

	apiTimer := time.NewTimer(0)
	defer apiTimer.Stop()

	for index, wallet := range wallets {

		<-apiTimer.C

		amount := generateRandomFloat(order.MinAmount, order.MaxAmount)

		deposit := &deposits[index]

		externalID, rateLimitRemain, rateLimitReset, err := s.Kc.Withdraw(context.Background(), &acc.Key, &kucoinapi.WithdrawReq{
			Currency:     "SOL",
			Address:      wallet.PublicKey,
			Amount:       amount,
			WithdrawType: "ADDRESS",
			Chain:        "sol",
			Remark:       fmt.Sprintf("Deposit to wallet %s", wallet.PublicKey),
			FeeDeduct:    "EXTERNAL",
		})

		deposit.ExternalID = externalID

		apiTimer.Reset(rateLimitReset + 100*time.Millisecond)

		if err != nil {
			errCh <- fmt.Errorf("wallet %s: %w", wallet.PublicKey, err)
			deposit.Status = model.DepositFailed
			_ = s.DepositRepository.Update(context.Background(), deposit)

			continue
		}

		acc.WithdrawLimit = rateLimitRemain

		wg.Add(1)

		go func(dep *model.Deposit, extID string, wKey string) {
			defer wg.Done()

			var status *kucoinapi.WithdrawalStatusResp

			status, err = s.Kc.GetWithdrawalStatus(context.Background(), &acc.Key, extID)

			if err == nil {
				dep.TransactionID = status.WalletTxId
			}

			err = s.DepositRepository.Update(context.Background(), dep)

			if err != nil {
				errCh <- fmt.Errorf("wallet %s: %w", wKey, err)
				return
			}

			err = s.AccountRepository.Update(context.Background(), acc)
			if err != nil {
				errCh <- fmt.Errorf("wallet %s: %w", wKey, err)
				return
			}

			timer := time.NewTicker(15 * time.Second)
			defer timer.Stop()
			for {
				<-timer.C
				status, err = s.Kc.GetWithdrawalStatus(context.Background(), &acc.Key, extID)
				if err != nil {
					continue
				}
				switch status.Status {
				case kucoinapi.WithdrawalStatusSuccess:
					dep.TransactionID = status.WalletTxId
					err = s.DepositRepository.Update(context.Background(), dep)
					if err != nil {
						errCh <- fmt.Errorf("wallet %s: %w", wKey, err)
					}
					return
				case kucoinapi.WithdrawalStatusFailure:
					failedCh <- wKey
					dep.Status = model.DepositFailed
					err = s.DepositRepository.Update(context.Background(), dep)
					if err != nil {
						errCh <- fmt.Errorf("wallet %s: %w", wKey, err)
					}
					return
				}
			}
		}(deposit, externalID, wallet.PublicKey)

		if rateLimitRemain == 0 {
			errCh <- fmt.Errorf("account %d: rate limit reached", acc.ID)
			break
		}

	}

	go func() {
		for {
			select {
			case <-signal:
				if err := s.saveDepositUserAction(wallets, acc, order); err != nil {
					errCh <- err
				}

				return
			default:
				time.Sleep(20 * time.Second)

				if err := s.saveDepositUserAction(wallets, acc, order); err != nil {
					errCh <- err
				}
			}
		}
	}()

	wg.Wait()
	close(signal)
	order.Status = model.DepositCompleted
	err = s.DepositOrderRepository.Update(context.Background(), order)
	if err != nil {
		log.Err(err).Msg("Failed to update deposit order")
	}
}

func (s *DepositService) saveDepositUserAction(wallets []model.Wallet, acc *model.Account, order *model.DepositOrder) error {
	history, err := s.DepositRepository.GetDepositHistoryByProjectID(context.Background(), order.ProjectID, acc.UserID)

	if err != nil {
		return err
	}

	if history == nil {
		return errors.New("deposit history not found")
	}

	success := 0

	for _, t := range history.Transactions {
		if t.Status == model.DepositCompleted {
			success++
		}
	}

	message := fmt.Sprintf("Deposit %d/%d wallets with %s from %s with name %s", success, len(wallets), history.ProjectName, acc.ExchangeName, acc.ExchangeApiName)

	action := model.NewUserHistory(acc.UserID, model.ActionWalletDeposit, message)
	err = s.UserActionRepository.Create(context.Background(), action)

	if err != nil {
		return err
	}

	return nil
}
