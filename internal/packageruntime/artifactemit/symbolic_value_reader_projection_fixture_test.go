package artifactemit

import (
	"encoding/json"
	"fmt"
	"strings"
)

func symbolicReaderFixture() []byte {
	indicators := make([]SymbolicValueContractIndicator, 0, 11)
	for index := 1; index <= 11; index++ {
		audiences := []string{"GOVERNOR"}
		if index <= 9 {
			audiences = append(audiences, "TOOL_AUTHOR")
		}
		if index <= 5 {
			audiences = append(audiences, "USER")
		}
		indicators = append(indicators, SymbolicValueContractIndicator{
			ID: fmt.Sprintf("source.indicator-%02d", index), Class: "DRIVER",
			ProofChoice: "FOUNDATION", MetaOperation: "fixture",
			Observed: 1, Expected: 1, Satisfied: true, Audiences: audiences,
		})
	}
	source := SymbolicValueReachability{
		Schema: "gooo/symbolic-invocation-value-reachability/v1",
		SubjectSHA: strings.Repeat("a", 40),
		MetricID: "gooo.metric.compiler.symbolic-value-reachability.v1",
		Decision: "PASS", Resolution: "SCHEMA_VALUE_REACHABILITY_ONLY",
		Source: SymbolicValueReachabilitySource{
			ArtifactDigest: "sha256:" + strings.Repeat("b", 64),
			ContractDigest: "sha256:" + strings.Repeat("c", 64),
		},
		Summary:    SymbolicValueReachabilitySummary{PolicyBranches: 3},
		Indicators: indicators,
		Views: []SymbolicValueContractView{
			{Audience: "USER", Resolution: "USER_VISIBLE", Satisfied: 5, Total: 5, BasisPoints: 10000},
			{Audience: "TOOL_AUTHOR", Resolution: "TOOL_CONTRACT", Satisfied: 9, Total: 9, BasisPoints: 10000},
			{Audience: "GOVERNOR", Resolution: "FULL_RECEIPT", Satisfied: 11, Total: 11, BasisPoints: 10000},
		},
	}
	source.Digest = symbolicReaderReachabilityDigest(source)
	payload, _ := json.Marshal(source)
	return payload
}

func symbolicReaderMutate(
	mutate func(*SymbolicValueReachability),
) []byte {
	var source SymbolicValueReachability
	_ = json.Unmarshal(symbolicReaderFixture(), &source)
	mutate(&source)
	source.Digest = symbolicReaderReachabilityDigest(source)
	payload, _ := json.Marshal(source)
	return payload
}
