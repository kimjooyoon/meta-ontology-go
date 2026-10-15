package bodycodegen

import (
	"go/types"
	"strconv"
)

// The interpreter models Go's truncation after typed integer operations.
// Keep the result in int64 storage without narrowing through a host integer.
func coerceBodyInteger(number int64, kind types.BasicKind) int64 {
	if kind == types.Uint8 {
		return number & 255
	}
	if kind == types.Int32 || (kind == types.Int && strconv.IntSize == 32) {
		low := number & ((1 << 32) - 1)
		if low >= 1<<31 {
			return low - 1<<32
		}
		return low
	}
	return number
}
