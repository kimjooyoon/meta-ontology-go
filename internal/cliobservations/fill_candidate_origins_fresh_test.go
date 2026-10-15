package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestGoooCandidateOrigins24731(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	dir := t.TempDir()
	for _, name := range []string{"source.gooo", "inspect.gooo", "gooo.workspace.json"} {
		raw, err := os.ReadFile(filepath.Join(root, "examples/fill-candidate-origins", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(dir, "source.gooo")
	var out, diag bytes.Buffer
	query := []string{"body-origins", "--activity", "Choose", "--fill-candidate", "wrong_type", source}
	if code := run(query, &out, &diag); code != exitOK {
		t.Fatal("inspect unselected type-invalid expression", code, diag.String(), out.String())
	}
	var report bodycodegen.SourceFillCandidateOrigins
	if json.Unmarshal(out.Bytes(), &report) != nil || report.Status != "RESOLVED" || len(report.Rows) != 2 || report.Rows[0].Expression != "true" {
		t.Fatal("declared origins", out.String())
	}
	input := filepath.Join(dir, "candidate24731.json")
	query = []string{"body-origins", "--activity", "Choose", "--fill-candidate", "wrong_type", "--inputs-for", "origin/fill-candidates:Inspect", source}
	out.Reset()
	diag.Reset()
	if code := run(query, &out, &diag); code != exitOK {
		t.Fatal("export primitive candidate rows", code, diag.String())
	}
	if err := os.WriteFile(input, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	checkCandidateOriginReader24731(t, dir, input, "declared24731", []string{"DECLARED_ASSIGNMENT", "DECLARED_ASSIGNMENT"})
	original, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	ambiguous := strings.Replace(string(original), "return __GOOO_BODY_HOLE_value__", "return __GOOO_BODY_HOLE_value__ + __GOOO_BODY_HOLE_value__", 1)
	if ambiguous == string(original) {
		t.Fatal("fixture use was not changed")
	}
	if err := os.WriteFile(source, []byte(ambiguous), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	if code := run(query, &out, &diag); code == exitOK || out.Len() == 0 {
		t.Fatal("ambiguous use gained successful mapping or lost its observation", code, diag.String())
	}
	if err := os.WriteFile(input, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	checkCandidateOriginReader24731(t, dir, input, "ambiguous24731", []string{"UNRESOLVED_USE", "DECLARED_ASSIGNMENT"})
	out.Reset()
	diag.Reset()
	bad := []string{"body-origins", "--activity", "Choose", "--fill-candidate", "wrong_type", "--field", "urn:field24731", "--inputs-for", "origin/fill-candidates:Inspect", source}
	if code := run(bad, &out, &diag); code == exitOK || !strings.Contains(diag.String(), "typed value origins") {
		t.Fatal("typed graph selector was accepted for candidate locations", code, diag.String())
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"help", "body-origins"}, &out, &diag); code != exitOK || !strings.Contains(out.String(), "--fill-candidate") {
		t.Fatal("candidate help", code, diag.String())
	}
	t.Log("reader24731:actual CLI maps explicit unselected candidate2 rows;Gooo classifies mapped2 and ambiguous1+mapped1;finite0,model0;candidate producer not executed")
}

func checkCandidateOriginReader24731(t *testing.T, dir, input, label string, categories []string) {
	t.Helper()
	var out, diag bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--inputs", input, filepath.Join(dir, "gooo.workspace.json")}, &out, &diag)
	if target := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); target != "" {
		f, err := os.Create(filepath.Join(target, label+".json.gz"))
		if err != nil {
			t.Fatal(err)
		}
		z := gzip.NewWriter(f)
		if _, err := z.Write(out.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if code != exitOK {
		t.Fatal("Gooo candidate reader", code, diag.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if json.Unmarshal(out.Bytes(), &receipt) != nil || receipt.Result == nil || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 || len(receipt.Result.Runtime.Traces) != len(categories) {
		t.Fatal("input-only reader scope", out.String())
	}
	for i, trace := range receipt.Result.Runtime.Traces {
		var actual struct{ Category, CandidateID, HoleID, Expression, Reason string }
		if len(trace.Deliveries) != 1 {
			t.Fatal("reader delivery count")
		}
		var values map[string]string
		if json.Unmarshal(trace.Deliveries[0].Actual, &values) != nil {
			t.Fatal("reader actual output")
		}
		actual.Category, actual.CandidateID, actual.HoleID = values["category"], values["candidate_id"], values["hole_id"]
		actual.Expression, actual.Reason = values["expression"], values["reason"]
		if actual.Category != categories[i] || actual.CandidateID != "wrong_type" || (i == 0 && (actual.HoleID != "value" || actual.Expression != "true")) || (i == 1 && actual.HoleID != "bias") {
			t.Fatal("Gooo interpretation differs from the declared assignment", actual)
		}
		if actual.Category == "UNRESOLVED_USE" && actual.Reason != "TEMPLATE_HOLE_NOT_UNIQUE" {
			t.Fatal("unresolved reason lost", actual)
		}
	}
}

func TestGoooEightCandidateOrigins25717(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	dir := t.TempDir()
	for _, name := range []string{"inspect.gooo", "gooo.workspace.json"} {
		raw, err := os.ReadFile(filepath.Join(root, "examples/fill-candidate-origins", name))
		if err != nil || os.WriteFile(filepath.Join(dir, name), raw, 0600) != nil {
			t.Fatal("reader25717 source", err)
		}
	}
	var source strings.Builder
	source.WriteString("package complete25717\nnamespace complete25717\nentity Integer id \"urn:gooo:type:integer\"\nactivity Choose(Integer)->Integer id \"urn:complete25717:choose\" computes \"return __GOOO_BODY_HOLE_w0__\" assembling {\nsource_fill intent \"Read every declared hole.\" {\n")
	for i := range 8 {
		fmt.Fprintf(&source, "hole \"w%d\"\n", i)
	}
	for _, candidate := range []string{"bad25717", "other25717"} {
		fmt.Fprintf(&source, "candidate %q {\n", candidate)
		for i := range 8 {
			fmt.Fprintf(&source, "fill \"w%d\" \"input + 25717\"\n", i)
		}
		source.WriteString("}\n")
	}
	source.WriteString("}\ncase \"53\" -> \"59\"\n}\n")
	file, input := filepath.Join(dir, "전체25717.gooo"), filepath.Join(dir, "inputs25717.json")
	if err := os.WriteFile(file, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := run([]string{"body-origins", "--activity", "Choose", "--fill-candidate", "bad25717", "--inputs-for", "origin/fill-candidates:Inspect", file}, &out, &diag); code == exitOK || out.Len() == 0 {
		t.Fatal("seven missing uses must retain unresolved output", code, diag.String())
	}
	var rows struct {
		Inputs []map[string]bodycodegen.SourceFillCandidateInput `json:"inputs"`
	}
	if json.Unmarshal(out.Bytes(), &rows) != nil || len(rows.Inputs) != 8 || !rows.Inputs[0]["origin/fill-candidates:Inspect"].HasUse {
		t.Fatal("eight primitive rows were not preserved", out.String())
	}
	if err := os.WriteFile(input, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"package", "execute", "--json", "--inputs", input, filepath.Join(dir, "gooo.workspace.json")}, &out, &diag); code != exitOK {
		t.Fatal("Gooo eight-row reader", code, diag.String())
	}
	var receipt packageExecutionReceipt
	if json.Unmarshal(out.Bytes(), &receipt) != nil || receipt.Result == nil || len(receipt.Result.Runtime.Traces) != 8 || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("eight-row input-only scope")
	}
	for i, trace := range receipt.Result.Runtime.Traces {
		var actual map[string]string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("eight-row output")
		}
		category := "UNRESOLVED_USE"
		if i == 0 {
			category = "DECLARED_ASSIGNMENT"
		}
		if actual["category"] != category || actual["candidate_id"] != "bad25717" || actual["hole_id"] != fmt.Sprintf("w%d", i) {
			t.Fatal("Gooo eight-row interpretation", i, actual)
		}
	}
	if target := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); target != "" {
		if err := os.WriteFile(filepath.Join(target, "complete25717.json.gz"), gzipBytes25717(out.Bytes()), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("reader25717:all8 declared holes transported;Gooo mapped1 and unresolved7;finite0/model0;source candidate not executed")
}

func gzipBytes25717(raw []byte) []byte {
	var out bytes.Buffer
	z := gzip.NewWriter(&out)
	_, err := z.Write(raw)
	if err != nil {
		panic(err)
	}
	if err := z.Close(); err != nil {
		panic(err)
	}
	return out.Bytes()
}
