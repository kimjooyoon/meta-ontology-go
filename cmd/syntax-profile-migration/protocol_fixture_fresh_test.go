package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func profileCLIFamily(t *testing.T, root, afterPath string, files map[string][]byte) {
	t.Helper()
	const primary, secondary = "examples/cli1854/main.gooo", "examples/cli1854/echo.gooo"
	cut := bytes.Index(files[primary], []byte("activity Echo"))
	files[secondary] = append([]byte("package cli1854\nnamespace cli1854\n"), files[primary][cut:]...)
	files[primary] = files[primary][:cut]
	after, err := os.ReadFile(afterPath)
	if err != nil {
		t.Fatal(err)
	}
	profileCLIWrite1854(t, afterPath, after[:bytes.Index(after, []byte("activity Echo"))])
	const corpus = "examples/language-syntax-roundtrip/corpus.json"
	var registry languagesyntax.Registry
	if err := json.Unmarshal(files[corpus], &registry); err != nil {
		t.Fatal(err)
	}
	registry.CaseSources = []languagesyntax.CaseSourceSet{{CaseID: "cli1854", Members: []string{primary, secondary}}}
	files[corpus], _ = json.Marshal(registry)
	const native = "internal/meta/languagereadiness/languagesyntax/owner1854.go"
	memberField := fmt.Sprintf("CaseSources: []CaseSourceSet{{CaseID: %q, Members: []string{%q, %q}}}, Cases:",
		"cli1854", primary, secondary)
	files[native], err = format.Source(bytes.Replace(files[native], []byte("Cases:"), []byte(memberField), 1))
	if err != nil {
		t.Fatal(err)
	}
	for path, raw := range files {
		profileCLIWrite1854(t, filepath.Join(root, path), raw)
	}
}

func profileCLIFixture1854(t *testing.T) (string, string, string, string, map[string][]byte) {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "input")
	const contractPath = "internal/meta/syntaxregistration/profile_migration.gooo"
	contract, err := os.ReadFile(filepath.Join("../..", contractPath))
	if err != nil {
		t.Fatal(err)
	}
	definition := languagesyntax.CaseDefinition{ID: "cli1854", Path: "examples/cli1854/main.gooo",
		Kind: languagesyntax.KindValid, ExpectedDecision: languagesyntax.DecisionPass, ProofChoice: "COHERENCE",
		Scope: languagesyntax.ScopeLanguageCapability, MetaOperation: "replay-language-syntax", EntityFields: true}
	request := syntaxregistration.ProfileMigrationRequest{Schema: syntaxregistration.ProfileMigrationSchema,
		Case: definition, TargetProfile: syntax.EntityFieldsV2ProfileID, BaseVersion: 93}
	registry := languagesyntax.Registry{Schema: languagesyntax.RegistrySchema,
		MetaSources: []string{contractPath}, Cases: []languagesyntax.CaseDefinition{definition}}
	corpus, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	native, err := format.Source([]byte(`package languagesyntax
func expectedRegistry() Registry { return Registry{
MetaSources: []string{"` + contractPath + `"}, Cases: []CaseDefinition{{
ID:"cli1854", Path:"examples/cli1854/main.gooo", Kind:"VALID", ExpectedDecision:"PASS",
ProofChoice:"COHERENCE", Scope:"LANGUAGE_CAPABILITY", MetaOperation:"replay-language-syntax", EntityFields:true,
}}} }
`))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{contractPath: contract, definition.Path: profileCLIProgram1854("string"),
		"examples/language-syntax-roundtrip/corpus.json":                                           corpus,
		"internal/meta/languagereadiness/languagesyntax/owner1854.go":                              native,
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator-v93.json": []byte("{}\n")}
	for path, raw := range files {
		profileCLIWrite1854(t, filepath.Join(root, filepath.FromSlash(path)), raw)
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	requestPath, afterPath := filepath.Join(base, "request.json"), filepath.Join(base, "after.gooo")
	profileCLIWrite1854(t, requestPath, raw)
	profileCLIWrite1854(t, afterPath, profileCLIProgram1854("boolean"))
	return root, requestPath, afterPath, filepath.Join(base, "proposal"), files
}

func profileCLIProgram1854(kind string) []byte {
	return []byte(`package cli1854
namespace cli1854
entity Value id "urn:cli1854:value" fields {
field value id "urn:cli1854:field" type ` + kind + ` required one
}
activity Echo(Value) -> Value id "urn:cli1854:echo" computes "return input"
`)
}

func profileCLIWrite1854(t *testing.T, path string, raw []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
