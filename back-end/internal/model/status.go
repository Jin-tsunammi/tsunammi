package model

import (
	"database/sql/driver"
	"fmt"
)

type Status string

const (
	AccountActive        Status = "active"
	AccountInactive      Status = "inactive"
	AccountPending       Status = "pending"
	AccountDeleted       Status = "deleted"
	DefaultAccountStatus        = AccountPending

	Success         Status = "success"
	CreationPending Status = "creation_pending"
	ImportPending   Status = "import_pending"
)

var InitializedAccountStatuses = []Status{AccountActive, AccountInactive}

func (s Status) String() string {
	return string(s)
}

func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *Status) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into Status", value)
	}
	*s = Status(b)
	return nil
}
