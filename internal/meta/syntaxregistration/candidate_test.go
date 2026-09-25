package syntaxregistration

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"reflect"
	"runtime"
	"strings"
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
	request.ExecutionIdentity, err = ObserveExecutionIdentity()
	if err != nil {
		t.Fatal(err)
	}
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

const internalMetaRegistrationSource = `package metapolicyrevision
namespace metapolicyrevision

entity PolicySource id "gooo://meta-policy-revision/source"
entity PolicyDecisionRevision id "gooo://meta-policy-revision/request"
entity PolicyDecisionProposal id "gooo://meta-policy-revision/proposal"

activity ProposePolicyDecisionRevision(PolicySource, PolicyDecisionRevision) -> PolicyDecisionProposal
`

func internalMetaFixture(t *testing.T) (fstest.MapFS, Request) {
	t.Helper()
	data, request := fixture(t)
	delete(data, request.Case.Path)
	request.Case.ID = "policy-revision-operation-contract"
	request.Case.Path = "internal/meta/policycompilation/revision-contract.gooo"
	request.Case.EntityFields = false
	data[request.Case.Path] = &fstest.MapFile{Data: []byte(internalMetaRegistrationSource)}
	pin(t, data, &request)
	return data, request
}

func TestRegistrationSourceDomains(t *testing.T) {
	for _, test := range []struct {
		path    string
		allowed bool
	}{
		{"examples/new/main.gooo", true},
		{"internal/meta/new/contract.gooo", true},
		{"../examples/main.gooo", false},
		{"examples/../main.gooo", false},
		{"/absolute/main.gooo", false},
		{"internal/meta/../main.gooo", false},
		{"internal/meta-other/main.gooo", false},
		{"internal/other/main.gooo", false},
		{".github/workflows/main.gooo", false},
		{"docs/main.gooo", false},
		{"internal/meta/main.go", false},
		{"", false},
	} {
		t.Run(test.path, func(t *testing.T) {
			if registrationSourcePath(test.path) != test.allowed {
				t.Fatalf("unexpected source-domain admission for %q", test.path)
			}
			if !test.allowed {
				request := Request{BaseVersion: 30, Case: languagesyntax.CaseDefinition{Path: test.path}}
				_, _, err := InspectInputs(fstest.MapFS{}, request)
				requireFailure(t, err, "REFUTED", "")
				if err.Error() != "REFUTED/REGISTRATION_INPUT_INVALID" {
					t.Fatalf("source was not rejected before input observation: %v", err)
				}
			}
		})
	}
}

func TestInternalMetaSourceRegistrationPreservesAllNineRoles(t *testing.T) {
	data, request := internalMetaFixture(t)
	before := digestValue(data)
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
	if candidate.RequiredArtifacts != 9 || len(candidate.Artifacts) != 9 ||
		candidate.Required != candidate.Emitted || candidate.Emitted != len(candidate.Members) ||
		candidate.InputDigest != request.SnapshotDigest || candidate.RequestDigest != digestValue(request) ||
		candidate.ExecutionBinding.Identity != request.ExecutionIdentity ||
		candidate.ApplyAuthorized || candidate.PromotionAllowed || candidate.RepositoryWrites != 0 ||
		candidate.State != "PROPOSAL_ONLY" || candidate.Admission != "UNASSESSED" {
		t.Fatal("internal source registration lost its complete role, input, or authority boundary")
	}
	if digestValue(data) != before {
		t.Fatal("internal source registration changed the input snapshot")
	}
	for _, member := range candidate.Members {
		if member.Path == request.Case.Path {
			t.Fatal("registration attempted to rewrite its source")
		}
		if strings.HasPrefix(member.Path, closureRoot+"evidence/") &&
			(member.Path != denominatorPath(request.BaseVersion+1) || member.BeforeDigest != "ABSENT") {
			t.Fatal("registration attempted to rewrite denominator history")
		}
	}
}

func promotedMetaFixture(t *testing.T) (fstest.MapFS, Request) {
	t.Helper()
	data, request := fixture(t)
	delete(data, request.Case.Path)
	request.Case.ID = "syntax-registration-meta-contract"
	request.Case.Path = "internal/meta/syntaxregistration/contract.gooo"
	request.Case.EntityFields = false
	request.PromoteMetaSource = true
	source, err := os.ReadFile("contract.gooo")
	if err != nil {
		t.Fatal(err)
	}
	data[request.Case.Path] = &fstest.MapFile{Data: source}
	pin(t, data, &request)
	return data, request
}

func promotionRegistry(t *testing.T, raw []byte) languagesyntax.Registry {
	t.Helper()
	var registry languagesyntax.Registry
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestMetaSourcePromotionPreservesExactCohortAndNineRoles(t *testing.T) {
	data, request := promotedMetaFixture(t)
	before := digestValue(data)
	baseline := promotionRegistry(t, data[corpusPath].Data)
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeRequest(raw)
	if err != nil || !decoded.PromoteMetaSource || !reflect.DeepEqual(decoded, request) {
		t.Fatalf("explicit promotion request was lost: %v", err)
	}
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
	if candidate.RequiredArtifacts != 9 || len(candidate.Artifacts) != 9 ||
		candidate.RequestDigest != digestValue(request) || candidate.InputDigest != request.SnapshotDigest ||
		candidate.ApplyAuthorized || candidate.PromotionAllowed || candidate.RepositoryWrites != 0 ||
		candidate.State != "PROPOSAL_ONLY" || candidate.Admission != "UNASSESSED" || digestValue(data) != before {
		t.Fatal("promotion lost its role, request, read-only input, or authority boundary")
	}
	corpusMembers := 0
	for _, member := range candidate.Members {
		if member.Path == request.Case.Path {
			t.Fatal("promotion rewrote its existing source")
		}
		if strings.HasPrefix(member.Path, closureRoot+"evidence/") &&
			(member.Path != denominatorPath(request.BaseVersion+1) || member.BeforeDigest != "ABSENT") {
			t.Fatal("promotion rewrote denominator history")
		}
		if member.Path == corpusPath {
			corpusMembers++
			assertPromotedCohort(t, baseline, promotionRegistry(t, member.Content), request)
		}
	}
	if corpusMembers != 1 {
		t.Fatal("promotion did not produce exactly one corpus")
	}
}

func assertPromotedCohort(t *testing.T, before, after languagesyntax.Registry, request Request) {
	t.Helper()
	expected := []string{}
	for _, path := range before.MetaSources {
		if path != request.Case.Path {
			expected = append(expected, path)
		}
	}
	if len(after.Cases) != len(before.Cases)+1 || !reflect.DeepEqual(after.MetaSources, expected) ||
		!reflect.DeepEqual(after.PackageUnits, before.PackageUnits) {
		t.Fatal("promotion changed unrelated membership")
	}
	if !reflect.DeepEqual(after.Cases[:len(before.Cases)], before.Cases) ||
		!reflect.DeepEqual(after.Cases[len(before.Cases)], request.Case) {
		t.Fatal("promotion changed existing cases or did not bind its exact requested case")
	}
}

func TestMetaSourcePromotionRequiresExplicitUniqueMembership(t *testing.T) {
	for _, name := range []string{"not-requested", "not-registered", "duplicate-meta-source", "already-executed", "package-member"} {
		t.Run(name, func(t *testing.T) {
			data, request := promotedMetaFixture(t)
			registry := promotionRegistry(t, data[corpusPath].Data)
			reason := "REGISTRATION_SOURCE_ALREADY_REGISTERED"
			switch name {
			case "not-requested":
				request.PromoteMetaSource = false
			case "not-registered":
				registry.MetaSources = nil
				reason = "REGISTRATION_META_SOURCE_NOT_REGISTERED"
			case "duplicate-meta-source":
				registry.MetaSources = append(registry.MetaSources, request.Case.Path)
				reason = "REGISTRATION_META_SOURCE_NOT_UNIQUE"
			case "already-executed":
				registry.Cases = append(registry.Cases, request.Case)
				reason = "REGISTRATION_CASE_ALREADY_EXISTS"
			case "package-member":
				registry.PackageUnits = append(registry.PackageUnits, languagesyntax.PackageDefinition{
					ID: "conflicting-package", Members: []string{request.Case.Path},
				})
			}
			raw, err := json.Marshal(registry)
			if err != nil {
				t.Fatal(err)
			}
			data[corpusPath].Data = raw
			pin(t, data, &request)
			_, err = Compile(data, request)
			requireFailure(t, err, "REFUTED", "")
			if err.Error() != "REFUTED/"+reason {
				t.Fatalf("unexpected rejection: %v", err)
			}
		})
	}
}

func TestMetaSourcePromotionRejectsResealedMembership(t *testing.T) {
	data, request := promotedMetaFixture(t)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"keep-promoted-meta-source", "drop-unrelated-meta-source"} {
		t.Run(name, func(t *testing.T) {
			candidate, err := plan.Generate(data)
			if err != nil {
				t.Fatal(err)
			}
			for index := range candidate.Members {
				member := &candidate.Members[index]
				if member.Path != corpusPath {
					continue
				}
				registry := promotionRegistry(t, member.Content)
				if name == "keep-promoted-meta-source" {
					registry.MetaSources = append(registry.MetaSources, request.Case.Path)
				} else {
					if len(registry.MetaSources) == 0 {
						t.Fatal("fixture has no unrelated source")
					}
					registry.MetaSources = registry.MetaSources[1:]
				}
				member.Content, err = json.Marshal(registry)
				if err != nil {
					t.Fatal(err)
				}
				member.AfterDigest = digest(member.Content)
			}
			requireFailure(t, plan.ValidateCandidate(data, candidate), "REFUTED", "")
		})
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
