package generator

import (
	"strings"
	"testing"
)

func assertEntityFieldsDeferred(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), entityFieldsDeferredDiagnostic) {
		t.Fatalf("expected deterministic deferred error, got %v", err)
	}
}
