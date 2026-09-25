package analyzer

import (
	"strings"
	"testing"
)

func mutateGeneratedBillingSlotBody(t *testing.T, source SourceFile) SourceFile {
	t.Helper()
	const body = "\treturn Payment{}\n"
	mutated := strings.Replace(string(source.Source), body, "\treturn Payment{} // slot body mutation\n", 1)
	if mutated == string(source.Source) {
		t.Fatal("generated billing slot body was not found")
	}
	result := source
	result.Source = []byte(mutated)
	return result
}
