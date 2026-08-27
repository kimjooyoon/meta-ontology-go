package languageresourcebudget

import "reflect"

const (
	Producer = "scripts/meta-resource-budget"
	Consumer = "cmd/meta-resource-budget-reducer"
)

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, ID: "meta-resource-budget-v1",
		SourcePaths: []string{"examples/meta-resource-budget/activity.gooo", "examples/meta-resource-budget/entities.gooo"},
		Entry:       "PayOrder", SamplesPerOp: 3, Indicators: ExpectedIndicators,
		Operations: []Operation{
			{ID: "source-check", Stage: "LOWER", Step: "parse-source", MetaOperation: "observe-source-receipt", ProofChoice: "FOUNDATION", Output: "RECEIPT"},
			{ID: "project-manifest", Stage: "PROJECT", Step: "project-operation-manifest", MetaOperation: "project-generated-operation", ProofChoice: "COHERENCE", Output: "GENERATED"},
			{ID: "replay-manifest", Stage: "REPLAY", Step: "replay-generated-artifact", MetaOperation: "prove-generated-replay", ProofChoice: "REGRESSION", Output: "GENERATED"},
		},
		Limits: Limits{WallTimeMS: 2000, PeakRSSKiB: 131072, ReceiptBytes: 8192, GeneratedBytes: 16384},
		NotClaimed: []string{
			"cross-run performance improvement", "machine-independent resource bounds",
			"business correctness", "production readiness", "hermetic build guarantee",
		},
		References: []Reference{
			{ID: "github-hosted-runners", URL: "https://docs.github.com/en/actions/reference/runners/github-hosted-runners"},
			{ID: "bazel-hermeticity", URL: "https://bazel.build/basics/hermeticity"},
		},
	}
}

func validContract(value Contract) bool { return reflect.DeepEqual(value, CanonicalContract()) }

func operation(value string, contract Contract) (Operation, bool) {
	for _, item := range contract.Operations {
		if item.ID == value {
			return item, true
		}
	}
	return Operation{}, false
}

func positiveSHA(value string) bool { return len(value) == 40 && isHex(value) }

func isHex(value string) bool {
	for _, char := range value {
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') && !(char >= 'A' && char <= 'F') {
			return false
		}
	}
	return true
}
