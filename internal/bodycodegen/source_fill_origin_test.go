package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceFillOriginFragments3371(t *testing.T) {
	source := []byte(`package fragment3371
namespace fragment3371
entity Integer id "urn:gooo:type:integer"
activity Adjust(Integer) -> Integer id "urn:fragment3371:adjust" computes ` + "`" + `let bonus = __GOOO_BODY_HOLE_bonus__
if __GOOO_BODY_HOLE_gate__ { return input + bonus }; return input - 5` + "`" + ` assembling {
 source_fill intent "작은 입력에 보너스\n규칙을 적용한다" {
  hole "bonus"
  hole "gate"
  candidate "primary3371" { fill "bonus" "70000000000" fill "gate" "input < 137" }
  candidate "twin3371" { fill "bonus" "70000000000" fill "gate" "input < 137" }
 }
 case "4" -> "70000000004"
 case "140" -> "135"
}
`)
	prior := generateOriginFill3371(t, "fragment3371.gooo", source, "Adjust")
	origin := prior.Report.BodyFill.SourceOrigin
	if origin == nil || origin.SourceSHA256 != digest(source) || origin.CandidateSource == nil ||
		string(source[origin.CandidateSource.Start:origin.CandidateSource.End]) != prior.Report.BodyFill.SelectedCandidateID ||
		string(source[origin.IntentSource.Start:origin.IntentSource.End]) != `작은 입력에 보너스\n규칙을 적용한다` {
		t.Fatal("selected candidate or authored intent missing", origin)
	}
	checkOriginFragments3371(t, source, prior)
	if _, err := RealizeSourceAssembly(context.Background(), "fragment3371.gooo", source, prior); err != nil {
		t.Fatal("fresh origin realization", err)
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	var changed Result
	if err := json.Unmarshal(encoded, &changed); err != nil {
		t.Fatal(err)
	}
	changed.Report.BodyFill.SourceOrigin.IntentSource.Start++
	if _, err := RealizeSourceAssembly(context.Background(), "fragment3371.gooo", source, changed); err == nil {
		t.Fatal("changed origin was not noticed")
	}
	saveFillOrigin3371(t, "fragment3371", source, prior)
	t.Log("fragment3371: identical candidate expressions retain selected ID; two filled fragments and template fragments located; escaped intent range exact; changed origin rejected; model0")
}

func checkOriginFragments3371(t *testing.T, source []byte, prior Result) {
	t.Helper()
	origin := prior.Report.BodyFill.SourceOrigin
	filled, cursor := 0, 0
	for _, fragment := range origin.Fragments {
		if fragment.SelectedDecoded.Start != cursor || fragment.SelectedDecoded.End <= cursor {
			t.Fatal("fragment body layout has a gap", fragment)
		}
		cursor = fragment.SelectedDecoded.End
		if fragment.Kind == "declared_candidate" {
			filled++
			if fragment.HoleDeclaration == nil || fragment.ExpressionSource == nil ||
				string(source[fragment.ExpressionSource.Start:fragment.ExpressionSource.End]) != fragment.Expression {
				t.Fatal("declared fragment source missing", fragment)
			}
		}
	}
	if filled != 2 || len(prior.Report.BodyConditions) != 1 {
		t.Fatal("fragment or condition count", filled)
	}
	link := origin.ConditionOrigin(prior.Report.BodyConditions[0])
	if link == nil || len(link.Fragments) != 1 || link.Fragments[0].HoleID != "gate" {
		t.Fatal("selected condition did not lead back to its gate hole", link)
	}
	link.Fragments[0].ExpressionSource.Start++
	if again := origin.ConditionOrigin(prior.Report.BodyConditions[0]); again.Fragments[0].ExpressionSource.Start == link.Fragments[0].ExpressionSource.Start {
		t.Fatal("condition view aliases original fragment memory")
	}
}

func TestSourceFillDerivedOrigin3491(t *testing.T) {
	source := []byte(`package derived3491
namespace derived3491
entity Integer id "urn:gooo:type:integer"
activity Compose(Integer) -> Integer id "urn:derived3491:compose" computes "let left = __GOOO_BODY_HOLE_left__; return left + __GOOO_BODY_HOLE_right__" assembling {
 source_fill intent "후보를 조립한 문법을 기억한다" {
  hole "left"
  hole "right"
  derive grammar "integer-offset-constant/v1" max_expressions "4" max_candidates "4"
 }
 case "5" -> "16"
 case "10" -> "21"
}
`)
	prior := generateOriginFill3371(t, "derived3491.gooo", source, "Compose")
	origin := prior.Report.BodyFill.SourceOrigin
	if origin == nil || origin.CandidateSource != nil {
		t.Fatal("derived candidate claimed authored position")
	}
	derived := 0
	for _, fragment := range origin.Fragments {
		if fragment.Kind != "derived_grammar" {
			continue
		}
		derived++
		if fragment.ExpressionSource != nil || fragment.GrammarSource == nil ||
			string(source[fragment.GrammarSource.Start:fragment.GrammarSource.End]) != fragment.Grammar {
			t.Fatal("derived expression did not cite its grammar", fragment)
		}
	}
	if derived != 2 || prior.Report.BodyFill.CandidateGeneration == nil {
		t.Fatal("derived origin count", derived)
	}
	saveFillOrigin3371(t, "derived3491", source, prior)
	t.Log("derived3491: two expressions cite grammar and assembly input; no fabricated expression/candidate source range; generation scoring retained separately; model0")
}

func generateOriginFill3371(t *testing.T, filename string, source []byte, activity string) Result {
	t.Helper()
	spec, err := SourceAssembly(context.Background(), filename, source, activity)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithSourceIRBodyFill(context.Background(), filename, source, activity, spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prior.GoooSource, "source_fill") || prior.Report.BodyFill.LocalModelPredictions != 0 {
		t.Fatal("assembly removal or deterministic selection differs")
	}
	return prior
}

func saveFillOrigin3371(t *testing.T, name string, source []byte, prior Result) {
	t.Helper()
	output := os.Getenv("GOOO_FILL_ORIGIN_OUTPUT")
	if output == "" {
		return
	}
	raw, err := json.MarshalIndent(map[string]any{"source": string(source), "result": prior}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, name+".json"), append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}
