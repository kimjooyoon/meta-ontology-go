// Package textbound represents a source-declared decoded UTF-8 byte limit.
package textbound

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalid = errors.New("invalid Text UTF-8 byte bound")

// UTF8 is an optional uint32 limit stored by value in eight bytes. Zero means
// undeclared; limit+1 represents a declaration, including an explicit zero.
// Serialized declarations contain the actual limit, not this storage encoding.
type UTF8 uint64

func New(maximum uint32) UTF8 { return UTF8(maximum) + 1 }

func (b UTF8) IsSet() bool { return b != 0 }

func (b UTF8) Validate() error {
	if uint64(b) > uint64(1)<<32 {
		return ErrInvalid
	}
	return nil
}

// Resolve preserves the caller's existing limit when the source has no bound.
// Invalid carriers fail rather than wrapping to a smaller value or defaulting.
func (b UTF8) Resolve(fallback uint32) (uint32, error) {
	if err := b.Validate(); err != nil {
		return 0, err
	}
	if b == 0 {
		return fallback, nil
	}
	return uint32(b - 1), nil
}

// Parse accepts canonical unsigned decimal source text. Transport and aggregate
// budgets are independent of this value-domain declaration.
func Parse(value string) (UTF8, error) {
	if len(value) == 0 || len(value) > 10 || len(value) > 1 && value[0] == '0' {
		return 0, ErrInvalid
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, ErrInvalid
		}
	}
	maximum, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, ErrInvalid
	}
	return New(uint32(maximum)), nil
}

// Canonical is empty for an undeclared bound, preserving legacy fingerprints.
func (b UTF8) Canonical() string {
	maximum, err := b.Resolve(0)
	if err != nil {
		return fmt.Sprintf("invalid-max_utf8_bytes=%d", b)
	}
	if !b.IsSet() {
		return ""
	}
	return "max_utf8_bytes=" + strconv.FormatUint(uint64(maximum), 10)
}
