package pump_amm

import (
	"context"
	"fmt"
	"math/big"
	pumpammclient "mm/internal/client/pumpfun/amm/amm_client"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/pkg/apperrors"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	assoc "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/token"
)

var ProgramID = pumpammclient.ProgramID
var FeeProgramID = solana.MustPublicKeyFromBase58("pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ")

var GlobalConfigPubkey = solana.MustPublicKeyFromBase58("ADyA8hdefvWN2dbGGWFotbzWxrAvLW83WG6QCVXvJKqw")

type Pool = pumpammclient.Pool
type GlobalConfig = pumpammclient.GlobalConfig

func ParseAccount_Pool(accountData []byte) (*Pool, error) {
	return pumpammclient.ParseAccount_Pool(accountData)
}

func ParseAccount_GlobalConfig(accountData []byte) (*GlobalConfig, error) {
	return pumpammclient.ParseAccount_GlobalConfig(accountData)
}

type PoolStateWithReserve struct {
	ReserveA          *big.Int
	ReserveB          *big.Int
	PoolState         *Pool
	BaseMintDecimals  uint8
	QuoteMintDecimals uint8
}

type SwapParams struct {
	Pool                             *Pool
	GlobalConfig                     *GlobalConfig
	BaseMint                         *token.Mint
	QuoteMint                        *token.Mint
	BaseTokenProgram                 solana.PublicKey
	QuoteTokenProgram                solana.PublicKey
	ProtocolFeeRecipient             solana.PublicKey
	ProtocolFeeRecipientTokenAccount solana.PublicKey
	UserBaseTokenAccount             solana.PublicKey
	UserQuoteTokenAccount            solana.PublicKey
	CoinCreatorVaultAta              solana.PublicKey
	AssociatedTokenProgram           solana.PublicKey
	SystemProgram                    solana.PublicKey
	CoinCreatorVaultAuthority        solana.PublicKey
	GlobalVolumeAccumulator          solana.PublicKey
	UserVolumeAccumulator            solana.PublicKey
	FeeConfig                        solana.PublicKey
	EventAuthority                   solana.PublicKey
	PoolID                           solana.PublicKey
	PoolV2                           solana.PublicKey
	GlobalConfigID                   solana.PublicKey
}

func FetchAMMPoolState(ctx context.Context, rpc solanarpc.SolanaRPC, params *model.PoolParams, inputTokenMint, outputTokenMint solana.PublicKey) (*PoolStateWithReserve, error) {
	accounts, err := rpc.GetMultipleAccounts(ctx, params.InputTokenVault, params.OutputTokenVault, params.PoolID)
	if err != nil {
		return nil, err
	}

	if len(accounts.Value) != 3 {
		return nil, apperrors.Internal(fmt.Sprintf("failed to fetch pump amm accounts: %d accounts returned", len(accounts.Value)))
	}

	for i, acc := range accounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal(fmt.Sprintf("pump amm account at index %d is nil", i))
		}
	}

	pool, err := ParseAccount_Pool(accounts.Value[2].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse pump amm pool", err)
	}

	token0 := token.Account{}
	if err := token0.UnmarshalWithDecoder(bin.NewBinDecoder(accounts.Value[0].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse pump amm base vault", err)
	}

	token1 := token.Account{}
	if err := token1.UnmarshalWithDecoder(bin.NewBinDecoder(accounts.Value[1].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse pump amm quote vault", err)
	}

	mintAccounts, err := rpc.GetMultipleAccounts(ctx, pool.BaseMint, pool.QuoteMint)
	if err != nil {
		return nil, err
	}

	if len(mintAccounts.Value) != 2 {
		return nil, apperrors.Internal(fmt.Sprintf("failed to fetch pump amm mint accounts: %d accounts returned", len(mintAccounts.Value)))
	}

	for i, acc := range mintAccounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal(fmt.Sprintf("pump amm mint account at index %d is nil", i))
		}
	}

	baseMint := token.Mint{}
	if err := baseMint.UnmarshalWithDecoder(bin.NewBinDecoder(mintAccounts.Value[0].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse pump amm base mint", err)
	}

	quoteMint := token.Mint{}
	if err := quoteMint.UnmarshalWithDecoder(bin.NewBinDecoder(mintAccounts.Value[1].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse pump amm quote mint", err)
	}

	poolState := &PoolStateWithReserve{
		PoolState:         pool,
		BaseMintDecimals:  baseMint.Decimals,
		QuoteMintDecimals: quoteMint.Decimals,
	}

	baseReserve := new(big.Int).SetUint64(token0.Amount)
	quoteReserve := new(big.Int).SetUint64(token1.Amount)

	if pool.BaseMint.Equals(inputTokenMint) && pool.QuoteMint.Equals(outputTokenMint) {
		poolState.ReserveA = baseReserve
		poolState.ReserveB = quoteReserve
	} else {
		poolState.ReserveA = quoteReserve
		poolState.ReserveB = baseReserve
	}

	return poolState, nil
}

func FetchAMMSwapParams(ctx context.Context, rpc solanarpc.SolanaRPC, poolID, user solana.PublicKey) (*SwapParams, error) {
	accounts, err := rpc.GetMultipleAccounts(ctx, poolID, GlobalConfigPubkey)
	if err != nil {
		return nil, err
	}

	if len(accounts.Value) != 2 {
		return nil, apperrors.Internal(fmt.Sprintf("failed to fetch pump amm swap accounts: %d accounts returned", len(accounts.Value)))
	}

	for i, acc := range accounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal(fmt.Sprintf("pump amm swap account at index %d is nil", i))
		}
	}

	pool, err := ParseAccount_Pool(accounts.Value[0].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse pump amm pool", err)
	}

	globalConfig, err := ParseAccount_GlobalConfig(accounts.Value[1].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse pump amm global config", err)
	}

	mintAccounts, err := rpc.GetMultipleAccounts(ctx, pool.BaseMint, pool.QuoteMint)
	if err != nil {
		return nil, err
	}

	if len(mintAccounts.Value) != 2 {
		return nil, apperrors.Internal(fmt.Sprintf("failed to fetch pump amm mint accounts: %d accounts returned", len(mintAccounts.Value)))
	}

	for i, acc := range mintAccounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal(fmt.Sprintf("pump amm mint account at index %d is nil", i))
		}
	}

	baseMint := token.Mint{}
	if err := baseMint.UnmarshalWithDecoder(bin.NewBinDecoder(mintAccounts.Value[0].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse pump amm base mint", err)
	}

	quoteMint := token.Mint{}
	if err := quoteMint.UnmarshalWithDecoder(bin.NewBinDecoder(mintAccounts.Value[1].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse pump amm quote mint", err)
	}

	userBaseTokenAccount, _, err := FindATAWithTokenProgram(user, pool.BaseMint, mintAccounts.Value[0].Owner)
	if err != nil {
		return nil, apperrors.Internal("failed to derive user base token ATA", err)
	}

	userQuoteTokenAccount, _, err := FindATAWithTokenProgram(user, pool.QuoteMint, mintAccounts.Value[1].Owner)
	if err != nil {
		return nil, apperrors.Internal("failed to derive user quote token ATA", err)
	}

	protocolFeeRecipientTokenAccount, _, err := FindATAWithTokenProgram(globalConfig.ProtocolFeeRecipients[0], pool.QuoteMint, mintAccounts.Value[1].Owner)
	if err != nil {
		return nil, apperrors.Internal("failed to derive protocol fee recipient ATA", err)
	}

	coinCreatorVaultAuthority, _, err := FindCoinCreatorVaultAuthority(pool.CoinCreator)
	if err != nil {
		return nil, apperrors.Internal("failed to derive coin creator vault authority", err)
	}

	coinCreatorVaultAta, _, err := FindCoinCreatorVaultAta(coinCreatorVaultAuthority, pool.QuoteMint, mintAccounts.Value[1].Owner)
	if err != nil {
		return nil, apperrors.Internal("failed to derive coin creator vault ata", err)
	}

	globalVolumeAccumulator, _, err := FindGlobalVolumeAccumulator()
	if err != nil {
		return nil, apperrors.Internal("failed to derive global coin volume accumulator", err)
	}

	userVolumeAccumulator, _, err := FindUserVolumeAccumulator(user)
	if err != nil {
		return nil, apperrors.Internal("failed to derive user coin volume accumulator", err)
	}

	feeConfig, _, err := FindFeeConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to derive fee config", err)
	}

	eventAuthority, _, err := FindEventAuthority()

	if err != nil {
		return nil, apperrors.Internal("failed to derive event authority", err)
	}

	poolV2, _, err := FindPoolV2(pool.BaseMint)
	if err != nil {
		return nil, apperrors.Internal("failed to derive pool v2", err)
	}

	return &SwapParams{
		Pool:                             pool,
		PoolID:                           poolID,
		GlobalConfig:                     globalConfig,
		GlobalConfigID:                   GlobalConfigPubkey,
		BaseMint:                         &baseMint,
		QuoteMint:                        &quoteMint,
		BaseTokenProgram:                 mintAccounts.Value[0].Owner,
		QuoteTokenProgram:                mintAccounts.Value[1].Owner,
		ProtocolFeeRecipient:             globalConfig.ProtocolFeeRecipients[0],
		ProtocolFeeRecipientTokenAccount: protocolFeeRecipientTokenAccount,
		UserBaseTokenAccount:             userBaseTokenAccount,
		UserQuoteTokenAccount:            userQuoteTokenAccount,
		AssociatedTokenProgram:           assoc.ProgramID,
		SystemProgram:                    solana.SystemProgramID,
		CoinCreatorVaultAta:              coinCreatorVaultAta,
		GlobalVolumeAccumulator:          globalVolumeAccumulator,
		UserVolumeAccumulator:            userVolumeAccumulator,
		FeeConfig:                        feeConfig,
		CoinCreatorVaultAuthority:        coinCreatorVaultAuthority,
		EventAuthority:                   eventAuthority,
		PoolV2:                           poolV2,
	}, nil
}

func FindCoinCreatorVaultAuthority(coinCreator solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress([][]byte{
		[]byte("creator_vault"),
		coinCreator.Bytes(),
	}, ProgramID)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindCoinCreatorVaultAta(authority, quoteMint, quoteTokenProgram solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			authority.Bytes(),
			quoteTokenProgram.Bytes(),
			quoteMint.Bytes(),
		},
		assoc.ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindGlobalVolumeAccumulator() (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("global_volume_accumulator"),
		},
		ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindUserVolumeAccumulator(user solana.PublicKey) (solana.PublicKey, uint8,
	error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("user_volume_accumulator"),
			user.Bytes(),
		},
		ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindFeeConfig() (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("fee_config"),
			{
				12, 20, 222, 252, 130, 94, 198, 118,
				148, 37, 8, 24, 187, 101, 64, 101,
				244, 41, 141, 49, 86, 213, 113, 180,
				212, 248, 9, 12, 24, 233, 168, 99,
			},
		},
		FeeProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindATAWithTokenProgram(owner, mint, tokenProgram solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			owner.Bytes(),
			tokenProgram.Bytes(),
			mint.Bytes(),
		},
		assoc.ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindEventAuthority() (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("__event_authority"),
		},
		ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindPoolV2(baseMint solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("pool-v2"),
			baseMint.Bytes(),
		},
		ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func buildSwapInstruction(
	params *SwapParams,
	user solana.PublicKey,
	inputTokenMint solana.PublicKey,
	outputTokenMint solana.PublicKey,
	amountIn uint64,
	minOut uint64,
	trackVolume pumpammclient.OptionBool,
) (solana.Instruction, error) {
	isBuy := params.Pool.QuoteMint.Equals(inputTokenMint) && params.Pool.BaseMint.Equals(outputTokenMint)
	isSell := params.Pool.BaseMint.Equals(inputTokenMint) && params.Pool.QuoteMint.Equals(outputTokenMint)

	switch {
	case isBuy:
		return buildBuyExactQuoteInInstruction(
			params,
			user,
			amountIn,
			minOut,
			trackVolume,
		)
	case isSell:
		return buildSellInstruction(
			params,
			user,
			amountIn,
			minOut,
		)
	default:
		return nil, fmt.Errorf("input/output mints do not match pump amm pool")
	}
}

func buildBuyExactQuoteInInstruction(
	params *SwapParams,
	user solana.PublicKey,
	spendableQuoteIn uint64,
	minBaseAmountOut uint64,
	trackVolume pumpammclient.OptionBool,
) (solana.Instruction, error) {
	instruction, err := pumpammclient.NewBuyExactQuoteInInstruction(
		spendableQuoteIn,
		minBaseAmountOut,
		trackVolume,

		params.PoolID,
		user,
		params.GlobalConfigID,
		params.Pool.BaseMint,
		params.Pool.QuoteMint,
		params.UserBaseTokenAccount,
		params.UserQuoteTokenAccount,
		params.Pool.PoolBaseTokenAccount,
		params.Pool.PoolQuoteTokenAccount,
		params.ProtocolFeeRecipient,
		params.ProtocolFeeRecipientTokenAccount,
		params.BaseTokenProgram,
		params.QuoteTokenProgram,
		params.SystemProgram,
		params.AssociatedTokenProgram,
		params.EventAuthority,
		ProgramID,
		params.CoinCreatorVaultAta,
		params.CoinCreatorVaultAuthority,
		params.GlobalVolumeAccumulator,
		params.UserVolumeAccumulator,
		params.FeeConfig,
		FeeProgramID,
	)
	if err != nil {
		return nil, err
	}

	if genericInstruction, ok := instruction.(*solana.GenericInstruction); ok {
		genericInstruction.AccountValues = append(
			genericInstruction.AccountValues,
			solana.Meta(params.PoolV2),
		)
	}

	return instruction, nil
}

func buildSellInstruction(
	params *SwapParams,
	user solana.PublicKey,
	baseAmountIn uint64,
	minQuoteAmountOut uint64,
) (solana.Instruction, error) {
	instruction, err := pumpammclient.NewSellInstruction(
		baseAmountIn,
		minQuoteAmountOut,

		params.PoolID,
		user,
		params.GlobalConfigID,
		params.Pool.BaseMint,
		params.Pool.QuoteMint,
		params.UserBaseTokenAccount,
		params.UserQuoteTokenAccount,
		params.Pool.PoolBaseTokenAccount,
		params.Pool.PoolQuoteTokenAccount,
		params.ProtocolFeeRecipient,
		params.ProtocolFeeRecipientTokenAccount,
		params.BaseTokenProgram,
		params.QuoteTokenProgram,
		params.SystemProgram,
		params.AssociatedTokenProgram,
		params.EventAuthority,
		ProgramID,
		params.CoinCreatorVaultAta,
		params.CoinCreatorVaultAuthority,
		params.FeeConfig,
		FeeProgramID,
	)
	if err != nil {
		return nil, err
	}

	if genericInstruction, ok := instruction.(*solana.GenericInstruction); ok {
		genericInstruction.AccountValues = append(
			genericInstruction.AccountValues,
			solana.Meta(params.PoolV2),
		)
	}

	return instruction, nil
}
