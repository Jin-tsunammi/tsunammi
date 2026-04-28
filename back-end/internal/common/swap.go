package common

import (
	"crypto/rand"
	"errors"
	"math/big"
	"mm/pkg/apperrors"
)

const SignatureFeeReserveLamports uint64 = 5000

func SelectTxAmountInRange(remaining, min, max *big.Int) (*big.Int, error) {
	if remaining == nil || remaining.Sign() <= 0 {
		return nil, errors.New("budget exceeded")
	}

	if min.Cmp(max) > 0 {
		return nil, errors.New("min greater than max")
	}

	if remaining.Cmp(min) < 0 {
		return nil, errors.New("remaining below minimum transaction amount")
	}

	if remaining.Cmp(max) <= 0 {
		return remaining, nil
	}

	return RandBigIntRange(min, max)
}

func SolPayerReserveLamports(createOutputATA bool, ataRentLamports uint64, computeUnitLimit uint32, priorityFeeMicroLamports uint64) uint64 {
	reserve := ataRentLamports + SignatureFeeReserveLamports + priorityFeeLamports(computeUnitLimit, priorityFeeMicroLamports)
	if createOutputATA {
		reserve += ataRentLamports
	}

	return reserve
}

func priorityFeeLamports(computeUnitLimit uint32, priorityFeeMicroLamports uint64) uint64 {
	totalMicroLamports := uint64(computeUnitLimit) * priorityFeeMicroLamports
	return (totalMicroLamports + 999_999) / 1_000_000
}

func RandBigIntRange(min, max *big.Int) (*big.Int, error) {

	delta := new(big.Int).Sub(max, min)

	delta.Add(delta, big.NewInt(1))

	if delta.Sign() <= 0 {
		return nil, apperrors.Internal("max must be greater than min")
	}

	randomOffset, err := rand.Int(rand.Reader, delta)
	if err != nil {
		return nil, apperrors.Internal("failed to generate random offset", err)
	}

	return new(big.Int).Add(randomOffset, min), nil
}
