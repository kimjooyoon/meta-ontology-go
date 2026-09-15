package directorykind

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSourceSchemaBindsDirectoryKindOntology(t *testing.T) {
	source := SourceMetrics{
		Repository: "kimjooyoon/meta-ontology-go", CommitSHA: strings.Repeat("a", 40),
		Directories: []SourceDirectory{{Path: ".", SubjectKind: "PROJECT_ROOT"}},
		Meta: SourceMeta{Schema: indicatorSchema, Policy: SourcePolicy{
			Schema: policySchema, RequireHomogeneousDirectories: true,
			ExemptProjectRootTopology: true, ExemptProjectRootREADME: true,
		}},
	}
	payload, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSource(payload)
	if err != nil {
		t.Fatal(err)
	}
	ontologyDigest, err := validateOntology()
	if err != nil {
		t.Fatal(err)
	}
	if decoded.CommitSHA != source.CommitSHA || !strings.HasPrefix(ontologyDigest, "sha256:") {
		t.Fatalf("schema or ontology binding lost: %#v %q", decoded, ontologyDigest)
	}
}

func TestSourceSchemaRequiresExplicitProjectRoot(t *testing.T) {
	source := SourceMetrics{Repository: "repo", CommitSHA: "sha",
		Meta: SourceMeta{Schema: indicatorSchema, Policy: SourcePolicy{Schema: policySchema}}}
	if err := validateSource(source); err == nil {
		t.Fatal("missing project root was accepted")
	}
}
