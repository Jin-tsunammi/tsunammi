package ammv4

import (
	"encoding/binary"
	"fmt"
	"mm/internal/client/raydium"
	raydiumamm "mm/internal/client/raydium/ammv4/ammv4_client"
	"mm/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
)

// hundredOfBSPDenominator represents 100% in high-precision units (hundredths of a bip (10^-6)) (1,000,000 units = 100%).
const hundredOfBSPDenominator uint64 = 1000000

func numeratorWithDenominatorToHundredOfBSP(numerator uint64, denominator uint64) (uint64, error) {
	if denominator == 0 {
		return 0, apperrors.Internal("denominator cannot be zero")
	}
	if denominator > hundredOfBSPDenominator {
		return 0, apperrors.Internal(fmt.Sprintf("denominator cannot exceed 100%% (%d units)", hundredOfBSPDenominator))
	}

	return numerator * hundredOfBSPDenominator / denominator, nil
}

func ammGetAuthority(programID solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("amm authority")},
		programID,
	)
	if err != nil {
		return solana.PublicKey{}, err
	}

	return addr, err
}

func ammGetSerumVaultSigner(nonce uint64, marketId solana.PublicKey, marketProgramId solana.PublicKey) (solana.PublicKey, error) {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce)

	seeds := [][]byte{
		nonceBytes,
		marketId.Bytes(),
	}

	return solana.CreateProgramAddress(seeds, marketProgramId)
}

func ammBuildSwapInstruction(ammInfo *AMMInfoWithReservers, p raydium.SwapParams) (solana.Instruction, error) {
	amountInU64, err := raydium.SafeUint64(p.AmountIn)
	if err != nil {
		return nil, apperrors.Internal("AmountIn overflow: %w", err)
	}
	minAmountOutU64, err := raydium.SafeUint64(p.MinAmountOut)
	if err != nil {
		return nil, apperrors.Internal("MinAmountOut overflow: %w", err)
	}

	authority, err := ammGetAuthority(raydiumamm.ProgramID)

	if err != nil {
		return nil, apperrors.Internal("failed to get authority: %w", err)
	}

	serumVaultSigner, err := ammGetSerumVaultSigner(
		ammInfo.Market.VaultSignerNonce,
		ammInfo.Market.OwnAddress,
		ammInfo.PoolState.SerumDex,
	)

	instr, err := raydiumamm.NewSwapBaseInInstruction(
		// Swap Params
		amountInU64,     // Amount in (raw value)
		minAmountOutU64, // Min amount out (slippage protection)

		// System Programs
		solana.TokenProgramID,

		// Raydium AMM Accounts
		p.PoolID,                       // AMM Pool ID
		authority,                      // AMM Authority (PDA)
		ammInfo.PoolState.OpenOrders,   // AMM Open Orders
		ammInfo.PoolState.TargetOrders, // AMM Target Orders
		ammInfo.PoolState.TokenCoin,    // AMM Coin Vault (Base)
		ammInfo.PoolState.TokenPc,      // AMM PC Vault (Quote)

		// Serum/OpenBook Market Accounts
		ammInfo.PoolState.SerumDex, // Serum Program ID
		ammInfo.PoolState.Market,   // Market ID
		ammInfo.Market.Bids,        // Market Bids
		ammInfo.Market.Asks,        // Market Asks
		ammInfo.Market.EventQ,      // Market Event Queue
		ammInfo.Market.CoinVault,   // Market Coin Vault (Base)
		ammInfo.Market.PcVault,     // Market PC Vault (Quote)
		serumVaultSigner,           // Market Vault Signer (PDA from Nonce)

		// User Accounts
		p.UserSourceToken, // User Source (Input) - Direction defined here
		p.UserDestToken,   // User Destination (Output)
		p.UserWallet,      // User Owner (Signer)
	)
	if err != nil {
		return nil, apperrors.Internal("anchor-gen NewSwapBaseInputInstruction: %w", err)
	}
	return instr, nil
}
