package common

import (
	"math/big"
	"testing"
)

func TestSelectTxAmountInRange_ReturnsRemainingWhenItFitsRange(t *testing.T) {
	t.Parallel()

	got, err := SelectTxAmountInRange(big.NewInt(75), big.NewInt(50), big.NewInt(80))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cmp(big.NewInt(75)) != 0 {
		t.Fatalf("unexpected amount: got %s, want %s", got.String(), "75")
	}
}

func TestSelectTxAmountInRange_ReturnsRemainingWhenItEqualsMin(t *testing.T) {
	t.Parallel()

	got, err := SelectTxAmountInRange(big.NewInt(50), big.NewInt(50), big.NewInt(80))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cmp(big.NewInt(50)) != 0 {
		t.Fatalf("unexpected amount: got %s, want %s", got.String(), "50")
	}
}

func TestSelectTxAmountInRange_ReturnsRemainingWhenItEqualsMax(t *testing.T) {
	t.Parallel()

	got, err := SelectTxAmountInRange(big.NewInt(80), big.NewInt(50), big.NewInt(80))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cmp(big.NewInt(80)) != 0 {
		t.Fatalf("unexpected amount: got %s, want %s", got.String(), "80")
	}
}

func TestSelectTxAmountInRange_ReturnsRandomAmountWhenRemainingIsAboveMax(t *testing.T) {
	t.Parallel()

	min := big.NewInt(50)
	max := big.NewInt(80)

	got, err := SelectTxAmountInRange(big.NewInt(100), min, max)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cmp(min) < 0 || got.Cmp(max) > 0 {
		t.Fatalf("random amount out of range: got %s, min %s, max %s", got.String(), min.String(), max.String())
	}
}

func TestSelectTxAmountInRange_ReturnsErrorWhenRemainingIsNil(t *testing.T) {
	t.Parallel()

	_, err := SelectTxAmountInRange(nil, big.NewInt(50), big.NewInt(80))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSelectTxAmountInRange_ReturnsErrorWhenRemainingIsZero(t *testing.T) {
	t.Parallel()

	_, err := SelectTxAmountInRange(big.NewInt(0), big.NewInt(50), big.NewInt(80))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSelectTxAmountInRange_ReturnsErrorWhenMinIsGreaterThanMax(t *testing.T) {
	t.Parallel()

	_, err := SelectTxAmountInRange(big.NewInt(75), big.NewInt(81), big.NewInt(80))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRandBigIntRange_ReturnsValueWithinInclusiveRange(t *testing.T) {
	t.Parallel()

	min := big.NewInt(5)
	max := big.NewInt(8)

	got, err := RandBigIntRange(min, max)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cmp(min) < 0 || got.Cmp(max) > 0 {
		t.Fatalf("value out of range: got %s, min %s, max %s", got.String(), min.String(), max.String())
	}
}

func TestRandBigIntRange_ReturnsValueWhenMinEqualsMax(t *testing.T) {
	t.Parallel()

	got, err := RandBigIntRange(big.NewInt(7), big.NewInt(7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Cmp(big.NewInt(7)) != 0 {
		t.Fatalf("unexpected value: got %s, want %s", got.String(), "7")
	}
}

func TestRandBigIntRange_ReturnsErrorWhenMaxIsLessThanMin(t *testing.T) {
	t.Parallel()

	_, err := RandBigIntRange(big.NewInt(8), big.NewInt(7))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
