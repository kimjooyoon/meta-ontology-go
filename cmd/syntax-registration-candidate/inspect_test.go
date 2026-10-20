package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestInspectResponseCarriesDirectlyDecodablePinnedRequest(t *testing.T) {
	request := syntaxregistration.Request{
		Case: languagesyntax.CaseDefinition{ID: "bounded-text", Path: "examples/bounded-text/main.gooo",
			Kind: languagesyntax.KindValid, ProofChoice: "COHERENCE", EntityFieldsProfile: "gooo.entityfields.go-projection.v6"},
		BaseVersion: 99, PromoteMetaSource: true,
		SnapshotDigest: "stale", SourceDigest: "stale", Toolchain: "stale",
	}
	identity := syntaxregistration.ExecutionIdentity{GoVersion: "go1.27.3", GOOS: "linux", GOARCH: "amd64",
		ExecutableDigest: "sha256:generator", GoCommandDigest: "sha256:driver", CompilerDigest: "sha256:compiler"}
	response := newInspectionResponse(request, "sha256:snapshot", "sha256:source", "go1.27.3", identity)
	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	pinned, err := syntaxregistration.DecodeRequest(fields["pinned_request"])
	if err != nil {
		t.Fatalf("inspect output must be usable as a registration request: %v", err)
	}
	if !reflect.DeepEqual(pinned.Case, request.Case) || pinned.BaseVersion != request.BaseVersion ||
		pinned.PromoteMetaSource != request.PromoteMetaSource || pinned.SnapshotDigest != response.SnapshotDigest ||
		pinned.SourceDigest != response.SourceDigest || pinned.Toolchain != response.Toolchain ||
		pinned.ExecutionIdentity != identity || response.SemanticAdmission != "UNASSESSED" {
		t.Fatalf("inspect request lost source intent or observation: %+v", pinned)
	}
	if request.SnapshotDigest != "stale" || request.ExecutionIdentity != (syntaxregistration.ExecutionIdentity{}) {
		t.Fatal("inspect mutated the original request")
	}
}
