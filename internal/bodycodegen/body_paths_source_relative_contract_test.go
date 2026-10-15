package bodycodegen

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Separate synthetic weights exercise the connection without replaying a study.
func writeSourceRelativeContractModel(t *testing.T) string {
	t.Helper()
	var weights [contractdecision.InteractionRequirementParameterCount]float32
	for i := range weights {
		weights[i] = float32(i%19-9) / 1000
	}
	model, err := contractdecision.NewSourceRelativeRequirementConditioned(weights)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func checkSourceRelativeContractArrays(t *testing.T, before TypedPathContextExport, doc pathplan.Document) {
	t.Helper()
	prepared, err := doc.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	input, err := prepared.InitialContractInput(doc.TestCases)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Inputs) != len(doc.Plan.Decisions) || before.ContractCases == nil ||
		before.ContractCases.Count != input.CaseCount() || len(before.ContractCases.Features) != input.CaseCount() {
		t.Fatal("source-relative export omitted a choice or output goal")
	}
	for i, row := range before.Inputs {
		var want [contractdecision.OrderedFeatureDim]float32
		if err := input.SourceRelativeOrderedSourceFeaturesInto(doc.Plan.Decisions[i].ID, &want); err != nil ||
			row.SourceRelativeFeatures == nil || *row.SourceRelativeFeatures != want || row.CanonicalFeatures != nil ||
			row.OrderedFeatures != nil || row.FlowFeatures != nil || row.Bytes != 2112 || row.InputSHA != candidateFeatureDigest(want[:]) {
			t.Fatal("source-relative input differs from SDK input", i, err)
		}
		view, err := input.SourceRelativeConditions(doc.Plan.Decisions[i].ID)
		if err != nil || row.SourceRelativeConditions == nil || *row.SourceRelativeConditions != view {
			t.Fatal("condition relation explanation differs from source", i, err)
		}
		branch, err := prepared.CanonicalOrderedBranchContext(doc.Plan.Decisions[i].ID)
		if err != nil || row.CanonicalBranch == nil || *row.CanonicalBranch != branch || row.OrderedBranch != nil {
			t.Fatal("resolved branch explanation differs from source", i, err)
		}
	}
	for i, row := range before.ContractCases.Features {
		want, err := input.CaseFeatures(i)
		if err != nil || row != want {
			t.Fatal("declared output differs", i, err)
		}
	}
	if before.ContractConditions == nil || before.ContractConditions.Count != input.ConditionCount() ||
		before.ContractConditions.Bytes != input.ConditionCount()*128 || len(before.ContractConditions.Features) != input.ConditionCount() {
		t.Fatal("source-relative model lost declared conditions")
	}
	for i, row := range before.ContractConditions.Features {
		want, err := input.ConditionFeatures(i)
		if err != nil || row != want {
			t.Fatal("declared condition differs", i, err)
		}
	}
}

func TestSourceRelativeContractPreflightAndConcurrentConstruction(t *testing.T) {
	name := writeSourceRelativeContractModel(t)
	source := interactionAssignmentSource(t)
	doc := declaredContractDocument(t, source)
	before, err := ExportTypedPathModelContext(context.Background(), "relative.gooo", source, "Choose", doc, name, "")
	if err != nil {
		t.Fatal(err)
	}
	info := before.ModelCompatibility.Model
	if info.ModelSchema != contractdecision.SourceRelativeRequirementSchema || info.ResidentTensorBytes != 76136 ||
		info.FeatureVersion != contractdecision.SourceRelativeOrderedSourceFeatureVersion || before.ModelPredictions != 0 ||
		before.CandidateTests != 0 || before.SelectedEmission || before.RepositoryWrites != 0 ||
		before.Context.Schema != "gooo/compiler-source-relative-contract-model-context/v1" {
		t.Fatal("source-relative identity or zero-work inspection changed")
	}
	checkSourceRelativeContractArrays(t, before, doc)
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for range 3 {
		workers.Go(func() {
			checkOrderedContractGeneration(t, g, source, doc, before, "source_relative_requirement_contract_fp32")
		})
	}
	workers.Wait()
}

func TestSourceRelativeContractEveryConditionAndEmptyStream(t *testing.T) {
	checkContractConditionStreams(t, writeSourceRelativeContractModel(t), checkSourceRelativeContractArrays)
}

func TestSourceRelativeContractCancellationAndPartialProgress(t *testing.T) {
	checkContractCancellationAndContradictoryCases(t, writeSourceRelativeContractModel(t))
}
