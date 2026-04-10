package solanarpc

import (
	"encoding/json"

	"github.com/gagliardetto/solana-go"
)

const (
	VoteStateDiscriminatorV1 = 0
	VoteStateDiscriminatorV2 = 1
	VoteStateDiscriminatorV3 = 2
	VoteStateDiscriminatorV4 = 3
)

type WalletTransaction struct {
	Transaction string      `json:"transaction" example:"1234567890"`
	Amount      json.Number `json:"amount" swaggertype:"primitive,number" example:"100.294"`
	Timestamp   int64       `json:"timestamp" example:"1651366400"`
}

type GetTransactionsReq struct {
	Address solana.PublicKey
	Limit   int
	Before  solana.Signature
}

type VoteStateV3 struct {
	NodePubkey           solana.PublicKey
	AuthorizedWithdrawer solana.PublicKey
	Commission           uint8

	Votes []LandedVote

	RootSlot uint64

	AuthorizedVoters []AuthorizedVoterItem

	PriorVoterTuple [32]PriorVoter

	Idx uint64

	IsEmpty bool

	EpochCredits []EpochCredit

	Slot      uint64
	Timestamp int64
}

type LandedVote struct {
	Slot              uint64
	ConfirmationCount uint32
}

type AuthorizedVoterItem struct {
	Epoch            uint64
	AuthorizedSigner solana.PublicKey
}

type EpochCredit struct {
	Epoch       uint64
	Credits     uint64
	PrevCredits uint64
}

type PriorVoter struct {
	AuthorizedPubkey solana.PublicKey
	EpochStart       uint64
	EpochEnd         uint64
}
