package solanarpc

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"mm/config"
	"mm/pkg/apperrors"
	"slices"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/vote"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/sync/errgroup"
)

var _ SolanaRPC = (*solanaRPCClient)(nil)

const SolanaRawAmountMultiplier float64 = 1_000_000_000
const MaxGetMultipleAccountsKeys = 100
const MaxParallelGetMultipleAccountsRequests = 30
const AssociatedTokenAccountSize = 165

//go:generate mockgen -source=solana.go -destination=mocks/solana_mock.go -package=solMocks
type SolanaRPC interface {
	GetWalletBalance(ctx context.Context, wallet solana.PublicKey) (float64, error)
	GetWalletTransactions(ctx context.Context, req *GetTransactionsReq) ([]WalletTransaction, error)
	GetMulltipyWalletBalance(ctx context.Context, addresses []solana.PublicKey) ([]float64, error)
	GetLatestBlockhash(ctx context.Context) (*solana.Hash, error)
	WaitForTransactionConfirmation(ctx context.Context, txHash solana.Signature, timeout time.Duration) (int64, error)
	GetAccountInfo(ctx context.Context, pubkey solana.PublicKey) (*rpc.GetAccountInfoResult, error)
	GetTokenAccountBalance(ctx context.Context, pubkey solana.PublicKey) (*rpc.GetTokenAccountBalanceResult, error)
	GetMultipleAccounts(ctx context.Context, pubkeys ...solana.PublicKey) (*rpc.GetMultipleAccountsResult, error)
	SendTransaction(ctx context.Context, transaction *solana.Transaction) (*solana.Signature, error)
	SimulateTransaction(ctx context.Context, transaction *solana.Transaction, opts *rpc.SimulateTransactionOpts) (*rpc.SimulateTransactionResponse, error)
	GetSlotLeadersWithRange(ctx context.Context, start, end uint64) ([]solana.PublicKey, error)
	GetCurrentSlot(ctx context.Context) (uint64, error)
	GetMultipleAccountsWithNoLimits(ctx context.Context, pubkeys ...solana.PublicKey) ([]*rpc.GetMultipleAccountsResult, error)
	GetVoteAccountsNodeKeys(ctx context.Context, votePubkeys ...solana.PublicKey) ([]solana.PublicKey, error)
	GetATARentExemption(ctx context.Context) (uint64, error)
	GetRentExemption(ctx context.Context, size uint64) (uint64, error)
	GetAverageSlotTime(ctx context.Context) (map[time.Duration]time.Duration, error)
	GetTransaction(ctx context.Context, txHash solana.Signature, opts *rpc.GetTransactionOpts) (*rpc.GetTransactionResult, error)
	GetTokenAccountsByOwner(ctx context.Context, owner solana.PublicKey, conf *rpc.GetTokenAccountsConfig, opts *rpc.GetTokenAccountsOpts) (out *rpc.GetTokenAccountsResult, err error)
	SendTransactionWithOpts(ctx context.Context, transaction *solana.Transaction, opts rpc.TransactionOpts) (signature solana.Signature, err error)
}

type solanaRPCClient struct {
	*rpc.Client
	retries int
}

func NewSolanaRPCClient(c *config.Config) SolanaRPC {

	return &solanaRPCClient{
		Client:  rpc.New(c.App.SolanaRPCURL),
		retries: 1,
	}
}

func WithRetries(solanaClient SolanaRPC, retries int) SolanaRPC {
	cloned := *solanaClient.(*solanaRPCClient)
	cloned.retries = retries
	return &cloned
}

func (c *solanaRPCClient) GetATARentExemption(ctx context.Context) (uint64, error) {
	rent, err := c.GetRentExemption(ctx, AssociatedTokenAccountSize)
	if err != nil {
		return 0.0, apperrors.Internal("failed to get associated token account rent exemption", err)
	}
	return rent, nil
}

func (c *solanaRPCClient) GetRentExemption(ctx context.Context, size uint64) (uint64, error) {
	rent, err := c.GetMinimumBalanceForRentExemption(ctx, size, rpc.CommitmentConfirmed)
	if err != nil {
		return 0.0, apperrors.Internal("failed to get rent exemption", err)
	}
	return rent, nil
}

func (c *solanaRPCClient) GetWalletBalance(
	ctx context.Context,
	address solana.PublicKey,
) (float64, error) {
	return DoWithRetry[float64](ctx, c.retries, func() (float64, error) {
		resp, err := c.GetBalance(ctx, address, rpc.CommitmentConfirmed)
		if err != nil {
			return 0.0, apperrors.Internal("failed to get wallet balance", err)
		}

		return LamportsToSOL(int64(resp.Value)), nil
	})

}

func (c *solanaRPCClient) GetWalletTransactions(
	ctx context.Context,
	req *GetTransactionsReq,
) ([]WalletTransaction, error) {
	return DoWithRetry(ctx, c.retries, func() ([]WalletTransaction, error) {
		resp, err := c.GetSignaturesForAddressWithOpts(
			ctx,
			req.Address,
			&rpc.GetSignaturesForAddressOpts{
				Limit:  &req.Limit,
				Before: req.Before,
			},
		)
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprintf("failed to get signatures for solana wallet: %s", req.Address.String()), err)
		}

		var walletTransactions = make([]WalletTransaction, len(resp))
		for i, t := range resp {
			balanceChange, err := c.getTransactionBalanceChange(ctx, req.Address, t.Signature)
			if err != nil {
				return nil, err
			}

			walletTransactions[i] = WalletTransaction{
				Transaction: t.Signature.String(),
				Amount:      balanceChange,
				Timestamp:   t.BlockTime.Time().UTC().UnixMilli(),
			}
		}

		return walletTransactions, nil

	},
	)
}

func (c *solanaRPCClient) GetMulltipyWalletBalance(ctx context.Context, addresses []solana.PublicKey) ([]float64, error) {
	chunkedAddresses := ChunkSlice(addresses, MaxGetMultipleAccountsKeys)

	slices.Chunk(addresses, MaxGetMultipleAccountsKeys)

	balances := make([]float64, 0, len(addresses))
	accounts := make([]*rpc.GetMultipleAccountsResult, len(chunkedAddresses))
	errs := make([]error, len(chunkedAddresses))

	wg := sync.WaitGroup{}
	wg.Add(len(chunkedAddresses))

	semaphore := make(chan struct{}, MaxParallelGetMultipleAccountsRequests)

	for i := 0; i < len(chunkedAddresses); i++ {
		go func(index int) {
			defer wg.Done()

			semaphore <- struct{}{}

			defer func() {
				<-semaphore
			}()

			if err := ctx.Err(); err != nil {
				errs[index] = err
				return
			}

			account, err := c.GetMultipleAccounts(ctx, chunkedAddresses[index]...)

			if err != nil {
				errs[index] = err
				return
			}

			accounts[index] = account
		}(i)
	}

	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	for i := 0; i < len(chunkedAddresses); i++ {
		for j := 0; j < len(accounts[i].Value); j++ {

			account := accounts[i].Value[j]

			if account != nil {
				balances = append(balances, LamportsToSOL(int64(account.Lamports)))
			} else {
				balances = append(balances, 0.0)
			}
		}
	}

	return balances, nil
}

func (c *solanaRPCClient) GetLatestBlockhash(ctx context.Context) (*solana.Hash, error) {

	resp, err := c.Client.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, apperrors.Internal("failed to get latest blockhash", err)
	}

	return &resp.Value.Blockhash, nil

}

func (c *solanaRPCClient) WaitForTransactionConfirmation(ctx context.Context, sig solana.Signature, timeout time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("timeout waiting for %s", rpc.CommitmentConfirmed)
		case <-ticker.C:
			resp, err := c.GetSignatureStatuses(ctx, true, sig)
			if err != nil {
				return 0, fmt.Errorf("GetSignatureStatuses: %w", err)
			}
			if len(resp.Value) == 0 || resp.Value[0] == nil {
				continue
			}
			val := resp.Value[0]
			if val.Err != nil {
				return 0, fmt.Errorf("transaction failed: %v", val.Err)
			}

			if val.ConfirmationStatus == rpc.ConfirmationStatusConfirmed {

				tx, err := c.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
					Encoding:   solana.EncodingBase64,
					Commitment: rpc.CommitmentConfirmed,
				})

				if err != nil {
					return 0, fmt.Errorf("GetTransaction: %w", err)
				}
				return tx.BlockTime.Time().UTC().UnixMilli(), nil
			}
		}
	}
}

func (c *solanaRPCClient) GetTokenAccountBalance(ctx context.Context, pubkey solana.PublicKey) (*rpc.GetTokenAccountBalanceResult, error) {
	return DoWithRetry(ctx, c.retries, func() (*rpc.GetTokenAccountBalanceResult, error) {
		response, err := c.Client.GetTokenAccountBalance(ctx, pubkey, rpc.CommitmentConfirmed)
		if err != nil {
			return nil, apperrors.Internal("failed to get token account balance", err)
		}

		if response.Value == nil {
			return nil, apperrors.Internal("failed to get token account balance: empty response")
		}

		return response, nil
	},
	)
}

func (c *solanaRPCClient) GetMultipleAccountsWithNoLimits(ctx context.Context, pubkeys ...solana.PublicKey) ([]*rpc.GetMultipleAccountsResult, error) {

	if len(pubkeys) == 0 {
		return nil, errors.New("no pubkeys provided")
	}

	chunkedPubkeys := make([][]solana.PublicKey, 0, len(pubkeys)/MaxGetMultipleAccountsKeys+1)

	chunkedPubkeys = slices.AppendSeq(
		chunkedPubkeys,
		slices.Chunk(pubkeys, MaxGetMultipleAccountsKeys),
	)

	accounts := make([]*rpc.GetMultipleAccountsResult, len(chunkedPubkeys))

	eg, errctx := errgroup.WithContext(ctx)

	eg.SetLimit(10)

	for i := range chunkedPubkeys {
		eg.Go(func() error {

			account, err := c.GetMultipleAccounts(errctx, chunkedPubkeys[i]...)
			if err != nil {
				return err
			}

			accounts[i] = account

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, apperrors.Internal("failed to get vote accounts", err)
	}

	return accounts, nil
}

func (c *solanaRPCClient) SendTransaction(ctx context.Context, transaction *solana.Transaction) (*solana.Signature, error) {
	txSignature, err := c.Client.SendTransaction(ctx, transaction)
	if err != nil {
		return nil, err
	}

	return &txSignature, nil
}

func (c *solanaRPCClient) SimulateTransaction(ctx context.Context, transaction *solana.Transaction, opts *rpc.SimulateTransactionOpts) (*rpc.SimulateTransactionResponse, error) {
	return c.SimulateTransactionWithOpts(ctx, transaction, opts)
}

func (c *solanaRPCClient) GetSlotLeadersWithRange(ctx context.Context, start, end uint64) ([]solana.PublicKey, error) {
	const methodLimit uint64 = 5000

	if start >= end {
		return nil, apperrors.Internal("invalid epoch schedule range")
	}

	estimatedCap := (end - start) / methodLimit

	type seg struct {
		start, length uint64
	}

	segments := make([]seg, 0, estimatedCap+1)

	for current := start; current < end; {
		remaining := end - current

		currentLen := methodLimit
		if remaining < methodLimit {
			currentLen = remaining
		}

		segments = append(segments, seg{
			start:  current,
			length: currentLen,
		})

		current += currentLen
	}

	eg, erctx := errgroup.WithContext(ctx)

	eg.SetLimit(10)

	leaderSegments := make([][]solana.PublicKey, len(segments))

	for i, segment := range segments {
		eg.Go(func() error {

			leaders, err := c.GetSlotLeaders(erctx, segment.start, segment.length)

			if err != nil {
				return err
			}

			leaderSegments[i] = leaders

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, apperrors.Internal("failed to get slot leaders", err)
	}

	leaders := make([]solana.PublicKey, 0, end-start+1)

	for _, segment := range leaderSegments {
		leaders = append(leaders, segment...)
	}

	return leaders, nil
}

func (c *solanaRPCClient) GetVoteAccountsNodeKeys(ctx context.Context, votePubkeys ...solana.PublicKey) ([]solana.PublicKey, error) {

	validatorPubKeys := make([]solana.PublicKey, 0, len(votePubkeys))

	accounts, err := c.GetMultipleAccountsWithNoLimits(ctx, votePubkeys...)

	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.Value == nil {
			return nil, apperrors.Internal("failed to get vote accounts: empty response")
		}
		for _, accountInfo := range account.Value {

			if !accountInfo.Owner.Equals(vote.ProgramID) {
				return nil, apperrors.Internal("failed to get vote accounts: invalid account owner")
			}

			data := accountInfo.Data.GetBinary()

			if data == nil {
				return nil, apperrors.Internal("failed to unmarshal vote state: empty data")
			}

			if len(data) < 4 || len(data) < 36 {
				return nil, apperrors.Internal("failed to unmarshal vote state: invalid data length")
			}

			discriminator := binary.LittleEndian.Uint32(data[:4])

			var nodePubkey solana.PublicKey

			switch discriminator {
			case VoteStateDiscriminatorV3:
				v := &VoteStateV3{}
				err := bin.UnmarshalBorsh(v, data[4:])
				if err != nil {
					return nil, apperrors.Internal("failed to unmarshal vote state v3", err)
				}
				nodePubkey = v.NodePubkey
			default:
				return nil, apperrors.Internal("failed to unmarshal vote state: unknown discriminator")
			}

			validatorPubKeys = append(validatorPubKeys, nodePubkey)
		}
	}

	return validatorPubKeys, nil
}

func (c *solanaRPCClient) GetCurrentSlot(ctx context.Context) (uint64, error) {
	return c.GetSlot(ctx, rpc.CommitmentConfirmed)
}

func (c *solanaRPCClient) GetAverageSlotTime(ctx context.Context) (map[time.Duration]time.Duration, error) {

	roundFloat := func(val float64, precision uint) float64 {
		ratio := math.Pow(10, float64(precision))
		return math.Round(val*ratio) / ratio
	}

	intervals := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		45 * time.Minute,
		60 * time.Minute,
	}

	maxDuration := intervals[len(intervals)-1]

	limit := uint(maxDuration.Seconds()/60) + 10

	samples, err := c.GetRecentPerformanceSamples(ctx, &limit)
	if err != nil {
		return nil, fmt.Errorf("rpc error: %w", err)
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("no data returned")
	}

	results := make(map[time.Duration]time.Duration, len(intervals))

	var totalSlots uint64
	var totalSeconds uint64

	currentIntervalIdx := 0

	for _, sample := range samples {
		totalSlots += sample.NumSlots
		totalSeconds += uint64(sample.SamplePeriodSecs)

		currentDuration := time.Duration(totalSeconds) * time.Second

		for currentIntervalIdx < len(intervals) && currentDuration >= intervals[currentIntervalIdx] {

			targetInterval := intervals[currentIntervalIdx]

			avg := float64(totalSeconds) / float64(totalSlots)
			avg = roundFloat(avg, 4)
			results[targetInterval] = time.Duration(avg * float64(time.Second))

			currentIntervalIdx++
		}

		if currentIntervalIdx >= len(intervals) {
			break
		}
	}

	return results, nil

}
