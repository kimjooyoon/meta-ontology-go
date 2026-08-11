package fuzz

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const validSeed = `package billing
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment
`

const malformedSeed = `package billing
namespace billing
entity Payment id "billing://entity/payment
activity PayOrder(Order) Payment
`

func addMinimalSeeds(f *testing.F) {
	f.Helper()
	f.Add(validSeed)
	f.Add(malformedSeed)
}

func TestMinimalSeeds(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		wantDiagnostic bool
	}{
		{name: "valid", source: validSeed},
		{name: "malformed", source: malformedSeed, wantDiagnostic: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := runWithLimit(t, "seed parse", func() parseResult {
				file, diagnostics := syntax.ParseFile(fuzzFilename, test.source)
				return parseResult{File: file, Diagnostics: diagnostics}
			})
			if (len(result.Diagnostics) > 0) != test.wantDiagnostic {
				t.Fatalf("diagnostic presence = %t, want %t: %v", len(result.Diagnostics) > 0, test.wantDiagnostic, result.Diagnostics)
			}
			assertDiagnostics(t, test.source, result.Diagnostics)
			assertFile(t, test.source, result.File)
		})
	}
}
