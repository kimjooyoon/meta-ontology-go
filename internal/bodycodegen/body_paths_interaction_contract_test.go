package bodycodegen

import (
	"context"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Synthetic weights exercise the adapter. This is not a learned-quality study.
func writeInteractionContractModel(t *testing.T, pooling string) string {
	t.Helper()
	var weights [contractdecision.InteractionRequirementParameterCount]float32
	for i := range weights {
		weights[i] = float32(i%17-8) / 1000
	}
	m, err := contractdecision.NewInteractionRequirementConditioned(weights, pooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func TestInteractionContractPreflightAndConcurrentConstruction(t *testing.T) {
	for _, pooling := range []string{contractdecision.MeanPooling, contractdecision.ExtremePooling} {
		name := writeInteractionContractModel(t, pooling)
		source := conditionModelSource()
		doc := declaredContractDocument(t, source)
		before, err := ExportTypedPathModelContext(context.Background(), "interaction.gooo", source, "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		checkInteractionContractExport(t, before, doc)
		g, err := NewTypedPathGenerator(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for range 3 {
			wg.Go(func() { checkInteractionContractGeneration(t, g, source, doc, before) })
		}
		wg.Wait()
	}
}

func checkInteractionContractExport(t *testing.T, before TypedPathContextExport, doc pathplan.Document) {
	t.Helper()
	info := before.ModelCompatibility.Model
	if before.ModelPredictions != 0 || before.CandidateTests != 0 || before.SelectedEmission || before.RepositoryWrites != 0 ||
		info.ModelSchema != contractdecision.InteractionRequirementSchema || info.ResidentTensorBytes != 76136 ||
		info.FeatureVersion != contractdecision.OrderedSourceFeatureVersion || before.ContractConditions == nil {
		t.Fatal("interaction preflight identity or work differs")
	}
	checkOrderedContractArrays(t, before, doc)
}

func checkOrderedContractArrays(t *testing.T, before TypedPathContextExport, doc pathplan.Document) {
	t.Helper()
	p, err := doc.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	input, err := p.InitialContractInput(doc.TestCases)
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range before.Inputs {
		var want [contractdecision.OrderedFeatureDim]float32
		if err := input.OrderedSourceFeaturesInto(doc.Plan.Decisions[i].ID, &want); err != nil ||
			row.OrderedFeatures == nil || *row.OrderedFeatures != want || row.FlowFeatures != nil || row.Bytes != 2112 {
			t.Fatal("ordered source input differs", i, err)
		}
	}
	if before.ContractConditions.Count != input.ConditionCount() || before.ContractConditions.Bytes != input.ConditionCount()*128 {
		t.Fatal("condition stream was truncated")
	}
	for i, row := range before.ContractConditions.Features {
		want, err := input.ConditionFeatures(i)
		if err != nil || row != want {
			t.Fatal("condition input differs", i, err)
		}
	}
}

func checkInteractionContractGeneration(t *testing.T, g *TypedPathGenerator, source []byte,
	doc pathplan.Document, before TypedPathContextExport) {
	t.Helper()
	checkOrderedContractGeneration(t, g, source, doc, before, "interaction_requirement_contract_fp32")
}

func checkOrderedContractGeneration(t *testing.T, g *TypedPathGenerator, source []byte,
	doc pathplan.Document, before TypedPathContextExport, variant string) {
	t.Helper()
	r, err := g.Generate(context.Background(), "interaction.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1})
	if err != nil {
		t.Error(err)
		return
	}
	p := r.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || p.ContractRanking == nil ||
		p.ContractRanking.Calls != 1 || p.Search.Selection.ModelCalls != 1 ||
		p.Search.Selection.ModelVariant != variant || !reflect.DeepEqual(p.ModelContext, before.Context) {
		t.Error("interaction model did not reach one-ranking checked construction")
		return
	}
	rank := p.ContractRanking
	if rank.ConditionCount != before.ContractConditions.Count || "sha256:"+rank.ConditionFeatureSHA != before.ContractConditions.FeatureSHA ||
		rank.CaseSHA != before.ContractCases.CaseSHA || p.ContractProgress[0].Attempted != 0 {
		t.Error("ranking read different case or condition goals")
	}
	for i, row := range before.Inputs {
		if row.InputSHA != "sha256:"+rank.FeatureSHA[i] {
			t.Error("ranking read different source features")
		}
	}
	for _, progress := range p.ContractProgress {
		if progress.PredictionsThisAdvance != 0 || progress.RankingSHA != rank.SHA {
			t.Error("advancing changed or repeated the ranking")
		}
	}
	if err := VerifyTypedPathProjection(context.Background(), "interaction.gooo", source, doc, r); err != nil {
		t.Error(err)
	}
	checkOrderedContractScores(t, g, before, rank)
}
