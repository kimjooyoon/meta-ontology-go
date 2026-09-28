package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestMissingAndStaleInputsDoNotProduceCandidates(t *testing.T) {
	data, request := fixture(t)
	delete(data, request.Case.Path)
	_, err := Compile(data, request)
	requireFailure(t, err, "UNKNOWN", "DIRECT_MISSING")
	data, request = fixture(t)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	data[request.Case.Path].Data = append(data[request.Case.Path].Data, '\n')
	candidate, err := plan.Generate(data)
	requireFailure(t, err, "UNKNOWN", "STALE")
	if candidate.Emitted != 0 || candidate.ApplyAuthorized {
		t.Fatal("stale input emitted an accepted candidate")
	}
}

func TestDuplicateIdentityAndPathAreRefuted(t *testing.T) {
	for _, duplicate := range []string{"id", "path"} {
		data, request := fixture(t)
		var registry languagesyntax.Registry
		if err := json.Unmarshal(data[corpusPath].Data, &registry); err != nil {
			t.Fatal(err)
		}
		item := request.Case
		if duplicate == "id" {
			item.Path = "examples/different-source/main.gooo"
		} else {
			item.ID = "different-case"
		}
		registry.Cases = append(registry.Cases, item)
		raw, err := json.Marshal(registry)
		if err != nil {
			t.Fatal(err)
		}
		data[corpusPath] = &fstest.MapFile{Data: raw}
		pin(t, data, &request)
		_, err = Compile(data, request)
		requireFailure(t, err, "REFUTED", "")
	}
}

func TestCandidateCannotDropMembersRewriteHistoryOrAcquireAuthority(t *testing.T) {
	for _, mutation := range []string{"missing", "duplicate", "history", "content", "authority", "lowered", "role", "binding"} {
		data, request := fixture(t)
		plan, err := Compile(data, request)
		if err != nil {
			t.Fatal(err)
		}
		candidate, err := plan.Generate(data)
		if err != nil {
			t.Fatal(err)
		}
		switch mutation {
		case "missing":
			candidate.Members = candidate.Members[:len(candidate.Members)-1]
		case "duplicate":
			candidate.Members[len(candidate.Members)-1] = candidate.Members[0]
		case "history":
			for index := range candidate.Members {
				if candidate.Members[index].Path == denominatorPath(request.BaseVersion+1) {
					candidate.Members[index].Path = denominatorPath(request.BaseVersion)
				}
			}
		case "content":
			candidate.Members[0].Content = []byte("{}")
		case "authority":
			candidate.ApplyAuthorized = true
		case "lowered":
			candidate.Required--
		case "role":
			candidate.Artifacts = candidate.Artifacts[:8]
		case "binding":
			candidate.Members[0].ActivityIDs = nil
		}
		requireFailure(t, plan.ValidateCandidate(data, candidate), "REFUTED", "")
	}
}

func TestAmbiguousJSONAndForgedPlanAreRejected(t *testing.T) {
	if _, err := DecodeRequest([]byte(`{"base_version":30,"base_version":31}`)); err == nil {
		t.Fatal("duplicate JSON input was accepted")
	}
	if _, err := DecodeRequest([]byte(`{"case":{"id":"a","id":"b"}}`)); err == nil {
		t.Fatal("duplicate nested JSON input was accepted")
	}
	data, request := fixture(t)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	data[denominatorPath(request.BaseVersion)].Data = bytes.Replace(
		data[denominatorPath(request.BaseVersion)].Data, []byte("\"syntax\""), []byte("\"changed\""), 1)
	_, err = plan.Generate(data)
	requireFailure(t, err, "UNKNOWN", "STALE")
	_, err = (Plan{}).Generate(data)
	requireFailure(t, err, "REFUTED", "")
}
