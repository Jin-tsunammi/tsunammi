package solutil

import "github.com/gagliardetto/solana-go"

const NativeSolMint = "So11111111111111111111111111111111111111111"

var NativeSolPubkey = solana.MustPublicKeyFromBase58(NativeSolMint)

func IsSOLLikeMint(mint solana.PublicKey) bool {
	return mint.Equals(solana.WrappedSol) ||
		mint.Equals(NativeSolPubkey) // enables bonding curve
}
