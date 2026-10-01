package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestCompilerContextExportEqualsNativeRankingInput(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		source, document := conditionalPathFixture(t, language)
		before, _ := json.Marshal(document)
		exported, err := ExportTypedPathContext(context.Background(), "fixture.gooo", source, "ConditionalAssign", document)
		if err != nil {
			t.Fatal(err)
		}
		generated, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "ConditionalAssign", document,
			writePathContextContractModel(t))
		if err != nil {
			t.Fatal(err)
		}
		ranked := generated.Report.BodyPaths.ModelContext
		if exported.Context.Schema != "gooo/compiler-typed-path-context/v2" || !exported.SourceBinding.Equivalent ||
			exported.Context.RankedPlanSHA != ranked.RankedPlanSHA || !reflect.DeepEqual(exported.Context.Inputs, ranked.Inputs) ||
			exported.Context.MetadataSHA != "" || exported.ModelPredictions != 0 || exported.CandidateTests != 0 ||
			exported.SelectedEmission || exported.RepositoryWrites != 0 || len(exported.Inputs) != 6 {
			t.Fatal("export changed runtime input or executed selection")
		}
		for _, input := range exported.Inputs {
			if input.InputSHA != digest([]byte(input.Text)) || input.Bytes != len(input.Text) ||
				strings.Contains(input.Text, "expected") || !strings.Contains(input.Text, ";basis=source_fallback;") {
				t.Fatal("export hashes/authority differ")
			}
		}
		document.TestCases[0].Expected = 999
		document.Seed = "not-applied-to-export"
		changed, err := ExportTypedPathContext(context.Background(), "fixture.gooo", source, "ConditionalAssign", document)
		if err != nil || !reflect.DeepEqual(exported.Inputs, changed.Inputs) || exported.DocumentSHA256 == changed.DocumentSHA256 ||
			exported.TestSuiteSHA256 == changed.TestSuiteSHA256 {
			t.Fatal("finite target or seed leaked into input")
		}
		document.TestCases[0].Expected = generated.Report.BodyPaths.NativeCases[0].Expected
		document.Seed = ""
		after, _ := json.Marshal(document)
		if string(before) != string(after) {
			t.Fatal("export mutated caller")
		}
	}
}

func TestCompilerContextSourceRelativeReversedFallback(t *testing.T) {
	_, document := conditionalPathFixture(t, "en")
	for i := range document.Plan.Decisions {
		choice := &document.Plan.Decisions[i]
		choice.Fallback = choice.Options[1].Label
	}
	facts := fallbackContextFacts(document.Plan)
	for _, choice := range document.Plan.Decisions {
		text, _, reason := encodePathContext(facts, choice)
		if reason != "" || !strings.Contains(text, ";fallback="+choice.Fallback+";basis=source_fallback;") {
			t.Fatal("missing source fallback basis")
		}
		switch choice.Kind {
		case pathplan.OperandOrder, pathplan.BranchLayout:
			if !strings.Contains(text, choice.Options[0].Label+":reverse_source=true|") ||
				!strings.Contains(text, choice.Options[1].Label+":reverse_source=false|") {
				t.Fatalf("raw reverse flags misrepresented normalized source: %s", text)
			}
		case pathplan.LocalReference, pathplan.AssignmentTarget:
			if !strings.Contains(text, choice.Options[1].Label+":name="+choice.Options[1].Name+"|") {
				t.Fatal("name option lost")
			}
		case pathplan.RootOrder:
			if !strings.Contains(text, choice.Options[1].Label+":order=[") {
				t.Fatal("root order lost")
			}
		}
	}
	prepared, err := document.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(prepared.Fallback().GoooBody())
	source := []byte("package sample\nnamespace sample\nentity Integer id \"sample://entity/integer\"\n" +
		"activity ConditionalAssign(Integer) -> Integer computes " + string(body) + "\n")
	if exported, err := ExportTypedPathContext(context.Background(), "fixture.gooo", source, "ConditionalAssign", document); err != nil || !exported.SourceBinding.Equivalent || len(exported.Inputs) != 6 {
		t.Fatalf("reversed source export failed: %+v %v", exported, err)
	}
}

func TestCompilerContextExportDeclineIsAtomicAndRejectsUnboundSource(t *testing.T) {
	source, document := typedPathFixture(t)
	document.Plan.Decisions[1].Intent = strings.Repeat("가", 150)
	result, err := ExportTypedPathContext(context.Background(), "fixture.gooo", source, "Combined", document)
	if err != nil || result.Context.Status != "DECLINED_TO_DETERMINISTIC" || len(result.Inputs) != 0 ||
		len(result.Context.Inputs) != 2 || result.Context.Inputs[1].Bytes <= 512 || result.Context.Inputs[1].InputSHA != "" {
		t.Fatal("decline exposed a partial/truncated ranking input")
	}
	for _, data := range [][]byte{nil, []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))} {
		_, err := ExportTypedPathContext(context.Background(), "fixture.gooo", data, "Combined", document)
		var failure *BodyPathError
		if !errors.As(err, &failure) || failure.Receipt.ModelContext != nil {
			t.Fatal("export ran without binding original source")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ExportTypedPathContext(ctx, "fixture.gooo", source, "Combined", document); !errors.Is(err, context.Canceled) {
		t.Fatal("export ignored cancellation")
	}
}
