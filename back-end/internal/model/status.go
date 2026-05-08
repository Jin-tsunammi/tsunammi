package model

import (
	"database/sql/driver"
	"fmt"
)

var InitializedAccountStatuses = []AccountStatus{AccountStatusActive, AccountStatusInactive}

type AccountStatus string
type WalletStatus string
type SwapStatus string
type BuybackStatus string

const (
	AccountStatusActive   AccountStatus = "ACTIVE"
	AccountStatusPending  AccountStatus = "PENDING"
	AccountStatusInactive AccountStatus = "INACTIVE"
	AccountStatusDeleted  AccountStatus = "DELETED"
)
const (
	WalletStatusImportPending   WalletStatus = "IMPORT_PENDING"
	WalletStatusCreationPending WalletStatus = "CREATION_PENDING"
	WalletStatusSuccess         WalletStatus = "SUCCESS"
)

const (
	SwapStatusActive            SwapStatus = "ACTIVE"
	SwapStatusDone              SwapStatus = "DONE"
	SwapStatusTargetCompleted   SwapStatus = "TARGET_COMPLETED"
	SwapStatusBudgetDone        SwapStatus = "BUDGET_DONE"
	SwapStatusInsufficientFunds SwapStatus = "INSUFFICIENT_FUNDS"
	SwapStatusStop              SwapStatus = "STOP"
	SwapStatusError             SwapStatus = "ERROR"
)

const (
	BuybackStatusActive            BuybackStatus = "ACTIVE"
	BuybackStatusScheduled         BuybackStatus = "SCHEDULED"
	BuybackStatusDone              BuybackStatus = "DONE"
	BuybackStatusError             BuybackStatus = "ERROR"
	BuybackStatusBudgetDone        BuybackStatus = "BUDGET_DONE"
	BuybackStatusInsufficientFunds BuybackStatus = "INSUFFICIENT_FUNDS"
	BuybackStatusStop              BuybackStatus = "STOP"
)

func scanStatus[T ~string](dst *T, value any) error {
	if value == nil {
		*dst = ""
		return nil
	}
	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into status", value)
	}
	*dst = T(b)
	return nil
}

func (s *AccountStatus) Scan(v any) error            { return scanStatus(s, v) }
func (s AccountStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *SwapStatus) Scan(v any) error            { return scanStatus(s, v) }
func (s SwapStatus) Value() (driver.Value, error) { return string(s), nil }

func (s *BuybackStatus) Scan(v any) error            { return scanStatus(s, v) }
func (s BuybackStatus) Value() (driver.Value, error) { return string(s), nil }
