package jito

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"mm/config"
	pumpamm "mm/internal/client/pumpfun/amm/amm_client"
	bonding "mm/internal/client/pumpfun/bonding/bonding_client"
	raydiumamm "mm/internal/client/raydium/ammv4/ammv4_client"
	raydiumcpswap "mm/internal/client/raydium/cpmm/cpmm_client"
	"mm/internal/client/solanarpc"
	"mm/internal/client/solanaws"
	"mm/internal/model"
	"mm/internal/swaperror"
	"mm/pkg/apperrors"
	"mm/pkg/pool"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/weeaa/jito-go/clients/searcher_client"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	BundleLimit = 5
)

var FailedToProcessBundle = apperrors.Internal("failed to process bundle")
var BundleNotAccepted = apperrors.Internal("bundle not accepted")

type Client struct {
	restyClient   *resty.Client
	jitoClients   *pool.CloseableRoundRobin[*searcher_client.Client]
	solanaWS      *solanaws.Client
	solanaRPC     solanarpc.SolanaRPC
	tipAccounts   []solana.PublicKey
	networkURL    string
	bundleURL     string
	rpcURL        string
	logger        *zap.Logger
	bundleTimeout time.Duration
}

func NewClient(restyClient *resty.Client, solanaWS *solanaws.Client, solanaRPC solanarpc.SolanaRPC, jitoClients *pool.CloseableRoundRobin[*searcher_client.Client], tipAccounts []solana.PublicKey, config *config.Config, logger *zap.Logger) *Client {

	return &Client{
		restyClient:   restyClient,
		jitoClients:   jitoClients,
		tipAccounts:   tipAccounts,
		networkURL:    config.Jito.NetworkURL,
		bundleURL:     config.Jito.BundleURL,
		solanaWS:      solanaWS,
		solanaRPC:     solanaRPC,
		logger:        logger,
		rpcURL:        config.App.SolanaRPCURL,
		bundleTimeout: config.Jito.BundleTimeout,
	}
}

func (c *Client) GetTipAccount(ctx context.Context) (*solana.PublicKey, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		index := rand.Intn(len(c.tipAccounts))
		return &c.tipAccounts[index], nil
	}
}

func (c *Client) SendBundle(ctx context.Context, bundleTransactions []*solana.Transaction) (*SendBundleResponse, error) {

	jitoClient, err := c.jitoClients.Get(ctx)

	if err != nil {
		return nil, apperrors.Internal("failed to get a client from pool: ", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		bundle, err := jitoClient.SendBundle(bundleTransactions)
		if err != nil {
			return nil, apperrors.Internal(fmt.Sprint("failed to send bundle to region ", jitoClient.GrpcConn.Target()), err, FailedToProcessBundle)
		}
		return &SendBundleResponse{BundleId: "region " + jitoClient.GrpcConn.Target() + " " + bundle.GetUuid()}, nil

	}
}

func (c *Client) GetTipFloor(ctx context.Context) (*GetTipFloorResponse, error) {
	method := "api/v1/bundles/tip_floor"

	responseList := make([]getTipFloorResponse, 1)

	data, err := c.restyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetResult(&responseList).
		Get(c.bundleURL + method)

	if err != nil {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") {
			return nil, fmt.Errorf("failed to get tip floor: %w: %w", swaperror.ErrRateLimit, err)
		}
		if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
			return nil, fmt.Errorf("failed to get tip floor: %w: %w", swaperror.ErrGatewayTimeout, err)
		}
		return nil, fmt.Errorf("failed to get tip floor: %w", err)
	}

	if data.IsError() {
		if data.StatusCode() == 429 {
			return nil, fmt.Errorf("failed to get tip floor (HTTP %d): %w", data.StatusCode(), swaperror.ErrRateLimit)
		}
		if data.StatusCode() == 504 {
			return nil, fmt.Errorf("failed to get tip floor (HTTP %d): %w", data.StatusCode(), swaperror.ErrGatewayTimeout)
		}
		return nil, fmt.Errorf("failed to get tip floor (HTTP %d): %s", data.StatusCode(), data.String())
	}

	response := responseList[0]

	tipFloorResponse := GetTipFloorResponse{
		Time:                        response.Time,
		LandedTips25ThPercentile:    response.LandedTips25ThPercentile,
		LandedTips50ThPercentile:    response.LandedTips50ThPercentile,
		LandedTips75ThPercentile:    response.LandedTips75ThPercentile,
		LandedTips95ThPercentile:    response.LandedTips95ThPercentile,
		LandedTips99ThPercentile:    response.LandedTips99ThPercentile,
		EmaLandedTips50ThPercentile: response.EmaLandedTips50ThPercentile,
	}

	return &tipFloorResponse, nil
}

func (c *Client) GetValidatorsCurrentEpoch(ctx context.Context) ([]ValidatorResponse, error) {
	return c.getValidators(ctx, nil)
}

func (c *Client) GetValidators(ctx context.Context, epoch int64) ([]ValidatorResponse, error) {
	req := GetValidatorsRequest{Epoch: epoch}
	return c.getValidators(ctx, &req)
}

func (c *Client) getValidators(ctx context.Context, req *GetValidatorsRequest) ([]ValidatorResponse, error) {
	method := "api/v1/validators"

	response := GetValidatorsResponse{}

	data, err := c.restyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&response).
		Post(c.networkURL + method)

	if err != nil {
		return nil, apperrors.Internal("failed to get validators: ", err)
	}

	if data.IsError() {
		return nil, apperrors.Internal("failed to get data: ", data.Err)
	}

	return response.Validators, nil
}

func (c *Client) Close() error {
	return c.jitoClients.Close()
}

func (c *Client) CalculateTip(ctx context.Context, speed model.TransactionSpeed) (float64, error) {
	tipFloor, err := c.GetTipFloor(ctx)
	if err != nil {
		return 0.0, apperrors.Internal("failed to get tip floor: %w", err)
	}

	tipAmount, err := GetTipByTransactionSpeed(ctx, tipFloor, speed)
	if err != nil {
		return 0.0, apperrors.Internal("failed to calculate tip amount: %w", err)
	}

	return tipAmount, nil
}

func (c *Client) BroadcastBundle(
	ctx context.Context,
	bundle []*solana.Transaction,
	poolProgramID solana.PublicKey,
) error {
	gCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, errCtx := errgroup.WithContext(gCtx)

	tipTx := bundle[len(bundle)-1].Signatures[0]

	eg.Go(func() error {
		_, err := c.solanaWS.AwaitConfirmationTransaction(errCtx, tipTx, c.bundleTimeout)
		if err != nil {
			return fmt.Errorf("bundle rejected while awaiting confirmation: %w: %w", swaperror.ErrBundleRejected, err)
		}
		return nil
	})

	err := c.SimulateBundle(ctx, bundle, poolProgramID)
	if err != nil {
		cancel()
		c.logger.Error("failed to simulate bundle", zap.Error(err))
		return err
	}

	// bundleID, err := c.SendBundle(ctx, bundle)
	// if err != nil {
	// 	cancel()
	// 	c.logger.Error("failed to send bundle", zap.Error(err))
	// 	return apperrors.Internal("jito send bundle failed: %w", err)
	// }

	bundleID := "mock"
	c.logger.Info("bundle sent", zap.String("bundle_id", bundleID))
	c.logger.Info("tip transaction", zap.String("tip_tx", tipTx.String()))

	if err = eg.Wait(); err != nil {
		return err
	}

	return nil
}

func BuildTipTransaction(
	payer solana.PrivateKey,
	tipAmount float64,
	tipAccount solana.PublicKey,
	blockHash solana.Hash,
) (*solana.Transaction, error) {
	tipIx := system.NewTransferInstruction(
		solanarpc.SOLToLamports(tipAmount),
		payer.PublicKey(),
		tipAccount,
	).Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{tipIx},
		blockHash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		return nil, apperrors.Internal("failed to create tip tx: %w", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(payer.PublicKey()) {
			return &payer
		}
		return nil
	})

	return tx, nil
}

func (c *Client) SimulateBundle(ctx context.Context, tsx []*solana.Transaction, poolProgramID solana.PublicKey) error {

	encodedTxs := make([]string, len(tsx))
	for index, tx := range tsx {
		rawTx, err := tx.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to deserialize transaction %d: %w", index, err)
		}

		encodedTxs[index] = base64.StdEncoding.EncodeToString(rawTx)
	}

	accountConfig := map[string]interface{}{
		"addresses": []string{solana.TokenProgramID.String()},
		"encoding":  solana.EncodingBase64,
	}

	accountsConfigsList := make([]interface{}, len(encodedTxs))
	for i := range encodedTxs {
		accountsConfigsList[i] = accountConfig
	}

	payload := struct {
		JSONRPC string        `json:"jsonrpc"`
		ID      string        `json:"id"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "simulateBundle",
		Params: []interface{}{
			struct {
				EncodedTransactions []string `json:"encodedTransactions"`
			}{
				EncodedTransactions: encodedTxs,
			},
			struct {
				PreExecutionAccountsConfigs  []interface{} `json:"preExecutionAccountsConfigs"`
				PostExecutionAccountsConfigs []interface{} `json:"postExecutionAccountsConfigs"`
				SkipSigVerify                bool          `json:"skipSigVerify"`
				SimulationBank               interface{}   `json:"simulationBank"`
				TransactionEncoding          string        `json:"transactionEncoding"`
				ReplaceRecentBlockhash       bool          `json:"replaceRecentBlockhash"`
			}{
				PreExecutionAccountsConfigs:  accountsConfigsList,
				PostExecutionAccountsConfigs: accountsConfigsList,
				SkipSigVerify:                true,
				SimulationBank: map[string]interface{}{
					"commitment": map[string]string{"commitment": "processed"},
				},
				TransactionEncoding:    "base64",
				ReplaceRecentBlockhash: false,
			},
		},
	}

	var response simulateBundleResponse

	data, err := c.restyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&response).
		Post(c.rpcURL)

	if err != nil {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "429") || strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") {
			return fmt.Errorf("failed to send request: %w: %w", swaperror.ErrRateLimit, err)
		}
		if strings.Contains(message, "gateway timeout") || strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout") {
			return fmt.Errorf("failed to send request: %w: %w", swaperror.ErrGatewayTimeout, err)
		}
		return fmt.Errorf("failed to send request: %w", err)
	}

	if data.IsError() {
		if data.StatusCode() == 429 {
			return fmt.Errorf("RPC return the following (HTTP %d): %w", data.StatusCode(), swaperror.ErrRateLimit)
		}
		if data.StatusCode() == 504 {
			return fmt.Errorf("RPC return the following (HTTP %d): %w", data.StatusCode(), swaperror.ErrGatewayTimeout)
		}
		return fmt.Errorf("RPC return the following (HTTP %d): %s", data.StatusCode(), data.String())
	}

	summary := response.Result.Value.Summary

	switch summary := summary.(type) {
	case string:
		if summary == "success" {
			return nil
		}
	case map[string]interface{}:
		bytes, err := json.Marshal(summary)
		if err != nil {
			return fmt.Errorf("failed to marshal map: %w", err)
		}

		var simErr SimulationError
		err = json.Unmarshal(bytes, &simErr)
		if err != nil {
			return fmt.Errorf("failed to unmarshal into struct: %w", err)
		}

		code, err := simErr.ParseErrorCode()
		if err != nil {
			return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSimulationError, simErr)
		}

		var customError error
		var ok bool

		switch poolProgramID {
		case raydiumcpswap.ProgramID:
			customError, ok = raydiumcpswap.Errors[code]
			if ok {
				if customError == raydiumcpswap.ErrExceededSlippage {
					return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSlippageExceeded, customError)
				}
				if strings.Contains(strings.ToLower(customError.Error()), "compute") {
					return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrComputeBudgetExceeded, customError)
				}
				if strings.Contains(strings.ToLower(customError.Error()), "empty") {
					return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrInsufficientFunds, customError)
				}
				return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrCustomProgramError, customError)
			}
		case raydiumamm.ProgramID:
			customError, ok = raydiumamm.Errors[code]
			if ok {
				if customError == raydiumamm.ErrExceededSlippage {
					return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSlippageExceeded, customError)
				}
				if strings.Contains(strings.ToLower(customError.Error()), "empty funds") {
					return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrInsufficientFunds, customError)
				}
				return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrCustomProgramError, customError)
			}
		case pumpamm.ProgramID:
			if code == 6004 || code == 6040 {
				return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSlippageExceeded, simErr)
			}
			if code == 6039 {
				return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrInsufficientFunds, simErr)
			}
			return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrCustomProgramError, simErr)
		case bonding.ProgramID:
			if code == 6023 || code == 6040 || code == 6041 {
				return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrInsufficientFunds, simErr)
			}
			if code == 6042 {
				return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSlippageExceeded, simErr)
			}
			return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrCustomProgramError, simErr)
		default:
			return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSimulationError, simErr)
		}

		return fmt.Errorf("bundle simulation failed: %w: %w", swaperror.ErrSimulationError, simErr)

	default:
		return fmt.Errorf("failed to unmarshal into struct: %w", err)

	}

	return nil
}
