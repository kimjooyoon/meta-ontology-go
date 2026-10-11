package bodycodegen

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Synthetic weights check compiler integration, not trained-model quality.
func writeCanonicalContractModel(t *testing.T, pooling string) string {
	t.Helper()
	var weights [contractdecision.InteractionRequirementParameterCount]float32
	for i := range weights {
		weights[i] = float32(i%17-8) / 1000
	}
	model, err := contractdecision.NewCanonicalInteractionRequirementConditioned(weights, pooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func checkCanonicalContractArrays(t *testing.T, before TypedPathContextExport, doc pathplan.Document) {
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
		t.Fatal("canonical export omitted a choice or output goal")
	}
	for i, row := range before.ContractCases.Features {
		want, err := input.CaseFeatures(i)
		if err != nil || row != want {
			t.Fatal("declared output input differs", i, err)
		}
	}
	for i, row := range before.Inputs {
		var want [contractdecision.OrderedFeatureDim]float32
		if err := input.CanonicalOrderedSourceFeaturesInto(doc.Plan.Decisions[i].ID, &want); err != nil ||
			row.CanonicalFeatures == nil || *row.CanonicalFeatures != want || row.OrderedFeatures != nil ||
			row.FlowFeatures != nil || row.Bytes != 2112 || row.InputSHA != candidateFeatureDigest(want[:]) {
			t.Fatal("canonical input differs from SDK input", i, err)
		}
		view, err := prepared.CanonicalOrderedBranchContext(doc.Plan.Decisions[i].ID)
		if err != nil || row.CanonicalBranch == nil || *row.CanonicalBranch != view || row.OrderedBranch != nil {
			t.Fatal("canonical explanation differs from source", i, err)
		}
	}
	if before.ContractConditions == nil || before.ContractConditions.Count != input.ConditionCount() ||
		before.ContractConditions.Bytes != input.ConditionCount()*128 {
		t.Fatal("canonical model lost declared conditions")
	}
	for i, row := range before.ContractConditions.Features {
		want, err := input.ConditionFeatures(i)
		if err != nil || row != want {
			t.Fatal("condition row differs", i, err)
		}
	}
}

func TestCanonicalContractPreflightAndConcurrentConstruction(t *testing.T) {
	for _, pooling := range []string{contractdecision.MeanPooling, contractdecision.ExtremePooling} {
		name := writeCanonicalContractModel(t, pooling)
		source := interactionAssignmentSource(t)
		doc := declaredContractDocument(t, source)
		before, err := ExportTypedPathModelContext(context.Background(), "canonical.gooo", source, "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		info := before.ModelCompatibility.Model
		if info.ModelSchema != contractdecision.CanonicalInteractionRequirementSchema || info.ResidentTensorBytes != 76136 ||
			info.FeatureVersion != contractdecision.CanonicalOrderedSourceFeatureVersion || before.ModelPredictions != 0 ||
			before.CandidateTests != 0 || before.SelectedEmission || before.RepositoryWrites != 0 ||
			before.Context.Schema != "gooo/compiler-canonical-interaction-contract-model-context/v1" {
			t.Fatal("canonical identity or zero-work inspection changed")
		}
		checkCanonicalContractArrays(t, before, doc)
		generator, err := NewTypedPathGenerator(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
		var workers sync.WaitGroup
		for range 3 {
			workers.Go(func() {
				checkOrderedContractGeneration(t, generator, source, doc, before, "canonical_interaction_requirement_contract_fp32")
			})
		}
		workers.Wait()
	}
}

func TestCanonicalContractEveryConditionAndEmptyStream(t *testing.T) {
	checkContractConditionStreams(t, writeCanonicalContractModel(t, contractdecision.MeanPooling), checkCanonicalContractArrays)
}

func TestCanonicalContractCancellationAndPartialProgress(t *testing.T) {
	checkContractCancellationAndContradictoryCases(t, writeCanonicalContractModel(t, contractdecision.ExtremePooling))
}
