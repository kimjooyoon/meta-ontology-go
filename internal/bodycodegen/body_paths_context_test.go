package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Synthetic weights exercise the compiler-input contract, not trained quality.
func writePathContextContractModel(t *testing.T) string {
	return writePathContextContractModelVersion(t, decision.SplitContextIntentFeatureVersion)
}

func writePathContextContractModelVersion(t *testing.T, version string) string {
	t.Helper()
	name := writeTypedPathContractModel(t, false)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var metadata decision.Metadata
	if err = json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	metadata.FeatureVersion = version
	raw, err = json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestTypedPathCompilerContextOwnsValidatedFacts(t *testing.T) {
	_, document := conditionalPathFixture(t, "ko")
	original, _ := json.Marshal(document)
	prepared, err := document.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	model, err := decision.LoadPath(writePathContextContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	ranked, receipt, declined, err := preparePathModelContext(context.Background(), document, prepared,
		model, "sample://activity/conditional", "source-digest")
	if err != nil || declined || receipt == nil || receipt.Status != "ENCODED" || ranked == prepared {
		t.Fatalf("context not prepared: %+v, %v", receipt, err)
	}
	after, _ := json.Marshal(document)
	if string(after) != string(original) || receipt.OriginalPlanSHA != prepared.PlanSHA256() ||
		receipt.RankedPlanSHA != ranked.PlanSHA256() || len(receipt.Inputs) != 6 ||
		receipt.SourceSemanticSHA != "source-digest" || receipt.ActivityID != "sample://activity/conditional" {
		t.Fatal("caller ownership or source identity lost")
	}
	facts := fallbackContextFacts(document.Plan)
	for index, choice := range document.Plan.Decisions {
		text, input, reason := encodePathContext(facts, choice)
		suffix := choice.Intent[strings.LastIndex(choice.Intent, "intent: ")+len("intent: "):]
		if reason != "" || input != receipt.Inputs[index] || input.Bytes != len(text) ||
			input.InputSHA != digest([]byte(text)) || input.NaturalIntentSHA != digest([]byte(suffix)) ||
			!strings.HasSuffix(text, ";intent: "+suffix) || !strings.Contains(text, ";legal=") {
			t.Fatalf("intent/facts hash mismatch: %s", text)
		}
		choice.Intent = "untrusted compiler facts intent: " + suffix
		replaced, _, _ := encodePathContext(facts, choice)
		if replaced != text {
			t.Fatal("caller prefix influenced compiler facts")
		}
	}
	document.TestCases[0].Expected = 999
	other, _, _, err := preparePathModelContext(context.Background(), document, prepared, model, "display-renamed", "other")
	if err != nil || other.PlanSHA256() != ranked.PlanSHA256() {
		t.Fatal("test outcomes leaked into initial input")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, _, err = preparePathModelContext(ctx, document, prepared, model, "", ""); !errors.Is(err, context.Canceled) {
		t.Fatal("context construction ignored cancellation")
	}
}

func TestTypedPathCompilerContextUsesFallbackSnapshot(t *testing.T) {
	_, document := conditionalPathFixture(t, "en")
	original, _ := json.Marshal(document.Plan)
	for i := range document.Plan.Decisions {
		choice := &document.Plan.Decisions[i]
		choice.Fallback = choice.Options[1].Label
	}
	before, _ := json.Marshal(document.Plan)
	facts := fallbackContextFacts(document.Plan)
	prepared, err := document.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	for _, choice := range document.Plan.Decisions {
		switch choice.Kind {
		case pathplan.LocalReference:
			if facts.Expressions[choice.Target].Name != choice.Options[1].Name {
				t.Fatal("reference fallback missing")
			}
		case pathplan.AssignmentTarget:
			if facts.Statements[choice.Target].Name != choice.Options[1].Name {
				t.Fatal("assignment fallback missing")
			}
		case pathplan.OperandOrder:
			if facts.Expressions[choice.Target].Left != document.Plan.Base.Expressions[choice.Target].Right {
				t.Fatal("operand fallback missing")
			}
		case pathplan.BranchLayout:
			if !reflect.DeepEqual(facts.Statements[choice.Target].Then, document.Plan.Base.Statements[choice.Target].Else) {
				t.Fatal("branch fallback missing")
			}
		case pathplan.RootOrder:
			if !reflect.DeepEqual(facts.Root, choice.Options[1].Order) {
				t.Fatal("root fallback missing")
			}
		}
	}
	after, _ := json.Marshal(document.Plan)
	if string(before) != string(after) || string(original) == string(after) || prepared.Fallback() == nil {
		t.Fatal("fact snapshot mutated or double-applied search plan")
	}
}

func TestTypedPathCompilerContextNativeCallsAndLegacyIsolation(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		source, document := conditionalPathFixture(t, language)
		result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "ConditionalAssign", document,
			writePathContextContractModel(t))
		if err != nil {
			t.Fatal(err)
		}
		p := result.Report.BodyPaths
		if p.ModelContext == nil || p.ModelContext.Status != "ENCODED" || !p.SourceBaseMatched ||
			p.Search.Selection.ModelCalls != 6 || p.FunctionalCompleteness != 100 || p.Timing.ContextPrepareMS <= 0 ||
			p.ModelContext.SourceSemanticSHA != p.SourceBinding.SourceSemanticDigest || result.Report.RepositoryWrites != 0 {
			t.Fatalf("native context receipt: %+v", p)
		}
	}
	source, document := typedPathFixture(t)
	for _, model := range []string{"", writeTypedPathContractModel(t, false)} {
		result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, model)
		if err != nil || result.Report.BodyPaths.ModelContext != nil {
			t.Fatal("legacy input changed")
		}
	}
}

func TestTypedPathCompilerContextDeclinesWithoutLosingPartialBody(t *testing.T) {
	for _, intent := range []string{strings.Repeat("가", 150), "caller prefix intent: "} {
		source, document := conditionalPathFixture(t, "en")
		document.Plan.Decisions[0].Intent = intent
		document.TestCases[6].Expected = 999
		baseline, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "ConditionalAssign", document, "")
		if err != nil {
			t.Fatal(err)
		}
		document.Seed = "explicit-seed"
		result, err := GenerateWithTypedPathUnfixedFeedback(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, writePathContextContractModel(t), 8, 2, nil)
		if err != nil {
			t.Fatal(err)
		}
		p := result.Report.BodyPaths
		if result.Source != baseline.Source || p.ModelContext == nil || p.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" ||
			!p.ModelContext.SeedSkipped || !p.ModelContext.FeedbackSkipped || p.FeedbackUnfixed || len(p.Feedback) != 0 ||
			p.Search.Selection.ModelCalls != 0 || p.FunctionalCompleteness != 600.0/7 || p.Search.Evaluated != 64 ||
			p.ModelContext.RankedPlanSHA != "" || p.ModelContext.Inputs[0].InputSHA != "" {
			t.Fatalf("representation decline lost continuation: %+v", p)
		}
		if intent[0] != 'c' && p.ModelContext.Inputs[0].Bytes <= decision.InputMaxBytes {
			t.Fatal("attempted full input size lost")
		}
	}
}

func TestTypedPathCompilerContextBoundsAndRetainedConcurrency(t *testing.T) {
	var buffer pathContextBuffer
	buffer.integer(-9223372036854775808)
	buffer.add(strings.Repeat("x", decision.InputMaxBytes))
	buffer.add("suffix")
	if !buffer.bound || buffer.wanted != 20+decision.InputMaxBytes+6 {
		t.Fatal("bound accounting wrong")
	}
	source, document := typedPathFixture(t)
	generator, err := NewTypedPathGenerator(writePathContextContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	first, err := generator.Generate(context.Background(), "fixture.gooo", source, "Combined", document, TypedPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	expected := first.Report.BodyPaths.ModelContext.Inputs[0].InputSHA
	first.Report.BodyPaths.ModelContext.Inputs[0].InputSHA = "caller mutation"
	bad := []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))
	_, err = generator.Generate(context.Background(), "fixture.gooo", bad, "Combined", document, TypedPathOptions{})
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.ModelContext != nil || failure.Receipt.Search.Selection.ModelCalls != 0 {
		t.Fatal("prediction/context ran before source binding")
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			result, err := generator.Generate(context.Background(), "fixture.gooo", source, "Combined", document, TypedPathOptions{})
			if err != nil || result.Report.BodyPaths.ModelContext.Inputs[0].InputSHA != expected ||
				result.Source != first.Source || result.Report.BodyPaths.Search.Selection.ModelCalls != 3 {
				t.Error("retained request shared mutable context")
			}
		})
	}
	group.Wait()
}
