package solanarpc

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mm/pkg/apperrors"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func (c *solanaRPCClient) getTransactionBalanceChange(
	ctx context.Context,
	address solana.PublicKey,
	txHash solana.Signature,
) (json.Number, error) {
	// Fetch the transaction
	resp, err := c.Client.GetTransaction(
		ctx,
		txHash,
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: rpc.CommitmentConfirmed,
		},
	)
	if err != nil {
		return "", apperrors.Internal("failed to get transaction from solana node", err)
	}
	if resp == nil || resp.Transaction == nil || resp.Meta == nil {
		return "", apperrors.Internal("empty solana node response")
	}
	txn, err := resp.Transaction.GetTransaction()
	if err != nil {
		return "", apperrors.Internal("failed to get transaction from solana node", err)
	}

	keys, err := txn.Message.GetAllKeys()
	if err != nil {
		keys = txn.Message.AccountKeys
	}

	idx := -1
	for i, k := range keys {
		if k.Equals(address) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return "", apperrors.Internal("account not found in transaction account keys", err)
	}

	if idx >= len(resp.Meta.PreBalances) ||
		idx >= len(resp.Meta.PostBalances) {
		return "", apperrors.Internal("pre/post balances do not contain account index", err)
	}

	// Calculate diff
	diff := int64(resp.Meta.PostBalances[idx]) - int64(resp.Meta.PreBalances[idx])
	return json.Number(fmt.Sprintf("%.9f", LamportsToSOL(diff))), nil
}

func LamportsToSOL(lamports int64) float64 {
	return float64(lamports) / SolanaRawAmountMultiplier
}

func SOLToLamports(sol float64) uint64 {
	return uint64(sol * SolanaRawAmountMultiplier)
}

func ChunkSlice(items []solana.PublicKey, chunkSize int) [][]solana.PublicKey {

	if chunkSize <= 0 {
		return nil
	}

	totalItems := len(items)
	chunks := make([][]solana.PublicKey, 0, (totalItems/chunkSize)+1)

	for i := 0; i < totalItems; i += chunkSize {

		end := i + chunkSize

		if end > totalItems {
			end = totalItems
		}

		chunks = append(chunks, items[i:end])
	}

	return chunks
}

func DoWithRetry[T any](ctx context.Context, retry int, fn func() (T, error)) (T, error) {
	var zero T

	backoff := 250 * time.Millisecond

	for i := 0; i < retry; i++ {
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		t, err := fn()
		if err != nil {
			if i == retry-1 {
				return zero, fmt.Errorf("after %d attempts, last error: %w", retry, err)
			}
			j := randInt63n(int64(backoff))
			time.Sleep(backoff + time.Duration(j))
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			continue
		}
		return t, nil
	}
	return zero, errors.New("unreachable code")
}

func randInt63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return time.Now().UnixNano() % n
	}
	v := int64(binary.LittleEndian.Uint64(b[:]) & (1<<63 - 1))
	return v % n
}

func ToAtomicUnit(amount float64, decimals uint8) uint64 {
	multiplier := math.Pow(10, float64(decimals))
	result := amount * multiplier

	return uint64(result)
}

func FromAtomicUnit(amount uint64, decimals uint8) float64 {
	multiplier := math.Pow(10, float64(decimals))
	result := float64(amount) / multiplier

	return result
}

func TokenToSOLPrice(tokenUSDPrice float64, solUSDPrice float64) float64 {
	if tokenUSDPrice == 0 || solUSDPrice == 0 {
		return 0
	}

	return tokenUSDPrice / solUSDPrice
}

func TokenSOLToUSD(tokenPriceInSOL float64, solUSDPrice float64) float64 {
	return tokenPriceInSOL * solUSDPrice
}
