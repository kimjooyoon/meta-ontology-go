package claimledger

import "encoding/json"

func inScope(id, path, operator, route, stage, step string, expected json.RawMessage) ClaimSpec {
	return ClaimSpec{
		ID: id, Kind: "OBLIGATION", Modality: "MUST", Subject: "subject", Predicate: "predicate",
		Scope: "IN_SCOPE", ProofRoute: route, Coordinate: Coordinate{Stage: stage, Step: step},
		Evidence:      &EvidenceSpec{Source: "observation", Paths: []string{path}, Operator: operator, Expected: expected},
		UnknownReason: id + "_MISSING", RefutedReason: id + "_INVALID",
	}
}

func excluded(id, route string) ClaimSpec {
	return ClaimSpec{
		ID: id, Kind: "CANDIDATE", Modality: "MAY", Subject: "subject", Predicate: "predicate",
		Scope: "EXCLUDED", ProofRoute: route, Coordinate: Coordinate{Stage: "OBSERVE", Step: "exclude"},
		ExcludedReason: id + "_EXCLUDED",
	}
}

func testContract(claims []ClaimSpec, expected ExpectedMetrics) []byte {
	encoded, err := json.Marshal(Contract{
		Schema: ContractSchema, Metric: "gooo.metric.test.claim-ledger.v1", Expected: expected, Claims: claims,
	})
	if err != nil {
		panic(err)
	}
	return encoded
}

func rawString(value string) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}
