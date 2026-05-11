package service

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mm/config"
	"mm/internal/client/helius"
	"mm/internal/client/jito"
	"mm/internal/client/lighthouse"
	"mm/internal/client/pumpfun/bonding"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	"mm/pkg/apperrors"
	"time"

	"go.uber.org/zap"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/google/uuid"
	"resty.dev/v3"
)

const (
	pumpfunComputeUnitLimit uint32 = 200_000

	pumpfunTokenDecimals                  uint8  = 6
	pumpfunCreationFeeLamports            uint64 = 8_100_000
	pumpfunUserVolumeAccumulatorSize      uint64 = 90
	pumpfunTxFeeReserveLamports           uint64 = 10_000
	pumpfunToken2022ATARentBufferLamports uint64 = 100_000
)

type PumpfunLaunchService struct {
	rpc          solanarpc.SolanaRPC
	resty        *resty.Client
	lighthouse   *lighthouse.Client
	repo         *repository.PumpfunLaunchRepository
	walletRepo   *repository.WalletRepository
	keyStorage   *secret.KeyStorage
	jitoClient   *jito.Client
	heliusClient *helius.Client
	cfg          *config.Config
	logger       *zap.Logger
}

func NewPumpfunLaunchService(
	c *resty.Client,
	rpc solanarpc.SolanaRPC,
	lh *lighthouse.Client,
	repo *repository.PumpfunLaunchRepository,
	walletRepo *repository.WalletRepository,
	keyStorage *secret.KeyStorage,
	jitoClient *jito.Client,
	heliusClient *helius.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *PumpfunLaunchService {
	return &PumpfunLaunchService{
		rpc:          rpc,
		resty:        c,
		lighthouse:   lh,
		repo:         repo,
		walletRepo:   walletRepo,
		keyStorage:   keyStorage,
		jitoClient:   jitoClient,
		heliusClient: heliusClient,
		cfg:          cfg,
		logger:       logger,
	}
}

func (s *PumpfunLaunchService) fetchComputeBudgetInstrs(accounts []string) ([]solana.Instruction, error) {
	instrs, _, err := s.fetchComputeBudgetInstrsWithFee(accounts)
	return instrs, err
}

func (s *PumpfunLaunchService) fetchComputeBudgetInstrsWithFee(accounts []string) ([]solana.Instruction, uint64, error) {
	levels, err := s.heliusClient.GetPriorityFeeEstimate(accounts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get priority fee estimate: %w", err)
	}
	microLamports := uint64(levels.VeryHigh)
	estimatedPriorityFeeLamports := estimatePriorityFeeLamports(microLamports, pumpfunComputeUnitLimit)
	return []solana.Instruction{
		computebudget.NewSetComputeUnitLimitInstruction(pumpfunComputeUnitLimit).Build(),
		computebudget.NewSetComputeUnitPriceInstruction(microLamports).Build(),
	}, estimatedPriorityFeeLamports, nil
}

func estimatePriorityFeeLamports(microLamports uint64, computeUnitLimit uint32) uint64 {
	return (microLamports * uint64(computeUnitLimit)) / 1_000_000
}

func (s *PumpfunLaunchService) fetchWalletBuyComputeBudgetInstrs(
	ctx context.Context,
	userID uint64,
	mint solana.PublicKey,
	walletBuys []model.WalletBuyConfig,
) ([]solana.Instruction, uint64, error) {
	if len(walletBuys) == 0 {
		return []solana.Instruction{}, 0, nil
	}

	walletIDs := make([]uint64, 0, len(walletBuys))
	for _, wb := range walletBuys {
		walletIDs = append(walletIDs, wb.WalletID)
	}

	wallets, err := s.walletRepo.FetchWalletsByIdsAndUserID(ctx, walletIDs, userID)
	if err != nil {
		return nil, 0, err
	}
	if len(wallets) != len(walletIDs) {
		return nil, 0, fmt.Errorf("one or more wallets do not belong to user")
	}

	accounts := make([]string, 0, 2+len(wallets))
	accounts = append(accounts, bonding.ProgramID.String(), mint.String())
	for _, wallet := range wallets {
		accounts = append(accounts, wallet.PublicKey)
	}

	return s.fetchComputeBudgetInstrsWithFee(accounts)
}

func (s *PumpfunLaunchService) EstimateCreateTx(ctx context.Context, userID uint64, req *model.PumpfunEstimateCreateRequest) (*model.PumpfunEstimateCreateResponse, error) {
	if len(req.WalletBuys) > 3 {
		return nil, apperrors.BadRequest("wallet_buys supports up to 3 buys", nil)
	}
	if req.BuyInSol < 0 {
		return nil, apperrors.BadRequest("buy_in_sol cannot be negative", nil)
	}
	ownerBuyLamports, err := validatePumpfunEstimateBuys(req)
	if err != nil {
		return nil, err
	}

	owner, err := solana.PublicKeyFromBase58(req.OwnerPublicKey)
	if err != nil {
		return nil, apperrors.BadRequest("invalid owner public key", err)
	}

	mint := solana.NewWallet().PublicKey()
	cbAccounts := []string{
		bonding.ProgramID.String(),
		owner.String(),
		mint.String(),
	}
	_, ownerPriorityFeeLamports, err := s.fetchComputeBudgetInstrsWithFee(cbAccounts)
	if err != nil {
		return nil, apperrors.Internal("failed to get compute budget instructions", err)
	}

	_, walletPriorityFeeLamports, err := s.fetchWalletBuyComputeBudgetInstrs(ctx, userID, mint, req.WalletBuys)
	if err != nil {
		return nil, apperrors.Internal("failed to get compute budget instructions for wallet buys", err)
	}

	pumpfunCommissionLamports, err := s.estimatePumpfunCommission(ctx, len(req.WalletBuys))
	if err != nil {
		return nil, err
	}
	priorityFeeLamports, ok := addLamports(ownerPriorityFeeLamports, walletPriorityFeeLamports)
	if !ok {
		return nil, apperrors.BadRequest("priority fees overflow uint64", nil)
	}

	quoteParams, err := bonding.FetchInitialBuyQuoteParams(ctx, s.rpc)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch initial buy quote params", err)
	}
	if quoteParams.VirtualTokenReserves == 0 {
		return nil, apperrors.Internal("initial virtual token reserves are zero")
	}

	resp := &model.PumpfunEstimateCreateResponse{
		CreationFeeSOL:       solanarpc.LamportsToSOL(int64(pumpfunCreationFeeLamports)),
		JitoTipSOL:           solanarpc.LamportsToSOL(int64(s.cfg.Jito.PumpfunCreateTip)),
		PriorityFeeSOL:       solanarpc.LamportsToSOL(int64(priorityFeeLamports)),
		PumpfunCommissionSOL: solanarpc.LamportsToSOL(int64(pumpfunCommissionLamports)),
	}

	if req.BuyInSol > 0 {
		ownerTokensOut, tradingFeeLamports, err := estimatePumpfunBuy(ownerBuyLamports, quoteParams)
		if err != nil {
			return nil, apperrors.BadRequest("failed to estimate owner buy", err)
		}
		if err := addTokensAndCommission(resp, ownerTokensOut, tradingFeeLamports); err != nil {
			return nil, err
		}
	}

	for _, wb := range req.WalletBuys {
		tokensOut, tradingFeeLamports, err := estimatePumpfunBuy(solanarpc.SOLToLamports(wb.AmountSol), quoteParams)
		if err != nil {
			return nil, apperrors.BadRequest(fmt.Sprintf("failed to estimate wallet %d buy", wb.WalletID), err)
		}
		if err := addTokensAndCommission(resp, tokensOut, tradingFeeLamports); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (s *PumpfunLaunchService) PrepareCreateTx(ctx context.Context, userID uint64, req *model.PumpfunPrepareCreateTxRequest) (*model.PumpfunPrepareCreateTxResponse, error) {
	reader, err := req.Logo.Open()
	if err != nil {
		return nil, apperrors.Internal("failed to save logo", err)
	}

	logoUri, err := s.lighthouse.UploadReader(ctx, fmt.Sprintf("%s_%s_logo", req.Name, req.Ticker), req.Logo.Size, reader)
	if err != nil {
		return nil, apperrors.Internal("failet to save logo", err)
	}

	metadata := model.Metadata{
		Name:        req.Name,
		Symbol:      req.Ticker,
		Description: req.Description,
		Image:       fmt.Sprintf("%s%s", "https://fit-aardvark-makzo.lighthouseweb3.xyz/ipfs/", logoUri.Hash),
		ShowName:    true,
		CreatedOn:   "https://pump.fun",
		Twitter:     req.Twitter,
		Telegram:    req.Telegram,
		Discord:     req.Discord,
		Website:     req.Website,
	}

	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, apperrors.Internal("failed to save metadata", err)
	}

	metadataResp, err := s.lighthouse.UploadReader(ctx, fmt.Sprintf("%s_%s_metadata", req.Name, req.Ticker), int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		return nil, apperrors.Internal("failed tto save metadata", err)
	}

	metadataUri := fmt.Sprintf("%s%s", "https://fit-aardvark-makzo.lighthouseweb3.xyz/ipfs/", metadataResp.Hash)

	mintPrivKey := solana.NewWallet().PrivateKey

	reader.Close()

	user, err := solana.PublicKeyFromBase58(req.OwnerPublicKey)
	if err != nil {
		return nil, apperrors.BadRequest("invalid owner public key", err)
	}

	cbAccounts := []string{
		bonding.ProgramID.String(),
		req.OwnerPublicKey,
		mintPrivKey.PublicKey().String(),
	}
	cbInstrs, ownerPriorityFeeLamports, err := s.fetchComputeBudgetInstrsWithFee(cbAccounts)
	if err != nil {
		return nil, apperrors.Internal("failed to get compute budget instructions", err)
	}

	tipLamports := s.cfg.Jito.PumpfunCreateTip

	tipAccount, err := s.jitoClient.GetTipAccount(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get jito tip account", err)
	}

	tipInstr := system.NewTransferInstruction(
		tipLamports,
		user,
		*tipAccount,
	).Build()

	createInstrs := make([]solana.Instruction, 0, len(cbInstrs)+1)
	createInstrs = append(createInstrs, cbInstrs...)
	createInstrs = append(createInstrs, tipInstr)

	blockHash, err := s.rpc.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get blockhash", err)
	}

	tx, err := bonding.BuildCreateV2Transaction(req.Name, req.Ticker, metadataUri, user, mintPrivKey, *blockHash, req.Mayhem, req.CashbackRewards, createInstrs...)
	if err != nil {
		return nil, apperrors.Internal("failed to form transaction", err)
	}

	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return nil, apperrors.Internal("failed to marshal tx", err)
	}

	if len(req.WalletBuys) > 3 {
		return nil, apperrors.BadRequest("wallet_buys supports up to 3 buys", nil)
	}
	if req.BuyInSol < 0 {
		return nil, apperrors.BadRequest("buy_in_sol cannot be negative", nil)
	}
	ownerBuyLamports := solanarpc.SOLToLamports(req.BuyInSol)

	walletCbInstrs, walletPriorityFeeLamports, err := s.fetchWalletBuyComputeBudgetInstrs(ctx, userID, mintPrivKey.PublicKey(), req.WalletBuys)
	if err != nil {
		return nil, apperrors.Internal("failed to get compute budget instructions for wallet buys", err)
	}

	if err := s.validateWalletBuys(ctx, userID, user, req.WalletBuys, ownerBuyLamports, tipLamports, ownerPriorityFeeLamports, walletPriorityFeeLamports); err != nil {
		return nil, err
	}

	buyBlockHash, err := s.rpc.GetLatestBlockhash(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get blockhash for buy tx", err)
	}

	var rawStoredBuyTx []byte
	if ownerBuyLamports > 0 {
		buyTx, err := bonding.BuildInitialBuyTransaction(ctx, s.rpc, mintPrivKey.PublicKey(), user, user, *buyBlockHash, ownerBuyLamports, cbInstrs...)
		if err != nil {
			return nil, apperrors.Internal("failed to build initial buy transaction", err)
		}

		rawBuyTx, err := buyTx.MarshalBinary()
		if err != nil {
			return nil, apperrors.Internal("failed to marshal buy tx", err)
		}

		rawStoredBuyTx = rawBuyTx
	} else {
		rawStoredBuyTx = nil
	}

	walletBuyTxs, err := s.buildWalletBuyTransactions(ctx, userID, mintPrivKey.PublicKey(), user, *buyBlockHash, req.WalletBuys, walletCbInstrs)
	if err != nil {
		return nil, err
	}

	txid := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	err = s.repo.Create(ctx, &model.PumpfunLaunch{
		ID:           txid,
		UserID:       userID,
		CreateTx:     rawTx,
		BuyTx:        rawStoredBuyTx,
		WalletBuyTxs: walletBuyTxs,
		MintPubkey:   mintPrivKey.PublicKey().String(),
		Signer:       req.OwnerPublicKey,
		Status:       model.PumpfunLaunchStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(time.Second * 90),
	})
	if err != nil {
		return nil, apperrors.Internal("failed to save pending launch", err)
	}

	resp := &model.PumpfunPrepareCreateTxResponse{
		ID:                txid,
		CreateTransaction: base64.StdEncoding.EncodeToString(rawTx),
		BuyTransaction:    base64.StdEncoding.EncodeToString(rawStoredBuyTx),
		MintPubkey:        mintPrivKey.PublicKey().String(),
	}

	return resp, nil
}

func (s *PumpfunLaunchService) validateWalletBuys(
	ctx context.Context,
	userID uint64,
	owner solana.PublicKey,
	walletBuys []model.WalletBuyConfig,
	ownerBuyLamports uint64,
	tipLamports uint64,
	ownerPriorityFeeLamports uint64,
	walletPriorityFeeLamports uint64,
) error {
	ownerRequiredLamports, ok := addLamports(ownerBuyLamports, pumpfunCreationFeeLamports, tipLamports, ownerPriorityFeeLamports, pumpfunTxFeeReserveLamports)
	if !ok {
		return apperrors.BadRequest("owner buy plus creation fee, tip and fees overflows uint64", nil)
	}

	ownerBalance, err := s.rpc.GetWalletBalance(ctx, owner)
	if err != nil {
		return apperrors.Internal("failed to get owner balance", err)
	}
	if solanarpc.SOLToLamports(ownerBalance) < ownerRequiredLamports {
		return apperrors.BadRequest("owner balance is not enough for buy, creation fee, tip and fees", nil)
	}

	if len(walletBuys) == 0 {
		return nil
	}

	ataRentLamports, err := s.rpc.GetATARentExemption(ctx)
	if err != nil {
		return apperrors.Internal("failed to get token account rent exemption", err)
	}
	userVolumeAccumulatorRentLamports, err := s.rpc.GetRentExemption(ctx, pumpfunUserVolumeAccumulatorSize)
	if err != nil {
		return apperrors.Internal("failed to get user volume accumulator rent exemption", err)
	}
	creatorVaultRentLamports, err := s.rpc.GetRentExemption(ctx, 0)
	if err != nil {
		return apperrors.Internal("failed to get creator vault rent exemption", err)
	}

	walletFeeReserveLamports, ok := addLamports(
		ataRentLamports,
		pumpfunToken2022ATARentBufferLamports,
		userVolumeAccumulatorRentLamports,
		creatorVaultRentLamports,
		walletPriorityFeeLamports,
		pumpfunTxFeeReserveLamports,
	)
	if !ok {
		return apperrors.BadRequest("wallet rent plus fee reserve overflows uint64", nil)
	}

	walletIDs := make([]uint64, 0, len(walletBuys))
	seen := make(map[uint64]struct{}, len(walletBuys))
	requiredByWalletID := make(map[uint64]uint64, len(walletBuys))
	var totalWalletBuy uint64
	for _, wb := range walletBuys {
		if wb.WalletID == 0 {
			return apperrors.BadRequest("wallet_id is required", nil)
		}
		if wb.AmountSol <= 0 {
			return apperrors.BadRequest("wallet buy amount must be greater than zero", nil)
		}
		amountLamports := solanarpc.SOLToLamports(wb.AmountSol)
		if amountLamports == 0 {
			return apperrors.BadRequest("wallet buy amount is too small", nil)
		}
		if _, ok := seen[wb.WalletID]; ok {
			return apperrors.BadRequest(fmt.Sprintf("duplicate wallet buy for wallet %d", wb.WalletID), nil)
		}
		seen[wb.WalletID] = struct{}{}
		walletIDs = append(walletIDs, wb.WalletID)

		requiredLamports, ok := addLamports(amountLamports, walletFeeReserveLamports)
		if !ok {
			return apperrors.BadRequest(fmt.Sprintf("wallet %d buy plus rent overflows uint64", wb.WalletID), nil)
		}
		requiredByWalletID[wb.WalletID] = requiredLamports

		if amountLamports > ownerBuyLamports || totalWalletBuy > ownerBuyLamports-amountLamports {
			return apperrors.BadRequest("total wallet buy amount cannot exceed owner buy amount", nil)
		}
		totalWalletBuy += amountLamports
	}

	wallets, err := s.walletRepo.FetchWalletsByIdsAndUserID(ctx, walletIDs, userID)
	if err != nil {
		return apperrors.Internal("failed to fetch wallets", err)
	}
	if len(wallets) != len(walletIDs) {
		return apperrors.BadRequest("one or more wallets do not belong to user", nil)
	}

	walletByID := make(map[uint64]model.Wallet, len(wallets))
	publicKeys := make([]solana.PublicKey, 0, len(wallets))
	for _, w := range wallets {
		pubkey, err := solana.PublicKeyFromBase58(w.PublicKey)
		if err != nil {
			return apperrors.Internal(fmt.Sprintf("failed to parse wallet %d public key", w.ID), err)
		}
		walletByID[w.ID] = w
		publicKeys = append(publicKeys, pubkey)
	}

	balances, err := s.rpc.GetMulltipyWalletBalance(ctx, publicKeys)
	if err != nil {
		return apperrors.Internal("failed to get wallet balances", err)
	}

	balanceByWalletID := make(map[uint64]uint64, len(wallets))
	for i, w := range wallets {
		balanceByWalletID[w.ID] = solanarpc.SOLToLamports(balances[i])
	}
	for _, wb := range walletBuys {
		if _, ok := walletByID[wb.WalletID]; !ok {
			return apperrors.BadRequest(fmt.Sprintf("wallet %d not found", wb.WalletID), nil)
		}
		if balanceByWalletID[wb.WalletID] < requiredByWalletID[wb.WalletID] {
			return apperrors.BadRequest(fmt.Sprintf("wallet %d balance is not enough for buy, token-account rent and fees", wb.WalletID), nil)
		}
	}

	return nil
}

func (s *PumpfunLaunchService) estimatePumpfunCommission(
	ctx context.Context,
	walletBuyCount int,
) (uint64, error) {
	commission, ok := addLamports(pumpfunTxFeeReserveLamports)
	if !ok {
		return 0, apperrors.BadRequest("pumpfun commission overflows uint64", nil)
	}

	if walletBuyCount == 0 {
		return commission, nil
	}

	ataRentLamports, err := s.rpc.GetATARentExemption(ctx)
	if err != nil {
		return 0, apperrors.Internal("failed to get token account rent exemption", err)
	}
	userVolumeAccumulatorRentLamports, err := s.rpc.GetRentExemption(ctx, pumpfunUserVolumeAccumulatorSize)
	if err != nil {
		return 0, apperrors.Internal("failed to get user volume accumulator rent exemption", err)
	}
	creatorVaultRentLamports, err := s.rpc.GetRentExemption(ctx, 0)
	if err != nil {
		return 0, apperrors.Internal("failed to get creator vault rent exemption", err)
	}

	walletCommissionLamports, ok := addLamports(
		ataRentLamports,
		pumpfunToken2022ATARentBufferLamports,
		userVolumeAccumulatorRentLamports,
		creatorVaultRentLamports,
		pumpfunTxFeeReserveLamports,
	)
	if !ok {
		return 0, apperrors.BadRequest("wallet pumpfun commission overflows uint64", nil)
	}

	for range walletBuyCount {
		commission, ok = addLamports(commission, walletCommissionLamports)
		if !ok {
			return 0, apperrors.BadRequest("pumpfun commission overflows uint64", nil)
		}
	}

	return commission, nil
}

func addLamports(values ...uint64) (uint64, bool) {
	var total uint64
	for _, value := range values {
		if total > ^uint64(0)-value {
			return 0, false
		}
		total += value
	}
	return total, true
}

func validatePumpfunEstimateBuys(req *model.PumpfunEstimateCreateRequest) (uint64, error) {
	ownerBuyLamports := solanarpc.SOLToLamports(req.BuyInSol)
	if req.BuyInSol > 0 && ownerBuyLamports == 0 {
		return 0, apperrors.BadRequest("buy_in_sol is too small", nil)
	}

	seen := make(map[uint64]struct{}, len(req.WalletBuys))
	var totalWalletBuy uint64
	for _, wb := range req.WalletBuys {
		if wb.WalletID == 0 {
			return 0, apperrors.BadRequest("wallet_id is required", nil)
		}
		if wb.AmountSol <= 0 {
			return 0, apperrors.BadRequest("wallet buy amount must be greater than zero", nil)
		}
		amountLamports := solanarpc.SOLToLamports(wb.AmountSol)
		if amountLamports == 0 {
			return 0, apperrors.BadRequest("wallet buy amount is too small", nil)
		}
		if _, ok := seen[wb.WalletID]; ok {
			return 0, apperrors.BadRequest(fmt.Sprintf("duplicate wallet buy for wallet %d", wb.WalletID), nil)
		}
		seen[wb.WalletID] = struct{}{}
		if amountLamports > ownerBuyLamports || totalWalletBuy > ownerBuyLamports-amountLamports {
			return 0, apperrors.BadRequest("total wallet buy amount cannot exceed owner buy amount", nil)
		}
		totalWalletBuy += amountLamports
	}

	return ownerBuyLamports, nil
}

func estimatePumpfunBuy(amountLamports uint64, quoteParams *bonding.SwapParams) (uint64, uint64, error) {
	quote, err := bonding.QuoteBuyExactSolIn(quoteParams, amountLamports)
	if err != nil {
		return 0, 0, err
	}
	return quote.TokensOut, quote.TotalFeeLamports, nil
}

func addTokensAndCommission(resp *model.PumpfunEstimateCreateResponse, tokensOut uint64, commissionLamports uint64) error {
	resp.PumpfunCommissionSOL += solanarpc.LamportsToSOL(int64(commissionLamports))
	resp.TotalTokensOut += solanarpc.FromAtomicUnit(tokensOut, pumpfunTokenDecimals)
	return nil
}

func (s *PumpfunLaunchService) buildWalletBuyTransactions(
	ctx context.Context,
	userID uint64,
	mint solana.PublicKey,
	owner solana.PublicKey,
	blockHash solana.Hash,
	walletBuys []model.WalletBuyConfig,
	walletCbInstrs []solana.Instruction,
) ([]string, error) {
	if len(walletBuys) == 0 {
		return []string{}, nil
	}

	walletIDs := make([]uint64, len(walletBuys))
	for i, wb := range walletBuys {
		walletIDs[i] = wb.WalletID
	}

	wallets, err := s.walletRepo.FetchWalletsByIdsAndUserID(ctx, walletIDs, userID)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch wallets", err)
	}

	walletByID := make(map[uint64]model.Wallet, len(wallets))
	for _, w := range wallets {
		walletByID[w.ID] = w
	}

	result := make([]string, 0, len(walletBuys))
	for _, wb := range walletBuys {
		w, ok := walletByID[wb.WalletID]
		if !ok {
			return nil, apperrors.BadRequest(fmt.Sprintf("wallet %d not found", wb.WalletID), nil)
		}
		keyStr, err := s.keyStorage.Get(ctx, userID, w.PublicKey)
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to get key for wallet %d", wb.WalletID), err)
		}
		privKey, err := solana.PrivateKeyFromBase58(keyStr)
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to parse key for wallet %d", wb.WalletID), err)
		}

		amountLamports := solanarpc.SOLToLamports(wb.AmountSol)
		buyTx, err := bonding.BuildInitialBuyTransaction(ctx, s.rpc, mint, owner, privKey.PublicKey(), blockHash, amountLamports, walletCbInstrs...)
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to build buy tx for wallet %d", wb.WalletID), err)
		}
		_, err = buyTx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
			if key.Equals(privKey.PublicKey()) {
				return &privKey
			}
			return nil
		})
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to sign buy tx for wallet %d", wb.WalletID), err)
		}

		rawTx, err := buyTx.MarshalBinary()
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to marshal buy tx for wallet %d", wb.WalletID), err)
		}
		result = append(result, base64.StdEncoding.EncodeToString(rawTx))
	}

	return result, nil
}

func (s *PumpfunLaunchService) Launch(ctx context.Context, userID uint64, req *model.PumpfunProcessCreateRequest) (*model.PumpfunProcessCreateResponse, error) {
	saved, err := s.repo.GetByID(ctx, req.TxID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("unknown transaction")
		}
		return nil, apperrors.Internal("failed to get pending transaction", err)
	}

	if saved.UserID != userID {
		return nil, apperrors.NotFound("unknown transaction")
	}
	if req.MintPubkey != saved.MintPubkey {
		return nil, apperrors.BadRequest("mint pubkey mismatch", nil)
	}
	if saved.Status != model.PumpfunLaunchStatusPending {
		return nil, apperrors.UnprocessableEntity("launch is not pending")
	}
	if saved.ExpiresAt.Before(time.Now().UTC()) {
		return nil, apperrors.UnprocessableEntity("transaction expired")
	}

	signedCreateTx, err := s.verifyAndParseSignedTx(saved.CreateTx, saved.Signer, req.SignedCreateTx)
	if err != nil {
		return nil, err
	}

	bundle := make([]*solana.Transaction, 0, 2+len(saved.WalletBuyTxs))
	bundle = append(bundle, signedCreateTx)

	if len(saved.BuyTx) > 0 {
		if req.SignedBuyTx == "" {
			return nil, apperrors.BadRequest("signed_buy_tx is required", nil)
		}
		signedBuyTx, err := s.verifyAndParseSignedTx(saved.BuyTx, saved.Signer, req.SignedBuyTx)
		if err != nil {
			return nil, err
		}
		bundle = append(bundle, signedBuyTx)
	}

	for i, walletTx := range saved.WalletBuyTxs {
		rawWalletTx, err := base64.StdEncoding.DecodeString(walletTx)
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("invalid saved wallet buy tx encoding at index %d", i), err)
		}
		tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(rawWalletTx))
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to parse saved wallet buy tx at index %d", i), err)
		}
		bundle = append(bundle, tx)
	}

	if err := s.jitoClient.SimulateBundle(ctx, bundle, bonding.ProgramID); err != nil {
		s.logger.Error("bundle simulation failed", zap.Error(err))
		_ = s.repo.UpdateStatus(ctx, saved.ID, model.PumpfunLaunchStatusFailed)
		return nil, err
	}

	if err := s.jitoClient.BroadcastBundleNoSim(ctx, bundle); err != nil {
		_ = s.repo.UpdateStatus(ctx, saved.ID, model.PumpfunLaunchStatusFailed)
		return nil, apperrors.Internal("failed to broadcast bundle", err)
	}

	if err := s.repo.UpdateStatus(ctx, saved.ID, model.PumpfunLaunchStatusSuccess); err != nil {
		return nil, apperrors.Internal("failed to update launch status", err)
	}

	return &model.PumpfunProcessCreateResponse{
		MintPubkey: saved.MintPubkey,
		Status:     model.PumpfunLaunchStatusSuccess,
		Signatures: transactionSignatures(bundle),
	}, nil
}

func transactionSignatures(txs []*solana.Transaction) []string {
	signatures := make([]string, 0, len(txs))
	for _, tx := range txs {
		if len(tx.Signatures) == 0 {
			continue
		}
		signatures = append(signatures, tx.Signatures[0].String())
	}
	return signatures
}

func (s *PumpfunLaunchService) verifyAndParseSignedTx(savedTxBytes []byte, signer string, signedTxBase64 string) (*solana.Transaction, error) {
	savedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(savedTxBytes))
	if err != nil {
		return nil, apperrors.Internal("failed to parse saved tx", err)
	}

	rawSigned, err := base64.StdEncoding.DecodeString(signedTxBase64)
	if err != nil {
		return nil, apperrors.BadRequest("invalid signed tx encoding", err)
	}

	signedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(rawSigned))
	if err != nil {
		return nil, apperrors.BadRequest("failed to parse signed tx", err)
	}

	// Wallet adapters (Phantom etc.) refresh the blockhash and reorder accounts before signing,
	// so byte-by-byte comparison fails. Instead we verify the ed25519 signature and check that
	// the instruction data is intact.
	signedMsg, _ := signedTx.Message.MarshalBinary()

	pubKeyBytes := signedTx.Message.AccountKeys[0]
	sig := signedTx.Signatures[0]

	if !ed25519.Verify(pubKeyBytes[:], signedMsg, sig[:]) {
		return nil, apperrors.UnprocessableEntity("invalid signature")
	}

	if pubKeyBytes.String() != signer {
		return nil, apperrors.UnprocessableEntity("signer mismatch")
	}

	// Verify instruction data is preserved (program ID + data bytes, accounts may be reordered).
	if len(signedTx.Message.Instructions) != len(savedTx.Message.Instructions) {
		return nil, apperrors.UnprocessableEntity("instruction count mismatch")
	}
	for i, savedInstr := range savedTx.Message.Instructions {
		signedInstr := signedTx.Message.Instructions[i]
		savedProgram := savedTx.Message.AccountKeys[savedInstr.ProgramIDIndex]
		signedProgram := signedTx.Message.AccountKeys[signedInstr.ProgramIDIndex]
		if !savedProgram.Equals(signedProgram) {
			s.logger.Warn("instruction program mismatch",
				zap.Int("index", i),
				zap.String("saved", savedProgram.String()),
				zap.String("signed", signedProgram.String()),
			)
			return nil, apperrors.UnprocessableEntity("instruction program mismatch")
		}
		if !bytes.Equal(savedInstr.Data, signedInstr.Data) {
			s.logger.Warn("instruction data mismatch",
				zap.Int("index", i),
				zap.String("saved_hex", hex.EncodeToString(savedInstr.Data)),
				zap.String("signed_hex", hex.EncodeToString(signedInstr.Data)),
			)
			return nil, apperrors.UnprocessableEntity("instruction data mismatch")
		}
	}

	return signedTx, nil
}
