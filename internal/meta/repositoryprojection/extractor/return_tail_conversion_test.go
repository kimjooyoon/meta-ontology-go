package extractor

import (
	"errors"
	"strings"
	"testing"
)

func TestReturnTailConversionClassificationCohort(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{name: "C1 named-function-field-call-is-not-conversion", src: returnTailNamedFunctionFieldFixture()},
		{name: "C2 named-function-index-call-is-not-conversion", src: returnTailNamedFunctionIndexFixture()},
	}
	if len(cases) != 2 {
		t.Fatalf("conversion cohort denominator=%d, want 2", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := extractReturnTailRegressionFixture(t, tc.src)
			var failure Failure
			if !errors.As(err, &failure) || failure.UnknownClass != "DIRECT_MISSING" || failure.Reason != "CALLEE_EFFECTS_UNPROVEN" {
				t.Fatalf("conversion classification failure=%v", err)
			}
		})
	}
}

func returnTailNamedFunctionFieldFixture() string {
	return "package p\n\ntype namedCallback func() error\ntype callbackBox struct{ Callback namedCallback }\n\nfunc F() error {\n\tbox := callbackBox{}\n" + repeatReturnTailPadding() + "\treturn box.Callback()\n}\n"
}

func returnTailNamedFunctionIndexFixture() string {
	return "package p\n\ntype namedCallback func() error\n\nfunc F() error {\n\tcallbacks := []namedCallback{nil}\n" + repeatReturnTailPadding() + "\treturn callbacks[0]()\n}\n"
}

func repeatReturnTailPadding() string {
	return strings.Repeat("\t_ = 1\n", 72)
}
