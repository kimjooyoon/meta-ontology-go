package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
)

type originOutputFailureFresh1735 struct{ short bool }

func (w originOutputFailureFresh1735) Write(payload []byte) (int, error) {
	if w.short {
		return len(payload) - 1, nil
	}
	return 0, errors.New("caller output failed")
}

func TestCommandOriginOutputFailureFresh1735(t *testing.T) {
	for _, short := range []bool{true, false} {
		name := "error"
		if short {
			name = "short-write"
		}
		t.Run(name, func(t *testing.T) {
			strategy, verification := unresolvedInputsFresh1733(t, "upstream://operation/output")
			_, _, failure := metricprogram.Compile(strategy, verification)
			if failure == nil {
				t.Fatal("fixture did not refuse its unregistered operation")
			}
			var stderr bytes.Buffer
			if code := reportCompileFailure(failure, originOutputFailureFresh1735{short}, &stderr); code != 1 ||
				!strings.Contains(stderr.String(), "write unresolved origin:") {
				t.Fatalf("command lost delivery failure: %d, %s", code, stderr.String())
			}
		})
	}
}
