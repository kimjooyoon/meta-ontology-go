package syntaxregistration

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"reflect"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

const fixtureSource = `package registrationfixture
namespace registrationfixture

entity Observation id "gooo://registration-fixture/observation" fields {
    field State id "gooo://registration-fixture/state" type string required one
}
activity Capture(Observation) -> Observation computes "record.forward:v1"
`

func fixture(t *testing.T) (fstest.MapFS, Request) {
	t.Helper()
	repository := os.DirFS("../../..")
	versions, err := fs.Glob(repository, closureRoot+"evidence/denominator-v*.json")
	if err != nil {
		t.Fatal(err)
	}
	version := 0
	for _, path := range versions {
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		var observed denominator
		if json.Unmarshal(raw, &observed) == nil && observed.Version > version {
			version = observed.Version
		}
	}
	request := Request{BaseVersion: version, Toolchain: runtime.Version(),
		Case: languagesyntax.CaseDefinition{ID: "syntax-registration-native-fixture",
			Path: "examples/syntax-registration-native-fixture/main.gooo", Kind: languagesyntax.KindValid,
			ExpectedDecision: languagesyntax.DecisionPass, ProofChoice: "COHERENCE",
			MetaOperation: "replay-language-syntax", Scope: languagesyntax.ScopeLanguageCapability, EntityFields: true}}
	data := fstest.MapFS{request.Case.Path: {Data: []byte(fixtureSource)}}
	paths, err := sourceInputPaths(repository)
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, corpusPath, denominatorPath(version))
	history, err := fs.Glob(repository, closureRoot+"evidence/denominator*.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(paths, history...) {
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		data[path] = &fstest.MapFile{Data: raw}
	}
	pin(t, data, &request)
	return data, request
}

func pin(t *testing.T, data fs.FS, request *Request) {
	t.Helper()
	var err error
	request.SnapshotDigest, request.SourceDigest, err = InspectInputs(data, *request)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCandidateIsCompleteSourceBoundReplayWithoutAuthority(t *testing.T) {
	data, request := fixture(t)
	before := digestValue(data)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	first, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, replay) || first.Emitted != len(first.Members) || first.Required != first.Emitted ||
		first.RequiredArtifacts != 9 || len(first.Artifacts) != 9 ||
		first.ContractDigest == "" || first.SemanticDigest == "" || first.ActivityID == "" ||
		first.ApplyAuthorized || first.PromotionAllowed || first.RepositoryWrites != 0 ||
		first.State != "PROPOSAL_ONLY" || first.Admission != "UNASSESSED" {
		t.Fatalf("candidate lost source/replay/authority boundaries: %#v", first)
	}
	if err := plan.ValidateCandidate(data, first); err != nil {
		t.Fatal(err)
	}
	if digestValue(data) != before {
		t.Fatal("generator changed its input snapshot")
	}
	for _, artifact := range first.Artifacts {
		if artifact.ActivityID == "" || artifact.OutputID == "" || len(artifact.Paths) == 0 {
			t.Fatalf("semantic artifact role is unresolved: %#v", artifact)
		}
	}
	seen := map[string]bool{}
	for _, member := range first.Members {
		if seen[member.Path] || len(member.ActivityIDs) == 0 || member.AfterDigest != digest(member.Content) {
			t.Fatalf("candidate member is unbound or duplicated: %s", member.Path)
		}
		seen[member.Path] = true
	}
}

func requireFailure(t *testing.T, err error, state, class string) {
	t.Helper()
	var observed *Failure
	if !errors.As(err, &observed) || observed.State != state || observed.UnknownClass != class ||
		observed.Stage == "" || observed.Step == "" || observed.Reason == "" ||
		observed.NextOperation == "" || observed.BlockedBy == nil {
		t.Fatalf("expected complete %s/%s evidence, got %v", state, class, err)
	}
}
