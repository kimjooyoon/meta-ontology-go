package bodycodegen

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func candidateBoundarySource25037(count int) []byte {
	var b strings.Builder
	b.WriteString("package candidate25037\nnamespace candidate25037\nentity Integer id \"urn:candidate25037:integer\"\n")
	b.WriteString("activity Choose(Integer)->Integer computes \"return __GOOO_BODY_HOLE_h0__\" assembling {\n source_fill intent \"Inspect source boundaries.\" {\n")
	for i := range count {
		fmt.Fprintf(&b, "hole \"h%d\"\n", i)
	}
	b.WriteString(`candidate "\u0062ad25037" {` + "\n")
	for i := range count {
		fmt.Fprintf(&b, "fill \"h%d\" %s\n", i, strconv.Quote(`"메모\\t25037"`))
	}
	b.WriteString("}\ncandidate \"next25037\" {\n")
	for i := range count {
		fmt.Fprintf(&b, "fill \"h%d\" \"0\"\n", i)
	}
	b.WriteString("}\n}\ncase \"41\" -> \"51\"\n}\n")
	return []byte(b.String())
}

func TestSourceCandidateBoundaries25037(t *testing.T) {
	source := candidateBoundarySource25037(8)
	r, err := InspectSourceFillCandidateOrigins("경계25037.gooo", source, "Choose", "bad25037")
	if err != nil || len(r.Rows) != 8 || r.UnresolvedRows != 7 || r.Status != "UNRESOLVED" || r.Rows[0].TemplateSource == nil {
		t.Fatal("eight-hole boundary", r, err)
	}
	if string(source[r.CandidateSource.Start:r.CandidateSource.End]) != `\u0062ad25037` {
		t.Fatal("escaped candidate identity range", r.CandidateSource)
	}
	for _, row := range r.Rows {
		raw := string(source[row.ExpressionSource.Start:row.ExpressionSource.End])
		decoded, err := strconv.Unquote(`"` + raw + `"`)
		if err != nil || decoded != row.Expression || !strings.Contains(raw, `\"`) {
			t.Fatal("escaped expression lost raw byte origin", raw, decoded, row, err)
		}
	}
	if _, err := SourceFillCandidateOriginInputs(r, "origin/boundary25037:Inspect"); err != nil {
		t.Fatal("eight primitive rows", err)
	}
	if _, err := InspectSourceFillCandidateOrigins("too-many25037.gooo", candidateBoundarySource25037(9), "Choose", "bad25037"); err == nil {
		t.Fatal("ninth hole silently dropped")
	}
	grammar := []byte(`package grammar25037
namespace grammar25037
entity Integer id "urn:grammar25037:integer"
activity Choose(Integer)->Integer computes "return __GOOO_BODY_HOLE_value__" assembling {
 source_fill intent "Inspect grammar provenance." {
  hole "value"
  derive grammar "integer-offset-constant/v1" max_expressions "3" max_candidates "3"
 }
 case "41" -> "51"
}
`)
	if _, err := InspectSourceFillCandidateOrigins("grammar25037.gooo", grammar, "Choose", "derived25037"); err == nil || !strings.Contains(err.Error(), "generated grammar") {
		t.Fatal("grammar candidate invented a direct assignment location", err)
	}
	if _, err := InspectSourceFillCandidateOrigins("large25037.gooo", []byte(strings.Repeat(" ", (128<<10)+1)), "Choose", "bad25037"); err == nil {
		t.Fatal("source size boundary ignored")
	}
	t.Log("boundary25037:8 rows retained with7 explicit unresolved uses;9 holes/derived grammar/oversized source refused;escaped candidate and quoted Unicode expression ranges exact;model0")
}
