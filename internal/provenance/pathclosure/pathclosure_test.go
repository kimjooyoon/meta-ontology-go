package pathclosure_test

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestCompleteDeclarationDerivationProjectionVerificationPath(t *testing.T) {
	fixture := completeInferenceFixture()
	if err := fixture.path.Validate(); err != nil {
		t.Fatalf("complete path rejected: %v", err)
	}
	normalized, err := fixture.path.Normalized()
	if err != nil {
		t.Fatalf("complete path normalization failed: %v", err)
	}
	assertExactRecordSequence(t, normalized.Edges, fixture.edges)
	chain, err := semantic.NewInferencePathChain(fixture.edges...)
	if err != nil {
		t.Fatalf("complete chain rejected: %v", err)
	}
	assertExactRecordSequence(t, chain.Edges, fixture.edges)
	if got := []semantic.InferenceKind{
		chain.Edges[0].Kind, chain.Edges[1].Kind, chain.Edges[2].Kind, chain.Edges[3].Kind,
	}; !reflect.DeepEqual(got, []semantic.InferenceKind{
		semantic.InferenceAuthoritativeDeclaration,
		semantic.InferenceDeterministicDerivation,
		semantic.InferenceDerivedProjection,
		semantic.InferenceIndependentVerification,
	}) {
		t.Fatalf("chain kind sequence = %#v", got)
	}
}

func TestInsertionOrderReplayPreservesExactRecordSequence(t *testing.T) {
	fixture := completeInferenceFixture()
	reordered := reorderedFixture(fixture)
	original := clonePath(reordered)
	left, err := fixture.path.Normalized()
	if err != nil {
		t.Fatalf("original normalization failed: %v", err)
	}
	right, err := reordered.Normalized()
	if err != nil {
		t.Fatalf("reordered normalization failed: %v", err)
	}
	assertExactRecordSequence(t, right.Edges, left.Edges)
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatal("insertion order changed the normalized path receipt")
	}
	if !reflect.DeepEqual(reordered, original) {
		t.Fatal("normalization mutated the insertion-ordered fixture")
	}
}

func TestMissingEdgeIsNotAnEmptySuccessfulChain(t *testing.T) {
	fixture := completeInferenceFixture()
	incomplete := append([]semantic.InferenceEdge(nil), fixture.edges[:2]...)
	incomplete = append(incomplete, fixture.edges[3])
	if err := (semantic.InferencePathV1{
		Version: semantic.InferencePathSchemaVersion, Edges: incomplete, Evidence: fixture.evidence,
	}).Validate(); err != nil {
		t.Fatalf("record-valid incomplete path was rejected before topology check: %v", err)
	}
	if _, err := semantic.NewInferencePathChain(incomplete...); err == nil {
		t.Fatal("missing edge produced a successful chain")
	}
}

func TestZeroRequirementsIsNotASuccessfulPath(t *testing.T) {
	if _, err := semantic.NewInferencePathChain(); err == nil {
		t.Fatal("zero requirements produced a successful chain")
	}
}

func requirementForPath(pathID string, edges []semantic.InferenceEdge) pathclosure.Requirement {
	records := make([]semantic.ID, 0, len(edges))
	kinds := make([]semantic.InferenceKind, 0, len(edges))
	for _, edge := range edges {
		records = append(records, edge.RecordID)
		kinds = append(kinds, edge.Kind)
	}
	return pathclosure.Requirement{
		PathID:        fixtureID("path/" + pathID),
		RecordIDs:     records,
		ExpectedKinds: kinds,
		StartID:       edges[0].SubjectID,
		EndID:         edges[len(edges)-1].ObjectID,
	}
}

func assertEvaluation(t *testing.T, got pathclosure.Result, status pathclosure.Status, code string, numerator, denominator int) {
	t.Helper()
	if got.Status != status || got.Code != code {
		t.Fatalf("evaluation status/code = %s/%s, want %s/%s: %#v", got.Status, got.Code, status, code, got)
	}
	if got.Numerator != numerator || got.Denominator != denominator {
		t.Fatalf("evaluation coverage = %d/%d, want %d/%d: %#v", got.Numerator, got.Denominator, numerator, denominator, got)
	}
}

func TestEvaluatorRunsClosureOutcomePartitions(t *testing.T) {
	fixture := completeInferenceFixture()
	complete := requirementForPath("complete", fixture.edges)
	missing := complete
	missing.PathID = fixtureID("path/missing")
	missing.RecordIDs = append([]semantic.ID(nil), complete.RecordIDs...)
	missing.ExpectedKinds = append([]semantic.InferenceKind(nil), complete.ExpectedKinds...)
	missing.RecordIDs[1] = fixtureID("record/not-present")

	t.Run("complete", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{complete})
		assertEvaluation(t, got, pathclosure.PASS, pathclosure.CodePass, 1, 1)
		if !reflect.DeepEqual(got.Complete, []semantic.ID{complete.PathID}) {
			t.Fatalf("complete paths = %#v", got.Complete)
		}
	})
	t.Run("missing edge UNKNOWN", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{missing})
		assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeMissingRecord, 0, 1)
		if !reflect.DeepEqual(got.Missing, []semantic.ID{missing.PathID}) {
			t.Fatalf("missing paths = %#v", got.Missing)
		}
	})
	t.Run("zero requirements UNKNOWN", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, nil)
		assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeZeroDenominator, 0, 0)
	})
	t.Run("two paths one incomplete UNKNOWN", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{complete, missing})
		assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeMissingRecord, 1, 2)
		if !reflect.DeepEqual(got.Complete, []semantic.ID{complete.PathID}) ||
			!reflect.DeepEqual(got.Missing, []semantic.ID{missing.PathID}) {
			t.Fatalf("mixed paths = %#v", got)
		}
	})
	t.Run("wrong phase order FAIL_CLOSED", func(t *testing.T) {
		wrongOrder := requirementForPath("wrong-order", []semantic.InferenceEdge{fixture.edges[1], fixture.edges[0]})
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{wrongOrder})
		assertEvaluation(t, got, pathclosure.FAIL_CLOSED, pathclosure.CodeMalformed, 0, 1)
		if !reflect.DeepEqual(got.Malformed, []semantic.ID{wrongOrder.PathID}) {
			t.Fatalf("malformed paths = %#v", got.Malformed)
		}
	})
}

func TestEvaluatorTreatsMissingAndAmbiguousEvidenceAsFullSuiteUnknown(t *testing.T) {
	fixture := completeInferenceFixture()
	first := requirementForPath("first", fixture.edges)
	second := requirementForPath("second", fixture.edges)

	cases := []struct {
		name   string
		mutate func(*semantic.InferencePathV1)
	}{
		{
			name: "orphan evidence",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[0].Evidence[0].ID = fixtureID("evidence/not-present")
			},
		},
		{
			name: "ambiguous evidence",
			mutate: func(path *semantic.InferencePathV1) {
				path.Edges[0].Evidence = append(path.Edges[0].Evidence, path.Edges[0].Evidence[0])
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := clonePath(fixture.path)
			test.mutate(&path)
			got := pathclosure.Evaluate(path, []pathclosure.Requirement{first, second})
			assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeMissingEvidence, 0, 2)
			if !reflect.DeepEqual(got.Missing, []semantic.ID{first.PathID, second.PathID}) {
				t.Fatalf("full-suite missing paths = %#v", got.Missing)
			}
		})
	}
}
