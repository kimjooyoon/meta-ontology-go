package artifactemit

import (
	"encoding/json"
	"testing"
)

func symbolicReaderRequestSourceFixture(audience, resolution string) []byte {
	return []byte("package checkout\nnamespace checkout\n\n" +
		"entity ReaderProjection id \"urn:gooo:checkout:reader-projection\"\n" +
		"entity ReaderView id \"urn:gooo:checkout:reader-view\"\n" +
		"activity ReadProjection(ReaderProjection) -> ReaderView computes \"reader.project:" +
		audience + ":" + resolution + "\"\n")
}

func symbolicReaderRequestProjectionFixture() SymbolicValueReaderProjection {
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	value := SymbolicValueReaderProjection{
		Schema: symbolicReaderProjectionSchema, SubjectSHA: "fixture-sha",
		MetricID: symbolicReaderProjectionMetric, Decision: "PASS",
		Resolution: "READER_PROJECTION_ONLY", Reason: "CANONICAL_READER_PROJECTIONS_BOUND",
		Source: SymbolicValueReaderProjectionSource{
			Schema: "gooo/symbolic-invocation-value-reachability/v1",
			MetricID: "gooo.metric.compiler.symbolic-value-reachability.v1",
			Decision: "PASS", Resolution: "SCHEMA_VALUE_REACHABILITY_ONLY",
			ReachabilityDigest: digest, FileDigest: digest,
		},
		Readers: []SymbolicValueReaderProjectionView{
			symbolicReaderRequestReaderFixture("USER", "USER_VISIBLE", "DECISION_AND_COUNTS_ONLY", 5),
			symbolicReaderRequestReaderFixture("TOOL_AUTHOR", "TOOL_CONTRACT", "INDICATOR_CONTRACT_ONLY", 9),
			symbolicReaderRequestReaderFixture("GOVERNOR", "FULL_RECEIPT", "SOURCE_BOUND_RECEIPT_ONLY", 11),
		},
		Coordinates: SymbolicValueContractCoordinates{Satisfied: 18, Total: 18, BasisPoints: 10000},
		Effects: SymbolicValueContractEffects{},
	}
	value.Digest = symbolicReaderProjectionDigest(value)
	return value
}

func symbolicReaderRequestReaderFixture(
	audience, sourceResolution, effectiveResolution string,
	total int,
) SymbolicValueReaderProjectionView {
	ids := make([]string, total)
	for index := range ids {
		ids[index] = audience + ".indicator." + string(rune('a'+index))
	}
	return SymbolicValueReaderProjectionView{
		Audience: audience, SourceResolution: sourceResolution,
		EffectiveResolution: effectiveResolution, IndicatorIDs: ids,
		Coordinates: SymbolicValueContractCoordinates{Satisfied: total, Total: total, BasisPoints: 10000},
	}
}

func encodeSymbolicReaderRequestProjection(t *testing.T, value SymbolicValueReaderProjection) []byte {
	t.Helper()
	value.Digest = symbolicReaderProjectionDigest(value)
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
