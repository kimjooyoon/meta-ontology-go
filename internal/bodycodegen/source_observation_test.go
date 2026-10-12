package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

const sourceObserveFixture = `package observed
namespace observed
entity Integer id "observed://integer"
activity Fee(Integer) -> Integer id "observed://fee" computes "return input - 11" assembling {
 choice "direction" operand_order at "0" intent "11에서 입력을 뺀다. Subtract from eleven."
 case "11" -> "0"
 observe {
  input "17"
  input "-5"
  oracle "Reference"
  max_candidates "2"
  max_rounds "1"
  reuse_outputs "true"
  resolve_unique "true"
 }
 attempts "2"
}
activity Reference(Integer) -> Integer id "observed://reference" computes "return 11 - input"`

func TestSourceObserveConstructsAndReplaysWithoutExternalPlan(t *testing.T) {
	ctx := context.Background()
	source := []byte(sourceObserveFixture)
	doc, err := DecodeSourcePathDocument(ctx, "observed.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "observed.gooo", source, "Fee", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p == nil || p.Observation == nil || p.Observation.Status != "ONE_SURVIVING_CANDIDATE" ||
		p.Observation.EffectiveCases != 2 || p.Observation.OracleEvaluations != 2 ||
		p.DeclaredTestCases != 1 || !pathResolved(p) || p.LocalModelRequested ||
		!strings.Contains(result.Source, "11 - input") {
		t.Fatal("Gooo observation did not resolve the tied original suite", p)
	}
	if len(p.Observation.Rounds[0].Ranking.SurvivingMasks) != 2 || p.Observation.Rounds[0].Observation == nil {
		t.Fatal("original ambiguity and new answer were not retained")
	}
	if err := VerifyTypedPathProjection(ctx, "observed.gooo", source, doc, result); err != nil {
		t.Fatal("source-owned observation failed saved replay", err)
	}
	if err := verifySourcePathObservation(ctx, "observed.gooo", source, "Fee", nil); err == nil {
		t.Fatal("source observation could be omitted from replay")
	}
	prior := *p.Observation
	prior.Options.MaxRounds++
	if err := verifySourcePathObservation(ctx, "observed.gooo", source, "Fee", &prior); err == nil {
		t.Fatal("saved options replaced source meaning")
	}
	_, err = GenerateWithTypedPathOptions(ctx, "observed.gooo", source, "Fee", doc, "missing-model.json",
		TypedPathOptions{Observation: &PathObservationOptions{Inputs: []int64{3}, MaxCandidates: 2, MaxRounds: 1}})
	if err == nil || !strings.Contains(err.Error(), "omit external path observation") {
		t.Fatal("external options displaced source before model loading", err)
	}
}

func TestSourceObserveMissingOracleLeavesTheQuestionOpen(t *testing.T) {
	ctx := context.Background()
	source := []byte(strings.Replace(sourceObserveFixture, `  oracle "Reference"`, "", 1))
	doc, err := DecodeSourcePathDocument(ctx, "question.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "question.gooo", source, "Fee", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths.Observation
	if r == nil || r.Status != "ORACLE_UNAVAILABLE" || r.EffectiveCases != 1 || r.OracleEvaluations != 0 ||
		r.Rounds[0].Observation != nil || r.Rounds[0].Ranking.RecommendedIndex == nil || pathResolved(result.Report.BodyPaths) {
		t.Fatal("missing source answer was invented or ambiguity hidden", r)
	}
}

func TestSourceObserveRetainedGeneratorHonorsTheSameDeclaration(t *testing.T) {
	ctx := context.Background()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(ctx, "retained.gooo", []byte(sourceObserveFixture), "Fee")
	if err != nil || r.Report.BodyPaths == nil || !pathResolved(r.Report.BodyPaths) {
		t.Fatal("retained generation dropped source observe", err)
	}
}

func TestSourceObserveOracleErrorsPrecedeModelLoading(t *testing.T) {
	for _, change := range []struct{ from, to, reason string }{
		{`oracle "Reference"`, `oracle "Fee"`, "separate declared activity"},
		{`oracle "Reference"`, `oracle "Missing"`, "prepare declared Gooo oracle"},
		{`return 11 - input`, `return 12 - input`, "contradicts"},
	} {
		source := []byte(strings.Replace(sourceObserveFixture, change.from, change.to, 1))
		ctx := context.Background()
		doc, err := DecodeSourcePathDocument(ctx, "invalid.gooo", source, "Fee", nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = GenerateWithTypedPaths(ctx, "invalid.gooo", source, "Fee", doc, "missing-model.json")
		if err == nil || !strings.Contains(err.Error(), change.reason) {
			t.Fatal("oracle declaration did not fail before model loading", change, err)
		}
	}
}

func TestSourceObserveCandidateLimitKeepsUnseenAlternatives(t *testing.T) {
	source := []byte(strings.Replace(sourceObserveFixture, `max_candidates "2"`, `max_candidates "1"`, 1))
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "bounded.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "bounded.gooo", source, "Fee", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths.Observation
	if r == nil || r.Status != "CANDIDATE_BUDGET" || r.Rounds[0].Ranking.Unobserved != 1 || pathResolved(result.Report.BodyPaths) {
		t.Fatal("one observed candidate was mistaken for a complete palette", r)
	}
}

func TestSourceObserveReferenceCannotBeAnotherUnfinishedAssembly(t *testing.T) {
	source := []byte(sourceObserveFixture + ` assembling {
 choice "reference_direction" operand_order at "0" intent "Still undecided."
 case "11" -> "0"
 attempts "2"
}`)
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "unfinished-reference.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateWithTypedPaths(ctx, "unfinished-reference.gooo", source, "Fee", doc, "missing-model.json")
	if err == nil || !strings.Contains(err.Error(), "fixed computes activity") {
		t.Fatal("an unresolved choice supplied the reference answer", err)
	}
}
