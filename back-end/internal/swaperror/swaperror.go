package swaperror

import "errors"

var ErrSimulationError = errors.New("simulation error")
var ErrSlippageExceeded = errors.New("slippage exceeded")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrCustomProgramError = errors.New("custom program error")
var ErrRateLimit = errors.New("rate limit")
var ErrGatewayTimeout = errors.New("gateway timeout")
var ErrComputeBudgetExceeded = errors.New("compute budget exceeded")
var ErrBundleRejected = errors.New("bundle rejected")
