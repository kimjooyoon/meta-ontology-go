package bodycodegen

import (
	"bytes"
	"context"
	"testing"
)

func TestTypedTreeNormalizesSDKIntegerConstantsExactly(t *testing.T) {
	prefix := "package p\nfunc Assemble(input int64) int64 { return "
	for _, literal := range []string{"0", "9007199254740993", "9007199254740995", "-9007199254740995", "-9223372036854775808", "9223372036854775807"} {
		left, err := typedBodyTree(context.Background(), "Assemble", []byte(prefix+literal+" }"))
		if err != nil {
			t.Fatal(err)
		}
		right, err := typedBodyTree(context.Background(), "Assemble", []byte(prefix+"int64("+literal+") }"))
		if err != nil || !bytes.Equal(left, right) {
			t.Fatal("literal changed", literal, err)
		}
	}
	for _, call := range []string{"int64(input)", "other(0)", "int64(0, 1)", "int64(1.5)", "int64(9223372036854775808)", "int64(0...)"} {
		if _, err := typedBodyTree(context.Background(), "Assemble", []byte(prefix+call+" }")); err == nil {
			t.Fatal("invalid typed literal accepted", call)
		}
	}
}
