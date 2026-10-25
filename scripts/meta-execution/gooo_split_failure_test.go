package main

import (
	"errors"
	"runtime"
	"testing"
)

func TestGoooSplitFailureKeepsUnavailableAndRefutedSeparate(t *testing.T) {
	for _, mode := range []string{"positive-exit", "signal", "missing"} {
		if mode == "signal" && runtime.GOOS == "windows" {
			continue
		}
		process, err := runVerifierProcessFixture(t, mode, mode == "missing")
		failure := goooSplitProcessFailure("execute-cases", process, err)
		want := "DIRECT_MISSING"
		if mode == "positive-exit" {
			want = "KNOWN_CONTRADICTION"
		}
		if failure == nil || failure.class != want {
			t.Fatal("process result lost its observed classification", mode, failure)
		}
	}
	failure := goooSplitUnavailable("read-cases", "WORKSPACE_CASES_UNAVAILABLE", errors.New("unreadable"))
	if failure.class != "DIRECT_MISSING" || len(failure.diagnostics) != 1 {
		t.Fatal("missing input was turned into a semantic contradiction", failure)
	}
}
