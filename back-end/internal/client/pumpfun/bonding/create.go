package bonding

import (
	"context"
	"fmt"
	"math/big"
	pump_bonding "mm/internal/client/pumpfun/bonding/bonding_client"
	"mm/internal/client/solanarpc"
	"mm/pkg/solutil"

	assoc "github.com/gagliardetto/solana-go/programs/associated-token-account"

	"github.com/gagliardetto/solana-go"
)

const DefaultBuyExactSolInSlippageBps uint64 = 500

func BuildCreateV2Transaction(name, symbol, uri string, user solana.PublicKey, mintPrivKey solana.PrivateKey, blockHash solana.Hash, mayhem, cashback bool, extraInstrs ...solana.Instruction) (*solana.Transaction, error) {
	mint := mintPrivKey.PublicKey()

	mintAuthority, _, err := FindMintAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to get mint authority: %w", err)
	}

	bondingCurve, _, err := FindBondingCurve(mint)
	if err != nil {
		return nil, fmt.Errorf("failed to get bonding curve: %w", err)
	}

	associatedBondingCurve, _, err := FindAssociatedBondingCurve(bondingCurve, solana.Token2022ProgramID, mint)
	if err != nil {
		return nil, fmt.Errorf("failed to get associated bonding curve: %w", err)
	}

	eventAuthority, _, err := FindEventAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to get event authority: %w", err)
	}

	globalParams, _, err := FindMayhemGlobalParams()
	if err != nil {
		return nil, fmt.Errorf("failed to get mayhem global params: %w", err)
	}

	solVault, _, err := FindMayhemSolVault()
	if err != nil {
		return nil, fmt.Errorf("failed to get mayhem sol vault: %w", err)
	}

	mayhemState, _, err := FindMayhemState(mint)
	if err != nil {
		return nil, fmt.Errorf("failed to get mayhem state: %w", err)
	}

	mayhemTokenVault, _, err := FindAssociatedBondingCurve(mayhemState, solana.Token2022ProgramID, mint)
	if err != nil {
		return nil, fmt.Errorf("failed to get mayhem token vault: %w", err)
	}

	createInstr, err := pump_bonding.NewCreateV2Instruction(
		name, symbol, uri,
		user,
		mayhem, pump_bonding.OptionBool{V0: cashback},
		mint,
		mintAuthority,
		bondingCurve,
		associatedBondingCurve,
		GlobalConfigPubkey,
		user,
		solana.SystemProgramID,
		solana.Token2022ProgramID,
		assoc.ProgramID,
		MayhemProgramID,
		globalParams,
		solVault,
		mayhemState,
		mayhemTokenVault,
		eventAuthority,
		ProgramID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to form create instruction: %w", err)
	}

	instrs := make([]solana.Instruction, 0, 1+len(extraInstrs))
	instrs = append(instrs, extraInstrs...)
	instrs = append(instrs, createInstr)

	tx, err := solana.NewTransaction(
		instrs,
		blockHash,
		solana.TransactionPayer(user),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to form transaction: %w", err)
	}

	numSigners := int(tx.Message.Header.NumRequiredSignatures)
	tx.Signatures = make([]solana.Signature, numSigners)

	msgBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message for signing: %w", err)
	}
	mintSig, err := mintPrivKey.Sign(msgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with mint key: %w", err)
	}
	for i := range numSigners {
		if tx.Message.AccountKeys[i].Equals(mint) {
			tx.Signatures[i] = mintSig
			break
		}
	}

	return tx, nil
}

func BuildBuyTransaction(ctx context.Context, rpc solanarpc.SolanaRPC, mint, user solana.PublicKey, blockHash solana.Hash, buyInLamports uint64) (*solana.Transaction, error) {
	bondingCurve, _, err := FindBondingCurve(mint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bonding curve: %w", err)
	}

	params, err := FetchBondingSwapParams(ctx, rpc, bondingCurve, user, mint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swap params: %w", err)
	}

	createATAInstr, err := solutil.NewCreateAssociatedTokenAccountInstruction(user, user, mint, params.TokenProgram)
	if err != nil {
		return nil, fmt.Errorf("failed to build create ATA instruction: %w", err)
	}

	minTokensOut, err := quoteBuyExactSolInMinTokensOut(params, buyInLamports)
	if err != nil {
		return nil, fmt.Errorf("failed to quote buy: %w", err)
	}
	minTokensOut = floorAmountWithSlippage(minTokensOut, DefaultBuyExactSolInSlippageBps)

	buyInstr, err := buildBuyExactSolInInstruction(params, user, buyInLamports, minTokensOut, pump_bonding.OptionBool{})
	if err != nil {
		return nil, fmt.Errorf("failed to build buy instruction: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{createATAInstr, buyInstr},
		blockHash,
		solana.TransactionPayer(user),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to form transaction: %w", err)
	}

	tx.Signatures = make([]solana.Signature, int(tx.Message.Header.NumRequiredSignatures))

	return tx, nil
}

// BuildInitialBuyTransaction builds a buy transaction for a token that hasn't been created on-chain yet.
// It derives all accounts from PDAs without fetching the bonding curve (which doesn't exist yet).
// creator must be the same pubkey used as the token creator in the preceding CreateV2 transaction.
// extraInstrs are prepended before account creation and buy instructions (e.g. compute budget instructions).
func BuildInitialBuyTransaction(ctx context.Context, rpc solanarpc.SolanaRPC, mint, creator, user solana.PublicKey, blockHash solana.Hash, buyInLamports uint64, extraInstrs ...solana.Instruction) (*solana.Transaction, error) {
	feeConfig, _, err := FindFeeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to derive fee config: %w", err)
	}

	accounts, err := rpc.GetMultipleAccounts(ctx, GlobalConfigPubkey, feeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial buy configs: %w", err)
	}
	if len(accounts.Value) != 2 || accounts.Value[0] == nil || accounts.Value[0].Data == nil {
		return nil, fmt.Errorf("global config account not found")
	}
	if accounts.Value[1] == nil || accounts.Value[1].Data == nil {
		return nil, fmt.Errorf("fee config account not found")
	}

	globalState, err := ParseAccount_Global(accounts.Value[0].Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}
	feeState, err := ParseAccount_FeeConfig(accounts.Value[1].Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse fee config: %w", err)
	}
	fees := maxBuyExactSolInFees(feeState)

	// pump.fun v2 always uses Token2022
	tokenProgram := solana.Token2022ProgramID

	bondingCurve, _, err := FindBondingCurve(mint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bonding curve: %w", err)
	}

	associatedBondingCurve, _, err := FindAssociatedBondingCurve(bondingCurve, tokenProgram, mint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive associated bonding curve: %w", err)
	}

	associatedUserAccount, _, err := FindAssociatedBondingCurve(user, tokenProgram, mint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive associated user account: %w", err)
	}

	creatorVault, _, err := FindCreatorVault(creator)
	if err != nil {
		return nil, fmt.Errorf("failed to derive creator vault: %w", err)
	}

	eventAuthority, _, err := FindEventAuthority()
	if err != nil {
		return nil, fmt.Errorf("failed to derive event authority: %w", err)
	}

	globalVolumeAccumulator, _, err := FindGlobalVolumeAccumulator()
	if err != nil {
		return nil, fmt.Errorf("failed to derive global volume accumulator: %w", err)
	}

	userVolumeAccumulator, _, err := FindUserVolumeAccumulator(user)
	if err != nil {
		return nil, fmt.Errorf("failed to derive user volume accumulator: %w", err)
	}

	bondingCurveV2, _, err := FindBondingCurveV2(mint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive bonding curve v2: %w", err)
	}

	feeRecipientNew := solana.MustPublicKeyFromBase58("5YxQFdt3Tr9zJLvkFccqXVUwhdTWJQc1fFg2YPbxvxeD")

	params := &SwapParams{
		GlobalID:                GlobalConfigPubkey,
		BondingCurveID:          bondingCurve,
		BondingCurveV2:          bondingCurveV2,
		Mint:                    mint,
		VirtualTokenReserves:    globalState.InitialVirtualTokenReserves,
		VirtualSolReserves:      globalState.InitialVirtualSolReserves,
		ProtocolFeeBps:          fees.ProtocolFeeBps,
		CreatorFeeBps:           fees.CreatorFeeBps,
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
		FeeRecipientNew:         feeRecipientNew,
	}

	createATAInstr, err := solutil.NewCreateAssociatedTokenAccountInstruction(user, user, mint, tokenProgram)
	if err != nil {
		return nil, fmt.Errorf("failed to build create ATA instruction: %w", err)
	}

	minTokensOut, err := quoteBuyExactSolInMinTokensOut(params, buyInLamports)
	if err != nil {
		return nil, fmt.Errorf("failed to quote initial buy: %w", err)
	}

	minTokensOut = floorAmountWithSlippage(minTokensOut, DefaultBuyExactSolInSlippageBps)

	buyInstr, err := buildBuyExactSolInInstruction(params, user, buyInLamports, minTokensOut, pump_bonding.OptionBool{})
	if err != nil {
		return nil, fmt.Errorf("failed to build buy instruction: %w", err)
	}

	instrs := make([]solana.Instruction, 0, 2+len(extraInstrs))
	instrs = append(instrs, extraInstrs...)
	instrs = append(instrs, createATAInstr, buyInstr)

	tx, err := solana.NewTransaction(
		instrs,
		blockHash,
		solana.TransactionPayer(user),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to form transaction: %w", err)
	}

	tx.Signatures = make([]solana.Signature, int(tx.Message.Header.NumRequiredSignatures))

	return tx, nil
}

type BuyExactSolInQuote struct {
	TokensOut           uint64
	NetSol              uint64
	ProtocolFeeLamports uint64
	CreatorFeeLamports  uint64
	TotalFeeLamports    uint64
}

func maxBuyExactSolInFees(feeConfig *FeeConfig) pump_bonding.Fees {
	fees := feeConfig.FlatFees
	for _, tier := range feeConfig.FeeTiers {
		if tier.Fees.ProtocolFeeBps > fees.ProtocolFeeBps {
			fees.ProtocolFeeBps = tier.Fees.ProtocolFeeBps
		}
		if tier.Fees.CreatorFeeBps > fees.CreatorFeeBps {
			fees.CreatorFeeBps = tier.Fees.CreatorFeeBps
		}
	}
	return fees
}

func FetchInitialBuyQuoteParams(ctx context.Context, rpc solanarpc.SolanaRPC) (*SwapParams, error) {
	feeConfig, _, err := FindFeeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to derive fee config: %w", err)
	}

	accounts, err := rpc.GetMultipleAccounts(ctx, GlobalConfigPubkey, feeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial buy configs: %w", err)
	}
	if len(accounts.Value) != 2 || accounts.Value[0] == nil || accounts.Value[0].Data == nil {
		return nil, fmt.Errorf("global config account not found")
	}
	if accounts.Value[1] == nil || accounts.Value[1].Data == nil {
		return nil, fmt.Errorf("fee config account not found")
	}

	globalState, err := ParseAccount_Global(accounts.Value[0].Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}
	feeState, err := ParseAccount_FeeConfig(accounts.Value[1].Data.GetBinary())
	if err != nil {
		return nil, fmt.Errorf("failed to parse fee config: %w", err)
	}
	fees := maxBuyExactSolInFees(feeState)

	return &SwapParams{
		VirtualTokenReserves: globalState.InitialVirtualTokenReserves,
		VirtualSolReserves:   globalState.InitialVirtualSolReserves,
		ProtocolFeeBps:       fees.ProtocolFeeBps,
		CreatorFeeBps:        fees.CreatorFeeBps,
	}, nil
}

func quoteBuyExactSolInMinTokensOut(params *SwapParams, spendableSolIn uint64) (uint64, error) {
	quote, err := QuoteBuyExactSolIn(params, spendableSolIn)
	if err != nil {
		return 0, err
	}
	return quote.TokensOut, nil
}

func QuoteBuyExactSolIn(params *SwapParams, spendableSolIn uint64) (*BuyExactSolInQuote, error) {
	if spendableSolIn == 0 {
		return nil, fmt.Errorf("spendable SOL is zero")
	}
	if params.VirtualTokenReserves == 0 {
		return nil, fmt.Errorf("virtual token reserves are zero")
	}
	if params.VirtualSolReserves == 0 {
		return nil, fmt.Errorf("virtual SOL reserves are zero")
	}

	const feeDenominator = uint64(10_000)
	totalFeeBps := params.ProtocolFeeBps + params.CreatorFeeBps

	netSol := new(big.Int).SetUint64(spendableSolIn)
	netSol.Mul(netSol, new(big.Int).SetUint64(feeDenominator))
	netSol.Div(netSol, new(big.Int).SetUint64(feeDenominator+totalFeeBps))

	protocolFee := ceilDivBig(
		new(big.Int).Mul(new(big.Int).Set(netSol), new(big.Int).SetUint64(params.ProtocolFeeBps)),
		new(big.Int).SetUint64(feeDenominator),
	)
	creatorFee := ceilDivBig(
		new(big.Int).Mul(new(big.Int).Set(netSol), new(big.Int).SetUint64(params.CreatorFeeBps)),
		new(big.Int).SetUint64(feeDenominator),
	)

	fees := new(big.Int).Add(protocolFee, creatorFee)
	netSolPlusFees := new(big.Int).Add(new(big.Int).Set(netSol), fees)
	spendable := new(big.Int).SetUint64(spendableSolIn)
	if netSolPlusFees.Cmp(spendable) > 0 {
		netSol.Sub(netSol, new(big.Int).Sub(netSolPlusFees, spendable))
	}
	if netSol.Cmp(big.NewInt(1)) <= 0 {
		return nil, fmt.Errorf("spendable SOL is too small to buy tokens")
	}

	effectiveSol := new(big.Int).Sub(netSol, big.NewInt(1))
	numerator := new(big.Int).Mul(effectiveSol, new(big.Int).SetUint64(params.VirtualTokenReserves))
	denominator := new(big.Int).Add(new(big.Int).SetUint64(params.VirtualSolReserves), effectiveSol)
	tokensOut := numerator.Div(numerator, denominator)
	if tokensOut.Sign() == 0 {
		return nil, fmt.Errorf("quoted token amount is zero")
	}
	if !tokensOut.IsUint64() {
		return nil, fmt.Errorf("quoted token amount overflows uint64")
	}
	if !netSol.IsUint64() {
		return nil, fmt.Errorf("quoted net SOL amount overflows uint64")
	}
	if !protocolFee.IsUint64() {
		return nil, fmt.Errorf("quoted protocol fee overflows uint64")
	}
	if !creatorFee.IsUint64() {
		return nil, fmt.Errorf("quoted creator fee overflows uint64")
	}
	if !fees.IsUint64() {
		return nil, fmt.Errorf("quoted total fee overflows uint64")
	}

	return &BuyExactSolInQuote{
		TokensOut:           tokensOut.Uint64(),
		NetSol:              netSol.Uint64(),
		ProtocolFeeLamports: protocolFee.Uint64(),
		CreatorFeeLamports:  creatorFee.Uint64(),
		TotalFeeLamports:    fees.Uint64(),
	}, nil
}

func floorAmountWithSlippage(amount, slippageBps uint64) uint64 {
	const basisPointDenominator = uint64(10_000)
	if slippageBps >= basisPointDenominator {
		return 1
	}
	floored := new(big.Int).SetUint64(amount)
	floored.Mul(floored, new(big.Int).SetUint64(basisPointDenominator-slippageBps))
	floored.Div(floored, new(big.Int).SetUint64(basisPointDenominator))
	if floored.Sign() == 0 {
		return 1
	}
	return floored.Uint64()
}

func ceilDivBig(numerator, denominator *big.Int) *big.Int {
	if numerator.Sign() == 0 {
		return new(big.Int)
	}
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}
