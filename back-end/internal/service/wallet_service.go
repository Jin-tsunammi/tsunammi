package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mm/internal/client/jito"
	"mm/internal/client/raydium"
	"slices"

	"mm/internal/client/solanarpc"
	"mm/internal/crypto"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/pkg/apperrors"
	repo "mm/pkg/repository"
	"sync"

	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"

	"github.com/gagliardetto/solana-go"
	"go.uber.org/fx"
)

type WalletService struct {
	WalletEncryptor       crypto.Encryptor `name:"wallet_encryptor"`
	WalletRepository      *repository.WalletRepository
	ProjectRepository     *repository.ProjectRepository
	UserHistoryRepository *repository.UserHistoryRepository
	TransactionManager    *repo.TransactionManager
	SolanaRPC             solanarpc.SolanaRPC
	JitoRPC               *jito.Client
	Raydium               *raydium.Client
	KeyStorage            *secret.KeyStorage
}

type walletServiceParams struct {
	fx.In

	WalletEncryptor       crypto.Encryptor `name:"wallet_encryptor"`
	WalletRepository      *repository.WalletRepository
	ProjectRepository     *repository.ProjectRepository
	UserHistoryRepository *repository.UserHistoryRepository
	TransactionManager    *repo.TransactionManager
	SolanaRPC             solanarpc.SolanaRPC
	JitoRPC               *jito.Client
	Raydium               *raydium.Client
	KeyStorage            *secret.KeyStorage
}

func NewWalletService(p walletServiceParams) *WalletService {
	return &WalletService{
		WalletEncryptor:       p.WalletEncryptor,
		WalletRepository:      p.WalletRepository,
		ProjectRepository:     p.ProjectRepository,
		UserHistoryRepository: p.UserHistoryRepository,
		TransactionManager:    p.TransactionManager,
		SolanaRPC:             p.SolanaRPC,
		JitoRPC:               p.JitoRPC,
		Raydium:               p.Raydium,
		KeyStorage:            p.KeyStorage,
	}
}

func (s *WalletService) GenerateSolanaWallets(ctx context.Context, req *model.GenerateWalletsReq, userID uint64) ([]model.Wallet, error) {
	projects, err := s.ProjectRepository.FetchAllByIDs(ctx, req.ProjectIDs, userID)

	if err != nil {
		return nil, apperrors.Internal("failed to fetch projects", err)
	}

	history := make([]model.UserHistory, 0, len(projects))

	for _, project := range projects {
		message := fmt.Sprintf("created %d wallets with %s", req.Count, project.Name)

		event := model.NewUserHistory(userID, model.ActionWalletsBatchAdd, message)
		history = append(history, *event)
	}

	resCh := make(chan model.Wallet, req.Count)
	errCh := make(chan error, req.Count)

	var wg sync.WaitGroup
	wg.Add(req.Count)

	semaphore := make(chan struct{}, 100)

	for i := 0; i < req.Count; i++ {
		go func() {
			defer wg.Done()

			semaphore <- struct{}{}

			defer func() {
				<-semaphore
			}()

			wallet := solana.NewWallet()

			w := model.Wallet{
				UserID:     userID,
				PublicKey:  wallet.PublicKey().String(),
				PrivateKey: wallet.PrivateKey.String(),
				ProjectIDs: req.ProjectIDs,
				Status:     model.WalletStatusCreationPending,
			}

			resCh <- w
		}()
	}

	go func() {
		wg.Wait()
		close(resCh)
		close(errCh)
	}()

	for err = range errCh {
		if err != nil {
			return nil, apperrors.Internal("failed to generate wallets", err)
		}
	}

	res := make([]model.Wallet, 0, req.Count)
	for w := range resCh {
		res = append(res, w)
	}

	if len(res) == 0 {
		return nil, nil
	}

	res, err = s.consistentlySaveWallets(ctx, res, history)

	if err != nil {
		return nil, apperrors.Internal("failed to save wallets", err)
	}

	return res, nil
}

func (s *WalletService) MonitorSolanaWallets(
	ctx context.Context,
	req *model.MonitorWalletsReq,
) ([]model.MonitorWalletsResp, error) {

	wallets, err := s.WalletRepository.GetWallets(ctx, req.WalletIDs)
	if err != nil {
		return nil, apperrors.Internal("failed to get wallets", err)
	}

	var monitorWallets = make([]model.MonitorWalletsResp, len(wallets))
	for i, w := range wallets {
		pubKey, err := solana.PublicKeyFromBase58(w.PublicKey)
		if err != nil {
			return nil, apperrors.BadRequest("invalid solana pub key", err)
		}

		balance, err := s.SolanaRPC.GetWalletBalance(ctx, pubKey)
		if err != nil {
			return nil, err
		}

		before, err := solana.SignatureFromBase58(req.Before)
		if err != nil && req.Before != "" {
			return nil, apperrors.BadRequest("invalid 'after' transaction signature", err)
		}
		transactions, err := s.SolanaRPC.GetWalletTransactions(ctx, &solanarpc.GetTransactionsReq{
			Address: pubKey,
			Limit:   req.PageSize,
			Before:  before,
		})
		if err != nil {
			return nil, apperrors.Internal("failed to get wallet transactions", err)
		}

		monitorWallets[i] = model.MonitorWalletsResp{
			WalletID:     w.ID,
			PublicKey:    w.PublicKey,
			Balance:      balance,
			Transactions: transactions,
		}
	}

	return monitorWallets, nil
}

func (s *WalletService) FetchPrivateKeysByProjectID(ctx context.Context, projectID uint64, userID uint64) ([]model.Wallet, error) {
	projectWithWallets, err := s.ProjectRepository.FetchProjectWithWalletsByID(ctx, projectID, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch project with wallets", err)
	}

	if projectWithWallets == nil {
		return nil, apperrors.NotFound("project not found", nil)
	}

	for index := range projectWithWallets.Wallets {
		wallet := &projectWithWallets.Wallets[index]

		privateKey, err := s.KeyStorage.Get(ctx, userID, wallet.PublicKey)
		if err != nil {
			return nil, err
		}
		wallet.PrivateKey = privateKey
	}

	return projectWithWallets.Wallets, nil
}

func (s *WalletService) FetchPrivateKeyByWalletID(ctx context.Context, walletID, userID uint64) (string, error) {

	w, err := s.WalletRepository.FetchByIDAndUserID(ctx, walletID, userID)

	if err != nil {
		return "", apperrors.Internal("failed to get private key", err)
	}

	privateKey, err := s.KeyStorage.Get(ctx, userID, w.PublicKey)

	if err != nil {
		return "", apperrors.Internal("failed to get secret", err)
	}

	return privateKey, nil
}

func (s *WalletService) ImportWallets(ctx context.Context, req *model.ImportWalletsReq, userID uint64) ([]model.Wallet, error) {

	projects, err := s.ProjectRepository.FetchAllByIDs(ctx, req.ProjectIDs, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch projects", err)
	}

	if len(projects) == 0 {
		return nil, apperrors.BadRequest("no projects found", nil)
	}

	parsedWallets := make([]model.Wallet, 0, len(req.PrivateKeys))
	publicKeys := make([]string, 0, len(req.PrivateKeys))

	for index, privateKeyBase58 := range req.PrivateKeys {
		privateKey, err := solana.PrivateKeyFromBase58(privateKeyBase58)
		if err != nil {
			return nil, apperrors.BadRequest(fmt.Sprintf("invalid private key №%d", index+1), err)
		}

		pubKey := privateKey.PublicKey().String()
		parsedWallets = append(parsedWallets, model.Wallet{
			UserID:     userID,
			PublicKey:  pubKey,
			PrivateKey: privateKeyBase58,
			ProjectIDs: req.ProjectIDs,
			Status:     model.WalletStatusImportPending,
		})
		publicKeys = append(publicKeys, pubKey)
	}

	existingWallets, err := s.WalletRepository.FetchWalletsByPublicKeys(ctx, publicKeys, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch existing wallets", err)
	}

	existingMap := make(map[string]model.Wallet, len(existingWallets))
	for _, ew := range existingWallets {
		existingMap[ew.PublicKey] = ew
	}

	var walletsToInsert []model.Wallet
	var walletsToLink []model.Wallet
	var alreadyFullyLinked []model.Wallet
	var finalWallets []model.Wallet

	for _, w := range parsedWallets {
		if extWallet, exists := existingMap[w.PublicKey]; exists {
			missingProjectIDs := slices.DeleteFunc(slices.Clone(req.ProjectIDs), func(e uint64) bool {
				return slices.Contains(extWallet.ProjectIDs, e)
			})

			if len(missingProjectIDs) > 0 {
				extWallet.ProjectIDs = missingProjectIDs
				walletsToLink = append(walletsToLink, extWallet)
			} else {
				alreadyFullyLinked = append(alreadyFullyLinked, extWallet)
				finalWallets = append(finalWallets, extWallet)
			}
		} else {
			walletsToInsert = append(walletsToInsert, w)
		}
	}

	history := make([]model.UserHistory, 0, len(projects))
	for _, project := range projects {
		message := fmt.Sprintf("imported %d new, linked %d existing wallets to %s", len(walletsToInsert), len(walletsToLink), project.Name)
		event := model.NewUserHistory(userID, model.ActionWalletsBatchAdd, message)
		history = append(history, *event)
	}

	if len(walletsToInsert) > 0 {
		savedNew, err := s.consistentlySaveWallets(ctx, walletsToInsert, history)
		if err != nil {
			return nil, apperrors.Internal("failed to save new wallets", err)
		}
		finalWallets = append(finalWallets, savedNew...)
	} else if len(history) > 0 {
		err = s.UserHistoryRepository.CreateAll(ctx, history)
		if err != nil {
			return nil, err
		}
	}

	if len(walletsToLink) > 0 {
		err = s.WalletRepository.SaveProjectWallets(ctx, walletsToLink)
		if err != nil {
			return nil, apperrors.Internal("failed to save project wallets links", err)
		}
		finalWallets = append(finalWallets, walletsToLink...)
	}

	if len(alreadyFullyLinked) > 0 {
		return nil, apperrors.AlreadyExist(fmt.Sprintf("%d wallets were not imported (already exist)", len(alreadyFullyLinked)), nil)
	}

	return finalWallets, nil
}

func (s *WalletService) consistentlySaveWallets(ctx context.Context, wallets []model.Wallet, history []model.UserHistory) ([]model.Wallet, error) {
	const errGroupLimit = 100

	res, err := s.WalletRepository.SaveSolanaWallets(ctx, wallets)

	if err != nil {
		return nil, err
	}

	eg, errctx := errgroup.WithContext(ctx)

	eg.SetLimit(errGroupLimit)

	for _, wallet := range res {
		eg.Go(func() error {

			return s.KeyStorage.
				Save(ctx,
					secret.KeySecret{
						UserID:     wallet.UserID,
						PublicKey:  wallet.PublicKey,
						PrivateKey: wallet.PrivateKey,
						CreatedAt:  wallet.CreatedAt,
					},
				)

		})
	}

	if err = eg.Wait(); err != nil {
		eg, errctx = errgroup.WithContext(ctx)

		eg.SetLimit(errGroupLimit)

		for _, wallet := range res {
			eg.Go(func() error {
				return s.KeyStorage.Delete(errctx, wallet.UserID, wallet.PublicKey)
			})
		}

		if err1 := eg.Wait(); err1 != nil {
			return nil, errors.Join(err, err1)
		}

		ids := make([]uint64, 0, len(res))

		for _, wallet := range res {
			ids = append(ids, wallet.ID)
		}

		if err1 := s.WalletRepository.DeleteWalletsByIds(ctx, ids); err1 != nil {
			return nil, errors.Join(err, err1)
		}

		return nil, err
	}

	for i := 0; i < len(res); i++ {
		wallet := &res[i]
		wallet.Status = model.WalletStatusSuccess
	}

	_ = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {

			err = s.WalletRepository.WithTx(tx).UpdateAllStatus(ctx, res)
			if err != nil {
				return err
			}

			err = s.UserHistoryRepository.WithTx(tx).CreateAll(ctx, history)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return res, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, id, userID uint64) error {
	publicKey, err := s.WalletRepository.Delete(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.BadRequest("unknown wallet")
		}
		return apperrors.Internal("failed to delete wallet", err)
	}

	err = s.KeyStorage.Delete(ctx, userID, publicKey)
	if err != nil {
		return apperrors.Internal("failed to delete wallet", err)
	}

	return nil
}
