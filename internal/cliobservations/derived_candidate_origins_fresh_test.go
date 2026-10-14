package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestGoooDerivedOrigins26903(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	dir := t.TempDir()
	for _, name := range []string{"source.gooo", "inspect.gooo", "gooo.workspace.json"} {
		raw, err := os.ReadFile(filepath.Join(root, "examples/derived-candidate-origins", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(dir, "source.gooo")
	var out, diag bytes.Buffer
	if code := run([]string{"body-origins", "--activity", "Choose", "--fill-candidates", file}, &out, &diag); code != exitOK {
		t.Fatal("source candidate list", code, diag.String())
	}
	var list bodycodegen.SourceFillOriginCandidates
	if json.Unmarshal(out.Bytes(), &list) != nil || len(list.Candidates) != 4 || list.Generation == nil || list.Generation.AssignmentSpaceSize != 6 || list.Generation.AssignmentsOmitted != 2 {
		t.Fatal("retained assignment population", out.String())
	}
	id := list.Candidates[2].CandidateID
	query := []string{"body-origins", "--activity", "Choose", "--fill-candidate", id, "--inputs-for", "origin/derived-candidates:Inspect", file}
	out.Reset()
	diag.Reset()
	if code := run(query, &out, &diag); code != exitOK {
		t.Fatal("derived rows", code, diag.String())
	}
	input := filepath.Join(dir, "derived26903.json")
	if err := os.WriteFile(input, out.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"package", "execute", "--json", "--inputs", input, filepath.Join(dir, "gooo.workspace.json")}, &out, &diag); code != exitOK {
		t.Fatal("Gooo derived interpretation", code, diag.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if json.Unmarshal(out.Bytes(), &receipt) != nil || receipt.Result == nil || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 || len(receipt.Result.Runtime.Traces) != 2 {
		t.Fatal("derived input-only scope", out.String())
	}
	for i, trace := range receipt.Result.Runtime.Traces {
		var actual map[string]string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual["category"] != "DERIVED_ASSIGNMENT" || actual["candidate_id"] != id || actual["expression"] != list.Candidates[2].HoleFills[i].Expression || actual["grammar"] == "" {
			t.Fatal("grammar was interpreted as a direct literal or expression changed", actual)
		}
	}
	if target := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); target != "" {
		f, err := os.Create(filepath.Join(target, "derived26903.json.gz"))
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
	out.Reset()
	diag.Reset()
	if code := run([]string{"body-origins", "--activity", "Choose", "--fill-candidates", "--inputs-for", "origin/derived-candidates:Inspect", file}, &out, &diag); code == exitOK || !strings.Contains(diag.String(), "one assignment") {
		t.Fatal("whole list silently exported as one assignment", code, diag.String())
	}
	t.Log("derived26903:actual CLI lists4 of6 assignments;2 derived grammar/hole/use rows reach Gooo and retain expressions;finite0/model0;candidate producer not executed")
}

func TestDeclaredOriginWire26903(t *testing.T) {
	source := []byte(`package legacy26903
namespace legacy26903
entity Integer id "urn:gooo:type:integer"
activity Keep(Integer)->Integer id "urn:legacy26903:keep" computes "return __GOOO_BODY_HOLE_value__" assembling {
 source_fill intent "Read a direct declaration." {
  hole "value"
  candidate "keep26903" { fill "value" "input" }
  candidate "next26903" { fill "value" "input + 26903" }
 }
 case "71" -> "71"
}
`)
	file := filepath.Join(t.TempDir(), "直接26903.gooo")
	if err := os.WriteFile(file, source, 0600); err != nil {
		t.Fatal(err)
	}
	legacy, err := bodycodegen.InspectSourceFillCandidateOrigins(file, source, "Keep", "keep26903")
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := run([]string{"body-origins", "--activity", "Keep", "--fill-candidate", "keep26903", file}, &out, &diag); code != exitOK || !bytes.Equal(out.Bytes(), append(want, '\n')) {
		t.Fatal("legacy v1 declared candidate wire changed", code, diag.String(), out.String())
	}
	t.Log("legacy26903:distinct explicit source keeps exact v1 JSON bytes through the new CLI dispatcher;model0")
}
