package solanaws

import (
	"context"
	"fmt"
	"mm/config"
	"mm/pkg/apperrors"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type Client struct {
	*ws.Client
}

func NewClient(c *config.Config) (*Client, error) {
	wc, err := ws.Connect(context.Background(), c.App.SolanaWSURL)

	if err != nil {
		return nil, err
	}

	return &Client{Client: wc}, nil
}

func (c *Client) AwaitConfirmationTransaction(
	ctx context.Context,
	sig solana.Signature,
	timeout time.Duration,
) (*time.Time, error) {

	sub, err := c.Client.SignatureSubscribe(sig, rpc.CommitmentConfirmed)

	if err != nil {
		return nil, apperrors.Internal("failed to subscribe: %w", err)
	}

	defer sub.Unsubscribe()

	select {
	case _, ok := <-sub.Response():

		confirmedAt := time.Now()

		if !ok {
			return nil, apperrors.Internal("subscription channel closed unexpectedly")
		}
		return &confirmedAt, nil

	case err = <-sub.Err():
		return nil, apperrors.Internal("subscription error: %w", err)

	case <-time.After(timeout):
		return nil, apperrors.Internal(fmt.Sprintf("timeout waiting for transaction %s", sig))

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) SubscribeToSlotUpdate(_ context.Context) (*ws.SlotSubscription, error) {
	subscription, err := c.Client.SlotSubscribe()
	if err != nil {
		return nil, apperrors.Internal("failed to subscribe to slot updates", err)
	}

	return subscription, err
}
