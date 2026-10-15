package bodycodegen

import (
	"go/types"
	"math"
	"strconv"
	"testing"
)

func TestIntegerWidthKeepsTypedWrapSemantics(t *testing.T) {
	cases := []struct{ input, byteValue, runeValue int64 }{
		{0, 0, 0}, {255, 255, 255}, {256, 0, 256}, {-1, 255, -1}, {-256, 0, -256},
		{math.MaxInt32, 255, math.MaxInt32}, {1 << 31, 0, math.MinInt32},
		{math.MinInt32 - 1, 255, math.MaxInt32}, {1 << 32, 0, 0},
		{math.MaxInt64, 255, -1}, {math.MinInt64, 0, 0},
	}
	for _, tc := range cases {
		native := tc.input
		if strconv.IntSize == 32 {
			native = tc.runeValue
		}
		for kind, want := range map[types.BasicKind]int64{
			types.Uint8: tc.byteValue, types.Int32: tc.runeValue, types.Int: native,
			types.Int64: tc.input, types.UntypedInt: tc.input, types.UntypedRune: tc.input,
		} {
			got, err := coerceBodyValue(tc.input, types.Typ[kind])
			if err != nil || got != want {
				t.Fatalf("input %d kind %d = %v, %v; want %d", tc.input, kind, got, err, want)
			}
		}
	}
}
