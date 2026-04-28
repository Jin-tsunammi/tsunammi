package mtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"mm/pkg/apperrors"
)

type BigRat struct {
	big.Rat
}

func NewDBBigRat(r *big.Rat) BigRat {
	if r == nil {
		return BigRat{Rat: *new(big.Rat)}
	}
	return BigRat{Rat: *r}
}

func (r *BigRat) GetBigRat() *big.Rat {
	return &r.Rat
}

func (r BigRat) Value() (driver.Value, error) {
	return r.FloatString(22), nil
}

func (r *BigRat) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		r.SetString(string(v))
	case string:
		r.SetString(v)
	case float64:
		r.SetFloat64(v)
	case nil:
		r.Rat = *new(big.Rat)
	default:
		return apperrors.Internal(fmt.Sprintf("failed to scan BigRat: cannot convert %T to BigRat", value))
	}
	return nil
}

func (r *BigRat) String() string {
	return r.FloatString(20)
}

func (r *BigRat) MarshalJSON() ([]byte, error) {
	if r.Rat.Denom() == nil {
		return []byte("null"), nil
	}

	floatStr := r.FloatString(20)

	return json.Marshal(floatStr)
}

func (r *BigRat) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		if _, ok := r.SetString(value); !ok {
			return apperrors.Internal(fmt.Sprintf("failed to parse big.Rat from string: %s", value))
		}
	case float64:
		r.SetFloat64(value)
	case nil:
		r.Rat = *new(big.Rat)
	default:
		return apperrors.Internal(fmt.Sprintf("invalid type for BigRat: %T", v))
	}
	return nil
}

func (r *BigRat) ToAtomicUnits(decimals uint8) *big.Int {
	if r == nil {
		return new(big.Int)
	}

	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	scaled := new(big.Rat).Mul(
		new(big.Rat).Set(&r.Rat),
		new(big.Rat).SetInt(multiplier),
	)

	return new(big.Int).Quo(scaled.Num(), scaled.Denom())
}
