package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestPathObservationCLIProducesSourceFromDeclaredOracle(t *testing.T) {
	read := func(name string) []byte {
		raw, err := os.ReadFile("../../examples/body-codegen/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	reader := mapSourceReader{"body.gooo": read("path-observation.gooo.fixture"), "plan.json": read("path-observation-plan.json"), "observe.json": read("path-observation-inputs.json")}
	args := []string{"--json", "--path-plan", "plan.json", "--path-observation", "observe.json", "--activity", "Probe", "body.gooo"}
	var out, stderr bytes.Buffer
	code := runBodyCodegen(args, reader, &out, &stderr)
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(out.String(), err)
	}
	if code != exitOK || stderr.Len() != 0 || !strings.Contains(result.Source, "return (2 - input)") ||
		result.Report.BodyPaths.Observation.Status != "ONE_SURVIVING_CANDIDATE" {
		t.Fatal(code, out.String(), stderr.String())
	}
	for _, flags := range [][]string{{"--path-observation", "observe.json"},
		{"--path-plan", "plan.json", "--path-observation", "observe.json", "--path-observation", "observe.json"}} {
		out.Reset()
		stderr.Reset()
		if runBodyCodegen(append(flags, "--activity", "Probe", "body.gooo"), reader, &out, &stderr) != exitUsage {
			t.Fatal("invalid mode")
		}
	}
}

func TestPathObservationStrictFiniteRequest(t *testing.T) {
	valid := `{"schema":"gooo/path-observation-request/v1","inputs":[3],"max_candidates":2,"max_rounds":1}`
	for _, raw := range []string{valid + "{}", strings.Replace(valid, "[3]", "[null]", 1),
		strings.Replace(valid, "[3]", "[3.5]", 1), strings.Replace(valid, "[3]", "[]", 1),
		strings.Replace(valid, `"max_rounds":1`, `"max_rounds":9`, 1),
		strings.Replace(valid, `"max_rounds":1`, `"max_rounds":1,"max_rounds":2`, 1),
		strings.Replace(valid, `"max_rounds":1`, `"max_rounds":1,"expected":-1`, 1)} {
		if _, err := decodePathObservation([]byte(raw)); err == nil {
			t.Fatal("invalid request accepted", raw)
		}
	}
	if _, err := decodePathObservation([]byte(valid)); err != nil {
		t.Fatal(err)
	}
}
