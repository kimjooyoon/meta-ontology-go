package bodycodegen

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func writeDeclaredContractModel(t *testing.T) string {
	t.Helper()
	m, err := contractdecision.New([contractdecision.ParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func declaredContractDocument(t *testing.T, source []byte) pathplan.Document {
	t.Helper()
	doc, err := DecodeSourcePathDocument(context.Background(), "contract.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestDeclaredContractPreflightMatchesActualInputAndSearch(t *testing.T) {
	ctx := context.Background()
	source := conditionModelSource()
	doc := declaredContractDocument(t, source)
	name := writeDeclaredContractModel(t)
	before, err := ExportTypedPathModelContext(ctx, "contract.gooo", source, "Choose", doc, name, "")
	if err != nil {
		t.Fatal(err)
	}
	if before.ModelPredictions != 0 || before.CandidateTests != 0 || before.ContractCases == nil || before.ContractCases.Count != len(doc.TestCases) || before.ModelCompatibility.Model.ModelSchema != contractdecision.Schema || before.ModelCompatibility.Model.ResidentTensorBytes != 38984 {
		t.Fatal("declared contract preflight identity changed")
	}
	for i, row := range doc.TestCases {
		var want [decision.DeclaredCaseFeatureDim]float32
		if err := decision.DeclaredCaseFeaturesInto(row.Input, row.Expected, &want); err != nil || want != before.ContractCases.Features[i] {
			t.Fatal("declared int64 case differs from model input", i, err)
		}
	}
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() { checkDeclaredContractGeneration(t, g, source, doc, before) })
	}
	wg.Wait()
}

func checkDeclaredContractGeneration(t *testing.T, g *TypedPathGenerator, source []byte, doc pathplan.Document, before TypedPathContextExport) {
	t.Helper()
	r, err := g.Generate(context.Background(), "contract.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1})
	if err != nil {
		t.Error(err)
		return
	}
	p := r.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || p.Search.Selection.ModelCalls != 1 || p.Search.Selection.ModelVariant != "contract_fp32" || p.ContractRanking == nil || len(p.ContractProgress) < 2 || len(p.ConditionProgress) != 0 || !reflect.DeepEqual(p.ModelContext, before.Context) {
		t.Error("one initial contract judgment did not reach checked assembly")
		return
	}
	if p.ContractRanking.CaseSHA != before.ContractCases.CaseSHA || p.ContractRanking.Calls != 1 || p.ContractProgress[0].Attempted != 0 || len(p.Search.Attempts) != 4 {
		t.Error("initial ranking or finite continuation changed")
	}
	for i, input := range before.Inputs {
		if input.InputSHA != "sha256:"+p.ContractRanking.FeatureSHA[i] {
			t.Error("actual source arrays differ from preflight")
		}
	}
	for i, progress := range p.ContractProgress {
		if progress.RankingSHA != p.ContractRanking.SHA || progress.PredictionsThisAdvance != 0 || i > 0 && progress.PreviousSHA != p.ContractProgress[i-1].SessionProgress.SHA {
			t.Error("progress lost ranking identity or added model calls")
		}
	}
	if err := VerifyTypedPathProjection(context.Background(), "contract.gooo", source, doc, r); err != nil {
		t.Error(err)
	}
}

func TestDeclaredContractExpectedLargeIntegerChangesOnlyCaseInput(t *testing.T) {
	source := conditionModelSource()
	changed := []byte(strings.Replace(string(source), `case "9007199254740993" -> "9007199254740993"`, `case "9007199254740993" -> "9007199254740994"`, 1))
	name := writeDeclaredContractModel(t)
	var exports []TypedPathContextExport
	for _, raw := range [][]byte{source, changed} {
		doc := declaredContractDocument(t, raw)
		e, err := ExportTypedPathModelContext(context.Background(), "contract.gooo", raw, "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		exports = append(exports, e)
	}
	if !reflect.DeepEqual(exports[0].Inputs, exports[1].Inputs) || exports[0].ContractCases.CaseSHA == exports[1].ContractCases.CaseSHA || exports[0].ContractCases.FeatureSHA == exports[1].ContractCases.FeatureSHA || exports[0].ContractCases.Features[3] == exports[1].ContractCases.Features[3] {
		t.Fatal("one-unit goal difference above 2^53 was lost or contaminated source inputs")
	}
}

func TestDeclaredContractOptionsCancellationAndPartial(t *testing.T) {
	source := conditionModelSource()
	doc := declaredContractDocument(t, source)
	g, err := NewTypedPathGenerator(writeDeclaredContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Generate(context.Background(), "contract.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1, FeedbackRounds: 1})
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SearchStarted || !strings.Contains(err.Error(), "one initial ranking") {
		t.Fatal("unsupported feedback did not explain the model contract", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := g.Generate(cancelled, "contract.gooo", source, "Choose", doc, TypedPathOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation ignored", err)
	}
	partialSource := []byte(strings.Replace(string(source), `case "9" -> "9"`, `case "9" -> "10"`, 1))
	partialDoc := declaredContractDocument(t, partialSource)
	r, err := g.Generate(context.Background(), "contract.gooo", partialSource, "Choose", partialDoc, TypedPathOptions{StepAttempts: 1})
	if err != nil || r.Report.BodyPaths.FunctionalCompleteness >= 100 || r.Report.BodyPaths.Search.Status == "TRAINING_COMPLETE" || len(r.Report.BodyPaths.Search.Attempts) != 4 {
		t.Fatal("contradictory goal was promoted to completeness", err)
	}
}
