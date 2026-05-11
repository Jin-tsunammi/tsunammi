package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/cache"
	"mm/internal/storage/repository"
	"mm/pkg/apperrors"
	"mm/pkg/solutil"
	"slices"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

type ProjectService struct {
	ProjectRepository *repository.ProjectRepository
	ProjectStorage    cache.ProjectStorage
	Solana            solanarpc.SolanaRPC
	RateCache         cache.RateStorage
	log               *zap.Logger
}

func NewProjectService(
	projectRepository *repository.ProjectRepository,
	solana solanarpc.SolanaRPC,
	rateCache cache.RateStorage,
	projectStorage cache.ProjectStorage,
	log *zap.Logger,
) *ProjectService {
	return &ProjectService{
		ProjectStorage: projectStorage, ProjectRepository: projectRepository, Solana: solana, RateCache: rateCache, log: log}
}

func (s *ProjectService) CreateProject(ctx context.Context, req model.CreateProjectReq, userID uint64) (*model.Project, error) {
	project := &model.Project{
		UserID: userID,
		Name:   req.Name,
	}

	err := s.ProjectRepository.Create(ctx, project)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperrors.AlreadyExist("project with this name already exists")
		}
		return nil, apperrors.Internal("failed to create project", err)
	}

	return project, nil
}

func (s *ProjectService) FetchProjectsWithWallets(ctx context.Context, userID uint64, page, pageSize int, sortBy string, sortDesc bool) (*model.ProjectsWithPaginationResponse, error) {
	projects, total, err := s.ProjectRepository.FindAllByUserID(ctx, userID, page, pageSize, sortDesc)

	if err != nil {
		return nil, apperrors.Internal("failed to fetch all projects", err)
	}

	for i := 0; i < len(projects); i++ {
		balanceSOL, balanceUSD, err := s.fetchProjectWalletBalanceWithTotal(ctx, (&projects[i]).Wallets)

		if err != nil {
			return nil, err
		}

		projects[i].BalanceSOL = balanceSOL
		projects[i].TotalBalanceSOL = balanceSOL
		projects[i].TotalBalanceUSD = balanceUSD
	}

	projectsResponses := make([]model.ProjectWithWalletsResponse, 0, len(projects))

	for i := 0; i < len(projects); i++ {
		projectsResponse := projectToProjectResponse(&projects[i])
		projectsResponses = append(projectsResponses, *projectsResponse)
	}

	result, err := s.enrichProjectResponsesWithTokens(ctx, projectsResponses)

	if err != nil {
		return nil, err
	}

	return &model.ProjectsWithPaginationResponse{
		Projects: result,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *ProjectService) FetchProjectsWithWalletsWithoutBalance(ctx context.Context, userID uint64, page, pageSize int, sortBy string, sortDesc bool) (*model.ProjectsWithoutBalanceWithPaginationResponse, error) {
	projects, total, err := s.ProjectRepository.FindAllByUserID(ctx, userID, page, pageSize, sortDesc)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch all projects", err)
	}

	projectsResponses := make([]model.ProjectWithWalletsWithoutBalance, 0, len(projects))

	for _, project := range projects {
		walletResponse := make([]model.WalletWithoutBalance, 0, len(project.Wallets))
		for _, wallet := range project.Wallets {
			walletResponse = append(walletResponse, model.WalletWithoutBalance{
				ID:        wallet.ID,
				PublicKey: wallet.PublicKey,
				CreatedAt: wallet.CreatedAt,
			})
		}
		projectsResponses = append(projectsResponses, model.ProjectWithWalletsWithoutBalance{
			Wallets:  walletResponse,
			LastSync: time.Now(),
			Project: model.Project{
				ID:          project.ID,
				UserID:      project.UserID,
				WalletCount: project.WalletCount,
				CreatedAt:   project.CreatedAt,
				Name:        project.Name,
			},
		})
	}

	return &model.ProjectsWithoutBalanceWithPaginationResponse{
		Projects: projectsResponses,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *ProjectService) FetchProjectWithWalletsByMint(ctx context.Context, userID uint64, page, pageSize int, sortBy string, sortDesc bool, mint solana.PublicKey) (*model.ProjectsWithMintPaginationResponse, error) {
	projects, total, err := s.ProjectRepository.FindAllByUserID(ctx, userID, page, pageSize, sortDesc)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch all projects", err)
	}

	mintKeys := make([]solana.PublicKey, 0, len(projects)*50)
	errs := make([]error, len(projects))

	for indexP, project := range projects {
		mints := make([]solana.PublicKey, len(project.Wallets))
		walletErrs := make([]error, len(project.Wallets))

		for indexW, wallet := range project.Wallets {
			publicKey, err := solana.PublicKeyFromBase58(wallet.PublicKey)

			if err != nil {
				walletErrs[indexW] = err
				continue
			}

			var tokenAddress solana.PublicKey
			if mint.Equals(solana.SolMint) {
				tokenAddress = publicKey
			} else {
				tokenAddress, _, err = solutil.FindAssociatedTokenAddressWithProgram(publicKey, mint, solana.TokenProgramID)
				if err != nil {

					walletErrs[indexW] = err
					continue
				}
			}

			mints[indexW] = tokenAddress
		}

		errs[indexP] = errors.Join(walletErrs...)
		mintKeys = append(mintKeys, mints...)
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}

	if len(mintKeys) == 0 {
		return &model.ProjectsWithMintPaginationResponse{
			Projects: []model.ProjectWithMintWalletsResponse{},
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		}, nil
	}

	results, err := s.Solana.GetMultipleAccountsWithNoLimits(ctx, mintKeys...)
	if err != nil {
		return nil, err
	}

	accounts := make([]*rpc.Account, 0, len(results)*solanarpc.MaxGetMultipleAccountsKeys)

	for _, result := range results {
		accounts = append(accounts, result.Value...)
	}

	projectsResponses := make([]model.ProjectWithMintWalletsResponse, 0, len(projects))

	var rate float64
	if mint.Equals(solana.SolMint) {
		rate = 1
	} else {
		res, err := s.RateCache.GetSOLRates(ctx, mint)
		if err != nil {
			return nil, err
		}
		rate = res[mint]
	}

	decimal, err := s.RateCache.GetDecimal(ctx, mint)
	if err != nil {
		return nil, err
	}

	accountIdx := 0

	for i := 0; i < len(projects); i++ {
		project := &projects[i]
		wallets := make([]model.WalletMintResponse, 0, len(project.Wallets))

		totalBalance := 0.0
		totalBalanceSOL := 0.0

		for j := 0; j < len(project.Wallets); j++ {
			wallet := &project.Wallets[j]

			var account *rpc.Account
			if accountIdx < len(accounts) {
				account = accounts[accountIdx]
				accountIdx++
			}

			var balance float64

			if account != nil {
				accountData := account.Data.GetBinary()
				if mint.Equals(solana.SolMint) {
					lamports := account.Lamports
					if lamports > 1 {
						balance = solanarpc.FromAtomicUnit(lamports, solana.SolDecimals)
					}
				} else {
					tokenAccount := token.Account{}
					err = tokenAccount.UnmarshalWithDecoder(bin.NewBinDecoder(accountData))
					if err != nil {
						return nil, err
					}

					balance = solanarpc.FromAtomicUnit(tokenAccount.Amount, decimal)
				}
			}

			balanceSOL := balance * rate

			totalBalance += balance
			totalBalanceSOL += balanceSOL

			wallets = append(wallets, model.WalletMintResponse{
				ID:               wallet.ID,
				PublicKey:        wallet.PublicKey,
				TokensBalance:    balance,
				TokensBalanceSOL: balanceSOL,
				CreatedAt:        wallet.CreatedAt,
			})
		}
		projectsResponses = append(projectsResponses, model.ProjectWithMintWalletsResponse{
			ID:              project.ID,
			Name:            project.Name,
			UserID:          project.UserID,
			Wallets:         wallets,
			LastSync:        time.Now(),
			TotalBalance:    totalBalance,
			TotalBalanceSOL: totalBalanceSOL,
			WalletCount:     project.WalletCount,
		})
	}

	return &model.ProjectsWithMintPaginationResponse{
		Projects: projectsResponses,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *ProjectService) FetchProjectWithWalletByIDAndMint(ctx context.Context, id, userID uint64, mint solana.PublicKey) (*model.ProjectWithMintWalletsResponse, error) {
	project, err := s.ProjectRepository.FetchProjectWithWalletsByID(ctx, id, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch all projects", err)
	}

	mintKeys := make([]solana.PublicKey, len(project.Wallets))
	errs := make([]error, len(project.Wallets))

	for indexW, wallet := range project.Wallets {
		publicKey, err := solana.PublicKeyFromBase58(wallet.PublicKey)

		if err != nil {
			errs[indexW] = err
			continue
		}

		var tokenAddress solana.PublicKey
		if mint.Equals(solana.SolMint) {
			fmt.Println("solana.SolMint")
			tokenAddress = publicKey
		} else {
			tokenAddress, _, err = solutil.FindAssociatedTokenAddressWithProgram(publicKey, mint, solana.TokenProgramID)
			if err != nil {
				errs[indexW] = err
				continue
			}
		}

		mintKeys[indexW] = tokenAddress
	}

	if err = errors.Join(errs...); err != nil {
		return nil, err
	}

	results, err := s.Solana.GetMultipleAccountsWithNoLimits(ctx, mintKeys...)
	if err != nil {
		return nil, err
	}

	accounts := make([]*rpc.Account, 0, len(results)*solanarpc.MaxGetMultipleAccountsKeys)

	for _, result := range results {
		accounts = append(accounts, result.Value...)
	}

	var rate float64
	if mint.Equals(solana.SolMint) {
		rate = 1
	} else {
		res, err := s.RateCache.GetSOLRates(ctx, mint)
		if err != nil {
			return nil, err
		}
		rate = res[mint]
	}

	decimal, err := s.RateCache.GetDecimal(ctx, mint)
	if err != nil {
		return nil, err
	}

	wallets := make([]model.WalletMintResponse, 0, len(project.Wallets))

	totalBalance := 0.0
	totalBalanceSOL := 0.0

	for j := 0; j < len(project.Wallets); j++ {
		wallet := &project.Wallets[j]
		account := accounts[j]

		var balance float64

		if account != nil {

			accountData := account.Data.GetBinary()
			if mint.Equals(solana.SolMint) {
				lamports := accounts[j].Lamports
				if lamports > 1 {
					balance = solanarpc.FromAtomicUnit(lamports, solana.SolDecimals)
				}
			} else {
				tokenAccount := token.Account{}
				err = tokenAccount.UnmarshalWithDecoder(bin.NewBinDecoder(accountData))
				if err != nil {
					return nil, err
				}

				balance = solanarpc.FromAtomicUnit(tokenAccount.Amount, decimal)
			}
		}

		balanceSOL := balance * rate

		totalBalance += balance
		totalBalanceSOL += balanceSOL

		wallets = append(wallets, model.WalletMintResponse{
			ID:               wallet.ID,
			PublicKey:        wallet.PublicKey,
			TokensBalance:    balance,
			TokensBalanceSOL: balance,
			CreatedAt:        wallet.CreatedAt,
		})
	}
	return &model.ProjectWithMintWalletsResponse{
		ID:              project.ID,
		Name:            project.Name,
		UserID:          project.UserID,
		Wallets:         wallets,
		LastSync:        time.Now(),
		TotalBalance:    totalBalance,
		TotalBalanceSOL: totalBalanceSOL,
		WalletCount:     project.WalletCount,
	}, nil
}

func (s *ProjectService) FetchCachedProjectWithWalletsByID(ctx context.Context, id, userID uint64) (*model.ProjectWithWalletsResponse, error) {
	response, err := s.ProjectStorage.Get(ctx, id, userID)
	if err != nil {
		response, err = s.FetchProjectWithWalletsByID(ctx, id, userID)
		if err != nil {
			return nil, err
		}

		s.ProjectStorage.Set(ctx, id, userID, response)
	}

	return response, nil
}

func (s *ProjectService) FetchProjectWithWalletsByID(ctx context.Context, id, userID uint64) (*model.ProjectWithWalletsResponse, error) {
	project, err := s.ProjectRepository.FetchProjectWithWalletsByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	balanceSOL, balanceUSD, err := s.fetchProjectWalletBalanceWithTotal(ctx, project.Wallets)

	if err != nil {
		return nil, err
	}

	project.BalanceSOL = balanceSOL
	project.TotalBalanceSOL = balanceSOL
	project.TotalBalanceUSD = balanceUSD

	projectResponse := projectToProjectResponse(project)

	projectResponse.LastSync = time.Now()

	result, err := s.enrichProjectResponsesWithTokens(ctx, []model.ProjectWithWalletsResponse{*projectResponse})

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, apperrors.NotFound("project not found", nil)
	}

	projectRes := result[0]

	s.ProjectStorage.Set(ctx, id, userID, &projectRes)

	return &projectRes, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id, userID uint64) error {
	err := s.ProjectRepository.DeleteProjectByIDAndUserID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("project not found", err)
		}
		return apperrors.Internal("failed to delete project", err)
	}
	return nil
}

func (s *ProjectService) FetchProjects(ctx context.Context, userID uint64) ([]model.Project, error) {
	projectsRes, _, err := s.ProjectRepository.FindAllByUserID(ctx, userID, 0, 0, false)

	projects := make([]model.Project, 0, len(projectsRes))

	for _, project := range projectsRes {
		projects = append(projects, model.Project{
			ID:          project.ID,
			UserID:      project.UserID,
			WalletCount: project.WalletCount,
			CreatedAt:   project.CreatedAt,
			Name:        project.Name,
		})
	}

	if err != nil {
		return nil, apperrors.Internal("failed to fetch projects", err)
	}
	return projects, nil

}

func (s *ProjectService) EditProfile(ctx context.Context, id, userID uint64, req model.EditProjectReq) error {
	return s.ProjectRepository.UpdateByID(ctx, id, userID, req)
}

func (s *ProjectService) fetchProjectWalletBalanceWithTotal(ctx context.Context, wallets []model.Wallet) (totalBalanceSOL float64, totalBalanceUSD float64, err error) {
	rate, err := s.RateCache.GetUSDRate(ctx, solana.SolMint)
	if err != nil {
		return 0.0, 0.0, apperrors.Internal("failed to get rate", err)
	}

	return fetchWalletsBalanceWithTotal(ctx, wallets, s.Solana, rate)
}

type walletResult struct {
	key      string
	accounts []*rpc.TokenAccount
}

func (s *ProjectService) enrichProjectResponsesWithTokens(ctx context.Context, projectsResponses []model.ProjectWithWalletsResponse) ([]model.ProjectWithWalletsResponse, error) {
	if len(projectsResponses) == 0 {
		return projectsResponses, nil
	}

	uniqueWalletKeys := make(map[string]struct{})
	for _, project := range projectsResponses {
		for _, wallet := range project.Wallets {
			if wallet.BalanceSOL > 0.000001 {
				uniqueWalletKeys[wallet.PublicKey] = struct{}{}
			}
		}
	}

	if len(uniqueWalletKeys) == 0 {
		return projectsResponses, nil
	}

	walletTokenAccounts := make(map[string][]*token.Account, len(uniqueWalletKeys))
	uniqueTokenMints := make(map[solana.PublicKey]struct{}, len(uniqueWalletKeys))

	inCh := make(chan solana.PublicKey)
	accountsCh := s.processWallets(ctx, inCh)

	for k := range uniqueWalletKeys {
		key, err := solana.PublicKeyFromBase58(k)
		if err != nil {
			return nil, err
		}

		inCh <- key
	}
	close(inCh)

	for r := range accountsCh {
		for _, a := range r.accounts {
			data := a.Account.Data.GetBinary()
			if data == nil {
				continue
			}

			var tokenAccount token.Account
			if err := tokenAccount.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
				return nil, err
			}

			uniqueTokenMints[tokenAccount.Mint] = struct{}{}
			walletTokenAccounts[r.key] = append(walletTokenAccounts[r.key], &tokenAccount)
		}
	}

	uniqueMints := slices.Collect(maps.Keys(uniqueTokenMints))

	decimals, err := s.RateCache.GetDecimals(ctx, uniqueMints...)
	if err != nil {
		return nil, err
	}

	usdRates, err := s.RateCache.GetUSDRates(ctx, uniqueMints...)
	if err != nil {
		return nil, err
	}

	solRates, err := s.RateCache.GetSOLRates(ctx, uniqueMints...)
	if err != nil {
		return nil, err
	}

	rent, err := s.Solana.GetATARentExemption(ctx)
	if err != nil {
		return nil, err
	}

	for i := range projectsResponses {
		project := &projectsResponses[i]
		var projectRent uint64 = 0

		projectTotalBalanceSOL := 0.0
		projectTotalBalanceUSD := 0.0

		for index := range project.Wallets {
			totalBalanceSol := 0.0
			totalBalanceUsd := 0.0

			wallet := &project.Wallets[index]

			tokens := make([]model.WalletToken, 0, len(walletTokenAccounts[wallet.PublicKey]))
			walletTokens := walletTokenAccounts[wallet.PublicKey]

			var walletRent uint64 = 0

			for _, walletToken := range walletTokens {
				decimal := decimals[walletToken.Mint]
				usdRate := usdRates[walletToken.Mint]
				solRate := solRates[walletToken.Mint]
				amount := solanarpc.FromAtomicUnit(walletToken.Amount, decimal)

				balanceUSD := amount * usdRate
				balanceSOL := amount * solRate

				tokens = append(tokens, model.WalletToken{
					Mint:       walletToken.Mint.String(),
					Balance:    amount,
					BalanceUSD: balanceUSD,
					BalanceSOL: balanceSOL,
				})
				walletRent += rent
				totalBalanceSol += balanceSOL
				totalBalanceUsd += balanceUSD
			}

			projectRent += walletRent
			wallet.Tokens = tokens
			wallet.TokensBalanceSOL = totalBalanceSol
			wallet.TokensBalanceUSD = totalBalanceUsd
			projectTotalBalanceSOL += totalBalanceSol
			projectTotalBalanceUSD += totalBalanceUsd
			wallet.Rent = solanarpc.FromAtomicUnit(walletRent, solana.SolDecimals)
		}

		project.LastSync = time.Now()
		project.TotalBalanceSOL += projectTotalBalanceSOL
		project.TotalBalanceUSD += projectTotalBalanceUSD
		project.RentTotal = solanarpc.FromAtomicUnit(projectRent, solana.SolDecimals)
	}

	return projectsResponses, nil
}

func (s *ProjectService) processWallets(ctx context.Context, in <-chan solana.PublicKey) chan walletResult {
	resCh := make(chan walletResult)
	semaphore := make(chan struct{}, 5)
	var wg sync.WaitGroup
	go func() {
		for {
			v, ok := <-in
			if !ok {
				break
			}

			wg.Go(func() {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				res, err := s.Solana.GetTokenAccountsByOwner(ctx, v, &rpc.GetTokenAccountsConfig{ProgramId: &solana.TokenProgramID}, &rpc.GetTokenAccountsOpts{Encoding: solana.EncodingBase64, Commitment: rpc.CommitmentConfirmed})
				if err != nil {
					s.log.Error("failed to get token accounts", zap.Error(err))
					return
				}

				resCh <- walletResult{
					key:      v.String(),
					accounts: res.Value,
				}
			})

			wg.Go(func() {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				res, err := s.Solana.GetTokenAccountsByOwner(ctx, v, &rpc.GetTokenAccountsConfig{ProgramId: &solana.Token2022ProgramID}, &rpc.GetTokenAccountsOpts{Encoding: solana.EncodingBase64, Commitment: rpc.CommitmentConfirmed})
				if err != nil {
					s.log.Error("failed to get token accounts", zap.Error(err))
					return
				}

				resCh <- walletResult{
					key:      v.String(),
					accounts: res.Value,
				}
			})
		}

		wg.Wait()
		close(resCh)
	}()

	return resCh
}
