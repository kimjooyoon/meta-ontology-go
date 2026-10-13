package toolchainrelease

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const originContinuity3179 = `package dispatch3179
namespace dispatch3179
entity Integer id "urn:gooo:type:integer"
activity Route(Integer) -> Integer id "urn:dispatch3179:route" computes "if input < 83 { return input + 19 }; return input - 23"
`

func TestPackageProgramOriginContinuity3179(t *testing.T) {
	program := freshOriginProgram3179(t)
	current, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), current...)
	legacy := program
	legacy.ConditionOrigins = nil
	old, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if !samePackageProgramSource(old, current) || !samePackageProgramSource(current, current) {
		t.Fatal("fresh source origin or legacy absence was not retained")
	}
	if samePackageProgramSource(current, old) || !bytes.Equal(before, current) {
		t.Fatal("recorded origins discarded or comparison modified input")
	}
	checkOriginMutations3179(t, old, current)
	checkOriginJSON3179(t)
	saveOriginContinuity3179(t, program, old, current)
	t.Log("fresh3179: retained origin and absent origin compatible; recorded claim changes rejected; source unchanged; no native/model calls")
}

func freshOriginProgram3179(t *testing.T) workspaceexecution.Program {
	t.Helper()
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "dispatch3179", Activity: "Route"},
		Packages: []packageruntime.PackageSpec{{Path: "dispatch3179", Name: "dispatch3179",
			Sources: []packageruntime.Source{{Filename: "기원3179.gooo", Content: originContinuity3179}}}}}
	program, err := workspaceexecution.Prepare(manifest)
	if err != nil || len(program.ConditionOrigins) != 1 {
		t.Fatal("fresh source origin preparation", err, len(program.ConditionOrigins))
	}
	origin := program.ConditionOrigins[0].Input
	if origin.Status != "MAPPED" || origin.Filename != "기원3179.gooo" ||
		origin.Expression != "input < 83" || originContinuity3179[origin.Start:origin.End] != origin.Expression {
		t.Fatal("original file origin lost", origin)
	}
	return program
}

func checkOriginMutations3179(t *testing.T, old, current json.RawMessage) {
	t.Helper()
	for _, change := range []struct{ key, value string }{
		{"condition_origins", "null"}, {"condition_origins", "[]"},
		{"condition_origins", `[{"input_source":{"filename":"다른3179.gooo"}}]`},
		{"lowered_gooo_source", `"changed source"`}, {"entry", `{"activity":"Other"}`},
		{"workspace_image", `{}`}, {"activities", `[]`}, {"scope", `"other"`},
		{"unknown_origin_claim", `"new"`},
	} {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(current, &fields); err != nil {
			t.Fatal(err)
		}
		fields[change.key] = json.RawMessage(change.value)
		mutated, err := json.Marshal(fields)
		if err != nil || samePackageProgramSource(mutated, current) {
			t.Fatalf("recorded %s change accepted: %v", change.key, err)
		}
		if change.key != "condition_origins" && samePackageProgramSource(old, mutated) {
			t.Fatalf("absent origins hid current %s change", change.key)
		}
	}
}

func checkOriginJSON3179(t *testing.T) {
	t.Helper()
	for _, row := range []struct {
		recorded, current string
		want              bool
	}{
		{`{"n":9007199254740997}`, `{"n":9007199254740997,"condition_origins":[]}`, true},
		{`{"n":9007199254740997}`, `{"n":9007199254740996,"condition_origins":[]}`, false},
		{`{"n":3179}`, `{"n":3179}`, true},
		{`null`, `{}`, false}, {`{}`, `null`, false}, {`[]`, `{}`, false},
		{`{}`, `[]`, false}, {`{} {}`, `{}`, false}, {`{}`, `{} {}`, false},
	} {
		if got := samePackageProgramSource([]byte(row.recorded), []byte(row.current)); got != row.want {
			t.Fatalf("JSON %s / %s: got %v, want %v", row.recorded, row.current, got, row.want)
		}
	}
}

func saveOriginContinuity3179(t *testing.T, program workspaceexecution.Program, old, current json.RawMessage) {
	t.Helper()
	output := os.Getenv("GOOO_ORIGIN_CONTINUITY_OUTPUT")
	if output == "" {
		return
	}
	value := map[string]any{"source": originContinuity3179, "program": program,
		"recorded_without_origins": old, "prepared_with_origins": current,
		"origin_absence_preserved": true, "recorded_origins_compared": true,
		"native_executions": 0, "model_calls": 0, "historical_fixtures_replayed": 0}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, "continuity3179.json"), append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}
