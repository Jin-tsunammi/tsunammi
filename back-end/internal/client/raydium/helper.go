package raydium

import (
	"errors"
	"fmt"
	"math/big"
	"mm/internal/model"
	"mm/pkg/apperrors"
)

func SafeUint64(v *big.Int) (uint64, error) {
	if v == nil {
		return 0, apperrors.Internal("nil big.Int")
	}
	if v.BitLen() > 64 {
		return 0, apperrors.Internal(fmt.Sprintf("value too large for uint64: bitlen=%d", v.BitLen()))
	}
	if v.Sign() < 0 {
		return 0, apperrors.Internal("negative value for uint64")
	}
	return v.Uint64(), nil
}

func IsTargetReached(current, goal *big.Rat, taskType string) bool {
	return (taskType == model.TargetUpTaskType && current.Cmp(goal) > 0) ||
		(taskType == model.TargetDownTaskType && current.Cmp(goal) < 0)
}

func IsAllErrorAre(errs []error, err error) bool {

	for _, e := range errs {
		if !errors.Is(e, err) {
			return false
		}
	}

	return true
}
