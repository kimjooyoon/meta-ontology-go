package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func directCanonicalSource(t *testing.T, source []byte) []byte {
	t.Helper()
	prefix, body, ok := strings.Cut(string(source), ` computes "`)
	if !ok {
		t.Fatal("fixture has no computation")
	}
	_, suffix, ok := strings.Cut(body, `" assembling`)
	if !ok {
		t.Fatal("fixture has no assembly declaration")
	}
	return []byte(prefix + ` computes "if input < 0 { return input } else { return 0 - input }" assembling` + suffix)
}

func TestCanonicalContractEquivalentFormsReachSameRanking(t *testing.T) {
	model := writeCanonicalContractModel(t, contractdecision.ExtremePooling)
	generator, err := NewTypedPathGenerator(model)
	if err != nil {
		t.Fatal(err)
	}
	assignment := interactionAssignmentSource(t)
	var exports []TypedPathContextExport
	var ranks []*pathplan.ContractRanking
	for _, source := range [][]byte{assignment, directCanonicalSource(t, assignment)} {
		doc := declaredContractDocument(t, source)
		exported, err := ExportTypedPathModelContext(context.Background(), "equivalent.gooo", source, "Choose", doc, model, "")
		if err != nil {
			t.Fatal(err)
		}
		checkCanonicalContractArrays(t, exported, doc)
		result, err := generator.Generate(context.Background(), "equivalent.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1})
		if err != nil || result.Report.BodyPaths.FunctionalCompleteness != 100 || result.Report.BodyPaths.ContractRanking == nil {
			t.Fatal("equivalent source failed construction", err)
		}
		exports = append(exports, exported)
		ranks = append(ranks, result.Report.BodyPaths.ContractRanking)
	}
	if exports[0].OriginalSourceSHA256 == exports[1].OriginalSourceSHA256 ||
		exports[0].Context.OriginalPlanSHA == exports[1].Context.OriginalPlanSHA {
		t.Fatal("canonical input erased original source identity")
	}
	for i, row := range exports[0].Inputs {
		other := exports[1].Inputs[i]
		if *row.CanonicalFeatures != *other.CanonicalFeatures || row.InputSHA != other.InputSHA ||
			*row.CanonicalBranch != *other.CanonicalBranch {
			t.Fatal("equivalent assignment and return inputs differ", i)
		}
	}
	if ranks[0].Logits != ranks[1].Logits || ranks[0].Proposed != ranks[1].Proposed ||
		ranks[0].Calls != 1 || ranks[1].Calls != 1 {
		t.Fatal("equivalent forms changed or repeated ranking")
	}
}

func TestCanonicalContractKeepsLargeLiteralAndOperandOrder(t *testing.T) {
	model := writeCanonicalContractModel(t, contractdecision.MeanPooling)
	source := strings.Replace(string(interactionAssignmentSource(t)), "limit = 0", "limit = 9007199254740993", 1)
	var previous string
	for _, changed := range []string{source, strings.Replace(source, "result = limit - value", "result = value - limit", 1)} {
		doc := declaredContractDocument(t, []byte(changed))
		exported, err := ExportTypedPathModelContext(context.Background(), "large.gooo", []byte(changed), "Choose", doc, model, "")
		if err != nil || len(exported.Inputs) != 2 || exported.ModelPredictions != 0 || exported.CandidateTests != 0 {
			t.Fatal("canonical source inspection failed", err)
		}
		for _, row := range exported.Inputs {
			if row.CanonicalBranch == nil || row.CanonicalBranch.Expressions[0].Operands[1].Int != 9007199254740993 {
				t.Fatal("large literal lost precision")
			}
			raw, err := json.Marshal(row.CanonicalBranch)
			if err != nil || !strings.Contains(string(raw), "9007199254740993") {
				t.Fatal("exact literal absent from JSON", err)
			}
		}
		if previous == exported.Inputs[0].InputSHA {
			t.Fatal("canonical source erased operand order")
		}
		previous = exported.Inputs[0].InputSHA
	}
}

func TestCanonicalContractRejectsMislabeledInputFormat(t *testing.T) {
	name := writeCanonicalContractModel(t, contractdecision.MeanPooling)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(raw), contractdecision.CanonicalOrderedSourceFeatureVersion, contractdecision.OrderedSourceFeatureVersion, 1)
	if _, err := NewTypedPathGenerator(writeCandidateArtifact(t, []byte(changed))); err == nil {
		t.Fatal("canonical model accepted legacy input label")
	}
}
