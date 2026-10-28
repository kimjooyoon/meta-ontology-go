package syntaxregistration

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestGoooCaseCompositionOwnsProfileSelection(t *testing.T) {
	base := CaseIntent{ID: "agent-case", Path: "examples/agent-case/main.gooo",
		ProofChoice: "COHERENCE", BaseVersion: 99}
	cases := []struct {
		profile string
		flags   languagesyntax.CaseDefinition
	}{
		{"ordinary", languagesyntax.CaseDefinition{}},
		{"text-fields", languagesyntax.CaseDefinition{EntityFields: true}},
		{"implicit-ports", languagesyntax.CaseDefinition{ImplicitActivityPorts: true}},
		{"ordered-collections", languagesyntax.CaseDefinition{OrderedCollections: true}},
		{"gooo.entityfields.go-projection.v6", languagesyntax.CaseDefinition{EntityFieldsProfile: "gooo.entityfields.go-projection.v6"}},
		{"gooo.entityfields.go-projection.v7", languagesyntax.CaseDefinition{EntityFieldsProfile: "gooo.entityfields.go-projection.v7"}},
	}
	for _, row := range cases {
		t.Run(row.profile, func(t *testing.T) {
			intent := base
			intent.Profile = row.profile
			request, value, err := ComposeDraft(context.Background(), intent)
			if err != nil {
				t.Fatal(err)
			}
			want := row.flags
			want.ID, want.Path, want.Kind = intent.ID, intent.Path, "VALID"
			want.ExpectedDecision, want.ProofChoice = "PASS", intent.ProofChoice
			want.MetaOperation, want.Scope = "replay-language-syntax", "LANGUAGE_CAPABILITY"
			if !reflect.DeepEqual(request.Case, want) || request.CaseIntent == nil ||
				request.CaseSourceDigest != value.SourceSHA256 || value.ActivityID != "gooo://syntax-registration/compose-case" ||
				value.ProjectionSHA256 == "" || VerifyCaseComposition(context.Background(), request) != nil {
				t.Fatalf("Gooo case or origin differs: %+v %+v", request, value)
			}
		})
	}
}

func TestGoooCaseCompositionRejectsInvalidOrChangedIntent(t *testing.T) {
	base := CaseIntent{ID: "agent-case", Path: "examples/agent-case/main.gooo",
		ProofChoice: "COHERENCE", Profile: "ordinary", BaseVersion: 99}
	for _, changed := range []CaseIntent{{ID: base.ID, Path: base.Path, ProofChoice: base.ProofChoice,
		Profile: "unknown", BaseVersion: 99}, {ID: base.ID, Path: base.Path, ProofChoice: "UNKNOWN",
		Profile: base.Profile, BaseVersion: 99}} {
		if _, _, err := ComposeDraft(context.Background(), changed); err == nil {
			t.Fatal("invalid explicit intent was accepted")
		}
	}
	request, _, err := ComposeDraft(context.Background(), base)
	if err != nil {
		t.Fatal(err)
	}
	request.Case.Scope = "GOVERNANCE_OBSERVATION"
	if err := VerifyCaseComposition(context.Background(), request); err == nil {
		t.Fatal("a changed composed case retained its source authority")
	}
	request.Case.Scope = "LANGUAGE_CAPABILITY"
	request.CaseSourceDigest = "sha256:stale"
	if err := VerifyCaseComposition(context.Background(), request); err == nil {
		t.Fatal("a stale Gooo source identity was accepted")
	}
	if _, err := DecodeCaseIntent([]byte(`{"id":"a","id":"b"}`)); err == nil {
		t.Fatal("duplicate caller intent key was accepted")
	}
}

func TestGoooCaseCompositionSourceChangesTheActualValue(t *testing.T) {
	intent := CaseIntent{ID: "agent-case", Path: "examples/agent-case/main.gooo",
		ProofChoice: "COHERENCE", Profile: "text-fields", BaseVersion: 99}
	before, original, err := ComposeDraft(context.Background(), intent)
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(contractSource, []byte("result.entity_fields = true"),
		[]byte("result.ordered_collections = true"), 1)
	if bytes.Equal(changed, contractSource) {
		t.Fatal("source decision was not found")
	}
	after, altered, err := composeDraftSource(context.Background(), intent, changed)
	if err != nil || !before.Case.EntityFields || !after.Case.OrderedCollections ||
		before.Case.OrderedCollections || after.Case.EntityFields ||
		original.SourceSHA256 == altered.SourceSHA256 || original.ProjectionSHA256 == altered.ProjectionSHA256 {
		t.Fatalf("source edit did not change the composed value and origin: %+v %+v %v", before, after, err)
	}
	if err := VerifyCaseComposition(context.Background(), after); err == nil {
		t.Fatal("a request composed by different source was accepted by this executable")
	}
}

func TestComposedCaseEntersExistingPinnedRegistration(t *testing.T) {
	data, original := fixture(t)
	intent := CaseIntent{ID: original.Case.ID, Path: original.Case.Path,
		ProofChoice: original.Case.ProofChoice, Profile: "text-fields", BaseVersion: original.BaseVersion}
	request, _, err := ComposeDraft(context.Background(), intent)
	if err != nil || !reflect.DeepEqual(request.Case, original.Case) {
		t.Fatalf("Gooo did not construct the existing corpus case: %+v %v", request.Case, err)
	}
	request.SnapshotDigest, request.SourceDigest = original.SnapshotDigest, original.SourceDigest
	request.Toolchain, request.ExecutionIdentity = original.Toolchain, original.ExecutionIdentity
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	request.CaseIntent.Profile = "ordinary" // The compiled plan owns its own intent.
	candidate, err := plan.Generate(data)
	if err != nil || candidate.RequiredArtifacts != 9 || candidate.Emitted != candidate.Required {
		t.Fatalf("composed case lost registration roles: %+v %v", candidate, err)
	}
	if err := plan.ValidateCandidate(data, candidate); err != nil {
		t.Fatal(err)
	}
}

func TestContractCaseCompositionReplay(t *testing.T) {
	result := languagesyntax.ReplayCase(os.DirFS("../../.."), languagesyntax.CaseDefinition{
		Path: "internal/meta/syntaxregistration/contract.gooo", Kind: "VALID",
		EntityFieldsProfile: "gooo.entityfields.go-projection.v5"})
	if !result.ASTReplayed || !result.ByteReplayed || !result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("contract source no longer replays: %+v", result)
	}
	data, request := promotedMetaFixture(t)
	result = languagesyntax.ReplayCase(data, request.Case)
	if !result.ASTReplayed || !result.ByteReplayed || !result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("contract promotion snapshot no longer replays: %+v", result)
	}
}
