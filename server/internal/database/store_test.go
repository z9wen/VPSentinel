package database

import (
	"math"
	"testing"
)

func TestPostgresBigInt(t *testing.T) {
	valid := uint64(math.MaxInt64)
	converted, err := postgresBigInt(&valid)
	if err != nil {
		t.Fatalf("postgresBigInt: %v", err)
	}
	if converted != int64(math.MaxInt64) {
		t.Fatalf("converted = %v, want %d", converted, int64(math.MaxInt64))
	}

	overflow := uint64(math.MaxInt64) + 1
	if _, err := postgresBigInt(&overflow); err == nil {
		t.Fatal("postgresBigInt returned nil error for overflow")
	}
}
