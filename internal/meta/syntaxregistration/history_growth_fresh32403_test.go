package syntaxregistration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSourceOwnedHistoryGrowth32403(t *testing.T) {
	data, request := fixture(t)
	delete(data, request.Case.Path)
	// Bind the complete prepared source view rather than the public predecessor.
	prepared := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	var err error
	if prepared != "" {
		err = filepath.WalkDir(prepared, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(prepared, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if _, exists := data[rel]; exists {
				raw, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				data[rel] = &fstest.MapFile{Data: raw}
			}
			return nil
		})
	}
	if err != nil {
		t.Fatal(err)
	}
	baseVersion := request.BaseVersion
	baseTotal, _, baseCapability, err := corpusTotals(data[corpusPath].Data)
	if err != nil {
		t.Fatal(err)
	}
	initial := digestValue(data)
	initialCorpus := append([]byte(nil), data[corpusPath].Data...)
	var first Candidate
	for step := range 2 {
		request.BaseVersion = baseVersion + step
		request.Case.ID = []string{"historyorigin32403first", "historyorigin32403next"}[step]
		request.Case.Path = "examples/" + request.Case.ID + "/main.gooo"
		data[request.Case.Path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(fixtureSource, "registrationfixture", request.Case.ID))}
		pin(t, data, &request)
		plan, err := Compile(data, request)
		if err != nil {
			t.Fatal(err)
		}
		candidate, err := plan.Generate(data)
		if err != nil {
			t.Fatal(err)
		}
		if err := plan.ValidateCandidate(data, candidate); err != nil {
			t.Fatal(err)
		}
		if len(candidate.Artifacts) != 9 || candidate.ApplyAuthorized || candidate.RepositoryWrites != 0 {
			t.Fatal("history growth changed roles or authority")
		}
		for _, member := range candidate.Members {
			data[member.Path] = &fstest.MapFile{Data: member.Content}
		}
		inputs := map[string][]byte{}
		for name, file := range data {
			inputs[name] = file.Data
		}
		source, err := parseSourceUnits(inputs, closureRoot, "verticalsliceclosureshadow", false)
		if err != nil {
			t.Fatal(err)
		}
		freshAssertHistoryTarget32403(t, source, baseVersion, baseCapability)
		if step == 1 {
			freshAssertHistoryTarget32403(t, source, baseVersion+1, baseCapability+1)
		}
		var previous, current denominator
		if err := json.Unmarshal(data[denominatorPath(baseVersion+step)].Data, &previous); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data[denominatorPath(baseVersion+1+step)].Data, &current); err != nil {
			t.Fatal(err)
		}
		for index, expected := range previous.Boundaries {
			if index == 0 {
				expected.Target++
			}
			if current.Boundaries[index] != expected {
				t.Fatal("unrelated boundary or link target changed", index)
			}
		}
		if step == 0 {
			first = candidate
		}
		if out := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); out != "" {
			directory := filepath.Join(out, fmt.Sprintf("candidate-v%d", baseVersion+1+step))
			for name, file := range data {
				target := filepath.Join(directory, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, file.Data, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if initial == digestValue(data) {
		t.Fatal("history continuation fixture did not change")
	}
	var baseline, final struct {
		Cases []json.RawMessage `json:"cases"`
	}
	if err := json.Unmarshal(initialCorpus, &baseline); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data[corpusPath].Data, &final); err != nil {
		t.Fatal(err)
	}
	if len(final.Cases) != baseTotal+2 || !reflect.DeepEqual(final.Cases[:baseTotal], baseline.Cases) {
		t.Fatal("original110 source cases changed")
	}
	if out := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); out != "" {
		raw, err := json.MarshalIndent(map[string]any{"first_candidate": first, "final_version": baseVersion + 2, "first_target": baseCapability + 1, "final_target": baseCapability + 2, "preserved_baseline_target": baseCapability, "artifact_roles": 9, "model_calls": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "history32403.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func freshAssertHistoryTarget32403(t *testing.T, source *goSource, version, target int) {
	t.Helper()
	history, err := readHistoryOrigins(source)
	if err != nil {
		t.Fatal(err)
	}
	if value, exists := history.targets[version]; !exists || value != target {
		t.Fatal("historical source target missing or changed", version, value)
	}
}
