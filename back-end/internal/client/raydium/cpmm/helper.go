package cpmm

import (
	"mm/internal/client/raydium"
	raydiumcpswap "mm/internal/client/raydium/cpmm/cpmm_client"
	"mm/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
)

func cpmmGetAuthority(programID solana.PublicKey) (solana.PublicKey, error) {
	addr, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("vault_and_lp_mint_auth_seed")},
		programID,
	)
	if err != nil {
		return solana.PublicKey{}, err
	}

	return addr, err
}

func cpmmBuildSwapInstruction(ammConfigID, poolSourceToken, poolDestToken solana.PublicKey, p *raydium.SwapParams) (solana.Instruction, error) {
	amountInU64, err := raydium.SafeUint64(p.AmountIn)
	if err != nil {
		return nil, apperrors.Internal("AmountIn overflow: %w", err)
	}
	minAmountOutU64, err := raydium.SafeUint64(p.MinAmountOut)
	if err != nil {
		return nil, apperrors.Internal("MinAmountOut overflow: %w", err)
	}

	authority, err := cpmmGetAuthority(raydiumcpswap.ProgramID)

	if err != nil {
		return nil, apperrors.Internal("failed to get authority: %w", err)
	}

	instr, err := raydiumcpswap.NewSwapBaseInputInstruction(
		// Args:
		amountInU64,     // `amount_in`
		minAmountOutU64, // `minimum_amount_out`
		// Accounts:
		p.UserWallet,           // 'payer'
		authority,              // 'authority'
		ammConfigID,            // 'amm_config'
		p.PoolID,               // 'pool_state'
		p.UserSourceToken,      // 'input_token_account'
		p.UserDestToken,        // 'output_token_account'
		poolSourceToken,        // 'input_vault'
		poolDestToken,          // 'output_vault'
		p.InputTokenProgramID,  // 'input_token_program'
		p.OutputTokenProgramID, // 'output_token_program'
		p.InputTokenMint,       // 'input_token_mint'
		p.OutputTokenMint,      // 'output_token_mint'
		p.ObservationState,     // 'observation_state'
	)
	if err != nil {
		return nil, apperrors.Internal("anchor-gen NewSwapBaseInputInstruction: %w", err)
	}
	return instr, nil
}
