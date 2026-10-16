package syntaxregistration

import (
	"encoding/json"
	"go/format"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func profilePlanFixture1848(t *testing.T) (fstest.MapFS, ProfileMigrationRequest, []byte) {
	t.Helper()
	definition := languagesyntax.CaseDefinition{ID: "profile-plan1848", Path: "examples/profile1848/main.gooo",
		Kind: languagesyntax.KindValid, ExpectedDecision: languagesyntax.DecisionPass, ProofChoice: "COHERENCE",
		MetaOperation: "replay-language-syntax", Scope: languagesyntax.ScopeLanguageCapability, EntityFields: true}
	registry := languagesyntax.Registry{Schema: languagesyntax.RegistrySchema,
		MetaSources: []string{profileContractPath}, Cases: []languagesyntax.CaseDefinition{definition}}
	corpus, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	native, err := format.Source([]byte(`package languagesyntax
const profileID1848 = "profile-plan1848"
func expectedRegistry() Registry { return Registry{
 MetaSources: []string{"` + profileContractPath + `"}, Cases: owned1848(),
} }
func owned1848() []CaseDefinition { return []CaseDefinition{{
 ID: profileID1848, Path: "examples/profile1848/main.gooo", Kind: "VALID",
 ExpectedDecision: "PASS", ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax",
 Scope: "LANGUAGE_CAPABILITY", EntityFields: true,
}} }
func incidental1848() []CaseDefinition { return []CaseDefinition{{ID:"ignored1848"}} }
`))
	if err != nil {
		t.Fatal(err)
	}
	before := profileProgram1848("string", "return input")
	after := profileProgram1848("boolean", "return input")
	files := fstest.MapFS{corpusPath: {Data: corpus}, syntaxRoot + "owner1848.go": {Data: native},
		definition.Path: {Data: before}, profileContractPath: {Data: profileMigrationSource},
		denominatorPath(93): {Data: []byte("{\"retained\":93}\n")}}
	request := ProfileMigrationRequest{Schema: ProfileMigrationSchema, Case: definition,
		TargetProfile: syntax.EntityFieldsV2ProfileID, BaseVersion: 93, Toolchain: runtime.Version()}
	request.ExecutionIdentity, err = ObserveExecutionIdentity()
	if err != nil {
		t.Fatal(err)
	}
	request.SnapshotDigest, request.SourceDigest, request.ReplacementDigest, err =
		InspectProfileMigrationInputs(files, request, after)
	if err != nil {
		t.Fatal(err)
	}
	return files, request, after
}

func profileProgram1848(kind, body string) []byte {
	return []byte(`package profile1848
namespace profile1848
entity Observation id "urn:profile1848:observation" fields {
 field value id "urn:profile1848:value" type ` + kind + ` required one
}
activity Echo(Observation) -> Observation id "urn:profile1848:echo" computes "` + body + `"
`)
}
