package solutil

import (
	"fmt"
	"mm/pkg/apperrors"

	"github.com/gagliardetto/solana-go"
	assoc "github.com/gagliardetto/solana-go/programs/associated-token-account"
)

const NativeSolMint = "So11111111111111111111111111111111111111111"

var OldTokenProgramID = solana.TokenProgramID
var NewTokenProgramID = solana.Token2022ProgramID

var NativeSolPubkey = solana.MustPublicKeyFromBase58(NativeSolMint)

func IsSOLLikeMint(mint solana.PublicKey) bool {
	return mint.Equals(solana.WrappedSol) ||
		mint.Equals(NativeSolPubkey) // enables bonding curve
}

func FindAssociatedTokenAddressWithProgram(owner, mint, programID solana.PublicKey) (solana.PublicKey, uint8, error) {
	var ata solana.PublicKey
	var bump uint8
	var err error

	if programID.String() == OldTokenProgramID.String() {
		ata, bump, err = solana.FindAssociatedTokenAddress(owner, mint)
	} else if programID.String() == NewTokenProgramID.String() {
		ata, bump, err = solana.FindProgramAddress(
			[][]byte{
				owner.Bytes(),
				NewTokenProgramID.Bytes(),
				mint.Bytes(),
			},
			assoc.ProgramID,
		)
	} else {
		return solana.PublicKey{}, 0, apperrors.Internal(
			"unsupported token program",
			fmt.Errorf("program id: %s", programID.String()),
		)
	}

	if err != nil {
		return solana.PublicKey{}, 0, apperrors.Internal("failed to get user associated token address", err)
	}

	return ata, bump, nil
}

func NewCreateAssociatedTokenAccountInstruction(payer, owner, mint, tokenProgram solana.PublicKey) (solana.Instruction, error) {
	associatedTokenAddress, _, err := FindAssociatedTokenAddressWithProgram(owner, mint, tokenProgram)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		assoc.ProgramID,
		solana.AccountMetaSlice{
			solana.Meta(payer).WRITE().SIGNER(),
			solana.Meta(associatedTokenAddress).WRITE(),
			solana.Meta(owner),
			solana.Meta(mint),
			solana.Meta(solana.SystemProgramID),
			solana.Meta(tokenProgram),
		},
		nil,
	), nil
}
