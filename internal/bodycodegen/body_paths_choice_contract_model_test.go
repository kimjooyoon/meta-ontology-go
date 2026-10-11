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

func writeChoiceContractModel(t *testing.T, pooling string) string {
	t.Helper()
	var weights [contractdecision.ParameterCount]float32
	var contextWeights [contractdecision.ChoiceContextParameterCount]float32
	for i := range weights {
		weights[i] = float32(i%19-9) / 1000
	}
	for i := range contextWeights {
		contextWeights[i] = float32(i%13-6) / 1000
	}
	m, err := contractdecision.NewChoiceConditioned(weights, contextWeights, pooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return writeCandidateArtifact(t, raw)
}

func TestChoiceContractPreflightAndConcurrentConstruction(t *testing.T) {
	for _, pooling := range []string{contractdecision.MeanPooling, contractdecision.ExtremePooling} {
		name := writeChoiceContractModel(t, pooling)
		source := conditionModelSource()
		doc := declaredContractDocument(t, source)
		before, err := ExportTypedPathModelContext(context.Background(), "choice.gooo", source, "Choose", doc, name, "")
		if err != nil {
			t.Fatal(err)
		}
		info := before.ModelCompatibility.Model
		if before.ModelPredictions != 0 || before.CandidateTests != 0 || before.ContractCases == nil || info.ModelSchema != contractdecision.ChoiceSchema || info.ResidentTensorBytes != 51272 {
			t.Fatal("choice preflight identity", info)
		}
		for i, c := range doc.TestCases {
			var want [decision.DeclaredCaseFeatureDim]float32
			if err := decision.DeclaredCaseFeaturesInto(c.Input, c.Expected, &want); err != nil || want != before.ContractCases.Features[i] {
				t.Fatal("exact int64 input", i, err)
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
			wg.Go(func() { checkChoiceContractGeneration(t, g, source, doc, before) })
		}
		wg.Wait()
	}
}

func checkChoiceContractGeneration(t *testing.T, g *TypedPathGenerator, source []byte, doc pathplan.Document, before TypedPathContextExport) {
	t.Helper()
	r, err := g.Generate(context.Background(), "choice.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1})
	if err != nil {
		t.Error(err)
		return
	}
	p := r.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 || p.Search.Selection.ModelVariant != "choice_conditioned_contract_fp32" || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || !reflect.DeepEqual(p.ModelContext, before.Context) {
		t.Error("choice model did not reach full checked construction")
		return
	}
	if p.ContractRanking.CaseSHA != before.ContractCases.CaseSHA || p.ContractProgress[0].Attempted != 0 || len(p.ConditionProgress) != 0 {
		t.Error("preflight/initial ranking differs")
	}
	for i, row := range before.Inputs {
		if row.InputSHA != "sha256:"+p.ContractRanking.FeatureSHA[i] {
			t.Error("actual source inputs differ")
		}
	}
	for i, progress := range p.ContractProgress {
		if progress.PredictionsThisAdvance != 0 || progress.RankingSHA != p.ContractRanking.SHA || i > 0 && progress.PreviousSHA != p.ContractProgress[i-1].SessionProgress.SHA {
			t.Error("extra prediction or broken chain")
		}
	}
	if err := VerifyTypedPathProjection(context.Background(), "choice.gooo", source, doc, r); err != nil {
		t.Error(err)
	}
	checkChoiceContractScores(t, g, doc, before, p.ContractRanking)
}

func checkChoiceContractScores(t *testing.T, g *TypedPathGenerator, doc pathplan.Document, before TypedPathContextExport, rank *pathplan.ContractRanking) {
	t.Helper()
	p, err := doc.Prepare()
	if err != nil {
		t.Error(err)
		return
	}
	input, err := p.InitialContractInput(doc.TestCases)
	if err != nil {
		t.Error(err)
		return
	}
	var source [][contractdecision.FeatureDim]float32
	for _, row := range before.Inputs {
		source = append(source, *row.FlowFeatures)
	}
	var workspace contractdecision.ChoiceWorkspace
	var expected contractdecision.ChoicePrediction
	model, ok := g.condition.contract.(*contractdecision.ChoiceModel)
	if !ok {
		t.Error("retained model is not choice-conditioned")
		return
	}
	if err := model.PredictChoicesInto(source, input, &workspace, &expected); err != nil {
		t.Error(err)
		return
	}
	if expected.Logits != rank.Logits || rank.ModelFingerprint != model.Fingerprint() {
		t.Error("compiler ranking differs from retained choice model")
	}
}

func TestChoiceContractRestrictionsCancellationAndContradiction(t *testing.T) {
	source := conditionModelSource()
	doc := declaredContractDocument(t, source)
	g, err := NewTypedPathGenerator(writeChoiceContractModel(t, contractdecision.ExtremePooling))
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Generate(context.Background(), "choice.gooo", source, "Choose", doc, TypedPathOptions{StepAttempts: 1, FeedbackRounds: 1})
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SearchStarted || !strings.Contains(err.Error(), "one initial ranking") {
		t.Fatal("unsupported feedback", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = g.Generate(ctx, "choice.gooo", source, "Choose", doc, TypedPathOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(string(source), `case "9" -> "9"`, `case "9" -> "10"`, 1))
	partial, err := g.Generate(context.Background(), "choice.gooo", changed, "Choose", declaredContractDocument(t, changed), TypedPathOptions{StepAttempts: 1})
	if err != nil || partial.Report.BodyPaths.FunctionalCompleteness >= 100 || len(partial.Report.BodyPaths.Search.Attempts) != 4 || partial.Report.BodyPaths.ContractRanking.Calls != 1 {
		t.Fatal("contradiction accepted or extra prediction", err)
	}
}

func TestGlobalPoolingSchemaAndUnsupportedCaseABI(t *testing.T) {
	m, err := contractdecision.NewForPooling([contractdecision.ParameterCount]float32{}, contractdecision.ExtremePooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := writeCandidateArtifact(t, raw)
	source := conditionModelSource()
	doc := declaredContractDocument(t, source)
	before, err := ExportTypedPathModelContext(context.Background(), "pool.gooo", source, "Choose", doc, name, "")
	if err != nil || before.ModelCompatibility.Model.ModelSchema != contractdecision.PoolingSchema || before.ModelCompatibility.Model.ResidentTensorBytes != 38984 {
		t.Fatal("pooling identity", err)
	}
	r, err := GenerateWithTypedPaths(context.Background(), "pool.gooo", source, "Choose", doc, name)
	if err != nil || r.Report.BodyPaths.ContractRanking == nil || r.Report.BodyPaths.ContractRanking.Calls != 1 {
		t.Fatal("global pooling dispatch", err)
	}
	literal, err := contractdecision.NewForCaseFeatures([contractdecision.ParameterCount]float32{}, contractdecision.ExtremePooling, decision.SourceLiteralCaseFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = literal.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = NewTypedPathGenerator(writeCandidateArtifact(t, raw)); err == nil {
		t.Fatal("unsupported case encoding accepted")
	}
}

func TestChoiceContractRetainedModelReadsEachExactGoal(t *testing.T) {
	name := writeChoiceContractModel(t, contractdecision.ExtremePooling)
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	source := conditionModelSource()
	changed := []byte(strings.Replace(string(source), `case "9007199254740993" -> "9007199254740993"`, `case "9007199254740993" -> "9007199254740994"`, 1))
	var rankings []pathplan.ContractRanking
	for _, raw := range [][]byte{source, changed} {
		doc := declaredContractDocument(t, raw)
		r, err := g.Generate(context.Background(), "choice.gooo", raw, "Choose", doc, TypedPathOptions{StepAttempts: 1})
		if err != nil {
			t.Fatal(err)
		}
		rankings = append(rankings, *r.Report.BodyPaths.ContractRanking)
	}
	if rankings[0].FeatureSHA != rankings[1].FeatureSHA || rankings[0].CaseSHA == rankings[1].CaseSHA || rankings[0].Calls != 1 || rankings[1].Calls != 1 {
		t.Fatal("retained model reused another request's goal or changed source inputs")
	}
}
