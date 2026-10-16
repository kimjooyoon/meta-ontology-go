package unresolvedpolicy_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
	policyprojection "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy/projection"
	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func TestUnresolvedPolicyProjectionFresh1723(t *testing.T) {
	source, err := os.ReadFile("policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	target, err := os.ReadFile("disposition.go")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := policyprojection.Generate("policy.gooo", source, target)
	if err != nil {
		t.Fatal(err)
	}
	parts, err := generator.PartitionProjection(expected.Source, 75)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("policy.projection.json")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := projectionbundle.Decode(raw)
	if err != nil || len(manifest.Members) != len(parts) {
		t.Fatalf("projection member set: %v", err)
	}
	for index, member := range manifest.Members {
		actual, err := os.ReadFile(member.Path)
		if err != nil || !bytes.Equal(actual, parts[index].Source) {
			t.Fatalf("compiled source projection member differs: %s: %v", member.Path, err)
		}
	}
	if err := unresolvedpolicy.VerifyProjection(); err != nil {
		t.Fatal(err)
	}
}
