package bonding

import (
	"context"
	"math/big"
	pump_bonding "mm/internal/client/pumpfun/bonding/bonding_client"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/pkg/apperrors"
	"mm/pkg/solutil"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	assoc "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/token"
)

var ProgramID = pump_bonding.ProgramID
var FeeProgramID = solana.MustPublicKeyFromBase58("pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ")
var GlobalConfigPubkey = solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf")
var MplTokenMetadataProgramID = solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
var MayhemProgramID = solana.MustPublicKeyFromBase58("MAyhSmzXzV1pTf7LsNkrNwkWKTo4ougAJ1PPg47MD4e")

type Global = pump_bonding.Global
type BondingCurve = pump_bonding.BondingCurve
type FeeConfig = pump_bonding.FeeConfig

func ParseAccount_Global(accountData []byte) (*Global, error) {
	return pump_bonding.ParseAccount_Global(accountData)
}

func ParseAccount_BondingCurve(accountData []byte) (*BondingCurve, error) {
	return pump_bonding.ParseAccount_BondingCurve(accountData)
}

func ParseAccount_FeeConfig(accountData []byte) (*FeeConfig, error) {
	return pump_bonding.ParseAccount_FeeConfig(accountData)
}

type PoolStateWithReserve struct {
	ReserveA          *big.Int
	ReserveB          *big.Int
	GlobalState       *Global
	BondingCurveState *BondingCurve
	TokenMint         *token.Mint
	TokenMintAddress  solana.PublicKey
	BaseMintDecimals  uint8
	QuoteMintDecimals uint8
}

type SwapParams struct {
	GlobalID                solana.PublicKey
	BondingCurveID          solana.PublicKey
	BondingCurveV2          solana.PublicKey
	Mint                    solana.PublicKey
	VirtualTokenReserves    uint64
	VirtualSolReserves      uint64
	ProtocolFeeBps          uint64
	CreatorFeeBps           uint64
	FeeRecipient            solana.PublicKey
	AssociatedBondingCurve  solana.PublicKey
	AssociatedUserAccount   solana.PublicKey
	CreatorVault            solana.PublicKey
	EventAuthority          solana.PublicKey
	GlobalVolumeAccumulator solana.PublicKey
	UserVolumeAccumulator   solana.PublicKey
	FeeConfig               solana.PublicKey
	SystemProgram           solana.PublicKey
	TokenProgram            solana.PublicKey
	Program                 solana.PublicKey
	FeeProgram              solana.PublicKey
	FeeRecipientNew         solana.PublicKey
}

func FetchBondingCurveState(
	ctx context.Context,
	rpc solanarpc.SolanaRPC,
	params *model.PoolParams,
	inputTokenMint,
	outputTokenMint solana.PublicKey,
) (*PoolStateWithReserve, error) {
	tokenMint := inputTokenMint
	if solutil.IsSOLLikeMint(tokenMint) {
		tokenMint = outputTokenMint
	}

	accounts, err := rpc.GetMultipleAccounts(ctx, params.PoolID, GlobalConfigPubkey, tokenMint)
	if err != nil {
		return nil, err
	}

	if len(accounts.Value) != 3 {
		return nil, apperrors.Internal("failed to fetch pump bonding accounts")
	}

	for i, acc := range accounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal("pump bonding account is nil")
		}
		_ = i
	}

	bondingCurve, err := ParseAccount_BondingCurve(accounts.Value[0].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse bonding curve", err)
	}

	globalState, err := ParseAccount_Global(accounts.Value[1].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse bonding global config", err)
	}

	tokenMintState := token.Mint{}
	if err := tokenMintState.UnmarshalWithDecoder(bin.NewBinDecoder(accounts.Value[2].Data.GetBinary())); err != nil {
		return nil, apperrors.Internal("failed to parse bonding token mint", err)
	}

	// Bonding curve trades against native SOL. For swap math we use virtual reserves:
	// token reserve on one side and SOL reserve on the other.
	virtualTokenReserve := new(big.Int).SetUint64(bondingCurve.VirtualTokenReserves)
	virtualSOLReserve := new(big.Int).SetUint64(bondingCurve.VirtualSolReserves)

	state := &PoolStateWithReserve{
		GlobalState:       globalState,
		BondingCurveState: bondingCurve,
		TokenMint:         &tokenMintState,
		TokenMintAddress:  tokenMint,
		BaseMintDecimals:  tokenMintState.Decimals,
		QuoteMintDecimals: solana.SolDecimals,
	}

	if solutil.IsSOLLikeMint(inputTokenMint) {
		state.ReserveA = virtualSOLReserve
		state.ReserveB = virtualTokenReserve
	} else {
		state.ReserveA = virtualTokenReserve
		state.ReserveB = virtualSOLReserve
	}

	return state, nil
}

func FetchBondingSwapParams(
	ctx context.Context,
	rpc solanarpc.SolanaRPC,
	bondingCurveID, user, mint solana.PublicKey,
) (*SwapParams, error) {
	accounts, err := rpc.GetMultipleAccounts(ctx, bondingCurveID, GlobalConfigPubkey, mint)
	if err != nil {
		return nil, err
	}

	if len(accounts.Value) != 3 {
		return nil, apperrors.Internal("failed to fetch pump bonding swap accounts")
	}

	for _, acc := range accounts.Value {
		if acc == nil || acc.Data == nil {
			return nil, apperrors.Internal("pump bonding swap account is nil")
		}
	}

	bondingCurve, err := ParseAccount_BondingCurve(accounts.Value[0].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse bonding curve", err)
	}

	globalState, err := ParseAccount_Global(accounts.Value[1].Data.GetBinary())
	if err != nil {
		return nil, apperrors.Internal("failed to parse bonding global config", err)
	}

	tokenProgram := accounts.Value[2].Owner

	associatedBondingCurve, _, err := FindAssociatedBondingCurve(bondingCurveID, tokenProgram, mint)
	if err != nil {
		return nil, apperrors.Internal("failed to derive associated bonding curve", err)
	}

	associatedUserAccount, _, err := FindAssociatedBondingCurve(user, tokenProgram, mint)
	if err != nil {
		return nil, apperrors.Internal("failed to derive associated user account", err)
	}

	creatorVault, _, err := FindCreatorVault(bondingCurve.Creator)
	if err != nil {
		return nil, apperrors.Internal("failed to derive creator vault", err)
	}

	eventAuthority, _, err := FindEventAuthority()
	if err != nil {
		return nil, apperrors.Internal("failed to derive event authority", err)
	}

	globalVolumeAccumulator, _, err := FindGlobalVolumeAccumulator()
	if err != nil {
		return nil, apperrors.Internal("failed to derive global volume accumulator", err)
	}

	userVolumeAccumulator, _, err := FindUserVolumeAccumulator(user)
	if err != nil {
		return nil, apperrors.Internal("failed to derive user volume accumulator", err)
	}

	feeConfig, _, err := FindFeeConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to derive fee config", err)
	}

	bondingCurveV2, _, err := FindBondingCurveV2(mint)
	if err != nil {
		return nil, apperrors.Internal("failed to derive bonding curve v2", err)
	}

	feeRecipient, err := solana.PublicKeyFromBase58("5YxQFdt3Tr9zJLvkFccqXVUwhdTWJQc1fFg2YPbxvxeD")
	if err != nil {
		return nil, apperrors.Internal("failed to get fee recipient account public key")
	}

	return &SwapParams{
		GlobalID:                GlobalConfigPubkey,
		BondingCurveID:          bondingCurveID,
		BondingCurveV2:          bondingCurveV2,
		Mint:                    mint,
		VirtualTokenReserves:    bondingCurve.VirtualTokenReserves,
		VirtualSolReserves:      bondingCurve.VirtualSolReserves,
		ProtocolFeeBps:          globalState.FeeBasisPoints,
		CreatorFeeBps:           globalState.CreatorFeeBasisPoints,
		FeeRecipient:            globalState.FeeRecipient,
		AssociatedBondingCurve:  associatedBondingCurve,
		AssociatedUserAccount:   associatedUserAccount,
		CreatorVault:            creatorVault,
		EventAuthority:          eventAuthority,
		GlobalVolumeAccumulator: globalVolumeAccumulator,
		UserVolumeAccumulator:   userVolumeAccumulator,
		FeeConfig:               feeConfig,
		SystemProgram:           solana.SystemProgramID,
		TokenProgram:            tokenProgram,
		Program:                 ProgramID,
		FeeProgram:              FeeProgramID,
		FeeRecipientNew:         feeRecipient,
	}, nil
}

func FindAssociatedBondingCurve(curve, tokenProgram, mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			curve.Bytes(),
			tokenProgram.Bytes(),
			mint.Bytes(),
		},
		assoc.ProgramID,
	)
	return addr, bump, err
}

func FindCreatorVault(creator solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("creator-vault"),
			creator.Bytes(),
		},
		ProgramID,
	)

	return addr, bump, err
}

func FindEventAuthority() (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("__event_authority"),
		},
		ProgramID,
	)
	return addr, bump, err
}

func FindGlobalVolumeAccumulator() (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("global_volume_accumulator"),
		},
		ProgramID,
	)
	return addr, bump, err
}

func FindUserVolumeAccumulator(user solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("user_volume_accumulator"),
			user.Bytes(),
		},
		ProgramID,
	)
	return addr, bump, err
}

func FindFeeConfig() (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("fee_config"),
			{
				1, 86, 224, 246, 147, 102, 90, 207,
				68, 219, 21, 104, 191, 23, 91, 170,
				81, 137, 203, 151, 245, 210, 255, 59,
				101, 93, 43, 182, 253, 109, 24, 176,
			},
		},
		FeeProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindBondingCurveV2(mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	addr, bump, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("bonding-curve-v2"),
			mint.Bytes(),
		},
		ProgramID,
	)
	if err != nil {
		return solana.PublicKey{}, 0, err
	}

	return addr, bump, nil
}

func FindMintAuthority() (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("mint-authority"),
		},
		ProgramID,
	)
}

func FindBondingCurve(mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress(
		[][]byte{
			[]byte("bonding-curve"),
			mint.Bytes(),
		},
		ProgramID,
	)
}

func FindMayhemGlobalParams() (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress([][]byte{[]byte("global-params")}, MayhemProgramID)
}

func FindMayhemSolVault() (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress([][]byte{[]byte("sol-vault")}, MayhemProgramID)
}

func FindMayhemState(mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress([][]byte{[]byte("mayhem-state"), mint.Bytes()}, MayhemProgramID)
}
