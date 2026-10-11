package bodycodegen

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

func writeOrderedContractModel(t *testing.T, pooling string) string {
	t.Helper()
	var weights [contractdecision.OrderedRequirementParameterCount]float32
	for i := range weights {
		weights[i] = float32(i%13-6) / 1000
	}
	model, err := contractdecision.NewOrderedRequirementConditioned(weights, pooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func TestOrderedContractModelPreflightAndConstruction(t *testing.T) {
	for _, pooling := range []string{contractdecision.MeanPooling, contractdecision.ExtremePooling} {
		name := writeOrderedContractModel(t, pooling)
		source := interactionAssignmentSource(t)
		doc := declaredContractDocument(t, source)
		before, err := ExportTypedPathModelContext(context.Background(), "ordered.gooo", source, "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		info := before.ModelCompatibility.Model
		if info.ModelSchema != contractdecision.OrderedRequirementSchema || info.ResidentTensorBytes != 71560 ||
			before.ModelPredictions != 0 || before.CandidateTests != 0 || before.ContractConditions == nil ||
			before.Context.Schema != "gooo/compiler-ordered-contract-model-context/v1" {
			t.Fatal("ordered model identity, condition channel or zero-work inspection changed")
		}
		checkOrderedContractArrays(t, before, doc)
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
				checkOrderedContractGeneration(t, generator, source, doc, before, "ordered_requirement_contract_fp32")
			})
		}
		workers.Wait()
	}
}

func TestOrderedContractEveryConditionAndEmptyStream(t *testing.T) {
	checkContractConditionStreams(t, writeOrderedContractModel(t, contractdecision.MeanPooling), checkOrderedContractArrays)
}

func TestOrderedContractCancellationAndContradictoryCases(t *testing.T) {
	checkContractCancellationAndContradictoryCases(t, writeOrderedContractModel(t, contractdecision.ExtremePooling))
}
