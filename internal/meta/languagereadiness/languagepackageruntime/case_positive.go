package languagepackageruntime

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func executePositive(definition Definition, baseline packageruntime.Result) CaseResult {
	observed, err := packageruntime.Run(manifestFor(definition))
	result := CaseResult{ID: definition.ID, Kind: definition.Kind, Expected: "RUNTIME_IMAGE"}
	if err != nil {
		result.Observed, result.Reason = failureCode(err), err.Error()
		return result
	}
	replay, replayErr := packageruntime.Run(manifestFor(definition))
	result.Observed, result.RuntimeDigest = "RUNTIME_IMAGE", observed.ResultDigest
	measureRuntime(&result, observed)
	result.CanonicalReplays = boolCount(replayErr == nil && replay.ResultDigest == observed.ResultDigest)
	result.OrderInvariantReplays = boolCount(isPermutation(definition.Assertion) &&
		observed.ResultDigest == baseline.ResultDigest)
	result.Satisfied = positiveAssertion(definition.Assertion, observed, baseline) &&
		result.CanonicalReplays == 1
	if result.Satisfied { result.Reason = "RUNTIME_OBSERVATION_EXACT" } else { result.Reason = "RUNTIME_OBSERVATION_MISMATCH" }
	return result
}

func positiveAssertion(assertion string, observed, baseline packageruntime.Result) bool {
	switch assertion {
	case "PACKAGE_GRAPH": return len(observed.Image.Packages) == 4
	case "DIAMOND_ORDER": return strings.Join(observed.Image.InitOrder, ",") ==
		"example/core,example/left,example/right,example/app"
	case "MULTI_SOURCE": return packageSources(observed, "example/app") == 2
	case "ENTRY_CONTRACT": return observed.Image.Entry.Activity == "Run" &&
		len(observed.Image.Entry.Inputs) == 1 && observed.Image.Entry.Output == "Request"
	case "PACKAGE_PERMUTATION", "IMPORT_PERMUTATION", "SOURCE_PERMUTATION":
		return observed.ResultDigest == baseline.ResultDigest
	case "CANONICAL_REPLAY": return observed.ResultDigest != ""
	case "SEMANTIC_BINDING": return semanticBindings(observed) == 5
	case "ZERO_EFFECTS": return observed.Effects == 0 && observed.RepositoryWrites == 0 && !observed.MutationAuthorized
	}
	return false
}
