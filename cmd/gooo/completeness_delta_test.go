package main

import (
	"bytes"
	"testing"
)

func TestCompletenessDeltaArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"--before"}, {"--before", "x"}, {"--before", "x", "--before", "y"}, {"--unknown", "x"}} {
		var out, err bytes.Buffer
		if code := runCompletenessDelta(args, &out, &err); code != exitUsage || out.Len() != 0 {
			t.Fatal("invalid arguments accepted", code)
		}
	}
	var out, err bytes.Buffer
	if code := runCompletenessDelta([]string{"--before", t.TempDir(), "--after", t.TempDir()}, &out, &err); code != exitFailure || out.Len() != 0 {
		t.Fatal("directory accepted")
	}
}
