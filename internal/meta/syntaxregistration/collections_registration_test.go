package syntaxregistration

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func collectionRegistrationInputs8551(t *testing.T) (fstest.MapFS, int) {
	t.Helper()
	repository := os.DirFS("../../..")
	paths, err := sourceInputPaths(repository)
	if err != nil {
		t.Fatal(err)
	}
	history, err := fs.Glob(repository, closureRoot+"evidence/denominator*.json")
	if err != nil {
		t.Fatal(err)
	}
	data := fstest.MapFS{}
	for _, path := range append(append(paths, corpusPath), history...) {
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		data[path] = &fstest.MapFile{Data: raw}
	}
	// Fresh local observations use the unpublished source view without changing
	// the checked-out PR. CI reads the exact submitted repository directly.
	if pending := os.Getenv("GOOO_REGISTRATION_SOURCE_VIEW"); pending != "" {
		err := filepath.WalkDir(pending, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			relative, err := filepath.Rel(pending, path)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(relative)
			if _, exists := data[name]; !exists && !strings.HasPrefix(name, closureRoot) {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err == nil {
				data[name] = &fstest.MapFile{Data: raw}
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	version := 0
	for path, file := range data {
		if strings.HasPrefix(path, closureRoot+"evidence/denominator") {
			var value denominator
			if json.Unmarshal(file.Data, &value) == nil && value.Version > version {
				version = value.Version
			}
		}
	}
	return data, version
}

func TestOrderedCollectionRegistration8551(t *testing.T) {
	for index, primitive := range []string{"integer", "string", "boolean"} {
		t.Run(primitive, func(t *testing.T) {
			data, version := collectionRegistrationInputs8551(t)
			name := fmt.Sprintf("collectionregistration8551%d", index)
			request := Request{BaseVersion: version, Toolchain: runtime.Version(),
				Case: languagesyntax.CaseDefinition{ID: name, Path: "examples/" + name + "/main.gooo",
					Kind: languagesyntax.KindValid, ExpectedDecision: languagesyntax.DecisionPass,
					ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax",
					Scope: languagesyntax.ScopeLanguageCapability, OrderedCollections: true}}
			source := fmt.Sprintf("package %s\nnamespace %s\n"+
				"entity Tray id %q fields { field values id %q type %s required many }\n"+
				"activity Forward(Tray) -> Tray computes \"record.forward:v1\"\n",
				name, name, "urn:"+name+":tray", "urn:"+name+":values", primitive)
			data[request.Case.Path] = &fstest.MapFile{Data: []byte(source)}
			pin(t, data, &request)
			before := digestValue(data)
			plan, err := Compile(data, request)
			if err != nil {
				t.Fatal(err)
			}
			candidate, err := plan.Generate(data)
			if err != nil {
				t.Fatal(err)
			}
			assertCollectionRegistration8551(t, data, request, candidate)
			if err := plan.ValidateCandidate(data, candidate); err != nil {
				t.Fatal(err)
			}
			if digestValue(data) != before {
				t.Fatal("registration changed the source view")
			}
			for _, flags := range [][2]bool{{true, false}, {false, true}, {true, true}} {
				conflict := request
				conflict.Case.EntityFields, conflict.Case.ImplicitActivityPorts = flags[0], flags[1]
				_, err := Compile(data, conflict)
				if err == nil || !strings.Contains(err.Error(), "REGISTRATION_CASE_CONTRACT_MISMATCH") {
					t.Fatalf("mixed syntax profiles accepted: %v", err)
				}
			}
			t.Logf("8551: %s many; nine source-bound roles; syntax denominator +1; three conflicting profiles rejected; model calls=0", primitive)
		})
	}
}

func assertCollectionRegistration8551(t *testing.T, data fstest.MapFS, request Request, candidate Candidate) {
	t.Helper()
	if candidate.RequiredArtifacts != 9 || len(candidate.Artifacts) != 9 || candidate.Required != candidate.Emitted ||
		candidate.InputDigest != request.SnapshotDigest || candidate.RequestDigest != digestValue(request) ||
		candidate.ExecutionBinding.Identity != request.ExecutionIdentity || candidate.RepositoryWrites != 0 {
		t.Fatal("registration lost a source-bound artifact role")
	}
	seenCorpus, seenNative, seenDenominator := false, false, false
	var prior denominator
	if err := json.Unmarshal(data[denominatorPath(request.BaseVersion)].Data, &prior); err != nil {
		t.Fatal(err)
	}
	for _, member := range candidate.Members {
		if len(member.ActivityIDs) == 0 || digest(member.Content) != member.AfterDigest {
			t.Fatal("generated member has no origin", member.Path)
		}
		if member.Path == corpusPath {
			var registry languagesyntax.Registry
			if err := json.Unmarshal(member.Content, &registry); err != nil {
				t.Fatal(err)
			}
			seenCorpus = reflect.DeepEqual(registry.Cases[len(registry.Cases)-1], request.Case)
		}
		if strings.HasPrefix(member.Path, syntaxRoot) && strings.HasSuffix(member.Path, ".go") &&
			strings.Contains(string(member.Content), "ID: \""+request.Case.ID+"\"") {
			seenNative = strings.Contains(string(member.Content), "OrderedCollections: true")
		}
		if strings.HasPrefix(member.Path, closureRoot+"evidence/") {
			if member.Path != denominatorPath(request.BaseVersion+1) || member.BeforeDigest != "ABSENT" {
				t.Fatal("registration rewrote historical origin", member.Path)
			}
			var next denominator
			if err := json.Unmarshal(member.Content, &next); err != nil {
				t.Fatal(err)
			}
			prior.Version++
			prior.DenominatorID = denominatorID(prior.Version)
			prior.Boundaries[0].Target++
			seenDenominator = reflect.DeepEqual(prior, next)
		}
	}
	if !seenCorpus || !seenNative || !seenDenominator {
		t.Fatalf("collection profile or denominator lost: corpus=%v native=%v denominator=%v", seenCorpus, seenNative, seenDenominator)
	}
}
