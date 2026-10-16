package syntaxregistration

import (
	"encoding/json"
	"io/fs"
	"os"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRegisteredProfileProposalFresh1857(t *testing.T) {
	repository := os.DirFS("../../..")
	raw, err := fs.ReadFile(repository, corpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var registry languagesyntax.Registry
	if err := decodeStrict(raw, &registry); err != nil {
		t.Fatal(err)
	}
	request := ProfileMigrationRequest{Schema: ProfileMigrationSchema, Toolchain: runtime.Version(),
		TargetProfile: syntax.EntityFieldsV2ProfileID, BaseVersion: profileBaseVersion1857(t, repository)}
	for _, definition := range registry.Cases {
		if definition.ID == "unresolved-operation-source-origin" {
			request.Case = definition
		}
	}
	if request.Case.EntityFieldsProfile == request.TargetProfile {
		request.TargetProfile = syntax.EntityFieldsV3ProfileID
	}
	source, err := fs.ReadFile(repository, request.Case.Path)
	if err != nil {
		t.Fatal(err)
	}
	replacement := append(append([]byte{}, source...), '\n')
	request.ExecutionIdentity, err = ObserveExecutionIdentity()
	if err != nil {
		t.Fatal(err)
	}
	request.SnapshotDigest, request.SourceDigest, request.ReplacementDigest, err =
		InspectProfileMigrationInputs(repository, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := CompileProfileMigration(repository, request, replacement)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(repository)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateCandidate(repository, candidate); err != nil ||
		candidate.CasePopulation != len(registry.Cases) || candidate.RequiredMembers != 3 {
		t.Fatalf("actual source population or roles diverged: %v", err)
	}
	t.Logf("PROFILE_REGISTERED_PROPOSAL population=%d target_sources_replayed=2 native_program_executions=0 writes=0",
		candidate.CasePopulation)
}

func profileBaseVersion1857(t *testing.T, repository fs.FS) int {
	t.Helper()
	paths, err := fs.Glob(repository, closureRoot+"evidence/denominator-v*.json")
	if err != nil {
		t.Fatal(err)
	}
	version := 0
	for _, path := range paths {
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		var value denominator
		if json.Unmarshal(raw, &value) == nil && value.Version > version {
			version = value.Version
		}
	}
	return version
}

func TestProfileMigrationMixedLegacyFresh1858(t *testing.T) {
	for _, definition := range []languagesyntax.CaseDefinition{
		{EntityFields: true, ImplicitActivityPorts: true},
		{OrderedCollections: true, ImplicitActivityPorts: true},
		{OrderedCollections: true, EntityFields: true},
	} {
		request := ProfileMigrationRequest{Schema: ProfileMigrationSchema, Case: definition,
			TargetProfile: syntax.EntityFieldsV2ProfileID}
		request.Case.ID, request.Case.Path = "legacy1858", "examples/legacy1858/main.gooo"
		request.Case.Kind, request.Case.ExpectedDecision = languagesyntax.KindValid, languagesyntax.DecisionPass
		request.Case.Scope, request.Case.MetaOperation = languagesyntax.ScopeLanguageCapability, "replay-language-syntax"
		request.Case.ProofChoice = "COHERENCE"
		if err := validateProfileMigrationRequest(request, []byte("source")); err == nil {
			t.Fatal("mixed legacy source profiles accepted")
		}
	}
}
