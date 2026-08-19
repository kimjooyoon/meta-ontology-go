package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
	"testing"
)

func TestEntityFieldsCLIStateSwitchIsExhaustive(t *testing.T) {
	current := syntax.CurrentEntityFieldsSupport()
	for _, state := range []syntax.EntityFieldsState{syntax.EntityFieldsDeferred, syntax.EntityFieldsSupported} {
		current.State = state
		if err := validateCLIEntityFieldsSupport(current); err != nil {
			t.Fatalf("known state %q rejected: %v", state, err)
		}
	}
	current.State = "UNKNOWN"
	if err := validateCLIEntityFieldsSupport(current); err == nil || !strings.Contains(err.Error(), "GOOO-EF-V1-UNKNOWN-STATE") {
		t.Fatalf("unknown state was not rejected: %v", err)
	}
}
