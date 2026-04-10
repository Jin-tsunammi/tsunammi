package jito

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	JsonRpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type GetTipResponse struct {
	Result []string `json:"result"`
}

type SendTransactionResponse struct {
	Result string `json:"result"`
}

type getTipFloorResponse struct {
	Time                        time.Time `json:"time"`
	LandedTips25ThPercentile    float64   `json:"landed_tips_25th_percentile"`
	LandedTips50ThPercentile    float64   `json:"landed_tips_50th_percentile"`
	LandedTips75ThPercentile    float64   `json:"landed_tips_75th_percentile"`
	LandedTips95ThPercentile    float64   `json:"landed_tips_95th_percentile"`
	LandedTips99ThPercentile    float64   `json:"landed_tips_99th_percentile"`
	EmaLandedTips50ThPercentile float64   `json:"ema_landed_tips_50th_percentile"`
}

type GetTipFloorResponse struct {
	Time                        time.Time `json:"time"`
	LandedTips25ThPercentile    float64   `json:"landed_tips_25th_percentile"`
	LandedTips50ThPercentile    float64   `json:"landed_tips_50th_percentile"`
	LandedTips75ThPercentile    float64   `json:"landed_tips_75th_percentile"`
	LandedTips95ThPercentile    float64   `json:"landed_tips_95th_percentile"`
	LandedTips99ThPercentile    float64   `json:"landed_tips_99th_percentile"`
	EmaLandedTips50ThPercentile float64   `json:"ema_landed_tips_50th_percentile"`
}

type GetValidatorsRequest struct {
	Epoch int64 `json:"epoch"`
}

type GetValidatorsResponse struct {
	Validators []ValidatorResponse `json:"validators"`
}

type ValidatorResponse struct {
	IdentityAccount           string `json:"identity_account"`
	VoteAccount               string `json:"vote_account"`
	MevCommissionBps          int    `json:"mev_commission_bps"`
	MevRewards                int    `json:"mev_rewards"`
	PriorityFeeCommissionBps  int    `json:"priority_fee_commission_bps"`
	PriorityFeeRewards        int    `json:"priority_fee_rewards"`
	RunningJito               bool   `json:"running_jito"`
	RunningBam                bool   `json:"running_bam"`
	ActiveStake               int64  `json:"active_stake"`
	JitoDirectedStakeTarget   bool   `json:"jito_directed_stake_target"`
	JitoDirectedStakeLamports int    `json:"jito_directed_stake_lamports"`
}

type SendBundleResponse struct {
	BundleId string `json:"bundle_id"`
}

type simulateBundleResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Context struct {
			Slot       uint64 `json:"slot"`
			ApiVersion string `json:"apiVersion"`
		} `json:"context"`
		Value struct {
			Summary            any                 `json:"summary"` // "succeeded" or "failed"
			TransactionResults []transactionResult `json:"transactionResults"`
		} `json:"value"`
	} `json:"result"`
}

type transactionResult struct {
	Err                   interface{} `json:"err"`
	Logs                  []string    `json:"logs"`
	UnitsConsumed         uint64      `json:"unitsConsumed"`
	ReturnData            *returnData `json:"returnData"`
	PreExecutionAccounts  []account   `json:"preExecutionAccounts"`
	PostExecutionAccounts []account   `json:"postExecutionAccounts"`
}

type returnData struct {
	ProgramID string `json:"programId"`
	Data      string `json:"data"` // base64
}

type account struct {
	Executable bool     `json:"executable"`
	Owner      string   `json:"owner"`
	Lamports   uint64   `json:"lamports"`
	Data       []string `json:"data"` // [0] - base64 string, [1] - encoding
	Space      uint64   `json:"space"`
	RentEpoch  uint64   `json:"rentEpoch"`
}

type SimulationError struct {
	Failed struct {
		Error struct {
			TransactionFailure []interface{} `json:"TransactionFailure"`
		} `json:"error"`
		TxSignature string `json:"tx_signature"`
	} `json:"failed"`
}

func (e SimulationError) Error() string {
	transactionFailure := e.Failed.Error.TransactionFailure
	if len(transactionFailure) < 2 {
		return "unknown error code"
	}

	return transactionFailure[1].(string)
}

func (e SimulationError) ParseErrorCode() (int, error) {
	re := regexp.MustCompile(`0x[0-9a-fA-F]+`)
	match := re.FindString(e.Error())
	if match == "" {
		return 0, fmt.Errorf("no hex error code found")
	}

	code, err := strconv.ParseInt(strings.TrimPrefix(match, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}
	return int(code), nil
}
