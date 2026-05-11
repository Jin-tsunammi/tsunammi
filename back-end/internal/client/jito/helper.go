package jito

import (
	"context"
	"errors"
	"mm/internal/model"
)

func GetTipByTransactionSpeed(ctx context.Context, tipFloor *GetTipFloorResponse, transactionSpeed model.TransactionSpeed) (float64, error) {
	select {
	case <-ctx.Done():
		return 0.0, ctx.Err()
	default:
		switch transactionSpeed {
		case model.Default:
			return 0.001, nil
		case model.Fast:
			return tipFloor.LandedTips95ThPercentile, nil
		case model.Extra:
			return tipFloor.LandedTips99ThPercentile, nil
		default:
			return 0.0, errors.New("invalid transaction speed")
		}
	}
}
