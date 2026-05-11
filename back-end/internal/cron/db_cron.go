package cron

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"mm/config"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/pkg/apperrors"
	"mm/pkg/mtype"
	repo "mm/pkg/repository"
	"slices"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	rcron "github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

var timeoutDuration = 5 * time.Minute

type DBCron struct {
	AccountRepository            *repository.AccountRepository
	UserHistoryRepository        *repository.UserHistoryRepository
	WalletRepository             *repository.WalletRepository
	ProjectRepository            *repository.ProjectRepository
	SwapCampaignRepository       *repository.SwapCampaignRepository
	SwapTransactionRepository    *repository.SwapTransactionRepository
	BuybackTransactionRepository *repository.BuybackTransactionRepository
	DepositRepository            *repository.DepositRepository
	SecretStorage                secret.Storage
	TransactionManager           *repo.TransactionManager

	SolanaRPC solanarpc.SolanaRPC

	Cron *rcron.Cron

	Log *zap.Logger
	Cfg *config.Config
}

func NewDBCron(
	accountRepository *repository.AccountRepository,
	userHistoryRepository *repository.UserHistoryRepository,
	walletRepository *repository.WalletRepository,
	projectRepository *repository.ProjectRepository,
	swapCampaignRepository *repository.SwapCampaignRepository,
	swapTransactionRepository *repository.SwapTransactionRepository,
	buybackTransactionRepository *repository.BuybackTransactionRepository,
	depositRepository *repository.DepositRepository,
	secretStorage secret.Storage,
	transactionManager *repo.TransactionManager,

	solanaRPC solanarpc.SolanaRPC,

	cron *rcron.Cron,

	log *zap.Logger,
	cfg *config.Config,
) *DBCron {
	dbCron := &DBCron{
		AccountRepository:            accountRepository,
		UserHistoryRepository:        userHistoryRepository,
		WalletRepository:             walletRepository,
		ProjectRepository:            projectRepository,
		SwapCampaignRepository:       swapCampaignRepository,
		SwapTransactionRepository:    swapTransactionRepository,
		BuybackTransactionRepository: buybackTransactionRepository,
		DepositRepository:            depositRepository,
		SecretStorage:                secretStorage,
		TransactionManager:           transactionManager,

		SolanaRPC: solanaRPC,

		Cron: cron,

		Log: log,
		Cfg: cfg,
	}

	dbCron.registerJobs()
	dbCron.Log.Info("registered cron jobs")

	return dbCron
}

func (c *DBCron) registerJobs() {
	add := func(spec string, fn func()) {
		if _, err := c.Cron.AddFunc(spec, fn); err != nil {
			LogErr(c.Log, err)
		}
	}

	accountsPendingSpec := durationToCron(c.Cfg.Job.AccountsPendingCheckInterval)
	accountsDeletedSpec := durationToCron(c.Cfg.Job.AccountsDeletedCheckInterval)
	walletsPendingCreateSpec := durationToCron(c.Cfg.Job.WalletsPendingCreateCheckInterval)
	walletsPendingImportSpec := durationToCron(c.Cfg.Job.WalletsPendingImportCheckInterval)
	transactionPendingSpec := durationToCron(time.Minute)
	withdrawLimitSpec := durationToCron(c.Cfg.Job.WithdrawLimitCheckInterval)

	add(accountsPendingSpec, c.checkPendingAccounts)
	add(accountsDeletedSpec, c.checkDeletedAccounts)
	add(walletsPendingCreateSpec, c.checkPendingCreationWallets)
	add(walletsPendingImportSpec, c.checkPendingImportWallets)
	add(transactionPendingSpec, c.checkPendingTransactions)
	add(withdrawLimitSpec, c.updateWithdrawLimit)
}

func (c *DBCron) updateWithdrawLimit() {
	ctx := context.Background()

	accounts, err := c.AccountRepository.FindAll(ctx)

	if err != nil {
		c.Log.Error("failed to get accounts", zap.Error(err))
		return
	}

	for index := range accounts {
		account := accounts[index]
		account.WithdrawLimit = 50
	}

	err = c.AccountRepository.UpdateAll(ctx, accounts)
	if err != nil {
		c.Log.Error("failed to update withdraw limit", zap.Error(err))
	}

}

func (c *DBCron) checkDeletedAccounts() {
	ctx := context.Background()

	duration := c.Cfg.Job.AccountsDeletedCheckInterval

	accounts, err := c.AccountRepository.FindAllOlderThanByStatus(ctx, duration, model.AccountStatusDeleted)

	if err != nil {
		c.Log.Error("failed to get pending accounts", zap.Error(err))
		return
	}

	for i := 0; i < len(accounts); i++ {

		account := &accounts[i]

		path := secret.CreateAccountSecretPath(account.UserID, account.ExchangeAccountId, account.ID)
		_, err = c.SecretStorage.GetSecret(ctx, path)

		if errors.Is(err, secret.NotFoundError) {

			message := fmt.Sprintf("deleted account from %s with name %s", account.ExchangeName, account.Name)

			userEvent := model.NewUserHistory(account.UserID, model.ActionDeleteAPI, message)

			_ = c.TransactionManager.WithinTransaction(ctx,
				func(ctx context.Context, tx bun.Tx) error {

					err = c.AccountRepository.WithTx(tx).DeleteByIDAndUserID(ctx, account.ID, account.UserID)
					if err != nil {
						return err
					}

					err = c.UserHistoryRepository.WithTx(tx).Create(ctx, userEvent)
					if err != nil {
						return err
					}

					return nil
				},
			)
			return
		}

		account.Status = model.AccountStatusActive
		_ = c.AccountRepository.Update(ctx, account)

	}

}

func (c *DBCron) checkPendingAccounts() {
	ctx := context.Background()

	duration := c.Cfg.Job.AccountsPendingCheckInterval

	accounts, err := c.AccountRepository.FindAllOlderThanByStatus(ctx, duration, model.AccountStatusPending)

	if err != nil {
		c.Log.Error("failed to get pending accounts", zap.Error(err))
		return
	}

	for i := 0; i < len(accounts); i++ {

		account := &accounts[i]

		path := secret.CreateAccountSecretPath(account.UserID, account.ExchangeAccountId, account.ID)
		_, err = c.SecretStorage.GetSecret(ctx, path)

		if errors.Is(err, secret.NotFoundError) {
			_ = c.AccountRepository.DeleteByIDAndUserID(ctx, account.ID, account.UserID)
			return
		}

		message := fmt.Sprintf("created account from %s with name %s", account.ExchangeName, account.Name)

		userEvent := model.NewUserHistory(account.UserID, model.ActionAddAPI, message)

		_ = c.TransactionManager.WithinTransaction(ctx,
			func(ctx context.Context, tx bun.Tx) error {

				account.Status = model.AccountStatusActive
				err = c.AccountRepository.WithTx(tx).Update(ctx, account)
				if err != nil {
					return err
				}

				err = c.UserHistoryRepository.WithTx(tx).Create(ctx, userEvent)
				if err != nil {
					return err
				}

				return nil
			},
		)
	}
}

func (c *DBCron) checkPendingImportWallets() {
	ctx := context.Background()
	c.processPendingWallets(ctx, c.Cfg.Job.WalletsPendingImportCheckInterval, model.WalletStatusImportPending)
}

func (c *DBCron) checkPendingCreationWallets() {
	ctx := context.Background()
	c.processPendingWallets(ctx, c.Cfg.Job.WalletsPendingCreateCheckInterval, model.WalletStatusCreationPending)
}

func (c *DBCron) processPendingWallets(ctx context.Context, interval time.Duration, status model.WalletStatus) {
	c.Log.Info("processing pending wallets", zap.Duration("interval", interval))
	projects, err := c.ProjectRepository.FindAllOlderThanWithWalletsByStatus(ctx, interval, status)

	if err != nil {
		c.Log.Error("failed to get pending wallets", zap.Error(err))
		return
	}

	for i := 0; i < len(projects); i++ {

		project := &projects[i]
		wallets := project.Wallets

		idsToDelete := make([]uint64, 0, len(wallets))
		walletsToUpdate := make([]model.Wallet, 0, len(wallets))

		for j := 0; j < len(wallets); j++ {

			wallet := &wallets[j]

			path := secret.CreateWalletSecretPath(project.UserID, wallet.PublicKey)
			_, err = c.SecretStorage.GetSecret(ctx, path)

			if err != nil {
				if errors.Is(err, secret.NotFoundError) {
					idsToDelete = append(idsToDelete, wallet.ID)
				}
				continue
			}

			wallet.Status = model.WalletStatusSuccess
			walletsToUpdate = append(walletsToUpdate, *wallet)

		}

		if len(idsToDelete) > 0 {
			_ = c.WalletRepository.DeleteWalletsByIds(ctx, idsToDelete)
		}

		message := ""

		switch status {
		case model.WalletStatusImportPending:
			message = fmt.Sprintf("imported %d wallets with %s", len(wallets), project.Name)
		case model.WalletStatusCreationPending:
			message = fmt.Sprintf("created %d wallets with %s", len(wallets), project.Name)
		}

		if len(walletsToUpdate) > 0 {
			event := model.NewUserHistory(project.UserID, model.ActionWalletsBatchAdd, message)

			err = c.TransactionManager.WithinTransaction(ctx,
				func(ctx context.Context, tx bun.Tx) error {
					err = c.WalletRepository.WithTx(tx).UpdateAllStatus(ctx, walletsToUpdate)
					if err != nil {
						return err
					}

					err = c.UserHistoryRepository.WithTx(tx).Create(ctx, event)
					if err != nil {
						return err
					}

					return nil
				},
			)

			if err != nil {
				LogErr(c.Log, err)
			}

		}
	}

}

func (c *DBCron) checkPendingTransactions() {
	err := c.processPendingTransactions()
	c.Log.Error("INSIDE trans", zap.Error(err))
	if err != nil {
		LogErr(c.Log, err)
	}
}

func (c *DBCron) processPendingTransactions() error {
	ctx := context.Background()
	c.Log.Info("starting processing pending transactions", zap.Duration("interval", c.Cfg.Job.TransactionPendingCheckInterval))

	swapTransactions, err := c.SwapTransactionRepository.FindAllByStatus(ctx, "Pending")
	if err != nil {
		c.Log.Error("failed to get pending swap transactions", zap.Error(err))
		return err
	}
	c.Log.Info("fetched pending swap transactions", zap.Int("count", len(swapTransactions)))

	buybackTransactions, err := c.BuybackTransactionRepository.FindAllByStatus(ctx, "Pending")
	if err != nil {
		c.Log.Error("failed to get pending buyback transactions", zap.Error(err))
		return err
	}
	c.Log.Info("fetched pending buyback transactions", zap.Int("count", len(buybackTransactions)))

	deposits, err := c.DepositRepository.GetAllBySum(ctx, mtype.NewDBBigRat(big.NewRat(0, 1)))
	if err != nil {
		c.Log.Error("failed to get deposits without balance", zap.Error(err))
		return err
	}
	c.Log.Info("fetched deposits without balance", zap.Int("count", len(deposits)))

	totalToProcess := len(swapTransactions) + len(buybackTransactions) + len(deposits)
	if totalToProcess == 0 {
		c.Log.Info("no pending transactions to process")
		return nil
	}

	errs := make([]error, totalToProcess)
	transactionsSignatures := make([]solana.Signature, totalToProcess)

	for index, transaction := range swapTransactions {
		sig, tErr := solana.SignatureFromBase58(transaction.TransactionHash)
		if tErr != nil {
			c.Log.Warn("failed to parse swap signature", zap.String("hash", transaction.TransactionHash), zap.Error(tErr))
			errs[index] = tErr
			continue
		}
		transactionsSignatures[index] = sig
	}

	for index, transaction := range buybackTransactions {
		sig, tErr := solana.SignatureFromBase58(transaction.TransactionHash)
		if tErr != nil {
			c.Log.Warn("failed to parse buyback signature", zap.String("hash", transaction.TransactionHash), zap.Error(tErr))
			errs[index] = tErr
			continue
		}
		transactionsSignatures[index] = sig
	}

	for index, transaction := range deposits {
		sig, tErr := solana.SignatureFromBase58(transaction.TransactionID)
		if tErr != nil {
			c.Log.Warn("failed to parse deposit signature", zap.String("hash", transaction.TransactionID), zap.Error(tErr))
			errs[index+len(swapTransactions)] = tErr
			continue
		}
		transactionsSignatures[index+len(swapTransactions)] = sig
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	amounts := make(map[string]map[string]mtype.BigRat)

	semaphore := make(chan struct{}, 4)

	for i, sig := range transactionsSignatures {
		if sig.Equals(solana.Signature{}) {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(index int, transactionSignature solana.Signature) {
			defer wg.Done()
			defer func() { <-semaphore }()

			sigStr := transactionSignature.String()
			c.Log.Debug("requesting transaction from RPC", zap.String("sig", sigStr))

			transaction, tErr := c.SolanaRPC.GetTransaction(ctx, transactionSignature, &rpc.GetTransactionOpts{
				Encoding:   solana.EncodingBase64,
				Commitment: rpc.CommitmentConfirmed,
			})

			if tErr != nil {
				c.Log.Warn("RPC error for transaction", zap.String("sig", sigStr), zap.Error(tErr))
				errs[index] = tErr
				return
			}

			if transaction == nil || transaction.Meta == nil {
				c.Log.Warn("transaction or meta not found in Solana", zap.String("sig", sigStr))
				errs[index] = errors.New("transaction not found")
				return
			}

			result := make(map[string]mtype.BigRat)

			if len(transaction.Meta.PreBalances) > 0 && len(transaction.Meta.PostBalances) > 0 {
				preBalance := transaction.Meta.PreBalances[0]
				postBalance := transaction.Meta.PostBalances[0]

				delta := int64(postBalance) - int64(preBalance) + int64(transaction.Meta.Fee)

				if delta != 0 {
					balanceStr := solanarpc.FromAtomicUnit(uint64(math.Abs(float64(delta))), solana.SolDecimals)
					ratBalance := new(big.Rat).SetFloat64(balanceStr)
					result[solana.SolMint.String()] = mtype.NewDBBigRat(ratBalance)
				}
			}

			preBalances := make(map[uint16]rpc.TokenBalance)
			postBalances := make(map[uint16]rpc.TokenBalance)
			uniqueKeys := make(map[uint16]struct{})

			for _, b := range transaction.Meta.PreTokenBalances {
				preBalances[b.AccountIndex] = b
				uniqueKeys[b.AccountIndex] = struct{}{}
			}
			for _, b := range transaction.Meta.PostTokenBalances {
				postBalances[b.AccountIndex] = b
				uniqueKeys[b.AccountIndex] = struct{}{}
			}

			for key := range uniqueKeys {
				postB, postOk := postBalances[key]
				preB, preOk := preBalances[key]

				var mint solana.PublicKey
				if postOk {
					mint = postB.Mint
				} else if preOk {
					mint = preB.Mint
				} else {
					continue
				}

				preAmountStr := "0"
				if preOk {
					preAmountStr = preB.UiTokenAmount.UiAmountString
				}

				postAmountStr := "0"
				if postOk {
					postAmountStr = postB.UiTokenAmount.UiAmountString
				}

				x, _ := new(big.Rat).SetString(postAmountStr)
				y, _ := new(big.Rat).SetString(preAmountStr)
				res := new(big.Rat).Sub(x, y)

				if res.Sign() != 0 {
					absRes := new(big.Rat).Abs(res)
					mintStr := mint.String()

					if existing, exists := result[mintStr]; !exists || absRes.Cmp(existing.GetBigRat()) > 0 {
						result[mintStr] = mtype.NewDBBigRat(absRes)
					}

					c.Log.Debug("token balance change detected",
						zap.String("sig", sigStr),
						zap.String("mint", mintStr),
						zap.String("diff", res.FloatString(6)))
				}
			}

			mu.Lock()
			amounts[sigStr] = result
			mu.Unlock()
		}(i, sig)
	}

	wg.Wait()

	if err = errors.Join(errs...); err != nil {
		c.Log.Warn("some transactions failed during RPC fetching", zap.Error(err))
	}

	var swapTransactionsToUpdate []model.SwapTransaction
	var buybackTransactionsToUpdate []model.BuybackTransaction
	for i := range swapTransactions {
		t := &swapTransactions[i]
		if amt, ok := amounts[t.TransactionHash]; ok {
			t.AmountTokenFrom = amt[t.TokenMintFrom]
			t.AmountTokenTo = amt[t.TokenMintTo]
			t.Status = "Success"
			swapTransactionsToUpdate = append(swapTransactionsToUpdate, *t)
			c.Log.Info("mapping swap result success", zap.String("hash", t.TransactionHash))
		} else {
			if time.Since(t.CreatedAt) > timeoutDuration {
				t.Status = "Failed"
				swapTransactionsToUpdate = append(swapTransactionsToUpdate, *t)
				c.Log.Warn("swap marked as failed due to timeout", zap.String("hash", t.TransactionHash))
			}
		}
	}

	for i := range buybackTransactions {
		t := &buybackTransactions[i]
		if amt, ok := amounts[t.TransactionHash]; ok {
			t.AmountTokenFrom = amt[t.TokenMintFrom]
			t.AmountTokenTo = amt[t.TokenMintTo]
			t.Status = "Success"
			buybackTransactionsToUpdate = append(buybackTransactionsToUpdate, *t)
			c.Log.Info("mapping buyback result success", zap.String("hash", t.TransactionHash))
		} else {
			if time.Since(t.CreatedAt) > timeoutDuration {
				t.Status = "Failed"
				buybackTransactionsToUpdate = append(buybackTransactionsToUpdate, *t)
				c.Log.Warn("buyback marked as failed due to timeout", zap.String("hash", t.TransactionHash))
			}
		}

	}

	var depositsToUpdate []model.Deposit
	for i := range deposits {
		t := &deposits[i]
		if amt, ok := amounts[t.TransactionID]; ok {
			t.Status = model.DepositCompleted
			t.Amount = amt[solana.SolMint.String()]
			depositsToUpdate = append(depositsToUpdate, *t)
			c.Log.Info("mapping deposit result success", zap.String("id", t.TransactionID))
		} else {
			if time.Since(t.CreatedAt) > timeoutDuration {
				t.Status = model.DepositFailed
				depositsToUpdate = append(depositsToUpdate, *t)
				c.Log.Warn("deposit marked as failed due to timeout", zap.String("id", t.TransactionID))
			}
		}
	}

	if len(swapTransactionsToUpdate) > 0 {
		chunks := slices.Collect(slices.Chunk(swapTransactionsToUpdate, 500))
		for _, chunk := range chunks {
			if err = c.SwapTransactionRepository.UpdateAll(ctx, chunk); err != nil {
				c.Log.Error("failed to update swap transactions batch", zap.Error(err))
			}
		}
	}

	if len(buybackTransactionsToUpdate) > 0 {
		chunks := slices.Collect(slices.Chunk(buybackTransactionsToUpdate, 500))
		for _, chunk := range chunks {
			if err = c.BuybackTransactionRepository.UpdateAll(ctx, chunk); err != nil {
				c.Log.Error("failed to update buyback transactions batch", zap.Error(err))
			}
		}
	}

	if len(depositsToUpdate) > 0 {
		if err = c.DepositRepository.UpdateAll(ctx, depositsToUpdate); err != nil {
			c.Log.Error("failed to update deposit transactions", zap.Error(err))
			return err
		}
	}

	c.Log.Info("successfully finished processing transactions")
	return nil
}

func (c *DBCron) start(_ context.Context) error {
	c.Log.Info("starting cron jobs")
	c.Cron.Start()
	return nil
}

func (c *DBCron) stop(_ context.Context) error {
	c.Log.Info("stopping cron jobs")
	c.Cron.Stop()
	return nil
}

func LogErr(l *zap.Logger, err error) {
	appErr, ok := apperrors.IsAppError(err)
	fmt.Println(err)
	if !ok {
		l.Error("cron job failed", zap.Error(err))
		return
	}

	if appErr.BaseError != nil {
		l.Error(
			appErr.Message,
			zap.String("err", appErr.BaseError.Error()),
			zap.String("occurred", appErr.Path()),
		)
	} else {
		l.Error(
			appErr.Message,
			zap.String("occurred", appErr.Path()),
		)
	}
}
