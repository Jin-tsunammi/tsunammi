package kucoinapi

import "time"

const (
	OK        = 200000
	FORBIDDEN = 400003
)

type WithdrawalStatus string

const (
	WithdrawalStatusReviewing        WithdrawalStatus = "REVIEW"
	WithdrawalStatusProcessing       WithdrawalStatus = "PROCESSING"
	WithdrawalStatusWalletProcessing WithdrawalStatus = "WALLET_PROCESSING"
	WithdrawalStatusSuccess          WithdrawalStatus = "SUCCESS"
	WithdrawalStatusFailure          WithdrawalStatus = "FAILURE"
)

type WithdrawalFailure string

const (
	WithdrawalFailureSystemException WithdrawalFailure = "SYSTEM_EXCEPTION"
	WithdrawalFailureNoBalance       WithdrawalFailure = "NO_BALANCE"
	WithdrawalFailureCancelByUser    WithdrawalFailure = "CANCEL"
	WithdrawalFailureRejected        WithdrawalFailure = "REJECTED"
	WithdrawalFailureWalletFailure   WithdrawalFailure = "WALLET_FAILURE"
)

type Response[T any] struct {
	Code               int           `json:"code,string"`
	Msg                string        `json:"msg"`
	Data               T             `json:"data"`
	RateLimitRemaining int           `json:"rate_limit_remaining"`
	RateLimitReset     time.Duration `json:"rate_limit_reset"`
}

type KeyInfo struct {
	UID     uint64 `json:"uid"`
	Perm    string `json:"permission"`
	ApiName string `json:"remark"`
}

type Currency struct {
	Fee float64 `json:"fee,string"`
}

type Balance struct {
	Available float64 `json:"available,string"`
}

type WithdrawQuotas struct {
	MinWithdraw         float64 `json:"withdrawMinSize,string"`
	MinFee              float64 `json:"withdrawMinFee,string"`
	Enabled             bool    `json:"isWithdrawEnabled"`
	Chain               string  `json:"chain"`
	RemainingDailyLimit float64 `json:"remainAmount,string"`
}

type WithdrawReq struct {
	Currency     string  `json:"currency"`
	Address      string  `json:"toAddress"`
	Amount       float64 `json:"amount,string"`
	WithdrawType string  `json:"withdrawType"`
	Chain        string  `json:"chain"`
	Remark       string  `json:"remark"`
	FeeDeduct    string  `json:"feeDeductType"`
}

type WithdrawResp struct {
	WithdrawalID string `json:"withdrawalId"`
}

type WithdrawalStatusResp struct {
	ID               string            `json:"id"`
	Status           WithdrawalStatus  `json:"status"`
	Remark           string            `json:"remark"`
	CreatedAt        int64             `json:"createdAt"`
	FailureReason    WithdrawalFailure `json:"failureReason"`
	FailureReasonMsg string            `json:"failureReasonMsg"`
	WalletTxId       string            `json:"walletTxId"`
}
