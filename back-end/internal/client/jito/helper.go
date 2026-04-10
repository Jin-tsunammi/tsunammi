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
			return 0.0012, nil
		case model.Extra:
			return 0.0015, nil
		default:
			return 0.0, errors.New("invalid transaction speed")
		}
	}
}
