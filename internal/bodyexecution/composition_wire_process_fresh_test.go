package bodyexecution

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCompositionWireProcessHelperFresh1877(t *testing.T) {
	switch os.Args[len(os.Args)-1] {
	case "wire-out":
		fmt.Print(strings.Repeat("x", 128<<10))
	case "wire-over":
		fmt.Print(strings.Repeat("x", (1<<20)+1))
	case "wire-diagnostic":
		fmt.Fprint(os.Stderr, strings.Repeat("x", 128<<10))
	default:
		return
	}
	os.Exit(0)
}

func TestCompositionWireProcessPolicyFresh1878(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"compiler_stdout", "composition_stdout", "composition_overflow", "composition_stderr"} {
		t.Run(mode, func(t *testing.T) {
			limit, arg := compositionOutputMaxBytes, "wire-out"
			switch mode {
			case "compiler_stdout":
				limit = 64 << 10
			case "composition_overflow":
				arg = "wire-over"
			case "composition_stderr":
				arg = "wire-diagnostic"
			}
			raw, r, err := processWithOutputLimit(context.Background(), 2*time.Second, limit, t.TempDir(), binary, nil,
				"-test.run=^TestCompositionWireProcessHelperFresh1877$", "--", arg)
			if mode == "composition_stdout" {
				if err != nil || !r.Completed || len(raw) != 128<<10 || r.OutputTruncated {
					t.Fatal(r, err)
				}
			} else if err == nil || r.Completed || !r.OutputTruncated || raw != nil {
				t.Fatal("overflow was returned as a successful value", r, err)
			}
		})
	}
}
