package workfrontier

import (
	"fmt"
)

type r4DecodeFailureKind uint8

const (
	r4DecodeMissingRequired r4DecodeFailureKind = iota + 1
	r4DecodeMalformed
)

type r4DecodeError struct {
	kind r4DecodeFailureKind
	err  error
}

func (e *r4DecodeError) Error() string {
	return fmt.Sprintf("decode r4 work frontier: %v", e.err)
}
func (e *r4DecodeError) Unwrap() error { return e.err }

// DecodeR4JSON accepts exactly one R4 envelope. Unknown fields, duplicate
// fields, missing required fields, and legacy schema versions are rejected.
func DecodeR4JSON(data []byte) (R4Input, error) {
	return decodeR4JSON(data)
}
