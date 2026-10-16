package main

import (
	"os"
	"path/filepath"
	"testing"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation"
)

func TestConsumerJSONBoundaryFresh1619(t *testing.T) {
	for _, c := range []struct {
		name, raw string
		valid     bool
	}{
		{"one-record", `{"value":1}`, true}, {"two-records", `{"value":1} {"value":2}`, false},
		{"trailing-junk", `{"value":1}broken`, false}, {"empty", ``, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.json")
			if err := os.WriteFile(path, []byte(c.raw), 0600); err != nil {
				t.Fatal(err)
			}
			var value map[string]int
			if err := readJSON(path, &value); (err == nil) != c.valid {
				t.Fatal("JSON boundary", err)
			}
		})
	}
}

func TestConsumerInvocationFresh1619(t *testing.T) {
	p := continuation.Prepared{Metadata: continuation.Metadata{ConsumerRunID: 9161901, ConsumerAttempt: 1}}
	if currentInvocation(p, options{run: 9161901, attempt: 1}) != nil {
		t.Fatal("own invocation rejected")
	}
	for _, o := range []options{{run: 9161902, attempt: 1}, {run: 9161901, attempt: 2}} {
		if currentInvocation(p, o) == nil {
			t.Fatal("foreign invocation accepted")
		}
	}
}
