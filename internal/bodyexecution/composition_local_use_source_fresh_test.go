package bodyexecution

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestGoooLocalUseSource10709(t *testing.T) {
	body := "let 원본=Pick(input);let copy=원본;원본=원본+211;let unused=223;return copy+원본+Pick(input)"
	source := `package use10709
namespace use10709
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:use10709:input" fields {
 field values id "urn:use10709:values" type integer required many
 field reverse id "urn:use10709:reverse" type boolean required one
}
activity Pick(Input) -> Integer id "urn:use10709:pick" computes "if input.reverse {let selected=input.values[1]+227;return selected}else{let selected=input.values[0]-229;return selected}"
activity Build(Input) -> Integer id "urn:use10709:build" computes ` + strconv.QuoteToASCII(body) + "\n"
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][2]string{
		{`{"values":[233,239],"reverse":false}`, `223`},
		{`{"values":[241,251],"reverse":true}`, `1645`},
		{`{"values":[257,257],"reverse":false}`, `295`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(row[0])}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(row[1])}})
	}
	observation := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "local-use-source10709.json", observation)
	composition, err := GenerateCompositionWithOptions(context.Background(), "사용기원10709.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	observation["composition"] = composition
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "사용기원10709.gooo", []byte(source), composition, suite, "")
	observation["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 3 || !runtime.RuntimeReplayed {
		t.Fatal("native results", err, runtime.FinitePassed)
	}
	starts := map[int]bool{}
	escaped, helper := 0, 0
	for _, site := range runtime.LocalUseSites {
		if site.Source == nil || site.Source.Status != "MAPPED" || site.Source.SourceSHA256 != digest([]byte(source)) || site.Source.Expression != site.Name {
			t.Fatal("Gooo use mapping", site)
		}
		if site.Source.Start < 0 || site.Source.End > len(source) || site.Source.End <= site.Source.Start || starts[site.Source.Start] {
			t.Fatal("distinct original use positions", site)
		}
		starts[site.Source.Start] = true
		if site.ProjectionSHA256 != runtime.GeneratedSHA256 || site.Declaration.SourceSHA256 != site.Source.SourceSHA256 || site.Declaration.Start == site.Source.Start {
			t.Fatal("coordinate scopes", site)
		}
		if site.Name == "원본" {
			escaped++
			if !strings.Contains(source[site.Source.Start:site.Source.End], `\u`) {
				t.Fatal("escaped carrier coordinate", site.Source)
			}
		}
		if site.ActivityID == "urn:use10709:pick" {
			helper++
		}
		if site.Name == "unused" {
			t.Fatal("synthetic unused-local read gained source origin")
		}
	}
	if escaped != 3 || helper != 2 || len(runtime.LocalUseSites) != 6 {
		t.Fatal("use-site populations", escaped, helper, len(runtime.LocalUseSites))
	}
	for _, trace := range runtime.Traces {
		if trace.LocalFrames != 3 || len(trace.LocalUses) != 6 {
			t.Fatal("repeated helper frames", trace.LocalFrames, len(trace.LocalUses))
		}
	}
	origins := append([]bodycodegen.BodyLocalOrigin(nil), composition.Steps[0].Generation.Report.BodyLocals...)
	for i := range origins {
		if len(origins[i].Uses) == 0 {
			continue
		}
		origins[i].Uses = append([]bodycodegen.BodyLocalUseOrigin(nil), origins[i].Uses...)
		origins[i].Uses[0].Start++
		if bodycodegen.EqualRetainedLocalOrigins(origins, composition.Steps[0].Generation.Report.BodyLocals) {
			t.Fatal("mutated retained use accepted")
		}
		break
	}
	t.Log("new Gooo native3/3;6 distinct source use sites;escaped Korean uses3;helper sibling sites2;repeated helper invocation frames;synthetic unused read omitted")
}
