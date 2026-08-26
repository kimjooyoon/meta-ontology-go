package symbolicinvocationusecase

import "testing"

func TestEvaluateClosesFixedUserUseCase(t *testing.T) {
	report, err := Evaluate(testInput())
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PASS" || report.Resolution != "EXACT" || report.Summary.Coordinates.Satisfied != 6 || report.Summary.Coordinates.Total != 6 || report.Summary.UserDecisions != 2 || report.PromotionCreditBPS != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestEvaluateUnknownProducerDecisionLowersResolution(t *testing.T) {
	input := testInput()
	input.ProducerReceipt.Decision = "UNKNOWN"
	report, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" || report.Reason != reasonDecisionUnknown || report.Summary.Coordinates.Satisfied != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestEvaluateArtifactLinkMismatchPreservesInvariant(t *testing.T) {
	input := testInput()
	input.ProducerArtifact.Digest = digest('9')
	report, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "INVARIANT_ONLY" || report.Reason != reasonLinkMismatch || report.Summary.Coordinates.Satisfied != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func testInput() Input {
	artifactDigest, schemaDigest, toolDigest := digest('a'), digest('b'), digest('c')
	return Input{SubjectSHA: string(repeat('1', 40)), Contract: CanonicalContract(), ProducerReceipt: ProducerReceipt{Schema: "gooo/symbolic-invocation-schema-receipt/v1", Decision: "PASS", Resolution: "EXACT", Reason: "EXTERNAL_SCHEMA_VALIDATION_OBSERVED", SubjectSHA: string(repeat('1', 40)), Compiler: CompilerEvidence{GoVersion: "1.27.0", BinaryDigest: digest('d'), BinaryBytes: 12492193, RegisteredEmitters: 3}, Source: SourceCoordinate{GoooFiles: 2, GoFiles: 0, GoooLines: 10, Files: 5, Directories: 0}, Artifact: ArtifactEvidence{Kind: "symbolic-invocation-schema", ArtifactSchema: "gooo/symbolic-invocation-schema-artifact/v1", Digest: artifactDigest, JSONSchemaDialect: "https://json-schema.org/draft/2020-12/schema", JSONSchemaDigest: schemaDigest}, Validation: ValidationEvidence{Tool: "github.com/santhosh-tekuri/jsonschema/cmd/jv@v0.7.0", ToolDigest: toolDigest, AcceptedInstances: 1, RejectedInstances: 1}, DeterministicReplays: 1, Resources: ResourceEvidence{Samples: []ResourceSample{{1, 5, 10000}, {2, 6, 10001}, {3, 5, 9999}, {4, 4, 9998}, {5, 5, 9997}}, SampleCount: 5, MaxWallMS: 6, MaxRSSKiB: 10001}, Effects: Effects{}, NotClaimed: CanonicalNonClaims()}, ProducerArtifact: ProducerArtifact{Schema: "gooo/symbolic-invocation-schema-artifact/v1", Decision: "PASS", Resolution: "SYMBOLIC_ONLY", Reason: "SYMBOLIC_INVOCATION_SCHEMA_EMITTED", Kind: "symbolic-invocation-schema", Extensions: ArtifactExtensions{RegisteredEmitters: 3, Kinds: []string{"operation-interface", "operation-manifest", "symbolic-invocation-schema"}}, Effects: Effects{}, Digest: artifactDigest}, Observation: Observation{Schema: "gooo/symbolic-invocation-usecase-observation/v1", Decision: "PASS", Resolution: "EXACT", Reason: "EXTERNAL_USER_VALIDATION_REPLAYED", SubjectSHA: string(repeat('1', 40)), ArtifactDigest: artifactDigest, JSONSchemaDigest: schemaDigest, ToolDigest: toolDigest, AcceptedInstances: 1, RejectedInstances: 1, Effects: Effects{}}}
}

func digest(value byte) string { return "sha256:" + string(repeat(value, 64)) }

func repeat(value byte, size int) []byte {
	result := make([]byte, size)
	for index := range result {
		result[index] = value
	}
	return result
}
