package languagesourcebindingpromotion

import (
	"encoding/json"
	"strings"
)

func fixtureInput() Input {
	head := strings.Repeat("a", 40)
	receipt := receiptEnvelope{Schema: "gooo/source-execution-receipt/v1", Decision: DecisionPass,
		Reason: "SOURCE_ACTIVITY_EXECUTED", Resolution: ResolutionExact, Filename: "examples/billing/main.gooo",
		SourceDigest: "sha256:" + strings.Repeat("b", 64), SemanticDigest: "sha256:" + strings.Repeat("c", 64),
		Entry: json.RawMessage(`{}`), Events: json.RawMessage(`[]`), Diagnostics: json.RawMessage(`[]`), Effects: json.RawMessage(`{}`)}
	receipt.Digest = digestJSON(receipt)
	producer := producerFixture(head, receipt.Digest)
	oracle := oracleFixture(head, receipt)
	unknownProducer := producer
	unknownProducer.Decision, unknownProducer.Digest = "UNKNOWN", ""
	unknownProducer.Digest = digestJSON(unknownProducer)
	unknownOracle := oracle
	unknownOracle.Decision, unknownOracle.Digest = "UNKNOWN", ""
	unknownOracle.Digest = digestJSON(unknownOracle)
	mismatchedOracle := oracle
	cases, _ := decodeView[[]oracleCase](oracle.Cases)
	cases[0].ArtifactDigest = "sha256:" + strings.Repeat("0", 64)
	mismatchedOracle.Cases, _ = json.Marshal(cases)
	mismatchedOracle.Digest = ""
	mismatchedOracle.Digest = digestJSON(mismatchedOracle)
	return Input{Contract: CanonicalContract(), HeadSHA: head, PolicySource: []byte("policy"),
		PolicyArtifact: []byte("artifact"), PolicyReplayArtifact: []byte("artifact"),
		Producer: jsonBytes(producer), Receipt: jsonBytes(receipt), Oracle: jsonBytes(oracle),
		UnknownProducer: jsonBytes(unknownProducer), UnknownOracle: jsonBytes(unknownOracle),
		MismatchedOracle: jsonBytes(mismatchedOracle),
		Independence:     IndependenceEvidence{Schema: IndependenceSchema}}
}

func jsonBytes(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
