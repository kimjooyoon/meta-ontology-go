package operationconformance_test

import (
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

func TestSplitGoV1ContractHasFixedDenominator(t *testing.T) {
	contract := operationconformance.SplitGoV1Contract()
	want := []string{
		"filesystem.atomic-replacement/v1",
		"go.filename.build-semantics/v1",
		"go.header.preserved/v1",
		"go.import.identity/v1",
		"go.initialization.order/v1",
		"go.package.conformance/v1",
	}
	if contract.Version != 1 || contract.Denominator != 6 || len(contract.Indicators) != 6 {
		t.Fatalf("version=%d denominator=%d indicators=%d", contract.Version, contract.Denominator, len(contract.Indicators))
	}
	if contract.UnknownPolicy != operationconformance.PreserveUnknownAndBlock {
		t.Fatalf("unknown policy = %q", contract.UnknownPolicy)
	}
	if got := operationconformance.SplitGoV1IndicatorIDs(); !slices.Equal(got, want) {
		t.Fatalf("indicator IDs = %v, want %v", got, want)
	}
	roles := map[operationconformance.IndicatorRole]int{}
	routes := map[operationconformance.ProofRoute]int{}
	for _, indicator := range contract.Indicators {
		roles[indicator.Role]++
		routes[indicator.ProofRoute]++
	}
	if roles[operationconformance.RoleOutcome] != 3 || roles[operationconformance.RoleDriver] != 1 || roles[operationconformance.RoleGuardrail] != 2 {
		t.Fatalf("roles = %v", roles)
	}
	if routes[operationconformance.RouteFoundation] != 4 || routes[operationconformance.RouteCoherence] != 1 || routes[operationconformance.RouteRegression] != 1 {
		t.Fatalf("routes = %v", routes)
	}
}

func TestSplitGoV1CorpusCoversEveryVerdict(t *testing.T) {
	corpus := operationconformance.SplitGoV1Corpus()
	if len(corpus) != 18 {
		t.Fatalf("corpus cases = %d, want 18", len(corpus))
	}
	known := map[string]bool{}
	for _, identifier := range operationconformance.SplitGoV1IndicatorIDs() {
		known[identifier] = true
	}
	seenCases := map[string]bool{}
	counts := map[string]map[operationconformance.ExpectedVerdict]int{}
	for _, item := range corpus {
		if seenCases[item.ID] || !known[item.IndicatorID] || len(item.Facts) == 0 {
			t.Fatalf("invalid corpus case: %+v", item)
		}
		seenCases[item.ID] = true
		if counts[item.IndicatorID] == nil {
			counts[item.IndicatorID] = map[operationconformance.ExpectedVerdict]int{}
		}
		counts[item.IndicatorID][item.Expected]++
	}
	for identifier := range known {
		got := counts[identifier]
		if got[operationconformance.VerdictPass] != 1 || got[operationconformance.VerdictFail] != 1 || got[operationconformance.VerdictUnknown] != 1 {
			t.Fatalf("%s verdict coverage = %v", identifier, got)
		}
	}
}
